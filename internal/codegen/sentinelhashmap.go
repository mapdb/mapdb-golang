package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

// shmData is the per key/value view the sentinelhashmap template iterates over.
//
// The sentinelhashmap family is an open-addressing hash map that, instead of an
// occupied []bool array, reserves two key values — 0 (empty) and 1 (removed) —
// as in-table slot-state markers. Because 0 and 1 are also valid USER keys they
// are stored in dedicated struct fields (zeroKey*/oneKey*) and every operation
// special-cases them. There is exactly ONE shape (base only).
//
// TWO independent float axes drive the type-dependent logic — they mirror the
// base hashmap's axes but the KEY axis additionally pulls in a -0.0 sentinel:
//
// KEY axis (KeyIsFloat / KeyBitsFn / KeyHashExpr):
//   - hashKey: identical to the base hashmap's keyHashExpr (int/char golden-mix
//     uint64(<cast>(key)); int32 double-casts through uint32; floats reinterpret
//     the bit pattern via math.FloatNbits).
//   - sentinel checks: for FLOAT keys the empty/0 sentinel is compared on the
//     bit pattern — KeyBitsFn(key) == KeyBitsFn(emptyKey) — and a SEPARATE -0.0
//     sentinel (KeyBitsFn(key) == <Map>NegZeroBits) is routed to its own
//     zeroKey-style field, since +0.0 collides with the empty marker. The 1
//     (removed) sentinel for floats is a plain value compare (key == removedKey).
//     For int/char keys all three are plain == against the typed sentinel const.
//   - probe-site key equality: float keys use KeyBitsFn(k) == KeyBitsFn(key);
//     int/char keys use ==.
//
// VALUE axis (ValueIsFloat / ValBitsFn): ContainsValue compares values; float
// values use bit-pattern equality, int/char values use ==. ValZero is the value
// zero literal ("0" or "0.0").
//
// Imports: "math" iff (KeyIsFloat || ValueIsFloat).
type shmData struct {
	KeyName  string // Int32, Float32, Char (key identifier stem)
	KeyType  string // int32, float32, uint16 (Go key type)
	KeySnake string // int32, float32, char (key file-name stem)
	KeyZero  string // key zero literal ("0" or "0.0")

	ValName  string // Int32, Float32, Char (value identifier stem)
	ValType  string // int32, float32, uint16 (Go value type)
	ValSnake string // int32, float32, char (value file-name stem)
	ValZero  string // value zero literal ("0" or "0.0")

	KeyIsFloat   bool
	KeyBitsFn    string // math.Float32bits / math.Float64bits (float keys only)
	KeyHashExpr  string // inner operand of the golden-ratio multiply in hashKey
	NegZeroType  string // uint32 / uint64 — the NegZeroBits const type (float keys only)
	NegZeroBits  string // 0x80000000 / 0x8000000000000000 (float keys only)
	NegZeroValue string // floatN(math.Copysign(0, -1)) literal for the -0.0 key

	ValueIsFloat bool
	ValBitsFn    string // math.Float32bits / math.Float64bits (float values only)

	NeedsMath bool

	MapName   string // Int32Float32 (exported type stem: Int32Float32)
	MapSnake  string // int32_float32 (file-name stem)
	EntryStem string // int32Float32 (lower-camel, used for unexported const stem)
}

// genSentinelHashMap writes the 7×7 = 49 prim×prim sentinel hash map sources
// (base only) into the current working directory. Invoked from
// sentinelhashmap/ via go:generate.
func genSentinelHashMap() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	base := template.Must(template.New("shm-base").Parse(sentinelHashMapTmpl))

	write := func(name string, data shmData) error {
		var buf bytes.Buffer
		if err := base.Execute(&buf, data); err != nil {
			return fmt.Errorf("execute %s: %w", name, err)
		}
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return fmt.Errorf("format %s: %w\n---\n%s", name, err, buf.String())
		}
		out := filepath.Join(cwd, name)
		return os.WriteFile(out, formatted, 0o644)
	}

	prims := Primitives()
	for _, k := range prims {
		for _, v := range prims {
			data := shmData{
				KeyName:  k.Name,
				KeyType:  k.GoType,
				KeySnake: k.SnakeName,
				KeyZero:  "0",

				ValName:  v.Name,
				ValType:  v.GoType,
				ValSnake: v.SnakeName,
				ValZero:  "0",

				KeyIsFloat:  k.IsFloating,
				KeyHashExpr: keyHashExpr(k),

				ValueIsFloat: v.IsFloating,

				NeedsMath: k.IsFloating || v.IsFloating,

				MapName:   k.Name + v.Name,
				MapSnake:  k.SnakeName + "_" + v.SnakeName,
				EntryStem: lowerFirst(k.Name) + v.Name,
			}
			if k.IsFloating {
				data.KeyZero = "0.0"
				if k.ByteSize == 8 {
					data.KeyBitsFn = "math.Float64bits"
					data.NegZeroType = "uint64"
					data.NegZeroBits = "0x8000000000000000"
					data.NegZeroValue = "math.Copysign(0, -1)"
				} else {
					data.KeyBitsFn = "math.Float32bits"
					data.NegZeroType = "uint32"
					data.NegZeroBits = "0x80000000"
					data.NegZeroValue = "float32(math.Copysign(0, -1))"
				}
			}
			if v.IsFloating {
				data.ValZero = "0.0"
				data.ValBitsFn = "math.Float32bits"
				if v.ByteSize == 8 {
					data.ValBitsFn = "math.Float64bits"
				}
			}

			if err := write(data.MapSnake+"_sentinel_hash_map.go", data); err != nil {
				return err
			}
		}
	}

	return nil
}

const sentinelHashMapTmpl = genHeader + `package sentinelhashmap

import (
	"fmt"
	"iter"
{{- if .NeedsMath}}
	"math"
{{- end}}
	"strings"
)

const (
	{{.EntryStem}}DefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
	{{.EntryStem}}EmptyKey   = {{.KeyType}}(0)
	{{.EntryStem}}RemovedKey = {{.KeyType}}(1)
{{- if .KeyIsFloat}}
	// {{.EntryStem}}NegZeroBits is the IEEE-754 bit pattern of -0.0, routed to a
	// dedicated field so -0.0 stays distinct from +0.0 (which collides with
	// the empty sentinel) and from the table.
	{{.EntryStem}}NegZeroBits = {{.NegZeroType}}({{.NegZeroBits}})
{{- end}}
)

// {{.MapName}} is a sentinel-based open-addressing hash map with {{.KeyType}} keys and {{.ValType}} values.
// It uses sentinel values (0=empty, 1=removed) to track slot state.
// Keys 0 and 1 are stored separately in dedicated fields.
type {{.MapName}} struct {
	keys   []{{.KeyType}}
	values []{{.ValType}}
	size   int

	// Sentinel key storage — keys 0 and 1 are valid user keys but also
	// serve as empty/removed markers in the table, so we store them separately.
	zeroKeyPresent bool
	zeroKeyValue   {{.ValType}}
{{- if .KeyIsFloat}}
	negZeroKeyPresent bool
	negZeroKeyValue   {{.ValType}}
{{- end}}
	oneKeyPresent  bool
	oneKeyValue    {{.ValType}}
}

// New{{.MapName}} creates a new empty {{.MapName}} with default capacity.
func New{{.MapName}}() *{{.MapName}} {
	return New{{.MapName}}WithCapacity({{.EntryStem}}DefaultCapacity)
}

// New{{.MapName}}WithCapacity creates a new empty {{.MapName}} with the given initial capacity.
func New{{.MapName}}WithCapacity(capacity int) *{{.MapName}} {
	cap := nextPowerOfTwo{{.MapName}}(capacity)
	return &{{.MapName}}{
		keys:   make([]{{.KeyType}}, cap),
		values: make([]{{.ValType}}, cap),
		size:   0,
	}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *{{.MapName}}) Put(key {{.KeyType}}, value {{.ValType}}) ({{.ValType}}, bool) {
	if {{if .KeyIsFloat}}{{.KeyBitsFn}}(key) == {{.KeyBitsFn}}({{.EntryStem}}EmptyKey){{else}}key == {{.EntryStem}}EmptyKey{{end}} {
		old := m.zeroKeyValue
		existed := m.zeroKeyPresent
		m.zeroKeyValue = value
		if !m.zeroKeyPresent {
			m.zeroKeyPresent = true
			m.size++
		}
		return old, existed
	}
{{- if .KeyIsFloat}}
	if {{.KeyBitsFn}}(key) == {{.EntryStem}}NegZeroBits {
		old := m.negZeroKeyValue
		existed := m.negZeroKeyPresent
		m.negZeroKeyValue = value
		if !m.negZeroKeyPresent {
			m.negZeroKeyPresent = true
			m.size++
		}
		return old, existed
	}
{{- end}}
	if key == {{.EntryStem}}RemovedKey {
		old := m.oneKeyValue
		existed := m.oneKeyPresent
		m.oneKeyValue = value
		if !m.oneKeyPresent {
			m.oneKeyPresent = true
			m.size++
		}
		return old, existed
	}
	return m.putRegular(key, value)
}

func (m *{{.MapName}}) putRegular(key {{.KeyType}}, value {{.ValType}}) ({{.ValType}}, bool) {
	if m.needsResize() {
		m.resize()
	}
	cap := len(m.keys)
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask
	empty := {{.EntryStem}}EmptyKey
	removed := {{.EntryStem}}RemovedKey
	firstRemoved := -1

	for {
		k := m.keys[idx]
		if k == empty {
			if firstRemoved >= 0 {
				idx = firstRemoved
			}
			m.keys[idx] = key
			m.values[idx] = value
			m.size++
			return {{.ValZero}}, false
		}
		if k == removed {
			if firstRemoved < 0 {
				firstRemoved = idx
			}
			idx = (idx + 1) & mask
			continue
		}
		if {{if .KeyIsFloat}}{{.KeyBitsFn}}(k) == {{.KeyBitsFn}}(key){{else}}k == key{{end}} {
			old := m.values[idx]
			m.values[idx] = value
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// Get returns the value for the given key and true if found, or the zero value and false if not.
func (m *{{.MapName}}) Get(key {{.KeyType}}) ({{.ValType}}, bool) {
	if {{if .KeyIsFloat}}{{.KeyBitsFn}}(key) == {{.KeyBitsFn}}({{.EntryStem}}EmptyKey){{else}}key == {{.EntryStem}}EmptyKey{{end}} {
		if m.zeroKeyPresent {
			return m.zeroKeyValue, true
		}
		return {{.ValZero}}, false
	}
{{- if .KeyIsFloat}}
	if {{.KeyBitsFn}}(key) == {{.EntryStem}}NegZeroBits {
		if m.negZeroKeyPresent {
			return m.negZeroKeyValue, true
		}
		return {{.ValZero}}, false
	}
{{- end}}
	if key == {{.EntryStem}}RemovedKey {
		if m.oneKeyPresent {
			return m.oneKeyValue, true
		}
		return {{.ValZero}}, false
	}
	cap := len(m.keys)
	if cap == 0 {
		return {{.ValZero}}, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask
	empty := {{.EntryStem}}EmptyKey

	for {
		k := m.keys[idx]
		if k == empty {
			return {{.ValZero}}, false
		}
		if {{if .KeyIsFloat}}{{.KeyBitsFn}}(k) == {{.KeyBitsFn}}(key){{else}}k == key{{end}} {
			return m.values[idx], true
		}
		idx = (idx + 1) & mask
	}
}

// GetOrDefault returns the value for the given key if present, or the default value otherwise.
func (m *{{.MapName}}) GetOrDefault(key {{.KeyType}}, defaultValue {{.ValType}}) {{.ValType}} {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// Remove deletes the entry for the given key. Returns the previous value and true if the key existed.
func (m *{{.MapName}}) Remove(key {{.KeyType}}) ({{.ValType}}, bool) {
	if {{if .KeyIsFloat}}{{.KeyBitsFn}}(key) == {{.KeyBitsFn}}({{.EntryStem}}EmptyKey){{else}}key == {{.EntryStem}}EmptyKey{{end}} {
		if m.zeroKeyPresent {
			old := m.zeroKeyValue
			m.zeroKeyPresent = false
			m.zeroKeyValue = {{.ValZero}}
			m.size--
			return old, true
		}
		return {{.ValZero}}, false
	}
{{- if .KeyIsFloat}}
	if {{.KeyBitsFn}}(key) == {{.EntryStem}}NegZeroBits {
		if m.negZeroKeyPresent {
			old := m.negZeroKeyValue
			m.negZeroKeyPresent = false
			m.negZeroKeyValue = {{.ValZero}}
			m.size--
			return old, true
		}
		return {{.ValZero}}, false
	}
{{- end}}
	if key == {{.EntryStem}}RemovedKey {
		if m.oneKeyPresent {
			old := m.oneKeyValue
			m.oneKeyPresent = false
			m.oneKeyValue = {{.ValZero}}
			m.size--
			return old, true
		}
		return {{.ValZero}}, false
	}
	return m.removeRegular(key)
}

func (m *{{.MapName}}) removeRegular(key {{.KeyType}}) ({{.ValType}}, bool) {
	cap := len(m.keys)
	if cap == 0 {
		return {{.ValZero}}, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask
	empty := {{.EntryStem}}EmptyKey

	for {
		k := m.keys[idx]
		if k == empty {
			return {{.ValZero}}, false
		}
		if {{if .KeyIsFloat}}{{.KeyBitsFn}}(k) == {{.KeyBitsFn}}(key){{else}}k == key{{end}} {
			old := m.values[idx]
			m.keys[idx] = {{.EntryStem}}RemovedKey
			m.values[idx] = {{.ValZero}}
			m.size--
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// ContainsKey returns true if the map contains the given key.
func (m *{{.MapName}}) ContainsKey(key {{.KeyType}}) bool {
	_, ok := m.Get(key)
	return ok
}

// ContainsValue returns true if the map contains the given value.
func (m *{{.MapName}}) ContainsValue(value {{.ValType}}) bool {
	empty := {{.EntryStem}}EmptyKey
	removed := {{.EntryStem}}RemovedKey
	if m.zeroKeyPresent && {{if .ValueIsFloat}}{{.ValBitsFn}}(m.zeroKeyValue) == {{.ValBitsFn}}(value){{else}}m.zeroKeyValue == value{{end}} {
		return true
	}
{{- if .KeyIsFloat}}
	if m.negZeroKeyPresent && {{if .ValueIsFloat}}{{.ValBitsFn}}(m.negZeroKeyValue) == {{.ValBitsFn}}(value){{else}}m.negZeroKeyValue == value{{end}} {
		return true
	}
{{- end}}
	if m.oneKeyPresent && {{if .ValueIsFloat}}{{.ValBitsFn}}(m.oneKeyValue) == {{.ValBitsFn}}(value){{else}}m.oneKeyValue == value{{end}} {
		return true
	}
	for i := range m.keys {
		if m.keys[i] != empty && m.keys[i] != removed && {{if .ValueIsFloat}}{{.ValBitsFn}}(m.values[i]) == {{.ValBitsFn}}(value){{else}}m.values[i] == value{{end}} {
			return true
		}
	}
	return false
}

// Len returns the number of elements. Use m.Len() == 0 to test for emptiness.
func (m *{{.MapName}}) Len() int {
	return m.size
}

// Clear removes all entries from the map.
func (m *{{.MapName}}) Clear() {
	for i := range m.keys {
		m.keys[i] = {{.KeyZero}}
		m.values[i] = {{.ValZero}}
	}
	m.zeroKeyPresent = false
	m.zeroKeyValue = {{.ValZero}}
{{- if .KeyIsFloat}}
	m.negZeroKeyPresent = false
	m.negZeroKeyValue = {{.ValZero}}
{{- end}}
	m.oneKeyPresent = false
	m.oneKeyValue = {{.ValZero}}
	m.size = 0
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *{{.MapName}}) All() iter.Seq2[{{.KeyType}}, {{.ValType}}] {
	return func(yield func({{.KeyType}}, {{.ValType}}) bool) {
		if m.zeroKeyPresent {
			if !yield({{.KeyZero}}, m.zeroKeyValue) {
				return
			}
		}
{{- if .KeyIsFloat}}
		if m.negZeroKeyPresent {
			if !yield({{.NegZeroValue}}, m.negZeroKeyValue) {
				return
			}
		}
{{- end}}
		if m.oneKeyPresent {
			if !yield({{.EntryStem}}RemovedKey, m.oneKeyValue) {
				return
			}
		}
		empty := {{.EntryStem}}EmptyKey
		removed := {{.EntryStem}}RemovedKey
		for i := range m.keys {
			if m.keys[i] != empty && m.keys[i] != removed {
				if !yield(m.keys[i], m.values[i]) {
					return
				}
			}
		}
	}
}

// Keys returns an iter.Seq that yields all keys.
func (m *{{.MapName}}) Keys() iter.Seq[{{.KeyType}}] {
	return func(yield func({{.KeyType}}) bool) {
		if m.zeroKeyPresent {
			if !yield({{.KeyZero}}) {
				return
			}
		}
{{- if .KeyIsFloat}}
		if m.negZeroKeyPresent {
			if !yield({{.NegZeroValue}}) {
				return
			}
		}
{{- end}}
		if m.oneKeyPresent {
			if !yield({{.EntryStem}}RemovedKey) {
				return
			}
		}
		empty := {{.EntryStem}}EmptyKey
		removed := {{.EntryStem}}RemovedKey
		for i := range m.keys {
			if m.keys[i] != empty && m.keys[i] != removed {
				if !yield(m.keys[i]) {
					return
				}
			}
		}
	}
}

// Values returns an iter.Seq that yields all values.
func (m *{{.MapName}}) Values() iter.Seq[{{.ValType}}] {
	return func(yield func({{.ValType}}) bool) {
		if m.zeroKeyPresent {
			if !yield(m.zeroKeyValue) {
				return
			}
		}
{{- if .KeyIsFloat}}
		if m.negZeroKeyPresent {
			if !yield(m.negZeroKeyValue) {
				return
			}
		}
{{- end}}
		if m.oneKeyPresent {
			if !yield(m.oneKeyValue) {
				return
			}
		}
		empty := {{.EntryStem}}EmptyKey
		removed := {{.EntryStem}}RemovedKey
		for i := range m.keys {
			if m.keys[i] != empty && m.keys[i] != removed {
				if !yield(m.values[i]) {
					return
				}
			}
		}
	}
}

// ForEach calls the given function for each key-value pair.
func (m *{{.MapName}}) ForEach(f func({{.KeyType}}, {{.ValType}})) {
	for k, v := range m.All() {
		f(k, v)
	}
}

// Select returns a new map containing only the key-value pairs that satisfy the predicate.
func (m *{{.MapName}}) Select(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}} {
	result := New{{.MapName}}()
	for k, v := range m.All() {
		if predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// Reject returns a new map containing only the key-value pairs that do not satisfy the predicate.
func (m *{{.MapName}}) Reject(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}} {
	result := New{{.MapName}}()
	for k, v := range m.All() {
		if !predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// AnySatisfy returns true if any key-value pair satisfies the predicate.
func (m *{{.MapName}}) AnySatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	for k, v := range m.All() {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all key-value pairs satisfy the predicate.
func (m *{{.MapName}}) AllSatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	for k, v := range m.All() {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no key-value pair satisfies the predicate.
func (m *{{.MapName}}) NoneSatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	return !m.AnySatisfy(predicate)
}

// String returns a string representation of the map.
func (m *{{.MapName}}) String() string {
	if m.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for k, v := range m.All() {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v: %v", k, v)
		first = false
	}
	sb.WriteString("}")
	return sb.String()
}

func (m *{{.MapName}}) hashKey(key {{.KeyType}}) uint64 {
	h := {{.KeyHashExpr}} * 0x9E3779B97F4A7C15
	return h ^ (h >> 32)
}

func (m *{{.MapName}}) needsResize() bool {
	// Count only regular entries (not sentinel entries) for load factor
	regularEntries := m.size
	if m.zeroKeyPresent {
		regularEntries--
	}
{{- if .KeyIsFloat}}
	if m.negZeroKeyPresent {
		regularEntries--
	}
{{- end}}
	if m.oneKeyPresent {
		regularEntries--
	}
	return (regularEntries+1)*4 >= len(m.keys)*3 // 0.75 load factor, integer math
}

func (m *{{.MapName}}) resize() {
	oldKeys := m.keys
	oldValues := m.values
	newCap := len(oldKeys) * 2
	if newCap == 0 {
		newCap = {{.EntryStem}}DefaultCapacity
	}

	// Save sentinel state
	savedZeroPresent := m.zeroKeyPresent
	savedZeroValue := m.zeroKeyValue
{{- if .KeyIsFloat}}
	savedNegZeroPresent := m.negZeroKeyPresent
	savedNegZeroValue := m.negZeroKeyValue
{{- end}}
	savedOnePresent := m.oneKeyPresent
	savedOneValue := m.oneKeyValue

	m.keys = make([]{{.KeyType}}, newCap)
	m.values = make([]{{.ValType}}, newCap)
	m.size = 0
	m.zeroKeyPresent = false
{{- if .KeyIsFloat}}
	m.negZeroKeyPresent = false
{{- end}}
	m.oneKeyPresent = false

	// Re-insert sentinel entries
	if savedZeroPresent {
		m.Put({{.KeyZero}}, savedZeroValue)
	}
{{- if .KeyIsFloat}}
	if savedNegZeroPresent {
		m.Put({{.NegZeroValue}}, savedNegZeroValue)
	}
{{- end}}
	if savedOnePresent {
		m.Put({{.EntryStem}}RemovedKey, savedOneValue)
	}

	// Re-insert regular entries
	empty := {{.EntryStem}}EmptyKey
	removed := {{.EntryStem}}RemovedKey
	for i := range oldKeys {
		if oldKeys[i] != empty && oldKeys[i] != removed {
			m.Put(oldKeys[i], oldValues[i])
		}
	}
}

func nextPowerOfTwo{{.MapName}}(n int) int {
	if n <= 0 {
		return 16
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32 // no-op on 32-bit platforms (Go shifts are width-defined), required on 64-bit
	n++
	return n
}
`
