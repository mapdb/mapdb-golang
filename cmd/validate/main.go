// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Cross-language validation runner for mapdb-golang.
//
// Reads a JSON scenario file from the shared mapdb-collection-spec
// suite, runs the described operations through the Go collections, and
// prints the assertion outputs in the canonical per-line
// `<key>: <value>` format consumed by validate.sh.
//
// Sibling implementations: mapdb-rust/src/bin/validate.rs,
// mapdb-zig/src/validate.zig, mapdb-typescript/src/validate.ts.
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/mapdb/mapdb-golang/arraylist"
	"github.com/mapdb/mapdb-golang/bag"
	"github.com/mapdb/mapdb-golang/hash"
	"github.com/mapdb/mapdb-golang/hashmap"
	"github.com/mapdb/mapdb-golang/hashset"
	"github.com/mapdb/mapdb-golang/hyperloglog"
	"github.com/mapdb/mapdb-golang/immutablesorted"
	"github.com/mapdb/mapdb-golang/multimap"
	"github.com/mapdb/mapdb-golang/rangev"
	"github.com/mapdb/mapdb-golang/treemap"
	"github.com/mapdb/mapdb-golang/treeset"
)

type scenario struct {
	Name       string                     `json:"name"`
	Collection string                     `json:"collection"`
	Operations []map[string]any           `json:"operations"`
	Assertions map[string]json.RawMessage `json:"assertions"`
	Other      *otherSpec                 `json:"other,omitempty"`
	// Query is the optional single top-level range (NavigableMap/Set) the
	// range_* assertions refer to; same range-builder shape as the `range`
	// field on a remove_range op.
	Query map[string]any `json:"query,omitempty"`
}

type otherSpec struct {
	Operations []map[string]any `json:"operations"`
}

// anyFail is set whenever any assertion mismatches; the process exits
// non-zero at the end so the harness treats assertion failures as the
// primary pass/fail signal.
var anyFail bool

// floatMode controls how an expected JSON value is rendered into the
// canonical string emitted by the runner, so the comparison is by bit
// pattern for floats (NaN == NaN, +0.0 != -0.0).
type floatMode int

const (
	modeNone     floatMode = iota // i32 collections
	modeF32Keyed                  // f32 map/set: only arrays are float-labelled
	modeF32List                   // f32 list: sum/min/max scalars + arrays are floats
)

// emit prints a computed assertion and compares it against the expected
// JSON value. Unrecognised keys (UNKNOWN_ASSERTION:*) are skipped silently
// per the README unknown-assertion-skip rule: no line, no failure.
func emit(name, key, computed string, expected json.RawMessage, mode floatMode) {
	if strings.HasPrefix(computed, "UNKNOWN_ASSERTION:") {
		return
	}
	fmt.Printf("%s: %s\n", key, computed)
	want := renderExpected(expected, key, mode)
	if computed != want && !looseNaNMatch(expected, mode, computed) {
		fmt.Printf("FAIL %s %s: expected=%s got=%s\n", name, key, want, computed)
		anyFail = true
	}
}

// looseNaNMatch implements the loose-NaN scalar rule: when the EXPECTED
// operand is a bare NaN *label* ("NaN"/"+NaN"/"-NaN") — NOT a {"bits":"0x.."}
// object and NOT an array element — the assertion passes against ANY NaN the
// runner computed, regardless of sign/payload. This covers impl/arch-defined
// arithmetic NaNs such as (+Inf)+(-Inf), whose bits differ across x86 vs ARM.
// {"bits"} operands stay bitwise-exact and array elements stay exact/positional
// (renderExpected is unchanged for both). See cross-language-validation/README.md
// §"Float operand encoding".
func looseNaNMatch(expected json.RawMessage, mode floatMode, computed string) bool {
	if mode == modeNone {
		return false
	}
	var v any
	if err := json.Unmarshal(expected, &v); err != nil {
		return false
	}
	s, ok := v.(string)
	if !ok || !math.IsNaN(float64(parseF32Label(s))) {
		return false
	}
	// Computed must itself be a NaN bit pattern (canonical "0x........").
	if len(computed) != 10 || (computed[:2] != "0x" && computed[:2] != "0X") {
		return false
	}
	bits, err := strconv.ParseUint(computed[2:], 16, 32)
	if err != nil {
		return false
	}
	return math.IsNaN(float64(math.Float32frombits(uint32(bits))))
}

// renderExpected renders an expected JSON assertion value into the same
// canonical string the runner emits for its computed value.
func renderExpected(raw json.RawMessage, key string, mode floatMode) string {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	switch t := v.(type) {
	case nil:
		return "null"
	case bool:
		return strconv.FormatBool(t)
	case float64:
		// JSON number. f32 list scalars (sum/min/max) render as floats;
		// the structural `size` count stays an integer. Under F32Keyed the
		// key-typed scalars min/max are also f32 and render via the f32
		// formatter (matching TS); other scalars stay i32.
		if (mode == modeF32List && key != "size") ||
			(mode == modeF32Keyed && (key == "min" || key == "max")) {
			return formatF32(float32(t))
		}
		return strconv.FormatInt(int64(t), 10)
	case string:
		if mode == modeNone {
			// Plain string scalar (e.g. Range lower_bound_type: "closed").
			return t
		}
		// Float label scalar (e.g. sum: "NaN").
		return formatF32(parseF32Label(t))
	case map[string]any:
		// Bits-escape float scalar (e.g. sum: {"bits":"0xffc00000"}).
		if mode != modeNone {
			return formatF32(parseF32(t))
		}
		return string(raw)
	case []any:
		parts := make([]string, len(t))
		for i, e := range t {
			switch mode {
			case modeF32Keyed:
				parts[i] = "\"" + formatF32(elementToF32(e)) + "\""
			case modeF32List:
				parts[i] = formatF32(elementToF32(e))
			default:
				// modeNone arrays are normally i32 (JSON numbers). The
				// HashMap<i64,i32> `sorted_keys` array is decimal STRINGS
				// (i64 keys exceed 2^53) — render those quoted to match the
				// runner's computed quoted-string array. NavigableMap/Set poll
				// logs (poll_first_keys / poll_first_values / ...) are
				// (int|null)[] — a nil element renders as the bare "null".
				switch ev := e.(type) {
				case nil:
					parts[i] = "null"
				case string:
					parts[i] = "\"" + ev + "\""
				default:
					parts[i] = strconv.FormatInt(int64(ev.(float64)), 10)
				}
			}
		}
		return "[" + strings.Join(parts, ",") + "]"
	}
	return string(raw)
}

func elementToF32(e any) float32 {
	switch x := e.(type) {
	case string:
		return parseF32Label(x)
	case float64:
		return float32(x)
	case map[string]any:
		// {"bits":"0x.."} escape inside an assertion array.
		return parseF32(e)
	}
	fatalf("unexpected float array element: %T", e)
	return 0
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: validate <scenario.json>")
		os.Exit(1)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fatalf("failed to read scenario file: %v", err)
	}
	// Decode with UseNumber() so JSON number tokens are preserved as
	// json.Number (their exact decimal string) rather than float64. This is
	// REQUIRED for i64 keys: a bare number like 9007199254740993 or
	// 9223372036854775807 would otherwise be rounded through f64 and corrupted.
	// All scalar parsers (asInt32/asInt/parseF32/parseI64Operand) handle the
	// json.Number case. (renderExpected re-unmarshals the raw assertion bytes
	// into its own `any` WITHOUT UseNumber, so the expected-value rendering
	// path is unaffected and still sees float64.)
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber()
	var s scenario
	if err := dec.Decode(&s); err != nil {
		fatalf("failed to parse JSON: %v", err)
	}

	fmt.Printf("=== scenario: %s ===\n", s.Name)

	switch s.Collection {
	case "HashMap<i32, i32>":
		runHashMap(s)
	case "HashMap<i64, i32>":
		runI64HashMap(s)
	case "ListMultimap<i64, i32>":
		runI64ListMultimap(s)
	case "SetMultimap<i64, i32>":
		runI64SetMultimap(s)
	case "ArrayList<i32>":
		runArrayList(s)
	case "HashSet<i32>":
		runHashSet(s)
	case "HashBag<i32>":
		runHashBag(s)
	case "TreeSet<i32>":
		runTreeSet(s)
	case "TreeMap<i32, i32>":
		runTreeMap(s)
	case "HashMap<f32, i32>":
		runF32HashMap(s)
	case "HashSet<f32>":
		runF32HashSet(s)
	case "TreeSet<f32>":
		runF32TreeSet(s)
	case "ArrayList<f32>":
		runF32ArrayList(s)
	case "Range<i32>":
		runRange(s)
	case "ImmutableSortedMap<i32, i32>":
		runImmutableSortedMap(s)
	case "ImmutableSortedSet<i32>":
		runImmutableSortedSet(s)
	case "HashPipeline":
		runHashPipeline(s)
	case "HyperLogLog":
		runHyperLogLog(s)
	default:
		// Forward-compat (README "unknown collection kinds skip"): a runner
		// that does not understand a collection kind must SKIP, not fail, so
		// newer scenarios never break an older runner. Mirrors the
		// unknown-assertion-key skip in emit.
		fmt.Fprintf(os.Stderr, "skip: unsupported collection kind (forward-compat): %s\n", s.Collection)
		return
	}

	if anyFail {
		os.Exit(1)
	}
}

// ---- shared helpers -------------------------------------------------------

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// sortedAssertionKeys returns assertion keys in stable order, skipping
// the "comment" field that scenario authors use for docs.
func sortedAssertionKeys(a map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(a))
	for k := range a {
		if k == "comment" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func formatArray(v []int32) string {
	parts := make([]string, len(v))
	for i, x := range v {
		parts[i] = strconv.FormatInt(int64(x), 10)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func asInt32(v any) int32 {
	switch n := v.(type) {
	case float64:
		return int32(int64(n))
	case json.Number:
		// Make a parse failure FATAL (consistent with the i64 decimal-string
		// rule) rather than silently coercing a bad operand to 0.
		i, err := n.Int64()
		if err != nil {
			fatalf("invalid integer operand %q: %v", n.String(), err)
		}
		return int32(i)
	}
	fatalf("expected integer, got %T (%v)", v, v)
	return 0
}

func asInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			fatalf("invalid integer operand %q: %v", n.String(), err)
		}
		return int(i)
	}
	fatalf("expected integer, got %T (%v)", v, v)
	return 0
}

// tryInt is the non-fatal variant of asInt: it returns ok=false on a missing
// (nil) or non-numeric operand instead of fataling. Used by builders that must
// SKIP a malformed scenario (forward-compat), mirroring the Rust runner's
// `as_u64()?` / `as_i64()?` which yield None rather than panicking.
func tryInt(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return 0, false
		}
		return int(i), true
	}
	return 0, false
}

// tryInt32 is the non-fatal variant of asInt32 (see tryInt).
func tryInt32(v any) (int32, bool) {
	switch n := v.(type) {
	case float64:
		return int32(int64(n)), true
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return 0, false
		}
		return int32(i), true
	}
	return 0, false
}

// Q4 float operand encoding (see cross-language-validation/README.md
// §"Float operand encoding"): JSON number, human-label string, or a
// {"bits":"0x........"} object reinterpreting 32 IEEE-754 bits.
func parseF32(v any) float32 {
	switch x := v.(type) {
	case string:
		return parseF32Label(x)
	case float64:
		return float32(x)
	case json.Number:
		// Bare JSON-number f32 operand under UseNumber(): parse the decimal
		// token to f64 then narrow to f32 (same as the float64 path).
		f, err := strconv.ParseFloat(x.String(), 64)
		if err != nil {
			fatalf("invalid f32 number literal: %q", x.String())
		}
		return float32(f)
	case map[string]any:
		if hx, ok := x["bits"].(string); ok {
			return math.Float32frombits(parseF32Bits(hx))
		}
		fatalf("expected {\"bits\":\"0x..\"} float object, got %v", v)
	}
	fatalf("expected f32 value, got %T (%v)", v, v)
	return 0
}

// parseF32Bits parses a 0x-prefixed, 8-hex-digit (case-insensitive) string
// into a raw 32-bit IEEE-754 pattern (NaN-payload / signed-bit escape).
func parseF32Bits(hx string) uint32 {
	body := hx
	if len(hx) >= 2 && (hx[:2] == "0x" || hx[:2] == "0X") {
		body = hx[2:]
	} else {
		fatalf("f32 bits literal must start with 0x: %q", hx)
	}
	if len(body) != 8 {
		fatalf("f32 bits literal must be 8 hex digits: %q", hx)
	}
	bits, err := strconv.ParseUint(body, 16, 32)
	if err != nil {
		fatalf("invalid f32 bits literal: %q", hx)
	}
	return uint32(bits)
}

// parseF32Label parses a human-label / decimal / hex-bits float string.
// Used for string operands and assertion-key suffixes (get_-NaN,
// contains_0.0, contains_0x7fc00001). Canonical NaN bits:
// +NaN=0x7FC00000, -NaN=0xFFC00000.
func parseF32Label(s string) float32 {
	switch s {
	case "NaN", "+NaN":
		return math.Float32frombits(0x7FC00000)
	case "-NaN":
		return math.Float32frombits(0xFFC00000)
	case "Infinity", "+Infinity":
		return float32(math.Inf(1))
	case "-Infinity":
		return float32(math.Inf(-1))
	case "0.0", "+0.0":
		return 0.0
	case "-0.0":
		return float32(math.Copysign(0, -1))
	case "pos_zero":
		return 0.0
	case "neg_zero":
		return float32(math.Copysign(0, -1))
	}
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		return math.Float32frombits(parseF32Bits(s))
	}
	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		fatalf("invalid f32 literal in key: %q", s)
	}
	return float32(f)
}

// formatF32 is the canonical, bit-faithful serialization. NaN (any
// sign/payload) and ±0.0 render as their 0x-hex bit pattern so distinct
// payloads and signed zeros stay distinguishable and every port emits the
// identical string; finite/inf values keep their human-readable label.
func formatF32(v float32) string {
	if math.IsNaN(float64(v)) || v == 0 {
		return fmt.Sprintf("0x%08x", math.Float32bits(v))
	}
	if math.IsInf(float64(v), 1) {
		return "Infinity"
	}
	if math.IsInf(float64(v), -1) {
		return "-Infinity"
	}
	if v == float32(math.Trunc(float64(v))) && math.Abs(float64(v)) < 1e16 {
		// Match Java/Rust "3.0" rendering for integer-valued floats.
		return fmt.Sprintf("%d.0", int64(v))
	}
	return strconv.FormatFloat(float64(v), 'g', -1, 32)
}

func unknown(key string) string {
	return "UNKNOWN_ASSERTION:" + key
}

// ---- HashPipeline (spec/features/hash-pipeline.md) ------------------------
//
// A stateless probe (not a stored collection): exactly ONE hash op carries the
// input + seed under test; the assertions read the deterministic hash output.
// Outputs are serialized as fixed-width, lower-case, 0x-prefixed hex strings
// (8 digits for a u32, 16 for a u64) so a 64-bit hash survives the JSON 2^53
// ceiling and every port's consensus diff is byte-identical. `positions` is an
// int[] in derivation order (NOT sorted). Unknown ops/keys SKIP (forward-compat).

// hashProbe is the probe a hash op builds. word32/word64 carry an
// already-encoded input word (so only the matching hash width is meaningful);
// i32/bytes carry the logical input so EITHER the 32- or 64-bit form (incl.
// lanes) can be asserted; positions carries the derived-position array.
type hashProbe struct {
	kind      string // "word32" | "word64" | "i32" | "bytes" | "positions"
	h32       uint32 // word32
	h64       uint64 // word64
	value     int32  // i32
	bytes     []byte // bytes
	seed      uint64 // i32 / bytes
	positions []uint32
}

func runHashPipeline(s scenario) {
	// Authoring rule: exactly ONE hash op. Zero or multiple => malformed =>
	// SKIP (like the sorted-table from_sorted rule). Forward-compat: an
	// unrecognised op kind also makes the scenario un-runnable here => SKIP.
	if len(s.Operations) != 1 {
		fmt.Fprintf(os.Stderr, "skip: hash-pipeline scenario must have exactly one op (forward-compat): got %d\n", len(s.Operations))
		return
	}
	op := s.Operations[0]
	var probe hashProbe
	switch op["op"] {
	case "hash_word32":
		raw := parseHexWord(op["word"])
		if raw > math.MaxUint32 {
			fatalf("hash_word32 `word` exceeds 32 bits: %#x", raw)
		}
		probe = hashProbe{kind: "word32", h32: hash.Hash32(uint32(raw), parseSeed(op["seed"]))}
	case "hash_word64":
		probe = hashProbe{kind: "word64", h64: hash.Hash64(parseHexWord(op["word"]), parseSeed(op["seed"]))}
	case "hash_i32":
		probe = hashProbe{kind: "i32", value: asInt32(op["value"]), seed: parseSeed(op["seed"])}
	case "hash_bytes":
		probe = hashProbe{kind: "bytes", bytes: parseHexBytes(op["bytes"]), seed: parseSeed(op["seed"])}
	case "positions":
		value := asInt32(op["value"])
		m := uint32(asInt(op["m"]))
		k := uint32(asInt(op["k"]))
		// The byte encoding of an i32 element drives positions: encode the i32
		// to its little-endian 4-byte form (the byte path the sketches use),
		// then derive. No op-level seed (the scheme fixes 0 / SALT2).
		w := uint32(value)
		b := []byte{byte(w), byte(w >> 8), byte(w >> 16), byte(w >> 24)}
		probe = hashProbe{kind: "positions", positions: hash.Positions(b, m, k)}
	default:
		fmt.Fprintf(os.Stderr, "skip: unknown hash-pipeline op (forward-compat): %v\n", op["op"])
		return
	}

	for _, key := range sortedAssertionKeys(s.Assertions) {
		emit(s.Name, key, probe.eval(key), s.Assertions[key], modeNone)
	}
}

func (p hashProbe) eval(key string) string {
	switch p.kind {
	case "word32":
		return evalH32Only(p.h32, key)
	case "word64":
		return evalH64Only(p.h64, key)
	case "positions":
		return evalPositions(p.positions, key)
	case "i32":
		switch key {
		case "hash32":
			return evalH32Only(hash.Hash32Int32(p.value, p.seed), "hash32")
		case "hash64", "hash64_hi", "hash64_lo":
			return evalH64Only(hash.Hash64Int32(p.value, p.seed), key)
		}
		return unknown(key)
	case "bytes":
		switch key {
		case "hash32":
			return evalH32Only(hash.Hash32Bytes(p.bytes, p.seed), "hash32")
		case "hash64", "hash64_hi", "hash64_lo":
			return evalH64Only(hash.Hash64Bytes(p.bytes, p.seed), key)
		}
		return unknown(key)
	}
	return unknown(key)
}

func evalH32Only(h uint32, key string) string {
	if key == "hash32" {
		return fmt.Sprintf("0x%08x", h)
	}
	return unknown(key)
}

func evalH64Only(h uint64, key string) string {
	switch key {
	case "hash64":
		return fmt.Sprintf("0x%016x", h)
	case "hash64_hi":
		return fmt.Sprintf("0x%08x", uint32(h>>32))
	case "hash64_lo":
		return fmt.Sprintf("0x%08x", uint32(h))
	}
	return unknown(key)
}

func evalPositions(p []uint32, key string) string {
	if key != "positions" {
		return unknown(key)
	}
	// Emitted in DERIVATION order (p_0 .. p_{k-1}), NOT sorted.
	parts := make([]string, len(p))
	for i, x := range p {
		parts[i] = strconv.FormatUint(uint64(x), 10)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// parseHexWord parses a 0x-prefixed hex word operand to a uint64 (used for
// `word` operands; the caller narrows to uint32 where the op needs a 32-bit
// word).
func parseHexWord(v any) uint64 {
	s, ok := v.(string)
	if !ok {
		fatalf("hash-pipeline `word` must be a 0x-hex string, got %T", v)
	}
	body, ok := strings.CutPrefix(s, "0x")
	if !ok {
		body, ok = strings.CutPrefix(s, "0X")
	}
	if !ok {
		fatalf("hash-pipeline word must start with 0x: %q", s)
	}
	n, err := strconv.ParseUint(body, 16, 64)
	if err != nil {
		fatalf("invalid hex word %q: %v", s, err)
	}
	return n
}

// parseSeed parses a `seed` operand: a DECIMAL STRING parsed straight to uint64
// (NEVER via float64), reusing the i64-suite's decimal-string discipline. A
// bare JSON number is also accepted for small seeds.
func parseSeed(v any) uint64 {
	switch t := v.(type) {
	case string:
		n, err := strconv.ParseUint(t, 10, 64)
		if err != nil {
			fatalf("invalid u64 decimal-string seed %q: %v", t, err)
		}
		return n
	case json.Number:
		n, err := strconv.ParseUint(t.String(), 10, 64)
		if err != nil {
			fatalf("invalid u64 seed %q: %v", t.String(), err)
		}
		return n
	case float64:
		return uint64(t)
	}
	fatalf("expected u64 seed (decimal string or number), got %T", v)
	return 0
}

// parseHexBytes parses a 0x-hex byte string (e.g. "0x01020304") to a byte slice.
func parseHexBytes(v any) []byte {
	s, ok := v.(string)
	if !ok {
		fatalf("hash-pipeline `bytes` must be a 0x-hex string, got %T", v)
	}
	body, ok := strings.CutPrefix(s, "0x")
	if !ok {
		body, ok = strings.CutPrefix(s, "0X")
	}
	if !ok {
		fatalf("hash-pipeline bytes must start with 0x: %q", s)
	}
	if len(body)%2 != 0 {
		fatalf("hash-pipeline bytes must have an even hex-digit count: %q", s)
	}
	out := make([]byte, len(body)/2)
	for i := range out {
		b, err := strconv.ParseUint(body[2*i:2*i+2], 16, 8)
		if err != nil {
			fatalf("invalid hex byte in %q: %v", s, err)
		}
		out[i] = byte(b)
	}
	return out
}

// ---- HyperLogLog (spec/features/hyperloglog.md) ---------------------------
//
// A stored cardinality sketch. The cross-language oracle is the INTEGER
// register array (via register_hex / nonzero_registers / max_register /
// register_at_N) -- NEVER the float estimate (float-quarantine Rule Q1; there
// is deliberately NO estimate assertion key here). Exactly one builder op,
// first: either a with_precision(p) (then zero or more add/merge) OR a single
// from_bytes. Zero/two builders or an add before the builder => malformed =>
// SKIP. A merge consumes the scenario's `other` HyperLogLog. Unknown
// ops/keys/kinds SKIP (forward-compat).

// buildHLL builds a HyperLogLog from an op list (used for the primary and the
// `other` block). Returns ok=false (=> caller SKIPs) when the op list is
// malformed for the harness: not starting with exactly one builder, an
// add/merge before the builder, an out-of-range with_precision, or a bad
// from_bytes.
func buildHLL(operations []map[string]any, other *otherSpec) (hyperloglog.HyperLogLog, bool) {
	if len(operations) == 0 {
		return hyperloglog.HyperLogLog{}, false
	}
	first := operations[0]
	var hll hyperloglog.HyperLogLog
	switch first["op"] {
	case "with_precision":
		// Non-fatal operand parse: a missing/mistyped `p` is a malformed
		// scenario -> SKIP (mirrors Rust build_hll's `first["p"].as_u64()?`,
		// which returns None rather than panicking).
		pv, ok := tryInt(first["p"])
		if !ok {
			fmt.Fprintln(os.Stderr, "skip: HyperLogLog with_precision needs an integer p (forward-compat)")
			return hyperloglog.HyperLogLog{}, false
		}
		p := uint8(pv)
		h, err := hyperloglog.NewHyperLogLogWithPrecision(p)
		if err != nil {
			// Out-of-range p is a construction error -> SKIP (the harness cannot
			// build the probe). The native tests pin the error path itself.
			fmt.Fprintf(os.Stderr, "skip: HyperLogLog with_precision error: %v\n", err)
			return hyperloglog.HyperLogLog{}, false
		}
		hll = h
	case "from_bytes":
		// from_bytes is the SOLE op when present (full state replacement, first
		// op or malformed). Reject any trailing ops.
		if len(operations) != 1 {
			fmt.Fprintln(os.Stderr, "skip: from_bytes must be the only op (forward-compat)")
			return hyperloglog.HyperLogLog{}, false
		}
		b := parseHexBytes(first["bytes"])
		h, err := hyperloglog.HyperLogLogFromBytes(b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip: HyperLogLog from_bytes error: %v\n", err)
			return hyperloglog.HyperLogLog{}, false
		}
		hll = h
	default:
		fmt.Fprintln(os.Stderr, "skip: HyperLogLog first op must be a builder (forward-compat)")
		return hyperloglog.HyperLogLog{}, false
	}

	for _, op := range operations[1:] {
		switch op["op"] {
		case "add":
			// Non-fatal operand parse (mirrors Rust `op["value"].as_i64()?`):
			// a missing/mistyped `value` => malformed => SKIP.
			v, ok := tryInt32(op["value"])
			if !ok {
				fmt.Fprintln(os.Stderr, "skip: HyperLogLog add needs an integer value (forward-compat)")
				return hyperloglog.HyperLogLog{}, false
			}
			hll.Add(v)
		case "merge":
			// Merge the scenario's `other` HyperLogLog (built by its own op list)
			// by element-wise register max.
			if other == nil {
				fmt.Fprintln(os.Stderr, "skip: HyperLogLog merge needs an `other` block (forward-compat)")
				return hyperloglog.HyperLogLog{}, false
			}
			otherHLL, ok := buildHLL(other.Operations, nil)
			if !ok {
				return hyperloglog.HyperLogLog{}, false
			}
			if err := hll.Merge(&otherHLL); err != nil {
				fmt.Fprintf(os.Stderr, "skip: HyperLogLog merge error: %v\n", err)
				return hyperloglog.HyperLogLog{}, false
			}
		default:
			fmt.Fprintf(os.Stderr, "skip: unknown HyperLogLog op (forward-compat): %v\n", op["op"])
			return hyperloglog.HyperLogLog{}, false
		}
	}
	return hll, true
}

func runHyperLogLog(s scenario) {
	hll, ok := buildHLL(s.Operations, s.Other)
	if !ok {
		fmt.Fprintln(os.Stderr, "skip: malformed HyperLogLog scenario (forward-compat)")
		return
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		emit(s.Name, key, evalHLLAssertion(key, &hll), s.Assertions[key], modeNone)
	}
}

func evalHLLAssertion(key string, hll *hyperloglog.HyperLogLog) string {
	switch {
	case key == "register_hex":
		// The PRIMARY integer oracle: the full serialized form (HLL1 + p +
		// register bytes) as a lower-case, 0x-prefixed hex string.
		var sb strings.Builder
		sb.WriteString("0x")
		for _, b := range hll.ToBytes() {
			fmt.Fprintf(&sb, "%02x", b)
		}
		return sb.String()
	case key == "nonzero_registers":
		return strconv.FormatUint(uint64(hll.NonzeroRegisters()), 10)
	case key == "max_register":
		return strconv.FormatUint(uint64(hll.MaxRegister()), 10)
	// NOTE: there is deliberately NO estimate key (float-quarantine Q1).
	case strings.HasPrefix(key, "register_at_"):
		n, err := strconv.Atoi(key[len("register_at_"):])
		if err != nil || n < 0 || n >= hll.RegisterCount() {
			return unknown(key)
		}
		return strconv.FormatUint(uint64(hll.Registers()[n]), 10)
	default:
		return unknown(key)
	}
}

// ---- HashMap<i32, i32> ---------------------------------------------------

func runHashMap(s scenario) {
	m := hashmap.NewInt32Int32()
	for _, op := range s.Operations {
		switch op["op"] {
		case "put":
			m.Put(asInt32(op["key"]), asInt32(op["value"]))
		case "remove":
			m.Remove(asInt32(op["key"]))
		case "addToValue":
			m.AddToValue(asInt32(op["key"]), asInt32(op["delta"]))
		case "clear":
			m.Clear()
		default:
			fatalf("unknown hashmap op: %v", op["op"])
		}
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		emit(s.Name, key, evalMapAssertion(key, m), s.Assertions[key], modeNone)
	}
}

func evalMapAssertion(key string, m *hashmap.Int32Int32) string {
	switch key {
	case "size":
		return strconv.Itoa(m.Len())
	case "is_empty":
		return strconv.FormatBool(m.Len() == 0)
	case "sorted_keys":
		keys := m.KeysToSlice()
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		return formatArray(keys)
	case "sorted_values":
		vals := m.ValuesToSlice()
		sort.Slice(vals, func(i, j int) bool { return vals[i] < vals[j] })
		return formatArray(vals)
	case "min":
		keys := m.KeysToSlice()
		if len(keys) == 0 {
			return "null"
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		return strconv.FormatInt(int64(keys[0]), 10)
	case "max":
		keys := m.KeysToSlice()
		if len(keys) == 0 {
			return "null"
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		return strconv.FormatInt(int64(keys[len(keys)-1]), 10)
	}
	if rest, ok := strings.CutPrefix(key, "get_"); ok {
		k, _ := strconv.ParseInt(rest, 10, 32)
		if v, found := m.Get(int32(k)); found {
			return strconv.FormatInt(int64(v), 10)
		}
		return "null"
	}
	if rest, ok := strings.CutPrefix(key, "contains_"); ok {
		k, _ := strconv.ParseInt(rest, 10, 32)
		return strconv.FormatBool(m.ContainsKey(int32(k)))
	}
	return unknown(key)
}

// ---- HashMap<i64, i32> ---------------------------------------------------

// parseI64Operand parses a Wide-integer (i64) key operand — see
// cross-language-validation/README.md §"Wide-integer (i64) operand encoding".
// An i64 KEY is a decimal STRING (small keys may also be bare JSON numbers),
// parsed straight to int64 via strconv.ParseInt(.,10,64) — never via f64. The
// value stays an i32 JSON number.
func parseI64Operand(v any) int64 {
	switch n := v.(type) {
	case string:
		i, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			fatalf("invalid or out-of-range i64 decimal-string key: %q", n)
		}
		return i
	case json.Number:
		// Lossless: parse the exact decimal token (never via f64). Fatal on
		// out-of-range so a bare 9223372036854775808 (2^63) can't silently wrap.
		i, err := strconv.ParseInt(n.String(), 10, 64)
		if err != nil {
			fatalf("invalid or out-of-range i64 number key: %q", n.String())
		}
		return i
	}
	fatalf("expected i64 key (decimal string or number), got %T (%v)", v, v)
	return 0
}

// runI64HashMap routes through the PRODUCTION Int64Int32HashMap (real i64 hash
// spread + key identity).
func runI64HashMap(s scenario) {
	m := hashmap.NewInt64Int32()
	for _, op := range s.Operations {
		switch op["op"] {
		case "put":
			m.Put(parseI64Operand(op["key"]), asInt32(op["value"]))
		case "remove":
			m.Remove(parseI64Operand(op["key"]))
		case "clear":
			m.Clear()
		default:
			fatalf("unknown i64-hashmap op: %v", op["op"])
		}
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		emit(s.Name, key, evalI64MapAssertion(key, m), s.Assertions[key], modeNone)
	}
}

func evalI64MapAssertion(key string, m *hashmap.Int64Int32) string {
	switch key {
	case "size":
		return strconv.Itoa(m.Len())
	case "is_empty":
		return strconv.FormatBool(m.Len() == 0)
	case "sorted_keys":
		// i64 keys exceed 2^53: serialize each as a plain decimal STRING in a
		// quoted array, sorted numerically as i64 ascending.
		keys := m.KeysToSlice()
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = "\"" + strconv.FormatInt(k, 10) + "\""
		}
		return "[" + strings.Join(parts, ",") + "]"
	}
	if rest, ok := strings.CutPrefix(key, "get_"); ok {
		k, err := strconv.ParseInt(rest, 10, 64)
		if err != nil {
			fatalf("invalid or out-of-range i64 get_ suffix: %q", rest)
		}
		if v, found := m.Get(k); found {
			return strconv.FormatInt(int64(v), 10)
		}
		return "null"
	}
	if rest, ok := strings.CutPrefix(key, "contains_"); ok {
		k, err := strconv.ParseInt(rest, 10, 64)
		if err != nil {
			fatalf("invalid or out-of-range i64 contains_ suffix: %q", rest)
		}
		return strconv.FormatBool(m.ContainsKey(k))
	}
	return unknown(key)
}

// ---- {List,Set}Multimap<i64, i32> ----------------------------------------

// i64Multimap is the subset of the production List/SetMultimap API the
// validator exercises. Both Int64Int32ListMultimap and Int64Int32SetMultimap
// satisfy it; the multimaps back onto the Go BUILTIN map[int64][]int32 (stdlib
// i64 hashing), NOT the production OpenHashMap high-bit fold — so this verifies
// full-range i64 keys keep their identity (stay distinct and retrievable)
// through the stdlib map. It checks key identity, not bucket-distribution
// quality (which is the stdlib hasher's job and a native-test concern).
type i64Multimap interface {
	Put(key int64, value int32)
	Get(key int64) []int32
	RemoveAll(key int64) []int32
	ContainsKey(key int64) bool
	KeysCount() int
	Keys() []int64
}

func runI64ListMultimap(s scenario) { runI64Multimap(s, multimap.NewInt64Int32List()) }
func runI64SetMultimap(s scenario)  { runI64Multimap(s, multimap.NewInt64Int32Set()) }

func runI64Multimap(s scenario, m i64Multimap) {
	for _, op := range s.Operations {
		switch op["op"] {
		case "put":
			m.Put(parseI64Operand(op["key"]), asInt32(op["value"]))
		case "removeAll":
			m.RemoveAll(parseI64Operand(op["key"]))
		default:
			fatalf("unknown i64-multimap op: %v", op["op"])
		}
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		emit(s.Name, key, evalI64MultimapAssertion(key, m), s.Assertions[key], modeNone)
	}
}

func evalI64MultimapAssertion(key string, m i64Multimap) string {
	switch key {
	case "distinct_key_count":
		return strconv.Itoa(m.KeysCount())
	case "sorted_keys":
		// DISTINCT keys, ascending i64, each a quoted decimal string (i64 keys
		// exceed 2^53) — same serialization as the i64-HashMap sorted_keys.
		keys := m.Keys()
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = "\"" + strconv.FormatInt(k, 10) + "\""
		}
		return "[" + strings.Join(parts, ",") + "]"
	}
	if rest, ok := strings.CutPrefix(key, "get_"); ok {
		k, err := strconv.ParseInt(rest, 10, 64)
		if err != nil {
			fatalf("invalid or out-of-range i64 get_ suffix: %q", rest)
		}
		// Values for the key as an ascending-sorted i32 array (sort a COPY);
		// absent/removed key => []. Same compact emit as sorted_values.
		vals := append([]int32(nil), m.Get(k)...)
		sort.Slice(vals, func(i, j int) bool { return vals[i] < vals[j] })
		return formatArray(vals)
	}
	if rest, ok := strings.CutPrefix(key, "contains_key_"); ok {
		k, err := strconv.ParseInt(rest, 10, 64)
		if err != nil {
			fatalf("invalid or out-of-range i64 contains_key_ suffix: %q", rest)
		}
		return strconv.FormatBool(m.ContainsKey(k))
	}
	return unknown(key)
}

// ---- ArrayList<i32> ------------------------------------------------------

func runArrayList(s scenario) {
	l := arraylist.NewInt32()
	for _, op := range s.Operations {
		switch op["op"] {
		case "add":
			l.Add(asInt32(op["value"]))
		case "add_at":
			idx := asInt(op["index"])
			v := asInt32(op["value"])
			// Int32ArrayList has no insert-at; rebuild as a slice and reload.
			cur := snapshotList(l)
			next := make([]int32, 0, len(cur)+1)
			next = append(next, cur[:idx]...)
			next = append(next, v)
			next = append(next, cur[idx:]...)
			l.Clear()
			l.AddAll(next...)
		case "remove":
			l.Remove(asInt32(op["value"]))
		case "clear":
			l.Clear()
		default:
			fatalf("unknown arraylist op: %v", op["op"])
		}
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		emit(s.Name, key, evalListAssertion(key, l), s.Assertions[key], modeNone)
	}
}

// addInt32Wrapping / mulInt32Wrapping route the wrapping reductions through the
// production InjectInto accumulator (int32, wraps two's-complement) so the
// harness proves the production overflow contract rather than a local copy.
func addInt32Wrapping(acc, v int32) int32 { return acc + v }
func mulInt32Wrapping(acc, v int32) int32 { return acc * v }

func snapshotList(l *arraylist.Int32) []int32 {
	out := make([]int32, 0, l.Len())
	l.ForEach(func(v int32) { out = append(out, v) })
	return out
}

func evalListAssertion(key string, l *arraylist.Int32) string {
	values := snapshotList(l)
	switch key {
	case "size":
		return strconv.Itoa(len(values))
	case "is_empty":
		return strconv.FormatBool(len(values) == 0)
	case "sum":
		// List Sum() widens into an int64 accumulator (IntList.sum(): long
		// parity) and does NOT wrap at i32 -- see algorithms.md "Integer
		// overflow contract" and scenarios/06-overflow/i32_sum_overflow.json.
		return strconv.FormatInt(l.Sum(), 10)
	case "inject_into_wrapping_product", "product":
		return strconv.FormatInt(int64(l.InjectInto(1, mulInt32Wrapping)), 10)
	case "max_minus_min":
		if l.Len() == 0 {
			return "null"
		}
		mn, _ := l.Min()
		mx, _ := l.Max()
		return strconv.FormatInt(int64(mx-mn), 10)
	case "min":
		if mn, ok := l.Min(); ok {
			return strconv.FormatInt(int64(mn), 10)
		}
		return "null"
	case "max":
		if mx, ok := l.Max(); ok {
			return strconv.FormatInt(int64(mx), 10)
		}
		return "null"
	case "to_sorted_array":
		// Sort a COPY so this assertion never mutates the production list --
		// otherwise a later order-sensitive assertion on the same list would
		// see the reordered elements. We still exercise the production Sort().
		sorted := arraylist.NewInt32()
		sorted.AddAll(values...)
		sorted.Sort()
		return formatArray(snapshotList(sorted))
	case "inject_into_sum":
		// injectInto with a + reduction accumulates in the i32 seed type and
		// wraps two's-complement at i32 -- via the production InjectInto.
		return strconv.FormatInt(int64(l.InjectInto(0, addInt32Wrapping)), 10)
	case "inject_into_product":
		var acc int64 = 1
		for _, v := range values {
			acc *= int64(v)
		}
		return strconv.FormatInt(acc, 10)
	case "any_satisfy_even":
		for _, v := range values {
			if v%2 == 0 {
				return "true"
			}
		}
		return "false"
	case "all_satisfy_even":
		for _, v := range values {
			if v%2 != 0 {
				return "false"
			}
		}
		return "true"
	case "none_satisfy_odd":
		for _, v := range values {
			if v%2 != 0 {
				return "false"
			}
		}
		return "true"
	case "count_even":
		c := 0
		for _, v := range values {
			if v%2 == 0 {
				c++
			}
		}
		return strconv.Itoa(c)
	case "count_odd":
		c := 0
		for _, v := range values {
			if v%2 != 0 {
				c++
			}
		}
		return strconv.Itoa(c)
	}
	if rest, ok := strings.CutPrefix(key, "get_at_"); ok {
		idx, _ := strconv.Atoi(rest)
		if idx < 0 || idx >= len(values) {
			return "null"
		}
		return strconv.FormatInt(int64(values[idx]), 10)
	}
	if rest, ok := strings.CutPrefix(key, "contains_"); ok {
		v, _ := strconv.ParseInt(rest, 10, 32)
		for _, x := range values {
			if x == int32(v) {
				return "true"
			}
		}
		return "false"
	}
	if rest, ok := strings.CutPrefix(key, "select_gt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		var v []int32
		for _, x := range values {
			if x > int32(t) {
				v = append(v, x)
			}
		}
		sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
		return formatArray(v)
	}
	if rest, ok := strings.CutPrefix(key, "reject_gt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		var v []int32
		for _, x := range values {
			if x <= int32(t) {
				v = append(v, x)
			}
		}
		sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
		return formatArray(v)
	}
	if rest, ok := strings.CutPrefix(key, "detect_gt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		for _, x := range values {
			if x > int32(t) {
				return strconv.FormatInt(int64(x), 10)
			}
		}
		return "null"
	}
	if rest, ok := strings.CutPrefix(key, "count_gt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		c := 0
		for _, x := range values {
			if x > int32(t) {
				c++
			}
		}
		return strconv.Itoa(c)
	}
	if rest, ok := strings.CutPrefix(key, "count_lt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		c := 0
		for _, x := range values {
			if x < int32(t) {
				c++
			}
		}
		return strconv.Itoa(c)
	}
	if rest, ok := strings.CutPrefix(key, "any_satisfy_gt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		for _, x := range values {
			if x > int32(t) {
				return "true"
			}
		}
		return "false"
	}
	if rest, ok := strings.CutPrefix(key, "all_satisfy_gt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		for _, x := range values {
			if x <= int32(t) {
				return "false"
			}
		}
		return "true"
	}
	if rest, ok := strings.CutPrefix(key, "none_satisfy_gt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		for _, x := range values {
			if x > int32(t) {
				return "false"
			}
		}
		return "true"
	}
	if rest, ok := strings.CutPrefix(key, "none_satisfy_lt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		for _, x := range values {
			if x < int32(t) {
				return "false"
			}
		}
		return "true"
	}
	return unknown(key)
}

// ---- HashSet<i32> --------------------------------------------------------

func runHashSet(s scenario) {
	set := hashset.NewInt32()
	for _, op := range s.Operations {
		switch op["op"] {
		case "add":
			set.Add(asInt32(op["value"]))
		case "remove":
			set.Remove(asInt32(op["value"]))
		case "clear":
			set.Clear()
		default:
			fatalf("unknown hashset op: %v", op["op"])
		}
	}
	var other *hashset.Int32
	if s.Other != nil {
		other = hashset.NewInt32()
		for _, op := range s.Other.Operations {
			if op["op"] == "add" {
				other.Add(asInt32(op["value"]))
			}
		}
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		emit(s.Name, key, evalSetAssertion(key, set, other), s.Assertions[key], modeNone)
	}
}

func setToSorted(set *hashset.Int32) []int32 {
	out := make([]int32, 0, set.Len())
	set.ForEach(func(v int32) { out = append(out, v) })
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func evalSetAssertion(key string, set, other *hashset.Int32) string {
	switch key {
	case "size":
		return strconv.Itoa(set.Len())
	case "is_empty":
		return strconv.FormatBool(set.Len() == 0)
	case "to_sorted_array":
		return formatArray(setToSorted(set))
	}
	if other != nil {
		switch key {
		case "union_sorted":
			seen := map[int32]struct{}{}
			set.ForEach(func(v int32) { seen[v] = struct{}{} })
			other.ForEach(func(v int32) { seen[v] = struct{}{} })
			v := make([]int32, 0, len(seen))
			for k := range seen {
				v = append(v, k)
			}
			sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
			return formatArray(v)
		case "intersect_sorted":
			var v []int32
			set.ForEach(func(x int32) {
				if other.Contains(x) {
					v = append(v, x)
				}
			})
			sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
			return formatArray(v)
		case "difference_sorted":
			var v []int32
			set.ForEach(func(x int32) {
				if !other.Contains(x) {
					v = append(v, x)
				}
			})
			sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
			return formatArray(v)
		case "symmetric_difference_sorted":
			var v []int32
			set.ForEach(func(x int32) {
				if !other.Contains(x) {
					v = append(v, x)
				}
			})
			other.ForEach(func(x int32) {
				if !set.Contains(x) {
					v = append(v, x)
				}
			})
			sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
			return formatArray(v)
		case "union_size":
			seen := map[int32]struct{}{}
			set.ForEach(func(v int32) { seen[v] = struct{}{} })
			other.ForEach(func(v int32) { seen[v] = struct{}{} })
			return strconv.Itoa(len(seen))
		case "intersect_size":
			c := 0
			set.ForEach(func(x int32) {
				if other.Contains(x) {
					c++
				}
			})
			return strconv.Itoa(c)
		case "difference_size":
			c := 0
			set.ForEach(func(x int32) {
				if !other.Contains(x) {
					c++
				}
			})
			return strconv.Itoa(c)
		case "symmetric_difference_size":
			seen := map[int32]struct{}{}
			set.ForEach(func(x int32) {
				if !other.Contains(x) {
					seen[x] = struct{}{}
				}
			})
			other.ForEach(func(x int32) {
				if !set.Contains(x) {
					seen[x] = struct{}{}
				}
			})
			return strconv.Itoa(len(seen))
		case "other_size":
			return strconv.Itoa(other.Len())
		}
	}
	if rest, ok := strings.CutPrefix(key, "contains_"); ok {
		v, _ := strconv.ParseInt(rest, 10, 32)
		return strconv.FormatBool(set.Contains(int32(v)))
	}
	return unknown(key)
}

// ---- HashBag<i32> --------------------------------------------------------

func runHashBag(s scenario) {
	b := bag.NewHashInt32()
	for _, op := range s.Operations {
		switch op["op"] {
		case "add":
			b.Add(asInt32(op["value"]))
		case "remove":
			b.Remove(asInt32(op["value"]))
		case "clear":
			b.Clear()
		default:
			fatalf("unknown hashbag op: %v", op["op"])
		}
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		emit(s.Name, key, evalBagAssertion(key, b), s.Assertions[key], modeNone)
	}
}

func evalBagAssertion(key string, b *bag.HashInt32) string {
	switch key {
	case "size":
		return strconv.Itoa(b.Len())
	case "size_distinct":
		return strconv.Itoa(b.SizeDistinct())
	case "is_empty":
		return strconv.FormatBool(b.Len() == 0)
	case "sorted_distinct":
		seen := map[int32]struct{}{}
		b.ForEach(func(v int32) { seen[v] = struct{}{} })
		v := make([]int32, 0, len(seen))
		for k := range seen {
			v = append(v, k)
		}
		sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
		return formatArray(v)
	case "to_sorted_array":
		flat := make([]int32, 0, b.Len())
		b.ForEachWithOccurrences(func(value int32, count int) {
			for i := 0; i < count; i++ {
				flat = append(flat, value)
			}
		})
		sort.Slice(flat, func(i, j int) bool { return flat[i] < flat[j] })
		return formatArray(flat)
	}
	if rest, ok := strings.CutPrefix(key, "occurrences_"); ok {
		v, _ := strconv.ParseInt(rest, 10, 32)
		return strconv.Itoa(b.OccurrencesOf(int32(v)))
	}
	if rest, ok := strings.CutPrefix(key, "contains_"); ok {
		v, _ := strconv.ParseInt(rest, 10, 32)
		return strconv.FormatBool(b.Contains(int32(v)))
	}
	return unknown(key)
}

// ---- NavigableMap / NavigableSet shared helpers --------------------------

// navLog records poll/remove_range return values in execution order while
// applying operations, so they are cross-language observable (see README
// §NavigableMap).
type navLog struct {
	pollFirstKeys    []*int32
	pollLastKeys     []*int32
	pollFirstValues  []*int32
	pollLastValues   []*int32
	removeRangeCount []int32
}

func i32p(v int32) *int32 { return &v }

// optArray renders an []*int32 as a JSON-ish array with null for absent.
func optArray(v []*int32) string {
	parts := make([]string, len(v))
	for i, x := range v {
		if x == nil {
			parts[i] = "null"
		} else {
			parts[i] = strconv.FormatInt(int64(*x), 10)
		}
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// navKeyPrefix recognises a floor_<k>/ceiling_<k>/lower_<k>/higher_<k> assertion
// key. <k> is parsed as a SIGNED base-10 i32 (leading '-' and the full i32 range
// allowed). Returns (kind, k, true) on a match.
func navKeyPrefix(key string) (string, int32, bool) {
	for _, prefix := range []string{"floor_", "ceiling_", "lower_", "higher_"} {
		if rest, ok := strings.CutPrefix(key, prefix); ok {
			n, err := strconv.ParseInt(rest, 10, 32)
			if err != nil {
				continue // not a nav key (suffix is not a signed i32)
			}
			return strings.TrimSuffix(prefix, "_"), int32(n), true
		}
	}
	return "", 0, false
}

// rankKey recognises a rank_<k> order-statistic assertion: <k> is a SIGNED
// base-10 i32 (exact ^rank_(-?[0-9]+)$, full i32 range incl. negatives). Returns
// (k, true) on a match. Rejects a leading '+' so the recogniser matches the
// documented regex (strconv.ParseInt would otherwise accept "+5").
func rankKey(key string) (int32, bool) {
	rest, ok := strings.CutPrefix(key, "rank_")
	if !ok {
		return 0, false
	}
	digits := strings.TrimPrefix(rest, "-")
	if digits == "" || !isAllDigits(digits) {
		return 0, false
	}
	n, err := strconv.ParseInt(rest, 10, 32)
	if err != nil {
		return 0, false
	}
	return int32(n), true
}

// selectIndex recognises a select_<i> order-statistic assertion: <i> is a
// NON-NEGATIVE base-10 index (exact ^select_([0-9]+)$). Returns (i, true) on a
// match. Must NOT match the functional predicate keys (select_gt_N, select_even,
// ...): those are not all-digits so they are rejected.
func selectIndex(key string) (int, bool) {
	rest, ok := strings.CutPrefix(key, "select_")
	if !ok {
		return 0, false
	}
	if rest == "" || !isAllDigits(rest) {
		return 0, false
	}
	n, err := strconv.Atoi(rest)
	if err != nil {
		return 0, false
	}
	return n, true
}

// isAllDigits reports whether s is non-empty and consists solely of ASCII
// digits 0-9 (no sign, no separators).
func isAllDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return s != ""
}

// ---- TreeSet<i32> --------------------------------------------------------

func runTreeSet(s scenario) {
	set := treeset.NewInt32()
	var log navLog
	for _, op := range s.Operations {
		switch op["op"] {
		case "add":
			set.Add(asInt32(op["value"]))
		case "remove":
			set.Remove(asInt32(op["value"]))
		case "clear":
			set.Clear()
		case "poll_first":
			if v, ok := set.PollFirst(); ok {
				log.pollFirstKeys = append(log.pollFirstKeys, i32p(v))
			} else {
				log.pollFirstKeys = append(log.pollFirstKeys, nil)
			}
		case "poll_last":
			if v, ok := set.PollLast(); ok {
				log.pollLastKeys = append(log.pollLastKeys, i32p(v))
			} else {
				log.pollLastKeys = append(log.pollLastKeys, nil)
			}
		case "remove_range":
			r := buildRangeObj(op["range"].(map[string]any))
			log.removeRangeCount = append(log.removeRangeCount, int32(set.RemoveRange(r)))
		default:
			// Forward-compat: an unknown op must not crash an older/newer runner
			// mix; skip it (mirrors unknown-collection/assertion skip).
		}
	}
	var query rangev.Int32Range
	hasQuery := s.Query != nil
	if hasQuery {
		query = buildRangeObj(s.Query)
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		val := func() string {
			switch key {
			case "size":
				return strconv.Itoa(set.Len())
			case "is_empty":
				return strconv.FormatBool(set.Len() == 0)
			case "min", "first":
				return optInt32Str(set.Min())
			case "max", "last":
				return optInt32Str(set.Max())
			case "to_sorted_array":
				return formatArray(set.ToSlice())
			case "descending_elements":
				return formatArray(set.Descending())
			case "range_elements":
				if hasQuery {
					return formatArray(set.RangeElements(query))
				}
				return unknown(key)
			case "range_elements_desc":
				if hasQuery {
					return formatArray(set.DescendingRangeElements(query))
				}
				return unknown(key)
			case "range_size":
				if hasQuery {
					return strconv.Itoa(len(set.RangeElements(query)))
				}
				return unknown(key)
			case "poll_first_keys":
				return optArray(log.pollFirstKeys)
			case "poll_last_keys":
				return optArray(log.pollLastKeys)
			case "remove_range_counts":
				return formatArray(log.removeRangeCount)
			}
			if kind, k, ok := navKeyPrefix(key); ok {
				switch kind {
				case "floor":
					return optInt32Str(set.Floor(k))
				case "ceiling":
					return optInt32Str(set.Ceiling(k))
				case "lower":
					return optInt32Str(set.Lower(k))
				case "higher":
					return optInt32Str(set.Higher(k))
				}
			}
			if k, ok := rankKey(key); ok {
				return strconv.Itoa(set.Rank(k))
			}
			if i, ok := selectIndex(key); ok {
				return optInt32Str(set.Select(i))
			}
			if rest, ok := strings.CutPrefix(key, "contains_"); ok {
				v, _ := strconv.ParseInt(rest, 10, 32)
				return strconv.FormatBool(set.Contains(int32(v)))
			}
			return unknown(key)
		}()
		emit(s.Name, key, val, s.Assertions[key], modeNone)
	}
}

// ---- TreeMap<i32, i32> ---------------------------------------------------

func runTreeMap(s scenario) {
	m := treemap.NewInt32Int32()
	var log navLog
	for _, op := range s.Operations {
		switch op["op"] {
		case "put":
			m.Put(asInt32(op["key"]), asInt32(op["value"]))
		case "remove":
			m.Remove(asInt32(op["key"]))
		case "clear":
			m.Clear()
		case "poll_first":
			if k, v, ok := m.PollFirstEntry(); ok {
				log.pollFirstKeys = append(log.pollFirstKeys, i32p(k))
				log.pollFirstValues = append(log.pollFirstValues, i32p(v))
			} else {
				log.pollFirstKeys = append(log.pollFirstKeys, nil)
				log.pollFirstValues = append(log.pollFirstValues, nil)
			}
		case "poll_last":
			if k, v, ok := m.PollLastEntry(); ok {
				log.pollLastKeys = append(log.pollLastKeys, i32p(k))
				log.pollLastValues = append(log.pollLastValues, i32p(v))
			} else {
				log.pollLastKeys = append(log.pollLastKeys, nil)
				log.pollLastValues = append(log.pollLastValues, nil)
			}
		case "remove_range":
			r := buildRangeObj(op["range"].(map[string]any))
			log.removeRangeCount = append(log.removeRangeCount, int32(m.RemoveRange(r)))
		default:
			// Forward-compat: skip unknown ops.
		}
	}
	var query rangev.Int32Range
	hasQuery := s.Query != nil
	if hasQuery {
		query = buildRangeObj(s.Query)
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		val := func() string {
			switch key {
			case "size":
				return strconv.Itoa(m.Len())
			case "is_empty":
				return strconv.FormatBool(m.Len() == 0)
			case "min", "first_key":
				k, _, ok := m.Min()
				return optInt32Str(k, ok)
			case "max", "last_key":
				k, _, ok := m.Max()
				return optInt32Str(k, ok)
			case "sorted_keys":
				var keys []int32
				for k := range m.Keys() {
					keys = append(keys, k)
				}
				return formatArray(keys)
			case "sorted_values":
				var vals []int32
				for _, v := range m.All() {
					vals = append(vals, v)
				}
				// Int32Int32 iterates in key order; values follow keys, which is
				// what "sorted_values" asks for in the cross-language contract.
				return formatArray(vals)
			case "descending_keys":
				var keys []int32
				for k := range m.DescendingKeys() {
					keys = append(keys, k)
				}
				return formatArray(keys)
			case "range_keys":
				if hasQuery {
					return formatArray(m.RangeKeysIn(query))
				}
				return unknown(key)
			case "range_keys_desc":
				if hasQuery {
					return formatArray(m.DescendingRangeKeys(query))
				}
				return unknown(key)
			case "range_size":
				if hasQuery {
					return strconv.Itoa(len(m.RangeKeysIn(query)))
				}
				return unknown(key)
			case "poll_first_keys":
				return optArray(log.pollFirstKeys)
			case "poll_last_keys":
				return optArray(log.pollLastKeys)
			case "poll_first_values":
				return optArray(log.pollFirstValues)
			case "poll_last_values":
				return optArray(log.pollLastValues)
			case "remove_range_counts":
				return formatArray(log.removeRangeCount)
			}
			if kind, k, ok := navKeyPrefix(key); ok {
				switch kind {
				case "floor":
					return optInt32Str(m.FloorKey(k))
				case "ceiling":
					return optInt32Str(m.CeilingKey(k))
				case "lower":
					return optInt32Str(m.LowerKey(k))
				case "higher":
					return optInt32Str(m.HigherKey(k))
				}
			}
			if k, ok := rankKey(key); ok {
				return strconv.Itoa(m.Rank(k))
			}
			if i, ok := selectIndex(key); ok {
				return optInt32Str(m.SelectKey(i))
			}
			if rest, ok := strings.CutPrefix(key, "get_"); ok {
				k, _ := strconv.ParseInt(rest, 10, 32)
				if v, ok := m.Get(int32(k)); ok {
					return strconv.FormatInt(int64(v), 10)
				}
				return "null"
			}
			if rest, ok := strings.CutPrefix(key, "contains_"); ok {
				k, _ := strconv.ParseInt(rest, 10, 32)
				return strconv.FormatBool(m.ContainsKey(int32(k)))
			}
			return unknown(key)
		}()
		emit(s.Name, key, val, s.Assertions[key], modeNone)
	}
}

// ---- HashMap<f32, i32> ---------------------------------------------------

func runF32HashMap(s scenario) {
	m := hashmap.NewFloat32Int32()
	for _, op := range s.Operations {
		switch op["op"] {
		case "put":
			m.Put(parseF32(op["key"]), asInt32(op["value"]))
		case "remove":
			m.Remove(parseF32(op["key"]))
		case "clear":
			m.Clear()
		default:
			fatalf("unknown f32-hashmap op: %v", op["op"])
		}
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		val := func() string {
			switch key {
			case "size":
				return strconv.Itoa(m.Len())
			case "is_empty":
				return strconv.FormatBool(m.Len() == 0)
			case "sorted_keys":
				var keys []float32
				for k := range m.Keys() {
					keys = append(keys, k)
				}
				sortFloat32Total(keys)
				parts := make([]string, len(keys))
				for i, x := range keys {
					parts[i] = "\"" + formatF32(x) + "\""
				}
				return "[" + strings.Join(parts, ",") + "]"
			}
			if rest, ok := strings.CutPrefix(key, "get_"); ok {
				probe := parseF32Label(rest)
				if v, ok := m.Get(probe); ok {
					return strconv.FormatInt(int64(v), 10)
				}
				return "null"
			}
			if rest, ok := strings.CutPrefix(key, "contains_"); ok {
				return strconv.FormatBool(m.ContainsKey(parseF32Label(rest)))
			}
			return unknown(key)
		}()
		emit(s.Name, key, val, s.Assertions[key], modeF32Keyed)
	}
}

// ---- HashSet<f32> --------------------------------------------------------

func runF32HashSet(s scenario) {
	set := hashset.NewFloat32()
	for _, op := range s.Operations {
		switch op["op"] {
		case "add":
			set.Add(parseF32(op["value"]))
		case "remove":
			set.Remove(parseF32(op["value"]))
		case "clear":
			set.Clear()
		default:
			fatalf("unknown f32-hashset op: %v", op["op"])
		}
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		val := func() string {
			switch key {
			case "size":
				return strconv.Itoa(set.Len())
			case "is_empty":
				return strconv.FormatBool(set.Len() == 0)
			case "sorted_values", "to_sorted_array":
				vals := make([]float32, 0, set.Len())
				set.ForEach(func(v float32) { vals = append(vals, v) })
				sortFloat32Total(vals)
				parts := make([]string, len(vals))
				for i, x := range vals {
					parts[i] = "\"" + formatF32(x) + "\""
				}
				return "[" + strings.Join(parts, ",") + "]"
			}
			if rest, ok := strings.CutPrefix(key, "contains_"); ok {
				return strconv.FormatBool(set.Contains(parseF32Label(rest)))
			}
			return unknown(key)
		}()
		emit(s.Name, key, val, s.Assertions[key], modeF32Keyed)
	}
}

// ---- TreeSet<f32> --------------------------------------------------------

// runF32TreeSet routes through the PRODUCTION treeset.Float32, whose
// node ordering is the cmpFloat32 sign-flip total order. The sorted output is
// the tree's in-order traversal (All()) -- NEVER sorted in the runner -- so
// this exercises the production float total-order comparator directly.
func runF32TreeSet(s scenario) {
	set := treeset.NewFloat32()
	for _, op := range s.Operations {
		switch op["op"] {
		case "add":
			set.Add(parseF32(op["value"]))
		case "remove":
			set.Remove(parseF32(op["value"]))
		case "clear":
			set.Clear()
		default:
			fatalf("unknown f32-treeset op: %v", op["op"])
		}
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		val := func() string {
			switch key {
			case "size":
				return strconv.Itoa(set.Len())
			case "is_empty":
				return strconv.FormatBool(set.Len() == 0)
			case "min":
				if mn, ok := set.Min(); ok {
					return formatF32(mn)
				}
				return "null"
			case "max":
				if mx, ok := set.Max(); ok {
					return formatF32(mx)
				}
				return "null"
			case "sorted", "sorted_values", "to_sorted_array":
				// In-order traversal straight from the production tree.
				var parts []string
				for v := range set.All() {
					parts = append(parts, "\""+formatF32(v)+"\"")
				}
				return "[" + strings.Join(parts, ",") + "]"
			}
			if rest, ok := strings.CutPrefix(key, "contains_"); ok {
				return strconv.FormatBool(set.Contains(parseF32Label(rest)))
			}
			return unknown(key)
		}()
		emit(s.Name, key, val, s.Assertions[key], modeF32Keyed)
	}
}

// ---- ArrayList<f32> ------------------------------------------------------

func runF32ArrayList(s scenario) {
	l := arraylist.NewFloat32()
	for _, op := range s.Operations {
		switch op["op"] {
		case "add":
			l.Add(parseF32(op["value"]))
		case "clear":
			l.Clear()
		default:
			fatalf("unknown f32-arraylist op: %v", op["op"])
		}
	}
	values := make([]float32, 0, l.Len())
	l.ForEach(func(v float32) { values = append(values, v) })
	for _, key := range sortedAssertionKeys(s.Assertions) {
		val := func() string {
			switch key {
			case "size":
				return strconv.Itoa(len(values))
			case "is_empty":
				return strconv.FormatBool(len(values) == 0)
			case "sum":
				// Production float sum (left-fold, IEEE arithmetic).
				return formatF32(l.Sum())
			case "min":
				// Production Float32ArrayList.Min now uses the total-order
				// comparator (cmpFloat32), so the NaN/±0 min scenarios are
				// proved against the real collection code.
				if mn, ok := l.Min(); ok {
					return formatF32(mn)
				}
				return "null"
			case "max":
				if mx, ok := l.Max(); ok {
					return formatF32(mx)
				}
				return "null"
			case "sorted", "to_sorted_array":
				// Sort a COPY through the production total-order Sort() so the
				// assertion proves conformance without mutating the live list.
				sorted := arraylist.NewFloat32()
				sorted.AddAll(values...)
				sorted.Sort()
				parts := make([]string, 0, sorted.Len())
				sorted.ForEach(func(x float32) { parts = append(parts, formatF32(x)) })
				return "[" + strings.Join(parts, ",") + "]"
			}
			return unknown(key)
		}()
		emit(s.Name, key, val, s.Assertions[key], modeF32List)
	}
}

// sortFloat32Total sorts in IEEE total order (matches Rust's
// `f32::total_cmp` and the bit-pattern ordering used in
// algorithms.md §"Float ordering for tree collections"). Required so
// NaN and +0/-0 sort deterministically across ports.
func sortFloat32Total(v []float32) {
	sort.Slice(v, func(i, j int) bool { return totalCmpF32(v[i], v[j]) < 0 })
}

func totalCmpF32(a, b float32) int {
	ai := int32(math.Float32bits(a))
	bi := int32(math.Float32bits(b))
	// Flip the sign bit so a lexicographic int32 compare matches the
	// IEEE total order. Same trick Rust's total_cmp uses internally.
	ai ^= int32(uint32(ai>>31) >> 1)
	bi ^= int32(uint32(bi>>31) >> 1)
	switch {
	case ai < bi:
		return -1
	case ai > bi:
		return 1
	}
	return 0
}

// ---- Range<i32> ----------------------------------------------------------

// The Bound/Range value model (spec/features/bound-range.md). Exactly ONE
// constructor op builds the range under test; an optional "other" block (same
// single-builder shape) supplies the second range for binary ops. Routed
// through the production rangev.Int32Range — every assertion is proved against
// the real cut algebra, not re-derived here.
func buildRange(ops []map[string]any) rangev.Int32Range {
	if len(ops) != 1 {
		fatalf("Range<i32> scenario must have exactly one constructor op, got %d", len(ops))
	}
	return buildRangeObj(ops[0])
}

// buildRangeObj builds an Int32Range from a single range-builder object (the
// 10-range op shape). Shared by the Range<i32> runner and the NavigableMap/Set
// `range`/`query` fields.
func buildRangeObj(op map[string]any) rangev.Int32Range {
	lower := func() int32 { return asInt32(op["lower"]) }
	upper := func() int32 { return asInt32(op["upper"]) }
	switch op["op"] {
	case "closed":
		return rangev.Closed(lower(), upper())
	case "open":
		return rangev.Open(lower(), upper())
	case "closed_open":
		return rangev.ClosedOpen(lower(), upper())
	case "open_closed":
		return rangev.OpenClosed(lower(), upper())
	case "at_least":
		return rangev.AtLeast(lower())
	case "greater_than":
		return rangev.GreaterThan(lower())
	case "at_most":
		return rangev.AtMost(upper())
	case "less_than":
		return rangev.LessThan(upper())
	case "all":
		return rangev.All()
	case "singleton":
		return rangev.Singleton(asInt32(op["value"]))
	default:
		fatalf("unknown range op: %v", op["op"])
		return rangev.Int32Range{}
	}
}

func boundTypeStr(bt rangev.BoundType, ok bool) string {
	if !ok {
		return "null"
	}
	return bt.String()
}

func optInt32Str(v int32, ok bool) string {
	if !ok {
		return "null"
	}
	return strconv.FormatInt(int64(v), 10)
}

func runRange(s scenario) {
	r := buildRange(s.Operations)
	var other rangev.Int32Range
	hasOther := false
	if s.Other != nil {
		other = buildRange(s.Other.Operations)
		hasOther = true
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		emit(s.Name, key, evalRangeAssertion(key, r, other, hasOther), s.Assertions[key], modeNone)
	}
}

func evalRangeAssertion(key string, r, other rangev.Int32Range, hasOther bool) string {
	switch key {
	case "is_empty":
		return strconv.FormatBool(r.IsEmpty())
	case "has_lower_bound":
		return strconv.FormatBool(r.HasLowerBound())
	case "has_upper_bound":
		return strconv.FormatBool(r.HasUpperBound())
	case "lower_bound_type":
		return boundTypeStr(r.LowerBoundType())
	case "upper_bound_type":
		return boundTypeStr(r.UpperBoundType())
	case "lower_endpoint":
		return optInt32Str(r.LowerEndpoint())
	case "upper_endpoint":
		return optInt32Str(r.UpperEndpoint())
	}
	if rest, ok := strings.CutPrefix(key, "contains_"); ok {
		n, err := strconv.ParseInt(rest, 10, 32)
		if err != nil {
			fatalf("invalid contains_<N> integer: %q", rest)
		}
		return strconv.FormatBool(r.Contains(int32(n)))
	}
	// ---- binary ops: require "other" ----------------------------------
	if !hasOther {
		return unknown(key)
	}
	switch key {
	case "encloses_other":
		return strconv.FormatBool(r.Encloses(other))
	case "is_connected_other":
		return strconv.FormatBool(r.IsConnected(other))
	case "span_lower":
		return optInt32Str(r.Span(other).LowerEndpoint())
	case "span_upper":
		return optInt32Str(r.Span(other).UpperEndpoint())
	case "span_lower_type":
		return boundTypeStr(r.Span(other).LowerBoundType())
	case "span_upper_type":
		return boundTypeStr(r.Span(other).UpperBoundType())
	case "intersection_is_none":
		_, ok := r.Intersection(other)
		return strconv.FormatBool(!ok)
	case "intersection_is_empty":
		i, ok := r.Intersection(other)
		return strconv.FormatBool(ok && i.IsEmpty())
	case "intersection_lower":
		if i, ok := r.Intersection(other); ok {
			return optInt32Str(i.LowerEndpoint())
		}
		return "null"
	case "intersection_upper":
		if i, ok := r.Intersection(other); ok {
			return optInt32Str(i.UpperEndpoint())
		}
		return "null"
	case "intersection_lower_type":
		if i, ok := r.Intersection(other); ok {
			return boundTypeStr(i.LowerBoundType())
		}
		return "null"
	case "intersection_upper_type":
		if i, ok := r.Intersection(other); ok {
			return boundTypeStr(i.UpperBoundType())
		}
		return "null"
	case "intersection_has_lower_bound":
		i, ok := r.Intersection(other)
		return strconv.FormatBool(ok && i.HasLowerBound())
	case "intersection_has_upper_bound":
		i, ok := r.Intersection(other)
		return strconv.FormatBool(ok && i.HasUpperBound())
	}
	return unknown(key)
}

// ---- ImmutableSortedMap<i32, i32> / ImmutableSortedSet<i32> --------------
//
// The compact immutable sorted table (spec/features/sorted-table-map.md). A
// sorted-table collection is built by a SINGLE bulk `from_sorted` op (it has no
// incremental mutators). The runner REQUIRES exactly one `from_sorted` op; zero
// or multiple is a malformed scenario and is SKIPped (forward-compat, not a
// failure), per the spec authoring rule. The optional top-level `query` names
// the Range the range_* assertions refer to.

// int32Array converts a decoded JSON array (a []any of json.Number under
// UseNumber) into an []int32. Used for the from_sorted keys/values/elements.
func int32Array(v any) []int32 {
	raw, ok := v.([]any)
	if !ok {
		fatalf("expected JSON array, got %T", v)
	}
	out := make([]int32, len(raw))
	for i, e := range raw {
		out[i] = asInt32(e)
	}
	return out
}

// singleFromSorted returns the lone from_sorted op, or (nil, false) if the
// operations are not exactly one from_sorted op (the caller then SKIPs).
func singleFromSorted(s scenario) (map[string]any, bool) {
	if len(s.Operations) != 1 {
		return nil, false
	}
	op := s.Operations[0]
	if op["op"] != "from_sorted" {
		return nil, false
	}
	return op, true
}

func runImmutableSortedMap(s scenario) {
	op, ok := singleFromSorted(s)
	if !ok {
		fmt.Fprintf(os.Stderr, "skip: malformed sorted-table scenario (need exactly one from_sorted op): %s\n", s.Name)
		return
	}
	keys := int32Array(op["keys"])
	values := int32Array(op["values"])
	m := immutablesorted.FromSortedInt32Int32(keys, values)

	var query rangev.Int32Range
	hasQuery := s.Query != nil
	if hasQuery {
		query = buildRangeObj(s.Query)
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		val := func() string {
			switch key {
			case "size":
				return strconv.Itoa(m.Size())
			case "is_empty":
				return strconv.FormatBool(m.IsEmpty())
			case "sorted_keys":
				return formatArray(m.Keys())
			case "sorted_values":
				// "all values, sorted ascending" (README). For this type the
				// value MULTISET sorted; Values() is in key order, so sort a copy.
				vals := m.Values()
				sort.Slice(vals, func(i, j int) bool { return vals[i] < vals[j] })
				return formatArray(vals)
			case "min", "first_key":
				return optInt32Str(m.FirstKey())
			case "max", "last_key":
				return optInt32Str(m.LastKey())
			case "descending_keys":
				return formatArray(m.DescendingKeys())
			case "range_keys":
				if hasQuery {
					return formatArray(m.RangeKeys(query))
				}
				return unknown(key)
			case "range_keys_desc":
				if hasQuery {
					return formatArray(m.DescendingRangeKeys(query))
				}
				return unknown(key)
			case "range_size":
				if hasQuery {
					return strconv.Itoa(len(m.RangeKeys(query)))
				}
				return unknown(key)
			}
			if kind, k, ok := navKeyPrefix(key); ok {
				switch kind {
				case "floor":
					return optInt32Str(m.FloorKey(k))
				case "ceiling":
					return optInt32Str(m.CeilingKey(k))
				case "lower":
					return optInt32Str(m.LowerKey(k))
				case "higher":
					return optInt32Str(m.HigherKey(k))
				}
			}
			if k, ok := rankKey(key); ok {
				return strconv.Itoa(m.Rank(k))
			}
			if i, ok := selectIndex(key); ok {
				return optInt32Str(m.SelectKey(i))
			}
			if rest, ok := strings.CutPrefix(key, "get_"); ok {
				k, err := strconv.ParseInt(rest, 10, 32)
				if err != nil {
					return unknown(key)
				}
				if v, ok := m.Get(int32(k)); ok {
					return strconv.FormatInt(int64(v), 10)
				}
				return "null"
			}
			if rest, ok := strings.CutPrefix(key, "contains_"); ok {
				k, err := strconv.ParseInt(rest, 10, 32)
				if err != nil {
					return unknown(key)
				}
				return strconv.FormatBool(m.ContainsKey(int32(k)))
			}
			return unknown(key)
		}()
		emit(s.Name, key, val, s.Assertions[key], modeNone)
	}
}

func runImmutableSortedSet(s scenario) {
	op, ok := singleFromSorted(s)
	if !ok {
		fmt.Fprintf(os.Stderr, "skip: malformed sorted-table scenario (need exactly one from_sorted op): %s\n", s.Name)
		return
	}
	elements := int32Array(op["elements"])
	set := immutablesorted.FromSortedInt32(elements)

	var query rangev.Int32Range
	hasQuery := s.Query != nil
	if hasQuery {
		query = buildRangeObj(s.Query)
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		val := func() string {
			switch key {
			case "size":
				return strconv.Itoa(set.Size())
			case "is_empty":
				return strconv.FormatBool(set.IsEmpty())
			case "to_sorted_array":
				return formatArray(set.Elements())
			case "min", "first":
				return optInt32Str(set.First())
			case "max", "last":
				return optInt32Str(set.Last())
			case "descending_elements":
				return formatArray(set.DescendingElements())
			case "range_elements":
				if hasQuery {
					return formatArray(set.RangeElements(query))
				}
				return unknown(key)
			case "range_elements_desc":
				if hasQuery {
					return formatArray(set.DescendingRangeElements(query))
				}
				return unknown(key)
			case "range_size":
				if hasQuery {
					return strconv.Itoa(len(set.RangeElements(query)))
				}
				return unknown(key)
			}
			if kind, k, ok := navKeyPrefix(key); ok {
				switch kind {
				case "floor":
					return optInt32Str(set.Floor(k))
				case "ceiling":
					return optInt32Str(set.Ceiling(k))
				case "lower":
					return optInt32Str(set.Lower(k))
				case "higher":
					return optInt32Str(set.Higher(k))
				}
			}
			if k, ok := rankKey(key); ok {
				return strconv.Itoa(set.Rank(k))
			}
			if i, ok := selectIndex(key); ok {
				return optInt32Str(set.Select(i))
			}
			if rest, ok := strings.CutPrefix(key, "contains_"); ok {
				k, err := strconv.ParseInt(rest, 10, 32)
				if err != nil {
					return unknown(key)
				}
				return strconv.FormatBool(set.Contains(int32(k)))
			}
			return unknown(key)
		}()
		emit(s.Name, key, val, s.Assertions[key], modeNone)
	}
}
