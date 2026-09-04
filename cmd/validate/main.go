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
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/mapdb/mapdb-golang/arraylist"
	"github.com/mapdb/mapdb-golang/bag"
	"github.com/mapdb/mapdb-golang/bloom"
	"github.com/mapdb/mapdb-golang/boundedlru"
	"github.com/mapdb/mapdb-golang/countmin"
	"github.com/mapdb/mapdb-golang/fenwick"
	"github.com/mapdb/mapdb-golang/hash"
	"github.com/mapdb/mapdb-golang/hashmap"
	"github.com/mapdb/mapdb-golang/hashset"
	"github.com/mapdb/mapdb-golang/hyperloglog"
	"github.com/mapdb/mapdb-golang/immutablesorted"
	"github.com/mapdb/mapdb-golang/multimap"
	"github.com/mapdb/mapdb-golang/pump"
	"github.com/mapdb/mapdb-golang/rangev"
	"github.com/mapdb/mapdb-golang/roaring"
	"github.com/mapdb/mapdb-golang/treemap"
	"github.com/mapdb/mapdb-golang/treeset"
)

type scenario struct {
	Name         string                     `json:"name"`
	Collection   string                     `json:"collection"`
	Construction string                     `json:"construction,omitempty"`
	Operations   []map[string]any           `json:"operations"`
	Assertions   map[string]json.RawMessage `json:"assertions"`
	Other        *otherSpec                 `json:"other,omitempty"`
	// Query is the optional single top-level range (NavigableMap/Set) the
	// range_* assertions refer to; same range-builder shape as the `range`
	// field on a remove_range op.
	Query map[string]any `json:"query,omitempty"`
	// MaxSize / TTL are the BoundedLruMap scenario config: max_size is the
	// required capacity; ttl is null/absent for a pure max-size map, otherwise a
	// u64 logical tick encoded as a decimal string (or plain number).
	MaxSize *json.Number    `json:"max_size,omitempty"`
	TTL     json.RawMessage `json:"ttl,omitempty"`
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
				// modeNone. Fenwick `tree` is an i64 decimal-string array in
				// JSON that the runner emits as a bare-decimal array; unquote
				// those strings. Every other array (bools, nested LRU
				// snapshot/eviction logs, i64 quoted keys) goes through the
				// recursive helper so a third nesting level or a bool cannot
				// panic on a float64 assertion.
				if key == "tree" {
					switch ev := e.(type) {
					case string:
						parts[i] = ev
					case float64:
						parts[i] = strconv.FormatInt(int64(ev), 10)
					default:
						parts[i] = renderModeNoneElement(e)
					}
				} else {
					parts[i] = renderModeNoneElement(e)
				}
			}
		}
		return "[" + strings.Join(parts, ",") + "]"
	}
	return string(raw)
}

// renderModeNoneElement renders one element of a modeNone (i32) assertion
// array into the runner's canonical string form. Scalars: numbers as integers,
// i64 decimal-string keys quoted, null bare. NESTED arrays (the BoundedLruMap
// eviction_log / snapshot_*_log assertions, shaped [[..],[..]] or [[[..]]]) are
// rendered recursively as compact `[a,b,...]` so they match the runner's
// computed nested-array strings.
func renderModeNoneElement(e any) string {
	switch ev := e.(type) {
	case nil:
		return "null"
	case bool:
		return strconv.FormatBool(ev)
	case string:
		return "\"" + ev + "\""
	case float64:
		return strconv.FormatInt(int64(ev), 10)
	case []any:
		parts := make([]string, len(ev))
		for i, inner := range ev {
			parts[i] = renderModeNoneElement(inner)
		}
		return "[" + strings.Join(parts, ",") + "]"
	}
	return fmt.Sprintf("%v", e)
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
	case "RangeSet<i32>":
		runRangeSet(s)
	case "RangeMap<i32, i32>":
		runRangeMap(s)
	case "ImmutableSortedMap<i32, i32>":
		runImmutableSortedMap(s)
	case "ImmutableSortedSet<i32>":
		runImmutableSortedSet(s)
	case "HashPipeline":
		runHashPipeline(s)
	case "Bloom":
		runBloom(s)
	case "HyperLogLog":
		runHyperLogLog(s)
	case "CountMin":
		runCountMin(s)
	case "SpaceSaving":
		runSpaceSaving(s)
	case "FenwickTree":
		runFenwick(s)
	case "RoaringU32":
		runRoaring(s)
	case "BoundedLruMap<i32, i32>":
		runBoundedLru(s)
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

// ---- Bloom (spec/features/bloom.md) --------------------------------------
//
// A new collection kind riding the hash pipeline's Positions() byte path. The
// scenario builds the filter via EXACTLY ONE with_params op (explicit (m,k);
// never optimal -- the float trap is quarantined to native tests) followed by
// add ops. A union scenario reads a second filter from the top-level "other"
// block (same with_params + add shape); the union_* assertions describe
// self.union(other). Unknown ops/keys/kinds SKIP (forward-compat). Out-of-range
// or invalid params (m=0, m/k/v outside u32/i32, union param-mismatch) SKIP for
// scenarios rather than panic.

// buildBloom replays a with_params + add op list into a *bloom.Bloom. It returns
// nil (with a skip reason) when the op list is malformed or out of the validated
// range -- the runner then SKIPs the scenario rather than panicking.
func buildBloom(ops []map[string]any) (*bloom.Bloom, string) {
	withParamsCount := 0
	for _, op := range ops {
		if op["op"] == "with_params" {
			withParamsCount++
		}
	}
	// Exactly one with_params builds the filter (the from_sorted/HashPipeline
	// rule); zero or multiple => malformed => SKIP.
	if withParamsCount != 1 {
		return nil, fmt.Sprintf("need exactly one with_params op, got %d", withParamsCount)
	}
	var b *bloom.Bloom
	for _, op := range ops {
		switch op["op"] {
		case "with_params":
			m, ok1 := bloomU32(op["m"])
			k, ok2 := bloomU32(op["k"])
			if !ok1 || !ok2 || m == 0 {
				return nil, "with_params m/k out of range (m must be >= 1, m,k in u32)"
			}
			b = bloom.NewBloomWithParams(m, k)
		case "add":
			if b == nil {
				return nil, "add before with_params"
			}
			v, ok := bloomI32(op["value"])
			if !ok {
				return nil, "add value out of i32 range"
			}
			b.Add(v)
		default:
			// Forward-compat: an unknown op makes the scenario un-runnable here.
			return nil, fmt.Sprintf("unknown bloom op (forward-compat): %v", op["op"])
		}
	}
	return b, ""
}

// bloomU32 parses an m/k operand to a uint32, reporting !ok when it is not an
// integer in [0, 2^32-1] (so the runner SKIPs rather than panicking).
func bloomU32(v any) (uint32, bool) {
	var i int64
	switch n := v.(type) {
	case float64:
		if n != math.Trunc(n) {
			return 0, false
		}
		i = int64(n)
	case json.Number:
		x, err := n.Int64()
		if err != nil {
			return 0, false
		}
		i = x
	default:
		return 0, false
	}
	if i < 0 || i > math.MaxUint32 {
		return 0, false
	}
	return uint32(i), true
}

// bloomI32 parses an element operand to an int32, reporting !ok when it is not an
// integer in the signed i32 range.
func bloomI32(v any) (int32, bool) {
	var i int64
	switch n := v.(type) {
	case float64:
		if n != math.Trunc(n) {
			return 0, false
		}
		i = int64(n)
	case json.Number:
		x, err := n.Int64()
		if err != nil {
			return 0, false
		}
		i = x
	default:
		return 0, false
	}
	if i < math.MinInt32 || i > math.MaxInt32 {
		return 0, false
	}
	return int32(i), true
}

func bloomHex(bytes []byte) string {
	var sb strings.Builder
	sb.WriteString("0x")
	for _, b := range bytes {
		fmt.Fprintf(&sb, "%02x", b)
	}
	return sb.String()
}

func bloomSortedSetBits(b *bloom.Bloom) string {
	// SetBits is already ascending; render as a JSON int array.
	parts := []string{}
	for _, x := range b.SetBits() {
		parts = append(parts, strconv.FormatUint(uint64(x), 10))
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func runBloom(s scenario) {
	b, skip := buildBloom(s.Operations)
	if b == nil {
		fmt.Fprintf(os.Stderr, "skip: malformed bloom scenario (%s): %s\n", skip, s.Name)
		return
	}
	// Optional union partner from the top-level "other" block. Build it eagerly;
	// a malformed/param-mismatched partner makes any union_* assertion un-runnable
	// (the union_* eval then SKIPs that key).
	var union *bloom.Bloom
	if s.Other != nil {
		other, oskip := buildBloom(s.Other.Operations)
		if other == nil {
			fmt.Fprintf(os.Stderr, "skip: malformed bloom other (%s): %s\n", oskip, s.Name)
		} else if other.MBits() != b.MBits() || other.K() != b.K() {
			// Param-mismatch union would panic in production; for a scenario we
			// leave union nil so union_* keys SKIP instead of crashing.
			fmt.Fprintf(os.Stderr, "skip: bloom union param mismatch in %s\n", s.Name)
		} else {
			union = b.Union(other)
		}
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		emit(s.Name, key, evalBloomAssertion(key, b, union), s.Assertions[key], modeNone)
	}
}

func evalBloomAssertion(key string, b, union *bloom.Bloom) string {
	switch key {
	case "m_bits":
		return strconv.FormatUint(uint64(b.MBits()), 10)
	case "k":
		return strconv.FormatUint(uint64(b.K()), 10)
	case "bit_count":
		return strconv.FormatUint(uint64(b.BitCount()), 10)
	case "is_empty":
		return strconv.FormatBool(b.IsEmpty())
	case "set_bits":
		return bloomSortedSetBits(b)
	case "bytes":
		return bloomHex(b.ToBytes())
	}
	// union_* keys require the "other" partner; SKIP when it is absent.
	switch key {
	case "union_bit_count":
		if union == nil {
			return unknown(key)
		}
		return strconv.FormatUint(uint64(union.BitCount()), 10)
	case "union_set_bits":
		if union == nil {
			return unknown(key)
		}
		return bloomSortedSetBits(union)
	case "union_bytes":
		if union == nil {
			return unknown(key)
		}
		return bloomHex(union.ToBytes())
	}
	if rest, ok := strings.CutPrefix(key, "union_contains_"); ok {
		if union == nil {
			return unknown(key)
		}
		v, ok := bloomI32SuffixToInt(rest)
		if !ok {
			return unknown(key)
		}
		return strconv.FormatBool(union.MightContain(v))
	}
	if rest, ok := strings.CutPrefix(key, "contains_"); ok {
		v, ok := bloomI32SuffixToInt(rest)
		if !ok {
			return unknown(key)
		}
		return strconv.FormatBool(b.MightContain(v))
	}
	return unknown(key)
}

// bloomI32SuffixToInt parses a signed-i32 decimal suffix (matches
// ^(-?[0-9]+)$ in the validated range); reports !ok for a non-i32 suffix so the
// key SKIPs cleanly.
func bloomI32SuffixToInt(s string) (int32, bool) {
	digits := strings.TrimPrefix(s, "-")
	if digits == "" || !isAllDigits(digits) {
		return 0, false
	}
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, false
	}
	return int32(n), true
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

// ---- FenwickTree (spec/features/fenwick.md) -------------------------------
//
// A fixed-size i32-element / i64-accumulator Binary Indexed Tree. Construction
// is EXACTLY ONE op (with_size or from_values) first, then any number of
// update/set point ops (all indices in-range; out-of-range traps are
// native-test-only). Sum-returning assertions (total, get_<i>, prefix_sum_<i>,
// range_sum_<lo>_<hi>, and each tree element) are i64 and wire-encoded as
// DECIMAL STRINGS (parsed straight to i64, never via f64); the runner accepts a
// bare JSON number too (renderExpected under modeNone). tree is the canonical
// 1-based BIT array in 1-based index order -- an explicit-order key, NOT sorted.
// Unknown ops / kinds / assertion keys SKIP (forward-compat).
//
// i64 wire encoding: the runner emits every Fenwick i64 result via
// strconv.FormatInt base 10 (a plain decimal). The JSON assertions carry the
// authoritative i64 values as DECIMAL STRINGS (or bare numbers when small);
// emit/renderExpected compare them as strings under modeNone, so a
// decimal-string expected ("total": "8589934588") and the runner's decimal
// output match without any f64 round-trip. Element operands (delta/value) stay
// plain JSON numbers (they are i32).
func runFenwick(s scenario) {
	// Authoring rule: the FIRST op MUST be exactly one construction op
	// (with_size OR from_values); a missing/late/duplicate construction op is a
	// malformed scenario => SKIP (forward-compat), like the hash-pipeline
	// single-op rule.
	if len(s.Operations) == 0 {
		fmt.Fprintln(os.Stderr, "skip: fenwick scenario must begin with a construction op (forward-compat)")
		return
	}
	first, _ := s.Operations[0]["op"].(string)
	var tree *fenwick.FenwickTree
	switch first {
	case "with_size":
		n := asInt(s.Operations[0]["n"])
		if n < 0 {
			fmt.Fprintf(os.Stderr, "skip: fenwick with_size negative n (malformed): %d\n", n)
			return
		}
		tree = fenwick.NewFenwickTreeWithSize(n)
	case "from_values":
		raw, ok := s.Operations[0]["values"].([]any)
		if !ok {
			fmt.Fprintln(os.Stderr, "skip: fenwick from_values needs values array (malformed)")
			return
		}
		vals := make([]int32, len(raw))
		for i, v := range raw {
			vals[i] = asInt32(v)
		}
		tree = fenwick.NewFenwickTreeFromValues(vals)
	default:
		fmt.Fprintf(os.Stderr, "skip: fenwick first op must be with_size/from_values (forward-compat): %s\n", first)
		return
	}
	// Any subsequent construction op is malformed => SKIP.
	for _, op := range s.Operations[1:] {
		switch op["op"] {
		case "update":
			tree.Update(asInt(op["index"]), asInt32(op["delta"]))
		case "set":
			tree.Set(asInt(op["index"]), asInt32(op["value"]))
		case "with_size", "from_values":
			fmt.Fprintln(os.Stderr, "skip: fenwick has a non-first construction op (malformed)")
			return
		default:
			fmt.Fprintf(os.Stderr, "skip: unknown fenwick op (forward-compat): %v\n", op["op"])
			return
		}
	}

	for _, key := range sortedAssertionKeys(s.Assertions) {
		emit(s.Name, key, evalFenwickAssertion(key, tree), s.Assertions[key], modeNone)
	}
}

func evalFenwickAssertion(key string, tree *fenwick.FenwickTree) string {
	switch {
	case key == "size":
		return strconv.Itoa(tree.Len())
	case key == "is_empty":
		return strconv.FormatBool(tree.IsEmpty())
	case key == "total":
		return strconv.FormatInt(tree.Total(), 10)
	case key == "tree":
		// Canonical 1-based BIT array, in 1-based index order (NOT sorted).
		ct := tree.CanonicalTree()
		parts := make([]string, len(ct))
		for i, v := range ct {
			parts[i] = strconv.FormatInt(v, 10)
		}
		return "[" + strings.Join(parts, ",") + "]"
	case strings.HasPrefix(key, "get_"):
		i, err := strconv.Atoi(key[len("get_"):])
		if err != nil {
			return "UNKNOWN_ASSERTION:" + key
		}
		return strconv.FormatInt(tree.Get(i), 10)
	case strings.HasPrefix(key, "prefix_sum_"):
		i, err := strconv.Atoi(key[len("prefix_sum_"):])
		if err != nil {
			return "UNKNOWN_ASSERTION:" + key
		}
		return strconv.FormatInt(tree.PrefixSum(i), 10)
	case strings.HasPrefix(key, "range_sum_"):
		// ^range_sum_([0-9]+)_([0-9]+)$
		rest := key[len("range_sum_"):]
		us := strings.IndexByte(rest, '_')
		if us < 0 {
			return "UNKNOWN_ASSERTION:" + key
		}
		lo, errLo := strconv.Atoi(rest[:us])
		hi, errHi := strconv.Atoi(rest[us+1:])
		if errLo != nil || errHi != nil {
			return "UNKNOWN_ASSERTION:" + key
		}
		return strconv.FormatInt(tree.RangeSum(lo, hi), 10)
	default:
		return "UNKNOWN_ASSERTION:" + key
	}
}

// ---- RoaringU32 (spec/features/roaring-u32.md) ---------------------------
//
// A sparse, compressed u32 set. The i32 scenario values are bit-reinterpreted
// to u32, and ordering is unsigned-u32 ascending throughout. Unknown ops/keys
// skip for forward compatibility.
func runRoaring(s scenario) {
	set, ok := buildRoaring(s.Operations)
	if !ok {
		return
	}
	var other *roaring.RoaringU32
	if s.Other != nil {
		other, ok = buildRoaring(s.Other.Operations)
		if !ok {
			return
		}
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		emit(s.Name, key, evalRoaringAssertion(key, set, other), s.Assertions[key], modeNone)
	}
}

func buildRoaring(ops []map[string]any) (*roaring.RoaringU32, bool) {
	for _, op := range ops {
		if op["op"] == "deserialize" {
			if len(ops) != 1 {
				fmt.Fprintln(os.Stderr, "skip: roaring deserialize op must be the only op (forward-compat)")
				return nil, false
			}
			hexStr, ok := op["bytes"].(string)
			if !ok {
				fmt.Fprintln(os.Stderr, "skip: roaring deserialize op missing string bytes (forward-compat)")
				return nil, false
			}
			raw, err := roaringParseHex(hexStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "skip: roaring deserialize bad hex: %v\n", err)
				return nil, false
			}
			set, err := roaring.Deserialize(raw)
			if err != nil {
				fmt.Fprintf(os.Stderr, "skip: roaring deserialize rejected bytes: %v\n", err)
				return nil, false
			}
			return set, true
		}
	}
	set := roaring.NewRoaringU32()
	for _, op := range ops {
		switch op["op"] {
		case "add":
			set.Add(uint32(asInt32(op["value"])))
		case "remove":
			set.Remove(uint32(asInt32(op["value"])))
		case "clear":
			set.Clear()
		case "add_range":
			from, to := uint32(asInt32(op["from"])), uint32(asInt32(op["to"]))
			if from > to {
				fmt.Fprintf(os.Stderr, "skip: roaring add_range reversed (%d > %d) (forward-compat)\n", from, to)
				return nil, false
			}
			for v := from; ; v++ {
				set.Add(v)
				if v == to {
					break
				}
			}
		case "remove_range":
			from, to := uint32(asInt32(op["from"])), uint32(asInt32(op["to"]))
			if from > to {
				fmt.Fprintf(os.Stderr, "skip: roaring remove_range reversed (%d > %d) (forward-compat)\n", from, to)
				return nil, false
			}
			for v := from; ; v++ {
				set.Remove(v)
				if v == to {
					break
				}
			}
		default:
			fmt.Fprintf(os.Stderr, "skip: unknown roaring op (forward-compat): %v\n", op["op"])
			return nil, false
		}
	}
	return set, true
}

func roaringParseHex(hx string) ([]byte, error) {
	if len(hx) < 2 || (hx[:2] != "0x" && hx[:2] != "0X") {
		return nil, fmt.Errorf("missing 0x prefix")
	}
	body := hx[2:]
	if len(body)%2 != 0 {
		return nil, fmt.Errorf("odd-length hex string")
	}
	out := make([]byte, len(body)/2)
	for i := 0; i < len(out); i++ {
		hi, ok1 := hexNibble(body[2*i])
		lo, ok2 := hexNibble(body[2*i+1])
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("invalid hex digit")
		}
		out[i] = hi<<4 | lo
	}
	return out, nil
}

func hexNibble(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

func toHex(b []byte) string {
	const hexDigits = "0123456789abcdef"
	buf := make([]byte, 2+2*len(b))
	buf[0], buf[1] = '0', 'x'
	for i, x := range b {
		buf[2+2*i] = hexDigits[x>>4]
		buf[2+2*i+1] = hexDigits[x&0x0f]
	}
	return string(buf)
}

func roaringSortedArray(set *roaring.RoaringU32) []int32 {
	vals := set.ToSortedSlice()
	out := make([]int32, len(vals))
	for i, v := range vals {
		out[i] = int32(v)
	}
	return out
}

func evalRoaringAssertion(key string, set, other *roaring.RoaringU32) string {
	switch key {
	case "cardinality":
		return strconv.FormatUint(set.Cardinality(), 10)
	case "is_empty":
		return strconv.FormatBool(set.IsEmpty())
	case "chunk_count":
		return strconv.Itoa(set.ChunkCount())
	case "serialized_len":
		return strconv.Itoa(len(set.Serialize()))
	case "serialized_hex":
		return toHex(set.Serialize())
	case "container_types":
		types := set.ContainerTypes()
		quoted := make([]string, len(types))
		for i, t := range types {
			quoted[i] = `"` + t + `"`
		}
		return "[" + strings.Join(quoted, ",") + "]"
	case "to_sorted_array":
		return formatArray(roaringSortedArray(set))
	case "min":
		if v, ok := set.Min(); ok {
			return strconv.FormatInt(int64(int32(v)), 10)
		}
		return "null"
	case "max":
		if v, ok := set.Max(); ok {
			return strconv.FormatInt(int64(int32(v)), 10)
		}
		return "null"
	}
	if other != nil {
		switch key {
		case "union_serialized_hex":
			return toHex(set.Or(other).Serialize())
		case "union_cardinality":
			return strconv.FormatUint(set.Or(other).Cardinality(), 10)
		case "intersect_serialized_hex":
			return toHex(set.And(other).Serialize())
		case "intersect_cardinality":
			return strconv.FormatUint(set.And(other).Cardinality(), 10)
		case "and_not_serialized_hex":
			return toHex(set.AndNot(other).Serialize())
		case "and_not_cardinality":
			return strconv.FormatUint(set.AndNot(other).Cardinality(), 10)
		case "xor_serialized_hex":
			return toHex(set.Xor(other).Serialize())
		case "xor_cardinality":
			return strconv.FormatUint(set.Xor(other).Cardinality(), 10)
		}
	}
	if rest, ok := strings.CutPrefix(key, "contains_"); ok {
		v, err := strconv.ParseInt(rest, 10, 32)
		if err != nil {
			return unknown(key)
		}
		return strconv.FormatBool(set.Contains(uint32(int32(v))))
	}
	return unknown(key)
}

// ---- HashMap<i32, i32> ---------------------------------------------------

func runHashMap(s scenario) {
	var m *hashmap.Int32Int32
	if s.Construction == "bulkLoadExact" {
		keys, vals := int32Pairs(s.Operations)
		var err error
		m, err = hashmap.Int32Int32BulkLoadExact(keys, vals, len(keys), pump.ErrorOnDuplicate)
		if err != nil {
			fatalf("bulkLoadExact failed: %v", err)
		}
	} else {
		m = hashmap.NewInt32Int32()
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
	}
	for _, key := range sortedAssertionKeys(s.Assertions) {
		emit(s.Name, key, evalMapAssertion(key, m), s.Assertions[key], modeNone)
	}
}

func int32Pairs(ops []map[string]any) ([]int32, []int32) {
	keys := make([]int32, 0, len(ops))
	vals := make([]int32, 0, len(ops))
	for _, op := range ops {
		keys = append(keys, asInt32(op["key"]))
		vals = append(vals, asInt32(op["value"]))
	}
	return keys, vals
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

func runI64ListMultimap(s scenario) {
	if s.Construction == "fromSortedKeyValues" {
		keys, vals := i64Pairs(s.Operations)
		m, err := multimap.NewInt64Int32ListFromSortedKeyValues(keys, vals)
		if err != nil {
			fatalf("fromSortedKeyValues failed: %v", err)
		}
		runI64MultimapAssertions(s, m)
		return
	}
	runI64Multimap(s, multimap.NewInt64Int32List())
}

func runI64SetMultimap(s scenario) {
	if s.Construction == "fromSortedKeyValues" {
		keys, vals := i64Pairs(s.Operations)
		m, err := multimap.NewInt64Int32SetFromSortedKeyValues(keys, vals)
		if err != nil {
			fatalf("fromSortedKeyValues failed: %v", err)
		}
		runI64MultimapAssertions(s, m)
		return
	}
	runI64Multimap(s, multimap.NewInt64Int32Set())
}

func i64Pairs(ops []map[string]any) ([]int64, []int32) {
	keys := make([]int64, 0, len(ops))
	vals := make([]int32, 0, len(ops))
	for _, op := range ops {
		keys = append(keys, parseI64Operand(op["key"]))
		vals = append(vals, asInt32(op["value"]))
	}
	return keys, vals
}

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
	runI64MultimapAssertions(s, m)
}

func runI64MultimapAssertions(s scenario, m i64Multimap) {
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
			// Production insert-at-index (EC MutableIntList.addAtIndex parity).
			l.AddAtIndex(asInt(op["index"]), asInt32(op["value"]))
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

// listToSorted renders a production list result in ascending order. Ordering is
// done by the production Sort() on a COPY so the source list is never mutated.
func listToSorted(l *arraylist.Int32) []int32 {
	sorted := arraylist.NewInt32()
	sorted.AddAll(l.ToSlice()...)
	sorted.Sort()
	return sorted.ToSlice()
}

func evalListAssertion(key string, l *arraylist.Int32) string {
	// Every assertion below is obtained from the production method the
	// assertion names (Len/Contains/Get/Select/Reject/Detect/Count/
	// AnySatisfy/AllSatisfy/NoneSatisfy/InjectInto/Sum/Min/Max/Sort). The one
	// documented exception is inject_into_product -- see its case.
	switch key {
	case "size":
		return strconv.Itoa(l.Len())
	case "is_empty":
		return strconv.FormatBool(l.Len() == 0)
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
		return formatArray(listToSorted(l))
	case "inject_into_sum":
		// injectInto with a + reduction accumulates in the i32 seed type and
		// wraps two's-complement at i32 -- via the production InjectInto.
		return strconv.FormatInt(int64(l.InjectInto(0, addInt32Wrapping)), 10)
	case "inject_into_product":
		// DELIBERATE runner-local reduction: this assertion demands a WIDENING
		// i64 product, and the production InjectInto is typed
		// InjectInto(int32, func(int32, int32) int32) -- its accumulator is
		// i32 and would wrap. There is no i64-accumulator reduction on
		// arraylist.Int32 (Sum() is the only widening one, and it is fixed to
		// addition), so the widening product has no production method to call.
		var acc int64 = 1
		for _, v := range l.ToSlice() {
			acc *= int64(v)
		}
		return strconv.FormatInt(acc, 10)
	case "any_satisfy_even":
		return strconv.FormatBool(l.AnySatisfy(isEvenInt32))
	case "all_satisfy_even":
		return strconv.FormatBool(l.AllSatisfy(isEvenInt32))
	case "none_satisfy_odd":
		return strconv.FormatBool(l.NoneSatisfy(isOddInt32))
	case "count_even":
		return strconv.Itoa(l.Count(isEvenInt32))
	case "count_odd":
		return strconv.Itoa(l.Count(isOddInt32))
	}
	if rest, ok := strings.CutPrefix(key, "get_at_"); ok {
		idx, _ := strconv.Atoi(rest)
		if idx < 0 || idx >= l.Len() {
			return "null"
		}
		return strconv.FormatInt(int64(l.Get(idx)), 10)
	}
	if rest, ok := strings.CutPrefix(key, "contains_"); ok {
		v, _ := strconv.ParseInt(rest, 10, 32)
		return strconv.FormatBool(l.Contains(int32(v)))
	}
	if rest, ok := strings.CutPrefix(key, "select_gt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		return formatArray(listToSorted(l.Select(greaterThanInt32(int32(t)))))
	}
	if rest, ok := strings.CutPrefix(key, "reject_gt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		return formatArray(listToSorted(l.Reject(greaterThanInt32(int32(t)))))
	}
	if rest, ok := strings.CutPrefix(key, "detect_gt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		if x, ok := l.Detect(greaterThanInt32(int32(t))); ok {
			return strconv.FormatInt(int64(x), 10)
		}
		return "null"
	}
	if rest, ok := strings.CutPrefix(key, "count_gt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		return strconv.Itoa(l.Count(greaterThanInt32(int32(t))))
	}
	if rest, ok := strings.CutPrefix(key, "count_lt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		return strconv.Itoa(l.Count(lessThanInt32(int32(t))))
	}
	if rest, ok := strings.CutPrefix(key, "any_satisfy_gt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		return strconv.FormatBool(l.AnySatisfy(greaterThanInt32(int32(t))))
	}
	if rest, ok := strings.CutPrefix(key, "all_satisfy_gt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		return strconv.FormatBool(l.AllSatisfy(greaterThanInt32(int32(t))))
	}
	if rest, ok := strings.CutPrefix(key, "none_satisfy_gt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		return strconv.FormatBool(l.NoneSatisfy(greaterThanInt32(int32(t))))
	}
	if rest, ok := strings.CutPrefix(key, "none_satisfy_lt_"); ok {
		t, _ := strconv.ParseInt(rest, 10, 32)
		return strconv.FormatBool(l.NoneSatisfy(lessThanInt32(int32(t))))
	}
	return unknown(key)
}

// Predicates handed to the production Select/Reject/Detect/Count/*Satisfy
// methods. They only decide membership; the traversal is the collection's.
func isEvenInt32(v int32) bool { return v%2 == 0 }
func isOddInt32(v int32) bool  { return v%2 != 0 }

func greaterThanInt32(t int32) func(int32) bool { return func(v int32) bool { return v > t } }
func lessThanInt32(t int32) func(int32) bool    { return func(v int32) bool { return v < t } }

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
	// ToSlice() is the production materializer; a hash set has no order, so
	// the harness sorts the RESULT for rendering only.
	out := set.ToSlice()
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
		// Every set-algebra assertion is computed by the production method of
		// the same name (hashset.Int32.Union / Intersect / Difference /
		// SymmetricDifference); the harness only reads Len() off the result or
		// sorts it for rendering.
		switch key {
		case "union_sorted":
			return formatArray(setToSorted(set.Union(other)))
		case "intersect_sorted":
			return formatArray(setToSorted(set.Intersect(other)))
		case "difference_sorted":
			return formatArray(setToSorted(set.Difference(other)))
		case "symmetric_difference_sorted":
			return formatArray(setToSorted(set.SymmetricDifference(other)))
		case "union_size":
			return strconv.Itoa(set.Union(other).Len())
		case "intersect_size":
			return strconv.Itoa(set.Intersect(other).Len())
		case "difference_size":
			return strconv.Itoa(set.Difference(other).Len())
		case "symmetric_difference_size":
			return strconv.Itoa(set.SymmetricDifference(other).Len())
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
		// AllDistinct() is the production distinct iterator; a hash bag has no
		// order, so the harness sorts the RESULT for rendering only.
		v := slices.Collect(b.AllDistinct())
		sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
		return formatArray(v)
	case "to_sorted_array":
		// ToSlice() is the production materializer and already repeats each
		// element per its occurrence count; sorted for rendering only.
		flat := b.ToSlice()
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
	var m *treemap.Int32Int32
	var log navLog
	if s.Construction == "fromSorted" {
		keys, vals := int32Pairs(s.Operations)
		var err error
		m, err = treemap.NewInt32Int32FromSorted(keys, vals, pump.ErrorOnDuplicate)
		if err != nil {
			fatalf("fromSorted failed: %v", err)
		}
	} else {
		m = treemap.NewInt32Int32()
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
				fatalf("unknown treemap op: %v", op["op"])
			}
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
				// Straight from the production Values() iterator, in the tree's
				// ascending-KEY order -- NOT re-sorted by value. The README's
				// one-line table says "all values, sorted ascending", but for
				// TreeMap the FIXTURES define the contract and they mean
				// key-order: 17-bulk-load/treemap_i32_from_sorted expects
				// [100,0,50,200,210] for keys [-10,0,5,20,21]. Sorting by value
				// here turns that scenario red. Rust's TreeMap runner
				// (mapdb-rust src/bin/validate.rs, `map.values()` with no sort)
				// agrees. Do not "fix" this to a value sort.
				return formatArray(slices.Collect(m.Values()))
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
				// KeysToSlice() is the production materializer; a hash map has
				// no order, so the harness sorts the RESULT (IEEE total order)
				// for rendering only.
				keys := m.KeysToSlice()
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
				// ToSlice() is the production materializer; a hash set has no
				// order, so the harness sorts the RESULT (IEEE total order) for
				// rendering only.
				vals := set.ToSlice()
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
				// ToSlice() is the production materializer and already returns
				// the tree's in-order (total-order) sequence -- no runner-side
				// sort, the harness only formats.
				elems := set.ToSlice()
				parts := make([]string, len(elems))
				for i, v := range elems {
					parts[i] = "\"" + formatF32(v) + "\""
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
	for _, key := range sortedAssertionKeys(s.Assertions) {
		val := func() string {
			switch key {
			case "size":
				return strconv.Itoa(l.Len())
			case "is_empty":
				return strconv.FormatBool(l.Len() == 0)
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
				sorted.AddAll(l.ToSlice()...)
				sorted.Sort()
				parts := make([]string, 0, sorted.Len())
				for _, x := range sorted.ToSlice() {
					parts = append(parts, formatF32(x))
				}
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

// ---- RangeSet<i32> / RangeMap<i32, i32> -----------------------------------
//
// The auto-coalescing RangeSet / piecewise RangeMap (spec/features/
// range-set-map.md). Routed through the PRODUCTION rangev.Int32RangeSet /
// Int32Int32RangeMap — every assertion is proved against the real cut-algebra
// coalescing/split/complement code, not re-derived here.
//
// A RangeSet/RangeMap is a STATEFUL structure built by a sequence of mutating
// ops, each naming a `range` via the shared 10-range builder object:
//
//	RangeSet: {"op":"add","range":{...}} / {"op":"remove_range","range":{...}}
//	          / {"op":"clear"}
//	RangeMap: {"op":"put","range":{...},"value":N}
//	          / {"op":"remove_range","range":{...}} / {"op":"clear"}
//
// An optional top-level `query` (same builder shape) supplies the range for
// encloses_query / intersects_query / sub_range_set_ranges /
// sub_range_map_entries. Unknown ops/keys/kinds SKIP (forward-compat).
//
// The as_ranges / complement_ranges / sub_range_set_ranges / as_map_of_ranges /
// sub_range_map_entries arrays are EXPLICIT-ORDER (ascending by lower cut), each
// element a fixed-shape range/entry object pinning the exact cut. Those object
// assertions go through emitJSON (compact-JSON comparison) since the standard
// renderExpected path does not canonicalise nested objects.

// optI32JSON renders an (int32, ok) endpoint as the i32 decimal or null.
func optI32JSON(v int32, ok bool) string {
	if !ok {
		return "null"
	}
	return strconv.FormatInt(int64(v), 10)
}

// boundTypeJSON renders a bound type as a quoted "open"/"closed" or null.
func boundTypeJSON(bt rangev.BoundType, ok bool) string {
	if !ok {
		return "null"
	}
	return "\"" + bt.String() + "\""
}

// rangeObjStr serialises an Int32Range as the fixed-shape assertion object
// {"lower":..,"lower_type":..,"upper":..,"upper_type":..} — endpoints are the
// i32 value or null when unbounded; *_type is "open"/"closed"/null. The key
// order matches the scenario JSON so emitJSON's compacted comparison agrees.
func rangeObjStr(r rangev.Int32Range) string {
	lv, lok := r.LowerEndpoint()
	uv, uok := r.UpperEndpoint()
	lt, ltok := r.LowerBoundType()
	ut, utok := r.UpperBoundType()
	return fmt.Sprintf(
		"{\"lower\":%s,\"lower_type\":%s,\"upper\":%s,\"upper_type\":%s}",
		optI32JSON(lv, lok), boundTypeJSON(lt, ltok),
		optI32JSON(uv, uok), boundTypeJSON(ut, utok),
	)
}

// entryObjStr serialises a (range, value) RangeMap entry: the range object plus
// a trailing "value":<i32>.
func entryObjStr(e rangev.Int32Int32Entry) string {
	r := e.Range
	lv, lok := r.LowerEndpoint()
	uv, uok := r.UpperEndpoint()
	lt, ltok := r.LowerBoundType()
	ut, utok := r.UpperBoundType()
	return fmt.Sprintf(
		"{\"lower\":%s,\"lower_type\":%s,\"upper\":%s,\"upper_type\":%s,\"value\":%d}",
		optI32JSON(lv, lok), boundTypeJSON(lt, ltok),
		optI32JSON(uv, uok), boundTypeJSON(ut, utok), e.Value,
	)
}

func rangeArrayStr(ranges []rangev.Int32Range) string {
	parts := make([]string, len(ranges))
	for i, r := range ranges {
		parts[i] = rangeObjStr(r)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func entryArrayStr(entries []rangev.Int32Int32Entry) string {
	parts := make([]string, len(entries))
	for i, e := range entries {
		parts[i] = entryObjStr(e)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// emitJSON prints a computed range-object assertion and compares it against the
// COMPACTED expected JSON. The standard emit/renderExpected path is bypassed
// because the expected value is a nested object (or array of objects); compact
// canonicalisation (which preserves source key order, matching rangeObjStr's
// fixed order) is the byte-for-byte oracle the Rust runner achieves via
// serde_json's to_string(). UNKNOWN_ASSERTION:* is skipped silently.
func emitJSON(name, key, computed string, expected json.RawMessage) {
	if strings.HasPrefix(computed, "UNKNOWN_ASSERTION:") {
		return
	}
	fmt.Printf("%s: %s\n", key, computed)
	var buf bytes.Buffer
	want := string(expected)
	if err := json.Compact(&buf, expected); err == nil {
		want = buf.String()
	}
	if computed != want {
		fmt.Printf("FAIL %s %s: expected=%s got=%s\n", name, key, want, computed)
		anyFail = true
	}
}

// signedI32Suffix parses a signed base-10 i32 suffix (leading '-' allowed,
// rejects '+') from a <prefix><N> key — the contains_<v> / get_<v> /
// range_containing_<v> / get_entry_<v> convention. Returns (n, true) on a match.
func signedI32Suffix(key, prefix string) (int32, bool) {
	rest, ok := strings.CutPrefix(key, prefix)
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

func runRangeSet(s scenario) {
	set := rangev.NewInt32RangeSet()
	for _, op := range s.Operations {
		switch op["op"] {
		case "add":
			set.Add(buildRangeObj(op["range"].(map[string]any)))
		case "remove_range":
			set.Remove(buildRangeObj(op["range"].(map[string]any)))
		case "clear":
			set.Clear()
		default:
			// Forward-compat: unknown op kinds skip (do not crash the runner).
		}
	}
	var query rangev.Int32Range
	hasQuery := s.Query != nil
	if hasQuery {
		query = buildRangeObj(s.Query)
	}
	span, hasSpan := set.Span()
	for _, key := range sortedAssertionKeys(s.Assertions) {
		// Object-shaped assertions go through emitJSON; scalar ones through emit.
		switch key {
		case "as_ranges":
			emitJSON(s.Name, key, rangeArrayStr(set.AsRanges()), s.Assertions[key])
			continue
		case "complement_ranges":
			emitJSON(s.Name, key, rangeArrayStr(set.Complement().AsRanges()), s.Assertions[key])
			continue
		case "sub_range_set_ranges":
			if hasQuery {
				emitJSON(s.Name, key, rangeArrayStr(set.SubRangeSet(query).AsRanges()), s.Assertions[key])
			}
			continue
		}
		if n, ok := signedI32Suffix(key, "range_containing_"); ok {
			if r, found := set.RangeContaining(n); found {
				emitJSON(s.Name, key, rangeObjStr(r), s.Assertions[key])
			} else {
				emitJSON(s.Name, key, "null", s.Assertions[key])
			}
			continue
		}
		val := func() string {
			switch key {
			case "is_empty":
				return strconv.FormatBool(set.IsEmpty())
			case "span_lower":
				if hasSpan {
					return optInt32Str(span.LowerEndpoint())
				}
				return "null"
			case "span_upper":
				if hasSpan {
					return optInt32Str(span.UpperEndpoint())
				}
				return "null"
			case "span_lower_type":
				if hasSpan {
					return boundTypeStr(span.LowerBoundType())
				}
				return "null"
			case "span_upper_type":
				if hasSpan {
					return boundTypeStr(span.UpperBoundType())
				}
				return "null"
			case "encloses_query":
				if hasQuery {
					return strconv.FormatBool(set.Encloses(query))
				}
				return unknown(key)
			case "intersects_query":
				if hasQuery {
					return strconv.FormatBool(set.Intersects(query))
				}
				return unknown(key)
			}
			if n, ok := signedI32Suffix(key, "contains_"); ok {
				return strconv.FormatBool(set.Contains(n))
			}
			return unknown(key)
		}()
		emit(s.Name, key, val, s.Assertions[key], modeNone)
	}
}

func runRangeMap(s scenario) {
	m := rangev.NewInt32Int32RangeMap()
	for _, op := range s.Operations {
		switch op["op"] {
		case "put":
			m.Put(buildRangeObj(op["range"].(map[string]any)), asInt32(op["value"]))
		case "remove_range":
			m.Remove(buildRangeObj(op["range"].(map[string]any)))
		case "clear":
			m.Clear()
		default:
			// Forward-compat: unknown op kinds skip.
		}
	}
	var query rangev.Int32Range
	hasQuery := s.Query != nil
	if hasQuery {
		query = buildRangeObj(s.Query)
	}
	span, hasSpan := m.Span()
	for _, key := range sortedAssertionKeys(s.Assertions) {
		switch key {
		case "as_map_of_ranges":
			emitJSON(s.Name, key, entryArrayStr(m.AsMapOfRanges()), s.Assertions[key])
			continue
		case "sub_range_map_entries":
			if hasQuery {
				emitJSON(s.Name, key, entryArrayStr(m.SubRangeMap(query).AsMapOfRanges()), s.Assertions[key])
			}
			continue
		}
		if n, ok := signedI32Suffix(key, "get_entry_"); ok {
			if r, v, found := m.GetEntry(n); found {
				emitJSON(s.Name, key, entryObjStr(rangev.Int32Int32Entry{Range: r, Value: v}), s.Assertions[key])
			} else {
				emitJSON(s.Name, key, "null", s.Assertions[key])
			}
			continue
		}
		val := func() string {
			switch key {
			case "is_empty":
				return strconv.FormatBool(m.IsEmpty())
			case "span_lower":
				if hasSpan {
					return optInt32Str(span.LowerEndpoint())
				}
				return "null"
			case "span_upper":
				if hasSpan {
					return optInt32Str(span.UpperEndpoint())
				}
				return "null"
			case "span_lower_type":
				if hasSpan {
					return boundTypeStr(span.LowerBoundType())
				}
				return "null"
			case "span_upper_type":
				if hasSpan {
					return boundTypeStr(span.UpperBoundType())
				}
				return "null"
			}
			if n, ok := signedI32Suffix(key, "get_"); ok {
				if v, found := m.Get(n); found {
					return strconv.FormatInt(int64(v), 10)
				}
				return "null"
			}
			return unknown(key)
		}()
		emit(s.Name, key, val, s.Assertions[key], modeNone)
	}
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

// ---- BoundedLruMap<i32, i32> ---------------------------------------------

func parseU64Tick(v any) uint64 {
	switch x := v.(type) {
	case string:
		n, err := strconv.ParseUint(x, 10, 64)
		if err != nil {
			fatalf("invalid u64 decimal-string tick: %q", x)
		}
		return n
	case json.Number:
		n, err := strconv.ParseUint(x.String(), 10, 64)
		if err != nil {
			fatalf("invalid u64 tick: %q", x.String())
		}
		return n
	case float64:
		return uint64(x)
	}
	fatalf("expected u64 tick (decimal string or number), got %T (%v)", v, v)
	return 0
}

type lruLog struct {
	putResults          []*int32
	getResults          []*int32
	getOrDefaultResults []int32
	containsResults     []bool
	removeResults       []*int32
	expiredCounts       []int32
	snapshotKeysLog     [][]int32
	snapshotValuesLog   [][]int32
	snapshotEntriesLog  [][]boundedlru.Entry
}

func runBoundedLru(s scenario) {
	if s.MaxSize == nil {
		fatalf("BoundedLruMap scenario needs a non-negative max_size")
	}
	maxSize, err := s.MaxSize.Int64()
	if err != nil || maxSize < 0 {
		fatalf("BoundedLruMap max_size must be a non-negative integer: %v", s.MaxSize)
	}

	builder := boundedlru.NewBuilderBoundedLruInt32Int32Map().MaxSize(int(maxSize))
	if len(s.TTL) > 0 && string(s.TTL) != "null" {
		var raw any
		dec := json.NewDecoder(strings.NewReader(string(s.TTL)))
		dec.UseNumber()
		if err := dec.Decode(&raw); err != nil {
			fatalf("invalid ttl: %v", err)
		}
		builder = builder.TTL(parseU64Tick(raw))
	}

	var evictLog []boundedlru.Entry
	var evictCauses []boundedlru.EvictionCause
	m := builder.OnEvict(func(k, v int32, c boundedlru.EvictionCause) {
		evictLog = append(evictLog, boundedlru.Entry{Key: k, Value: v})
		evictCauses = append(evictCauses, c)
	}).Build()

	var log lruLog
	for _, op := range s.Operations {
		switch op["op"] {
		case "put":
			k, v := asInt32(op["key"]), asInt32(op["value"])
			var prev int32
			var ok bool
			if now, present := op["now"]; present && now != nil {
				prev, ok = m.PutAt(k, v, parseU64Tick(now))
			} else {
				prev, ok = m.Put(k, v)
			}
			log.putResults = append(log.putResults, optPtr(prev, ok))
		case "put_at":
			k, v := asInt32(op["key"]), asInt32(op["value"])
			prev, ok := m.PutAt(k, v, parseU64Tick(op["now"]))
			log.putResults = append(log.putResults, optPtr(prev, ok))
		case "get":
			v, ok := m.Get(asInt32(op["key"]))
			log.getResults = append(log.getResults, optPtr(v, ok))
		case "get_or_default":
			d := asInt32(op["default"])
			log.getOrDefaultResults = append(log.getOrDefaultResults, m.GetOrDefault(asInt32(op["key"]), d))
		case "contains_key":
			log.containsResults = append(log.containsResults, m.ContainsKey(asInt32(op["key"])))
		case "remove":
			v, ok := m.Remove(asInt32(op["key"]))
			log.removeResults = append(log.removeResults, optPtr(v, ok))
		case "clear":
			m.Clear()
		case "expire_entries":
			log.expiredCounts = append(log.expiredCounts, int32(m.ExpireEntries(parseU64Tick(op["now"]))))
		case "snapshot_keys":
			log.snapshotKeysLog = append(log.snapshotKeysLog, m.Keys())
		case "snapshot_values":
			log.snapshotValuesLog = append(log.snapshotValuesLog, m.Values())
		case "snapshot_entries":
			log.snapshotEntriesLog = append(log.snapshotEntriesLog, m.Entries())
		default:
		}
	}

	for _, key := range sortedAssertionKeys(s.Assertions) {
		emit(s.Name, key, evalLruAssertion(key, m, &log, evictLog, evictCauses), s.Assertions[key], modeNone)
	}
}

func optPtr(v int32, ok bool) *int32 {
	if !ok {
		return nil
	}
	return &v
}

func boolArray(v []bool) string {
	parts := make([]string, len(v))
	for i, b := range v {
		parts[i] = strconv.FormatBool(b)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func arrayOfInt32Arrays(v [][]int32) string {
	parts := make([]string, len(v))
	for i, inner := range v {
		parts[i] = formatArray(inner)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func arrayOfPairArrays(v [][]boundedlru.Entry) string {
	outer := make([]string, len(v))
	for i, inner := range v {
		pairs := make([]string, len(inner))
		for j, e := range inner {
			pairs[j] = fmt.Sprintf("[%d,%d]", e.Key, e.Value)
		}
		outer[i] = "[" + strings.Join(pairs, ",") + "]"
	}
	return "[" + strings.Join(outer, ",") + "]"
}

func evalLruAssertion(key string, m *boundedlru.BoundedLruInt32Int32Map, log *lruLog, evictLog []boundedlru.Entry, evictCauses []boundedlru.EvictionCause) string {
	switch key {
	case "size":
		return strconv.Itoa(m.Size())
	case "is_empty":
		return strconv.FormatBool(m.IsEmpty())
	case "lru_order_keys":
		return formatArray(m.Keys())
	case "lru_order_values":
		return formatArray(m.Values())
	case "eviction_log":
		parts := make([]string, len(evictLog))
		for i, e := range evictLog {
			parts[i] = fmt.Sprintf("[%d,%d,%q]", e.Key, e.Value, evictCauses[i].String())
		}
		return "[" + strings.Join(parts, ",") + "]"
	case "put_results":
		return optArray(log.putResults)
	case "get_results":
		return optArray(log.getResults)
	case "get_or_default_results":
		return formatArray(log.getOrDefaultResults)
	case "contains_results":
		return boolArray(log.containsResults)
	case "remove_results":
		return optArray(log.removeResults)
	case "expired_counts":
		return formatArray(log.expiredCounts)
	case "snapshot_keys_log":
		return arrayOfInt32Arrays(log.snapshotKeysLog)
	case "snapshot_values_log":
		return arrayOfInt32Arrays(log.snapshotValuesLog)
	case "snapshot_entries_log":
		return arrayOfPairArrays(log.snapshotEntriesLog)
	}
	if rest, ok := strings.CutPrefix(key, "get_"); ok {
		k, err := strconv.ParseInt(rest, 10, 32)
		if err != nil {
			return unknown(key)
		}
		// Peek is the production NON-TOUCH read: Get would refresh recency and
		// corrupt the very LRU order that later assertions in this same
		// scenario observe (README: assertion reads must not mutate the
		// collection).
		if v, ok := m.Peek(int32(k)); ok {
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
}

// ---- CountMin (spec/features/count-min.md) -------------------------------
//
// A d x w integer counter matrix. Built by exactly ONE `with_params` op, first
// (zero or multiple => malformed => SKIP, like HashPipeline); never `optimal`
// (the float-derivation trap is kept out of the shared suite -- an
// optimal/epsilon/delta op is unknown here => SKIP). Subsequent `add` ops carry
// an i32 `value` and a `count` DECIMAL STRING (omitted => 1; may exceed 2^53).
// Counters / estimate_<v> / total are u64 DECIMAL STRINGS (the 2^64 range
// exceeds JSON-safe 2^53); depth/width are plain ints. `counters` is the
// row-major (explicit-order, NOT sorted) primary oracle. Unknown ops/keys SKIP
// (forward-compat).

// parseCountOpt parses a `count` operand: a DECIMAL STRING parsed straight to
// uint64 (never via f64), reusing the i64-suite's wide-integer discipline. A
// bare JSON number is also accepted for small counts. Returns ok=false if
// malformed (negative, non-numeric, or exceeding u64::MAX) so the caller SKIPs.
// A nil/absent operand means count omitted => 1 (the add_one shape).
func parseCountOpt(v any) (uint64, bool) {
	switch n := v.(type) {
	case nil:
		return 1, true
	case string:
		c, err := strconv.ParseUint(n, 10, 64)
		return c, err == nil
	case json.Number:
		c, err := strconv.ParseUint(n.String(), 10, 64)
		return c, err == nil
	case float64:
		if n < 0 || n != math.Trunc(n) {
			return 0, false
		}
		return uint64(n), true
	}
	return 0, false
}

// leadingOpIs reports whether the operations list has exactly one op named
// `name` and it is the first op (the "exactly one leading X" authoring rule).
func leadingOpIs(ops []map[string]any, name string) bool {
	count := 0
	for _, op := range ops {
		if op["op"] == name {
			count++
		}
	}
	return count == 1 && len(ops) > 0 && ops[0]["op"] == name
}

func runCountMin(s scenario) {
	if !leadingOpIs(s.Operations, "with_params") {
		fmt.Fprintln(os.Stderr, "skip: CountMin scenario needs exactly one leading `with_params` op (forward-compat)")
		return
	}
	ctor := s.Operations[0]
	d := uint32(asInt(ctor["d"]))
	w := uint32(asInt(ctor["w"]))
	cms := countmin.NewCountMinWithParams(d, w)

	for _, op := range s.Operations[1:] {
		switch op["op"] {
		case "add":
			value := asInt32(op["value"])
			count, ok := parseCountOpt(op["count"])
			if !ok {
				fmt.Fprintln(os.Stderr, "skip: CountMin add `count` is not a 0..=u64::MAX integer")
				return
			}
			cms.Add(value, count)
		default:
			fmt.Fprintf(os.Stderr, "skip: unknown CountMin op (forward-compat): %v\n", op["op"])
			return
		}
	}

	for _, key := range sortedAssertionKeys(s.Assertions) {
		emit(s.Name, key, evalCountMin(key, cms), s.Assertions[key], modeNone)
	}
}

func evalCountMin(key string, cms *countmin.CountMin) string {
	switch key {
	case "counters":
		// Row-major counter matrix, each u64 as a QUOTED decimal string;
		// explicit order (matches the assertion array of decimal strings).
		cs := cms.ToCounters()
		parts := make([]string, len(cs))
		for i, c := range cs {
			parts[i] = "\"" + strconv.FormatUint(c, 10) + "\""
		}
		return "[" + strings.Join(parts, ",") + "]"
	case "total":
		return strconv.FormatUint(cms.Total(), 10)
	case "depth":
		return strconv.FormatUint(uint64(cms.Depth()), 10)
	case "width":
		return strconv.FormatUint(uint64(cms.Width()), 10)
	}
	if v, ok := estimateKey(key); ok {
		return strconv.FormatUint(cms.Estimate(v), 10)
	}
	return unknown(key)
}

// estimateKey recognises an estimate_<v> assertion: <v> is a SIGNED base-10 i32
// (exact ^estimate_(-?[0-9]+)$, full i32 range incl. negatives). A leading `+`
// is rejected so the recogniser matches the documented regex.
func estimateKey(key string) (int32, bool) {
	rest, ok := strings.CutPrefix(key, "estimate_")
	if !ok {
		return 0, false
	}
	return parseSignedI32(rest)
}

// parseSignedI32 parses a signed base-10 i32, rejecting a leading `+` and any
// non-digit body (matching the Rust runner's signed-suffix recogniser).
func parseSignedI32(rest string) (int32, bool) {
	digits := strings.TrimPrefix(rest, "-")
	if digits == "" || strings.IndexFunc(digits, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
		return 0, false
	}
	v, err := strconv.ParseInt(rest, 10, 32)
	if err != nil {
		return 0, false
	}
	return int32(v), true
}

// ---- SpaceSaving (spec/features/count-min.md) ----------------------------
//
// A bounded heavy-hitters summary. Built by exactly ONE `with_capacity` op,
// first (zero or multiple => SKIP). Subsequent `add` ops are applied IN LISTED
// ORDER (Space-Saving is order-dependent -- a runner MUST NOT reorder). `value`
// is an i32; `count` is a u64 decimal string (omitted => 1). monitored_set /
// top_k_<k> are explicit-order arrays of [item, count_str, error_str] triples
// in canonical order (count DESC, signed item ASC). count/error are u64 decimal
// strings (2^64 range); size/capacity plain ints. Unknown ops/keys SKIP.

func runSpaceSaving(s scenario) {
	if !leadingOpIs(s.Operations, "with_capacity") {
		fmt.Fprintln(os.Stderr, "skip: SpaceSaving scenario needs exactly one leading `with_capacity` op (forward-compat)")
		return
	}
	m := uint32(asInt(s.Operations[0]["m"]))
	ss := countmin.NewSpaceSaving(m)

	for _, op := range s.Operations[1:] {
		switch op["op"] {
		case "add":
			value := asInt32(op["value"])
			count, ok := parseCountOpt(op["count"])
			if !ok {
				fmt.Fprintln(os.Stderr, "skip: SpaceSaving add `count` is not a 0..=u64::MAX integer")
				return
			}
			ss.Add(value, count)
		default:
			fmt.Fprintf(os.Stderr, "skip: unknown SpaceSaving op (forward-compat): %v\n", op["op"])
			return
		}
	}

	for _, key := range sortedAssertionKeys(s.Assertions) {
		emit(s.Name, key, evalSpaceSaving(key, ss), s.Assertions[key], modeNone)
	}
}

// formatSSTriples renders an SSEntry list as a JSON array of
// [item, "count", "error"] (item int, count/error u64 quoted decimal strings).
func formatSSTriples(triples []countmin.SSEntry) string {
	parts := make([]string, len(triples))
	for i, t := range triples {
		parts[i] = fmt.Sprintf("[%d,\"%s\",\"%s\"]",
			t.Item, strconv.FormatUint(t.Count, 10), strconv.FormatUint(t.Error, 10))
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func evalSpaceSaving(key string, ss *countmin.SpaceSaving) string {
	switch key {
	case "monitored_set":
		return formatSSTriples(ss.MonitoredSet())
	case "size":
		return strconv.FormatUint(uint64(ss.Size()), 10)
	case "capacity":
		return strconv.FormatUint(uint64(ss.Capacity()), 10)
	}
	if k, ok := topKKey(key); ok {
		return formatSSTriples(ss.TopK(k))
	}
	if v, ok := ssSignedKey(key, "count_"); ok {
		return strconv.FormatUint(ss.Count(v), 10)
	}
	if v, ok := ssSignedKey(key, "error_"); ok {
		return strconv.FormatUint(ss.Error(v), 10)
	}
	return unknown(key)
}

// topKKey recognises a top_k_<k> assertion: <k> is a NON-NEGATIVE base-10 int
// (exact ^top_k_([0-9]+)$).
func topKKey(key string) (uint32, bool) {
	rest, ok := strings.CutPrefix(key, "top_k_")
	if !ok || rest == "" {
		return 0, false
	}
	if strings.IndexFunc(rest, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
		return 0, false
	}
	v, err := strconv.ParseUint(rest, 10, 32)
	if err != nil {
		return 0, false
	}
	return uint32(v), true
}

// ssSignedKey recognises a <prefix><v> assertion whose <v> is a SIGNED base-10
// i32 (full range incl. negatives; leading `+` rejected). Used for
// count_<v>/error_<v>.
func ssSignedKey(key, prefix string) (int32, bool) {
	rest, ok := strings.CutPrefix(key, prefix)
	if !ok {
		return 0, false
	}
	return parseSignedI32(rest)
}
