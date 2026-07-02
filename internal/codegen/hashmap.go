package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

// hmData is the per key/value view the hashmap templates iterate over.
//
// The hashmap family is an open-addressing hash map and the K×V analog of the
// hashset family. Each slot holds a key AND a value. It has THREE variants:
// base, immutable wrapper, synchronized wrapper.
//
// TWO independent float axes drive the only type-dependent logic:
//
// KEY axis (KeyIsFloat / KeyHashExpr / KeyBitsFn):
//   - hashKey: int/char keys golden-ratio-mix uint64(<unsignedcast>(key));
//     int32 alone double-casts through uint32; float keys reinterpret the bit
//     pattern via unsafe before mixing. Each KeyHashExpr is captured verbatim.
//   - key equality at the probe sites (Put/Get/Remove/AndModify): float keys
//     use math.Float{32,64}bits(a) == math.Float{32,64}bits(b); int/char keys
//     use ==.
//
// VALUE axis (ValueIsFloat / ValBitsFn):
//   - ContainsValue and Equals compare values; float values use bit-pattern
//     equality, int/char values use ==.
//
// Zero literals: KeyZero/ValZero ("0" or "0.0") fill the key/value zero slots
// (Put/Get/Remove returns, Remove's slot clearing). Detect always returns the
// literal "0, 0" regardless of key/value type — matching the hand-written
// originals verbatim — so it is hardcoded in the template, not parameterised.
//
// Imports: "math" is needed iff (KeyIsFloat || ValueIsFloat) — used by key
// hashing/equality (math.FloatNbits in the float hashKey body) and/or value
// equality. The synchronized wrapper always imports "unsafe" (pointer-address
// lock ordering in Equals), independent of types.
type hmData struct {
	KeyName  string // Int32, Float32, Char (key identifier stem)
	KeyType  string // int32, float32, uint16 (Go key type)
	KeySnake string // int32, float32, char (key file-name stem)
	KeyZero  string // key zero literal ("0" or "0.0")

	ValName  string // Int32, Float32, Char (value identifier stem)
	ValType  string // int32, float32, uint16 (Go value type)
	ValSnake string // int32, float32, char (value file-name stem)
	ValZero  string // value zero literal ("0" or "0.0")

	// KeyIsFloat selects bit-pattern key equality at the probe sites and the
	// math.FloatNbits-based float hashKey body; it also (with ValueIsFloat)
	// gates the math import.
	KeyIsFloat bool
	// KeyBitsFn is the float bit-pattern function for key equality
	// (math.Float32bits / math.Float64bits). Float keys only.
	KeyBitsFn string
	// KeyHashExpr is the inner operand of the golden-ratio multiply in
	// hashKey: h := <KeyHashExpr> * 0x9E3779B97F4A7C15. Captured per key type
	// because the integer/char/float reinterpretations differ (int32 alone
	// double-casts through uint32; floats reinterpret via math.FloatNbits).
	KeyHashExpr string

	// ValueIsFloat selects bit-pattern value equality in ContainsValue/Equals.
	ValueIsFloat bool
	// ValBitsFn is the float bit-pattern function for value equality. Float
	// values only.
	ValBitsFn string

	// NeedsMath drives the base file's import block.
	NeedsMath bool

	// MapName is the combined identifier stem, e.g. Int32Float32 (used for the
	// exported type Int32Float32HashMap); MapSnake is the file-name stem,
	// e.g. int32_float32.
	MapName  string // Int32Float32
	MapSnake string // int32_float32
	// EntryStem is the lower-camel struct stem, e.g. int32Float32 (used for
	// int32Float32HashMapEntry / the Entry type Int32Float32Entry uses MapName).
	EntryStem string // int32Float32

	// RevName is the transposed (value→key) identifier stem, e.g. Float32Int32
	// for an Int32Float32 bimap. The bidirectional map's reverse inner map is a
	// *<RevName>HashMap, and Inverse() returns *<RevName>BiMap.
	RevName string // Float32Int32
}

// keyHashExpr returns the per-key-type inner operand of the golden-ratio
// multiply in hashKey, reproduced verbatim from the hand-written originals.
func keyHashExpr(p Primitive) string {
	switch {
	case p.IsFloating && p.ByteSize == 4:
		return "uint64(math.Float32bits(key))"
	case p.IsFloating && p.ByteSize == 8:
		return "math.Float64bits(key)"
	case p.GoType == "int32":
		return "uint64(uint32(key))"
	default:
		return "uint64(key)"
	}
}

// objHMData is the per-type view the object hash map templates iterate over.
//
// The object hash map family has TWO distinct generic shapes that do NOT mix:
//
// Shape A — Object<Value>HashMap[K comparable] (object KEY, prim value): the
// key is a generic comparable, hashed with the shared hashComparable helper and
// compared with ==; the value is a pure prim payload (no value-comparison
// methods). The only type-dependent bit is the value Go type name and its zero
// literal (ValZero: "0" or "0.0"). There is NO float-value branch.
//
// Shape B — <Key>ObjectHashMap[V any] (prim KEY, object value): the key axis is
// IDENTICAL to genHashMap's key axis (golden-ratio hashKey with the per-key-type
// KeyHashExpr, and bit-pattern equality for float keys); the value is V any, a
// pure payload whose zero is the generic `var zero V`. The math/unsafe imports
// are gated by KeyIsFloat alone.
type objHMData struct {
	// PrimName/PrimType/PrimSnake describe the prim half (the value for Shape A,
	// the key for Shape B). PrimZero is its zero literal ("0" or "0.0").
	PrimName  string // Int32, Float32, Char
	PrimType  string // int32, float32, uint16 (Go type)
	PrimSnake string // int32, float32, char
	PrimZero  string // "0" or "0.0"

	// MapName / MapSnake / EntryStem are the combined identifiers, e.g. for
	// Shape A value=int32: ObjectInt32 / object_int32 / objectInt32; for Shape B
	// key=int32: Int32Object / int32_object / int32Object.
	MapName   string
	MapSnake  string
	EntryStem string

	// KeyIsFloat (Shape B only) selects bit-pattern key equality and the
	// math.FloatNbits-based float hashKey body; it also gates the math import.
	KeyIsFloat  bool
	KeyBitsFn   string // math.Float32bits / math.Float64bits (float keys only)
	KeyHashExpr string // inner operand of the golden-ratio multiply in hashKey
	NeedsMath   bool
}

// genHashMap writes the 7×7 = 49 prim×prim hash map sources for each of the
// base, immutable, and synchronized variants (147 files), the 49 prim×prim
// bidirectional maps (*_hash_bi_map.go), and the 28 object-keyed/valued generic
// maps — Shape A Object<Value>HashMap[K comparable] (object_<value>_hash_map.go
// + immutable) and Shape B <Key>ObjectHashMap[V any] (<key>_object_hash_map.go
// + immutable), 7 prims × 2 shapes × 2 variants — into the current working
// directory. Invoked from hashmap/ via go:generate.
func genHashMap() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	base := template.Must(template.New("hm-base").Parse(hashMapTmpl))
	immutable := template.Must(template.New("hm-immutable").Parse(immutableHashMapTmpl))
	synchronized := template.Must(template.New("hm-sync").Parse(synchronizedHashMapTmpl))
	bimap := template.Must(template.New("hm-bimap").Parse(hashBiMapTmpl))
	objKey := template.Must(template.New("hm-objkey").Parse(objectKeyHashMapTmpl))
	objKeyImm := template.Must(template.New("hm-objkey-imm").Parse(immutableObjectKeyHashMapTmpl))
	objVal := template.Must(template.New("hm-objval").Parse(objectValueHashMapTmpl))
	objValImm := template.Must(template.New("hm-objval-imm").Parse(immutableObjectValueHashMapTmpl))

	writeData := func(name string, tmpl *template.Template, data any) error {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("execute %s: %w", name, err)
		}
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return fmt.Errorf("format %s: %w\n---\n%s", name, err, buf.String())
		}
		out := filepath.Join(cwd, name)
		return os.WriteFile(out, formatted, 0o644)
	}
	write := func(name string, tmpl *template.Template, data hmData) error {
		return writeData(name, tmpl, data)
	}

	prims := Primitives()
	for _, k := range prims {
		for _, v := range prims {
			data := hmData{
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
				RevName:   v.Name + k.Name,
			}
			if k.IsFloating {
				data.KeyZero = "0.0"
				data.KeyBitsFn = "math.Float32bits"
				if k.ByteSize == 8 {
					data.KeyBitsFn = "math.Float64bits"
				}
			}
			if v.IsFloating {
				data.ValZero = "0.0"
				data.ValBitsFn = "math.Float32bits"
				if v.ByteSize == 8 {
					data.ValBitsFn = "math.Float64bits"
				}
			}

			if err := write(data.MapSnake+"_hash_map.go", base, data); err != nil {
				return err
			}
			if err := write("immutable_"+data.MapSnake+"_hash_map.go", immutable, data); err != nil {
				return err
			}
			if err := write("synchronized_"+data.MapSnake+"_hash_map.go", synchronized, data); err != nil {
				return err
			}
			if err := write(data.MapSnake+"_hash_bi_map.go", bimap, data); err != nil {
				return err
			}
		}
	}

	// Object maps: two distinct generic shapes, 7 prims × 2 variants each.
	for _, p := range prims {
		// Shape A: object KEY, prim value — Object<Value>HashMap[K comparable].
		a := objHMData{
			PrimName:  p.Name,
			PrimType:  p.GoType,
			PrimSnake: p.SnakeName,
			PrimZero:  "0",
			MapName:   "Object" + p.Name,
			MapSnake:  "object_" + p.SnakeName,
			EntryStem: "object" + p.Name,
		}
		if p.IsFloating {
			a.PrimZero = "0.0"
		}
		if err := writeData(a.MapSnake+"_hash_map.go", objKey, a); err != nil {
			return err
		}
		if err := writeData("immutable_"+a.MapSnake+"_hash_map.go", objKeyImm, a); err != nil {
			return err
		}

		// Shape B: prim KEY, object value — <Key>ObjectHashMap[V any].
		b := objHMData{
			PrimName:    p.Name,
			PrimType:    p.GoType,
			PrimSnake:   p.SnakeName,
			PrimZero:    "0",
			MapName:     p.Name + "Object",
			MapSnake:    p.SnakeName + "_object",
			EntryStem:   lowerFirst(p.Name) + "Object",
			KeyIsFloat:  p.IsFloating,
			KeyHashExpr: keyHashExpr(p),
			NeedsMath:   p.IsFloating,
		}
		if p.IsFloating {
			b.PrimZero = "0.0"
			b.KeyBitsFn = "math.Float32bits"
			if p.ByteSize == 8 {
				b.KeyBitsFn = "math.Float64bits"
			}
		}
		if err := writeData(b.MapSnake+"_hash_map.go", objVal, b); err != nil {
			return err
		}
		if err := writeData("immutable_"+b.MapSnake+"_hash_map.go", objValImm, b); err != nil {
			return err
		}
	}

	return nil
}

const hashMapTmpl = genHeader + `package hashmap

import (
	"fmt"
	"iter"
{{- if .NeedsMath}}
	"math"
{{- end}}
	"strings"

	"github.com/mapdb/mapdb-golang/pump"
)

const (
	{{.EntryStem}}DefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
)

// {{.EntryStem}}Entry holds a single slot in the hash map for cache locality.
type {{.EntryStem}}Entry struct {
	key      {{.KeyType}}
	value    {{.ValType}}
	occupied bool
}

// {{.MapName}} is an open-addressing hash map with {{.KeyType}} keys and {{.ValType}} values.
type {{.MapName}} struct {
	entries []{{.EntryStem}}Entry
	size    int
}

// New{{.MapName}} creates a new empty {{.MapName}} with default capacity.
func New{{.MapName}}() *{{.MapName}} {
	return New{{.MapName}}WithCapacity({{.EntryStem}}DefaultCapacity)
}

// New{{.MapName}}WithCapacity creates a new empty {{.MapName}} with the given initial capacity.
func New{{.MapName}}WithCapacity(capacity int) *{{.MapName}} {
	cap := nextPowerOfTwo{{.MapName}}(capacity)
	return &{{.MapName}}{
		entries: make([]{{.EntryStem}}Entry, cap),
		size:    0,
	}
}

// {{.MapName}}Of creates a new {{.MapName}} from key-value pairs.
func {{.MapName}}Of(pairs ...struct {
	Key   {{.KeyType}}
	Value {{.ValType}}
}) *{{.MapName}} {
	m := New{{.MapName}}WithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return m
}

// {{.MapName}}BulkLoad builds a {{.MapName}} from keys/values in a single pass,
// presizing the table once to fit len(keys) at the 0.75 load factor. keys[i] and
// values[i] form one entry, so the slices must have equal length (a mismatch is a
// programmer error and panics). The input need not be sorted.
//
// On a duplicate key it returns pump.ErrDuplicateKey unless policy is
// pump.IgnoreDuplicates, in which case the first value for a key is kept.
// The result is observably identical to the same entries inserted one-by-one
// with Put into a table of the same final capacity (same probe layout, same
// iteration order). The size is a hint: this constructor may grow if the source
// has more distinct keys than expected — use {{.MapName}}BulkLoadExact for the
// zero-rehash guarantee.
func {{.MapName}}BulkLoad(keys []{{.KeyType}}, values []{{.ValType}}, policy pump.DuplicatePolicy) (*{{.MapName}}, error) {
	if len(keys) != len(values) {
		panic("mapdb: {{.MapName}}BulkLoad: len(keys) != len(values)")
	}
	m := &{{.MapName}}{entries: make([]{{.EntryStem}}Entry, {{.MapName}}bulkCap(len(keys)))}
	if err := m.bulkInsert(keys, values, policy); err != nil {
		return nil, err
	}
	return m, nil
}

// {{.MapName}}BulkLoadExact is like {{.MapName}}BulkLoad but guarantees zero
// mid-load rehash: the table is sized for exactly n consumed entries. It returns
// pump.ErrTooManyElements if the source yields more than n entries, even when
// the extra entries are duplicate keys skipped by pump.IgnoreDuplicates. n must
// be non-negative (negative panics).
func {{.MapName}}BulkLoadExact(keys []{{.KeyType}}, values []{{.ValType}}, n int, policy pump.DuplicatePolicy) (*{{.MapName}}, error) {
	if len(keys) != len(values) {
		panic("mapdb: {{.MapName}}BulkLoadExact: len(keys) != len(values)")
	}
	if n < 0 {
		panic("mapdb: {{.MapName}}BulkLoadExact: negative n")
	}
	m := &{{.MapName}}{entries: make([]{{.EntryStem}}Entry, {{.MapName}}bulkCap(n))}
	if len(keys) > n {
		return nil, pump.ErrTooManyElements
	}
	for i := range keys {
		dup, err := m.bulkPut(keys[i], values[i], policy)
		if err != nil {
			return nil, err
		}
		_ = dup
	}
	return m, nil
}

// bulkInsert inserts every (key, value) pair into a presized table, growing if a
// duplicate-free run would cross the load factor (only possible in the hint path).
func (m *{{.MapName}}) bulkInsert(keys []{{.KeyType}}, values []{{.ValType}}, policy pump.DuplicatePolicy) error {
	for i := range keys {
		if m.needsResize() {
			m.resize()
		}
		if _, err := m.bulkPut(keys[i], values[i], policy); err != nil {
			return err
		}
	}
	return nil
}

// bulkPut inserts a single entry via the ordinary probe without a resize check
// (callers guarantee capacity). It applies the duplicate policy and reports
// whether the key was already present.
func (m *{{.MapName}}) bulkPut(key {{.KeyType}}, value {{.ValType}}, policy pump.DuplicatePolicy) (bool, error) {
	mask := len(m.entries) - 1
	idx := int(m.hashKey(key)) & mask
	for {
		if !m.entries[idx].occupied {
			m.entries[idx].key = key
			m.entries[idx].value = value
			m.entries[idx].occupied = true
			m.size++
			return false, nil
		}
		if {{if .KeyIsFloat}}{{.KeyBitsFn}}(m.entries[idx].key) == {{.KeyBitsFn}}(key){{else}}m.entries[idx].key == key{{end}} {
			if policy == pump.IgnoreDuplicates {
				return true, nil
			}
			return true, pump.ErrDuplicateKey
		}
		idx = (idx + 1) & mask
	}
}

// {{.MapName}}bulkCap returns the presized table capacity for n entries that
// avoids any mid-load rehash, using the shared zero-rehash formula. It widens to
// the family's default minimum so very small loads still get a usable table.
func {{.MapName}}bulkCap(n int) int {
	c := pump.HashCapacityFor(n)
	if c < {{.EntryStem}}DefaultCapacity {
		return {{.EntryStem}}DefaultCapacity
	}
	return c
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *{{.MapName}}) Put(key {{.KeyType}}, value {{.ValType}}) ({{.ValType}}, bool) {
	if m.needsResize() {
		m.resize()
	}
	cap := len(m.entries)
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask

	for {
		if !m.entries[idx].occupied {
			m.entries[idx].key = key
			m.entries[idx].value = value
			m.entries[idx].occupied = true
			m.size++
			return {{.ValZero}}, false
		}
		if {{if .KeyIsFloat}}{{.KeyBitsFn}}(m.entries[idx].key) == {{.KeyBitsFn}}(key){{else}}m.entries[idx].key == key{{end}} {
			old := m.entries[idx].value
			m.entries[idx].value = value
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// Get returns the value for the given key and true if found, or the zero value and false if not.
func (m *{{.MapName}}) Get(key {{.KeyType}}) ({{.ValType}}, bool) {
	cap := len(m.entries)
	if cap == 0 {
		return {{.ValZero}}, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask

	for {
		if !m.entries[idx].occupied {
			return {{.ValZero}}, false
		}
		if {{if .KeyIsFloat}}{{.KeyBitsFn}}(m.entries[idx].key) == {{.KeyBitsFn}}(key){{else}}m.entries[idx].key == key{{end}} {
			return m.entries[idx].value, true
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
	cap := len(m.entries)
	if cap == 0 {
		return {{.ValZero}}, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask

	for {
		if !m.entries[idx].occupied {
			return {{.ValZero}}, false
		}
		if {{if .KeyIsFloat}}{{.KeyBitsFn}}(m.entries[idx].key) == {{.KeyBitsFn}}(key){{else}}m.entries[idx].key == key{{end}} {
			old := m.entries[idx].value
			m.entries[idx].occupied = false
			m.entries[idx].key = {{.KeyZero}}
			m.entries[idx].value = {{.ValZero}}
			m.size--
			// Backward-shift deletion: the sibling of linear probing that
			// closes the hole by pulling each subsequent probed-past entry
			// one slot back until we reach an empty slot or an entry whose
			// preferred index equals its current index. This is distinct
			// from Robin Hood hashing (which is an insertion strategy).
			m.rehashFrom(idx, mask)
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
	for i := range m.entries {
		if m.entries[i].occupied && {{if .ValueIsFloat}}{{.ValBitsFn}}(m.entries[i].value) == {{.ValBitsFn}}(value){{else}}m.entries[i].value == value{{end}} {
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
	for i := range m.entries {
		m.entries[i] = {{.EntryStem}}Entry{}
	}
	m.size = 0
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *{{.MapName}}) All() iter.Seq2[{{.KeyType}}, {{.ValType}}] {
	return func(yield func({{.KeyType}}, {{.ValType}}) bool) {
		for i := range m.entries {
			if m.entries[i].occupied {
				if !yield(m.entries[i].key, m.entries[i].value) {
					return
				}
			}
		}
	}
}

// Keys returns an iter.Seq that yields all keys.
func (m *{{.MapName}}) Keys() iter.Seq[{{.KeyType}}] {
	return func(yield func({{.KeyType}}) bool) {
		for i := range m.entries {
			if m.entries[i].occupied {
				if !yield(m.entries[i].key) {
					return
				}
			}
		}
	}
}

// Values returns an iter.Seq that yields all values.
func (m *{{.MapName}}) Values() iter.Seq[{{.ValType}}] {
	return func(yield func({{.ValType}}) bool) {
		for i := range m.entries {
			if m.entries[i].occupied {
				if !yield(m.entries[i].value) {
					return
				}
			}
		}
	}
}

// ForEach calls the given function for each key-value pair.
func (m *{{.MapName}}) ForEach(f func({{.KeyType}}, {{.ValType}})) {
	for i := range m.entries {
		if m.entries[i].occupied {
			f(m.entries[i].key, m.entries[i].value)
		}
	}
}

// ForEachKey calls the given function for each key.
func (m *{{.MapName}}) ForEachKey(f func({{.KeyType}})) {
	for i := range m.entries {
		if m.entries[i].occupied {
			f(m.entries[i].key)
		}
	}
}

// ForEachValue calls the given function for each value.
func (m *{{.MapName}}) ForEachValue(f func({{.ValType}})) {
	for i := range m.entries {
		if m.entries[i].occupied {
			f(m.entries[i].value)
		}
	}
}

// Select returns a new map containing only the key-value pairs that satisfy the predicate.
func (m *{{.MapName}}) Select(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}} {
	result := New{{.MapName}}()
	for i := range m.entries {
		if m.entries[i].occupied && predicate(m.entries[i].key, m.entries[i].value) {
			result.Put(m.entries[i].key, m.entries[i].value)
		}
	}
	return result
}

// Reject returns a new map containing only the key-value pairs that do not satisfy the predicate.
func (m *{{.MapName}}) Reject(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}} {
	result := New{{.MapName}}()
	for i := range m.entries {
		if m.entries[i].occupied && !predicate(m.entries[i].key, m.entries[i].value) {
			result.Put(m.entries[i].key, m.entries[i].value)
		}
	}
	return result
}

// Detect returns the first key-value pair that satisfies the predicate, or zero values and false.
func (m *{{.MapName}}) Detect(predicate func({{.KeyType}}, {{.ValType}}) bool) ({{.KeyType}}, {{.ValType}}, bool) {
	for i := range m.entries {
		if m.entries[i].occupied && predicate(m.entries[i].key, m.entries[i].value) {
			return m.entries[i].key, m.entries[i].value, true
		}
	}
	return 0, 0, false
}

// AnySatisfy returns true if any key-value pair satisfies the predicate.
func (m *{{.MapName}}) AnySatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	for i := range m.entries {
		if m.entries[i].occupied && predicate(m.entries[i].key, m.entries[i].value) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all key-value pairs satisfy the predicate.
func (m *{{.MapName}}) AllSatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	for i := range m.entries {
		if m.entries[i].occupied && !predicate(m.entries[i].key, m.entries[i].value) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no key-value pair satisfies the predicate.
func (m *{{.MapName}}) NoneSatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	for i := range m.entries {
		if m.entries[i].occupied && predicate(m.entries[i].key, m.entries[i].value) {
			return false
		}
	}
	return true
}

// Count returns the number of key-value pairs that satisfy the predicate.
func (m *{{.MapName}}) Count(predicate func({{.KeyType}}, {{.ValType}}) bool) int {
	count := 0
	for i := range m.entries {
		if m.entries[i].occupied && predicate(m.entries[i].key, m.entries[i].value) {
			count++
		}
	}
	return count
}

// String returns a string representation of the map.
func (m *{{.MapName}}) String() string {
	if m.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for i := range m.entries {
		if m.entries[i].occupied {
			if !first {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%v: %v", m.entries[i].key, m.entries[i].value)
			first = false
		}
	}
	sb.WriteString("}")
	return sb.String()
}

// Equals returns true if the other map has the same key-value pairs.
func (m *{{.MapName}}) Equals(other *{{.MapName}}) bool {
	if m.size != other.size {
		return false
	}
	for i := range m.entries {
		if m.entries[i].occupied {
			v, ok := other.Get(m.entries[i].key)
			if !ok || !({{if .ValueIsFloat}}{{.ValBitsFn}}(v) == {{.ValBitsFn}}(m.entries[i].value){{else}}v == m.entries[i].value{{end}}) {
				return false
			}
		}
	}
	return true
}

// KeysToSlice returns all keys as a slice.
func (m *{{.MapName}}) KeysToSlice() []{{.KeyType}} {
	result := make([]{{.KeyType}}, 0, m.size)
	for i := range m.entries {
		if m.entries[i].occupied {
			result = append(result, m.entries[i].key)
		}
	}
	return result
}

// ValuesToSlice returns all values as a slice.
func (m *{{.MapName}}) ValuesToSlice() []{{.ValType}} {
	result := make([]{{.ValType}}, 0, m.size)
	for i := range m.entries {
		if m.entries[i].occupied {
			result = append(result, m.entries[i].value)
		}
	}
	return result
}

// ToImmutable returns an immutable copy of this map.
func (m *{{.MapName}}) ToImmutable() *Immutable{{.MapName}} {
	return Immutable{{.MapName}}From(m)
}

// InjectInto performs a left fold over all key-value pairs.
func (m *{{.MapName}}) InjectInto(initial {{.ValType}}, f func({{.ValType}}, {{.KeyType}}, {{.ValType}}) {{.ValType}}) {{.ValType}} {
	result := initial
	for i := range m.entries {
		if m.entries[i].occupied {
			result = f(result, m.entries[i].key, m.entries[i].value)
		}
	}
	return result
}

// AddToValue adds the given amount to the value for the key.
// If the key is not present, inserts it with the given amount as value.
// Returns the new value.
func (m *{{.MapName}}) AddToValue(key {{.KeyType}}, amount {{.ValType}}) {{.ValType}} {
	if v, ok := m.Get(key); ok {
		newVal := v + amount
		m.Put(key, newVal)
		return newVal
	}
	m.Put(key, amount)
	return amount
}

// UpdateValue updates the value for the key using the function.
// If key is absent, inserts initialValue first then applies the function.
// Returns the new value.
func (m *{{.MapName}}) UpdateValue(key {{.KeyType}}, initialValue {{.ValType}}, f func({{.ValType}}) {{.ValType}}) {{.ValType}} {
	if v, ok := m.Get(key); ok {
		newVal := f(v)
		m.Put(key, newVal)
		return newVal
	}
	newVal := f(initialValue)
	m.Put(key, newVal)
	return newVal
}

// PutReturning returns the map after putting the key-value pair (fluent API).
func (m *{{.MapName}}) PutReturning(key {{.KeyType}}, value {{.ValType}}) *{{.MapName}} {
	m.Put(key, value)
	return m
}

// RemoveKeyReturning returns the map after removing the key (fluent API).
func (m *{{.MapName}}) RemoveKeyReturning(key {{.KeyType}}) *{{.MapName}} {
	m.Remove(key)
	return m
}

// WithoutAllKeys removes all given keys (fluent API).
func (m *{{.MapName}}) WithoutAllKeys(keys []{{.KeyType}}) *{{.MapName}} {
	for _, k := range keys {
		m.Remove(k)
	}
	return m
}

// SumOfValues returns the sum of all values.
func (m *{{.MapName}}) SumOfValues() {{.ValType}} {
	var sum {{.ValType}}
	for i := range m.entries {
		if m.entries[i].occupied {
			sum += m.entries[i].value
		}
	}
	return sum
}

// Entry returns a handle for in-place check-and-modify operations on the
// given key. The handle is not thread-safe: external synchronisation (the
// Synchronized{{.MapName}} wrapper's Lock / RLock, or your own mutex) is required
// when multiple goroutines share the same underlying map. The name is
// modelled on Rust's std::collections::hash_map::Entry, not on Java's
// ConcurrentMap.compute; there is no internal locking, no CAS, and no
// atomicity guarantee across callback invocation.
func (m *{{.MapName}}) Entry(key {{.KeyType}}) {{.MapName}}Entry {
	return {{.MapName}}Entry{m: m, key: key}
}

// {{.MapName}}Entry provides in-place check-and-modify operations for a single
// key. Not thread-safe — see {{.MapName}}.Entry.
type {{.MapName}}Entry struct {
	m   *{{.MapName}}
	key {{.KeyType}}
}

// OrInsert inserts the default value if the key is absent, and returns the current value.
func (e {{.MapName}}Entry) OrInsert(defaultValue {{.ValType}}) {{.ValType}} {
	if v, ok := e.m.Get(e.key); ok {
		return v
	}
	e.m.Put(e.key, defaultValue)
	return defaultValue
}

// OrInsertWith inserts the value from the function if the key is absent,
// and returns the current value.
func (e {{.MapName}}Entry) OrInsertWith(f func() {{.ValType}}) {{.ValType}} {
	if v, ok := e.m.Get(e.key); ok {
		return v
	}
	val := f()
	e.m.Put(e.key, val)
	return val
}

// AndModify calls f with a pointer to the value if the key is present,
// and returns the entry for fluent chaining. If the key is absent, f is
// not called and the entry is returned unchanged.
//
// CAUTION: f must not call Put / OrInsert / OrInsertWith on the same map.
// Those calls may trigger a resize that reallocates the underlying
// entries slice, leaving f's pointer dangling into the old slice. To
// guard against silent data loss this path panics if it detects a
// resize happened during f — see the post-call check below.
func (e {{.MapName}}Entry) AndModify(f func(*{{.ValType}})) {{.MapName}}Entry {
	cap := len(e.m.entries)
	if cap == 0 {
		return e
	}
	mask := cap - 1
	idx := int(e.m.hashKey(e.key)) & mask
	for {
		if !e.m.entries[idx].occupied {
			return e
		}
		if {{if .KeyIsFloat}}{{.KeyBitsFn}}(e.m.entries[idx].key) == {{.KeyBitsFn}}(e.key){{else}}e.m.entries[idx].key == e.key{{end}} {
			// Detect backing-slice identity before and after the callback.
			// If the slice header changed (resize) or length changed (rehash),
			// the pointer we passed to f aliased the pre-resize storage and
			// the mutation is lost. Panic rather than silently dropping data.
			prevPtr := &e.m.entries[0]
			prevLen := len(e.m.entries)
			f(&e.m.entries[idx].value)
			if prevLen != len(e.m.entries) || prevPtr != &e.m.entries[0] {
				panic("{{.MapName}}Entry.AndModify: map was resized during callback — do not mutate the map from within AndModify")
			}
			return e
		}
		idx = (idx + 1) & mask
	}
}

func (m *{{.MapName}}) hashKey(key {{.KeyType}}) uint64 {
	h := {{.KeyHashExpr}} * 0x9E3779B97F4A7C15
	return h ^ (h >> 32)
}

func (m *{{.MapName}}) needsResize() bool {
	return (m.size+1)*4 >= len(m.entries)*3 // 0.75 load factor, integer math
}

func (m *{{.MapName}}) resize() {
	oldEntries := m.entries
	newCap := len(oldEntries) * 2
	if newCap == 0 {
		newCap = {{.EntryStem}}DefaultCapacity
	}
	m.entries = make([]{{.EntryStem}}Entry, newCap)
	m.size = 0

	for i := range oldEntries {
		if oldEntries[i].occupied {
			m.Put(oldEntries[i].key, oldEntries[i].value)
		}
	}
}

// rehashFrom fixes the invariant after a deletion using backward-shift.
func (m *{{.MapName}}) rehashFrom(deleted int, mask int) {
	c := len(m.entries)
	idx := (deleted + 1) & mask
	for m.entries[idx].occupied {
		ideal := int(m.hashKey(m.entries[idx].key)) & mask
		distCurrent := (idx - ideal + c) & mask
		distGap := (deleted - ideal + c) & mask
		if distCurrent > distGap {
			m.entries[deleted] = m.entries[idx]
			m.entries[idx] = {{.EntryStem}}Entry{}
			deleted = idx
		}
		idx = (idx + 1) & mask
		if idx == deleted {
			break
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

const immutableHashMapTmpl = genHeader + `package hashmap

import (
	"iter"
)

// Immutable{{.MapName}} is an immutable view of a {{.MapName}}.
// It exposes only read operations. Any attempt to modify requires
// creating a mutable copy first via ToMutable().
type Immutable{{.MapName}} struct {
	delegate *{{.MapName}}
}

// NewImmutable{{.MapName}} creates an immutable map from key-value pairs.
func NewImmutable{{.MapName}}(pairs ...struct {
	Key   {{.KeyType}}
	Value {{.ValType}}
}) *Immutable{{.MapName}} {
	m := New{{.MapName}}WithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return &Immutable{{.MapName}}{delegate: m}
}

// Immutable{{.MapName}}From creates an immutable copy of a mutable map.
func Immutable{{.MapName}}From(m *{{.MapName}}) *Immutable{{.MapName}} {
	copy := New{{.MapName}}WithCapacity(m.Len() * 2)
	m.ForEach(func(k {{.KeyType}}, v {{.ValType}}) {
		copy.Put(k, v)
	})
	return &Immutable{{.MapName}}{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *Immutable{{.MapName}}) Get(key {{.KeyType}}) ({{.ValType}}, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *Immutable{{.MapName}}) GetOrDefault(key {{.KeyType}}, defaultValue {{.ValType}}) {{.ValType}} {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *Immutable{{.MapName}}) ContainsKey(key {{.KeyType}}) bool {
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *Immutable{{.MapName}}) ContainsValue(value {{.ValType}}) bool {
	return m.delegate.ContainsValue(value)
}

// Len returns the number of elements. Use m.Len() == 0 to test for emptiness.
func (m *Immutable{{.MapName}}) Len() int {
	return m.delegate.Len()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *Immutable{{.MapName}}) All() iter.Seq2[{{.KeyType}}, {{.ValType}}] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *Immutable{{.MapName}}) Keys() iter.Seq[{{.KeyType}}] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *Immutable{{.MapName}}) Values() iter.Seq[{{.ValType}}] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *Immutable{{.MapName}}) ForEach(f func({{.KeyType}}, {{.ValType}})) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries that satisfy the predicate.
func (m *Immutable{{.MapName}}) Select(predicate func({{.KeyType}}, {{.ValType}}) bool) *Immutable{{.MapName}} {
	return &Immutable{{.MapName}}{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries that do not satisfy the predicate.
func (m *Immutable{{.MapName}}) Reject(predicate func({{.KeyType}}, {{.ValType}}) bool) *Immutable{{.MapName}} {
	return &Immutable{{.MapName}}{delegate: m.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *Immutable{{.MapName}}) AnySatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	return m.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *Immutable{{.MapName}}) AllSatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	return m.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *Immutable{{.MapName}}) NoneSatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	return m.delegate.NoneSatisfy(predicate)
}

// KeysToSlice returns all keys as a slice.
func (m *Immutable{{.MapName}}) KeysToSlice() []{{.KeyType}} {
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns all values as a slice.
func (m *Immutable{{.MapName}}) ValuesToSlice() []{{.ValType}} {
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *Immutable{{.MapName}}) String() string {
	return m.delegate.String()
}

// Equals returns true if the other immutable map has the same entries.
func (m *Immutable{{.MapName}}) Equals(other *Immutable{{.MapName}}) bool {
	return m.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this map.
func (m *Immutable{{.MapName}}) ToMutable() *{{.MapName}} {
	copy := New{{.MapName}}WithCapacity(m.Len() * 2)
	m.ForEach(func(k {{.KeyType}}, v {{.ValType}}) {
		copy.Put(k, v)
	})
	return copy
}
`

const synchronizedHashMapTmpl = genHeader + `package hashmap

import (
	"iter"
	"sync"
	"unsafe"
)

// Synchronized{{.MapName}} is a thread-safe wrapper around {{.MapName}}.
//
// Read methods hold RLock; writes hold Lock. Functional methods
// (ForEach, Select, AnySatisfy, …) snapshot (keys, values) under
// RLock and run the callback unlocked, so callbacks may freely call
// back into this wrapper.
//
// Methods whose signature takes a callback AND mutates (UpdateValue)
// hold the write lock while invoking the callback — the callback
// must not re-enter the wrapper in that case. This matches the
// Java EC synchronized-collection convention.
type Synchronized{{.MapName}} struct {
	delegate *{{.MapName}}
	mu       sync.RWMutex
}

// NewSynchronized{{.MapName}} wraps a mutable map with synchronization.
func NewSynchronized{{.MapName}}() *Synchronized{{.MapName}} {
	return &Synchronized{{.MapName}}{delegate: New{{.MapName}}()}
}

// NewSynchronized{{.MapName}}WithCapacity wraps a new map with the given initial capacity.
func NewSynchronized{{.MapName}}WithCapacity(capacity int) *Synchronized{{.MapName}} {
	return &Synchronized{{.MapName}}{delegate: New{{.MapName}}WithCapacity(capacity)}
}

// NewSynchronized{{.MapName}}From wraps an existing map with synchronization.
// The wrapper takes ownership — do not mutate the delegate directly.
func NewSynchronized{{.MapName}}From(m *{{.MapName}}) *Synchronized{{.MapName}} {
	return &Synchronized{{.MapName}}{delegate: m}
}

// snapshot returns (keys, values) slices in matching order, taken under RLock.
func (m *Synchronized{{.MapName}}) snapshot() (keys []{{.KeyType}}, values []{{.ValType}}) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.KeysToSlice(), m.delegate.ValuesToSlice()
}

// ── writes ────────────────────────────────────────────────────────────

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *Synchronized{{.MapName}}) Put(key {{.KeyType}}, value {{.ValType}}) ({{.ValType}}, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.Put(key, value)
}

// Remove deletes the entry for the given key. Returns the previous value and true if found.
func (m *Synchronized{{.MapName}}) Remove(key {{.KeyType}}) ({{.ValType}}, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.Remove(key)
}

// Clear removes all entries.
func (m *Synchronized{{.MapName}}) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.Clear()
}

// AddToValue increments the value for the given key by ` + "`amount`" + `,
// inserting it if absent. Holds the write lock; returns the new value.
func (m *Synchronized{{.MapName}}) AddToValue(key {{.KeyType}}, amount {{.ValType}}) {{.ValType}} {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.AddToValue(key, amount)
}

// UpdateValue applies f to the current (or initial) value under the
// write lock. The callback must not re-enter this wrapper — it will
// deadlock. Prefer Get + Put on caller side if re-entry is needed.
func (m *Synchronized{{.MapName}}) UpdateValue(key {{.KeyType}}, initial {{.ValType}}, f func({{.ValType}}) {{.ValType}}) {{.ValType}} {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.UpdateValue(key, initial, f)
}

// ── simple reads ──────────────────────────────────────────────────────

// Get returns the value for the given key and true if found.
func (m *Synchronized{{.MapName}}) Get(key {{.KeyType}}) ({{.ValType}}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *Synchronized{{.MapName}}) GetOrDefault(key {{.KeyType}}, defaultValue {{.ValType}}) {{.ValType}} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *Synchronized{{.MapName}}) ContainsKey(key {{.KeyType}}) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if any entry's value matches.
func (m *Synchronized{{.MapName}}) ContainsValue(value {{.ValType}}) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ContainsValue(value)
}

// Len returns the number of elements. Use m.Len() == 0 to test for emptiness.
func (m *Synchronized{{.MapName}}) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.Len()
}

// SumOfValues returns the sum of all values, under RLock.
func (m *Synchronized{{.MapName}}) SumOfValues() {{.ValType}} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.SumOfValues()
}

// KeysToSlice returns a copy of all keys.
func (m *Synchronized{{.MapName}}) KeysToSlice() []{{.KeyType}} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns a copy of all values.
func (m *Synchronized{{.MapName}}) ValuesToSlice() []{{.ValType}} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *Synchronized{{.MapName}}) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.String()
}

// ── iteration (snapshot-based) ────────────────────────────────────────

// All returns an iter.Seq2 over a snapshot of all key-value pairs.
// Iteration is lock-free.
func (m *Synchronized{{.MapName}}) All() iter.Seq2[{{.KeyType}}, {{.ValType}}] {
	keys, values := m.snapshot()
	return func(yield func({{.KeyType}}, {{.ValType}}) bool) {
		for i := range keys {
			if !yield(keys[i], values[i]) {
				return
			}
		}
	}
}

// Keys returns an iter.Seq over a snapshot of keys.
func (m *Synchronized{{.MapName}}) Keys() iter.Seq[{{.KeyType}}] {
	keys, _ := m.snapshot()
	return func(yield func({{.KeyType}}) bool) {
		for _, k := range keys {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iter.Seq over a snapshot of values.
func (m *Synchronized{{.MapName}}) Values() iter.Seq[{{.ValType}}] {
	_, values := m.snapshot()
	return func(yield func({{.ValType}}) bool) {
		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}

// ── functional (callback) methods over snapshot ──────────────────────

// ForEach iterates entries over a snapshot. Callback runs unlocked.
func (m *Synchronized{{.MapName}}) ForEach(f func({{.KeyType}}, {{.ValType}})) {
	keys, values := m.snapshot()
	for i := range keys {
		f(keys[i], values[i])
	}
}

// ForEachKey iterates keys over a snapshot. Callback runs unlocked.
func (m *Synchronized{{.MapName}}) ForEachKey(f func({{.KeyType}})) {
	keys, _ := m.snapshot()
	for _, k := range keys {
		f(k)
	}
}

// ForEachValue iterates values over a snapshot. Callback runs unlocked.
func (m *Synchronized{{.MapName}}) ForEachValue(f func({{.ValType}})) {
	_, values := m.snapshot()
	for _, v := range values {
		f(v)
	}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *Synchronized{{.MapName}}) AnySatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if every entry satisfies the predicate.
func (m *Synchronized{{.MapName}}) AllSatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if !predicate(keys[i], values[i]) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *Synchronized{{.MapName}}) NoneSatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return false
		}
	}
	return true
}

// Count returns the number of entries satisfying the predicate.
func (m *Synchronized{{.MapName}}) Count(predicate func({{.KeyType}}, {{.ValType}}) bool) int {
	keys, values := m.snapshot()
	n := 0
	for i := range keys {
		if predicate(keys[i], values[i]) {
			n++
		}
	}
	return n
}

// Detect returns any entry satisfying the predicate, or zero values and false.
func (m *Synchronized{{.MapName}}) Detect(predicate func({{.KeyType}}, {{.ValType}}) bool) ({{.KeyType}}, {{.ValType}}, bool) {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return keys[i], values[i], true
		}
	}
	var zeroK {{.KeyType}}
	var zeroV {{.ValType}}
	return zeroK, zeroV, false
}

// InjectInto folds entries into an accumulator, callback unlocked.
func (m *Synchronized{{.MapName}}) InjectInto(initial {{.ValType}}, f func({{.ValType}}, {{.KeyType}}, {{.ValType}}) {{.ValType}}) {{.ValType}} {
	keys, values := m.snapshot()
	acc := initial
	for i := range keys {
		acc = f(acc, keys[i], values[i])
	}
	return acc
}

// ── functional that return a new map ─────────────────────────────────

// Select returns a new (unsynchronized) map with entries satisfying predicate.
func (m *Synchronized{{.MapName}}) Select(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}} {
	keys, values := m.snapshot()
	result := New{{.MapName}}()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			result.Put(keys[i], values[i])
		}
	}
	return result
}

// Reject returns a new (unsynchronized) map with entries NOT satisfying predicate.
func (m *Synchronized{{.MapName}}) Reject(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}} {
	keys, values := m.snapshot()
	result := New{{.MapName}}()
	for i := range keys {
		if !predicate(keys[i], values[i]) {
			result.Put(keys[i], values[i])
		}
	}
	return result
}

// ── fluent mutators ───────────────────────────────────────────────────

func (m *Synchronized{{.MapName}}) PutReturning(key {{.KeyType}}, value {{.ValType}}) *Synchronized{{.MapName}} {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.PutReturning(key, value)
	return m
}

func (m *Synchronized{{.MapName}}) RemoveKeyReturning(key {{.KeyType}}) *Synchronized{{.MapName}} {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.RemoveKeyReturning(key)
	return m
}

// WithoutAllKeys is variadic for caller convenience; internally the
// slice is passed straight through since the underlying method already
// accepts a slice.
func (m *Synchronized{{.MapName}}) WithoutAllKeys(keys ...{{.KeyType}}) *Synchronized{{.MapName}} {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.WithoutAllKeys(keys)
	return m
}

// ── conversions & equals ──────────────────────────────────────────────

func (m *Synchronized{{.MapName}}) ToImmutable() *Immutable{{.MapName}} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ToImmutable()
}

// Equals compares by contents. Locks acquired in pointer-address
// order to prevent A.Equals(B) / B.Equals(A) deadlocks.
func (m *Synchronized{{.MapName}}) Equals(other *Synchronized{{.MapName}}) bool {
	if m == other {
		m.mu.RLock()
		defer m.mu.RUnlock()
		return m.delegate.Equals(other.delegate)
	}
	first, second := m, other
	if uintptr(unsafe.Pointer(m)) > uintptr(unsafe.Pointer(other)) {
		first, second = other, m
	}
	first.mu.RLock()
	defer first.mu.RUnlock()
	second.mu.RLock()
	defer second.mu.RUnlock()
	return m.delegate.Equals(other.delegate)
}

// Entry is NOT wrapped — the Entry API on the mutable map is designed
// for lock-free fast paths and would require returning a
// delegate-bound handle, which would race with other callers through
// the wrapper. If you need atomic check-and-modify under the synchronized
// wrapper, use UpdateValue or take the wrapper's lock externally and
// call into the delegate directly via a helper.
`

const hashBiMapTmpl = genHeader + `package hashmap

import (
	"fmt"
	"iter"
{{- if .KeyIsFloat}}
	"math"
{{- end}}
	"strings"

	"github.com/mapdb/mapdb-golang/pump"
)

// {{.MapName}}BiMap is a bidirectional map with {{.KeyType}} keys and {{.ValType}} values.
// Both key-to-value and value-to-key lookups are O(1).
type {{.MapName}}BiMap struct {
	forward *{{.MapName}}
	reverse *{{.RevName}}
}

// New{{.MapName}}BiMap creates a new empty {{.MapName}}BiMap with default capacity.
func New{{.MapName}}BiMap() *{{.MapName}}BiMap {
	return &{{.MapName}}BiMap{
		forward: New{{.MapName}}(),
		reverse: New{{.RevName}}(),
	}
}

// New{{.MapName}}BiMapWithCapacity creates a new empty {{.MapName}}BiMap with the given initial capacity.
func New{{.MapName}}BiMapWithCapacity(capacity int) *{{.MapName}}BiMap {
	return &{{.MapName}}BiMap{
		forward: New{{.MapName}}WithCapacity(capacity),
		reverse: New{{.RevName}}WithCapacity(capacity),
	}
}

// {{.MapName}}BiMapBulkLoad builds a {{.MapName}}BiMap from keys/values in a
// single pass, presizing both inner tables to fit len(keys) at the 0.75 load
// factor. keys[i] and values[i] form one entry (a length mismatch panics). A
// BiMap requires a bijection, so the duplicate policy DOES NOT apply: a
// duplicate key returns pump.ErrDuplicateKey and a duplicate value returns
// pump.ErrDuplicateValue, ALWAYS — even under IgnoreDuplicates and even for a
// fully identical (key, value) pair (a repeated key breaks the single-pass
// bijection build). The policy parameter is accepted for signature symmetry with
// the other bulk loaders but is intentionally ignored. The input need not be
// sorted; the result is identical to the same pairs inserted one-by-one with Put.
func {{.MapName}}BiMapBulkLoad(keys []{{.KeyType}}, values []{{.ValType}}, policy pump.DuplicatePolicy) (*{{.MapName}}BiMap, error) {
	if len(keys) != len(values) {
		panic("mapdb: {{.MapName}}BiMapBulkLoad: len(keys) != len(values)")
	}
	m := &{{.MapName}}BiMap{
		forward: New{{.MapName}}WithCapacity({{.MapName}}bulkCap(len(keys))),
		reverse: New{{.RevName}}WithCapacity({{.RevName}}bulkCap(len(values))),
	}
	// policy is intentionally ignored: a BiMap requires a bijection, so any
	// duplicate key or value is always an error (even an identical pair, which
	// repeats the key and breaks the single-pass bijection build).
	_ = policy
	for i := range keys {
		key, value := keys[i], values[i]
		_, hasKey := m.forward.Get(key)
		_, hasVal := m.reverse.Get(value)
		if hasKey {
			return nil, pump.ErrDuplicateKey
		}
		if hasVal {
			return nil, pump.ErrDuplicateValue
		}
		m.forward.Put(key, value)
		m.reverse.Put(value, key)
	}
	return m, nil
}

// Put inserts or updates a key-value pair in both directions.
// If the key already existed, the old value mapping is removed from the reverse map.
// If the value already existed as a value for a different key, that old key mapping is removed.
// Returns the previous value and true if the key existed.
func (m *{{.MapName}}BiMap) Put(key {{.KeyType}}, value {{.ValType}}) ({{.ValType}}, bool) {
	// If this value is already mapped to a different key, remove that old key->value pair
	if oldKey, ok := m.reverse.Get(value); ok {
		if !({{if .KeyIsFloat}}{{.KeyBitsFn}}(oldKey) == {{.KeyBitsFn}}(key){{else}}oldKey == key{{end}}) {
			m.forward.Remove(oldKey)
		}
	}

	// If this key already has a value, remove the old value->key reverse mapping
	oldVal, existed := m.forward.Get(key)
	if existed {
		m.reverse.Remove(oldVal)
	}

	m.forward.Put(key, value)
	m.reverse.Put(value, key)
	return oldVal, existed
}

// Get returns the value for the given key and true if found, or the zero value and false if not.
func (m *{{.MapName}}BiMap) Get(key {{.KeyType}}) ({{.ValType}}, bool) {
	return m.forward.Get(key)
}

// GetKey returns the key for the given value and true if found, or the zero value and false if not.
func (m *{{.MapName}}BiMap) GetKey(value {{.ValType}}) ({{.KeyType}}, bool) {
	return m.reverse.Get(value)
}

// Remove deletes the entry for the given key from both directions.
// Returns the previous value and true if the key existed.
func (m *{{.MapName}}BiMap) Remove(key {{.KeyType}}) ({{.ValType}}, bool) {
	oldVal, existed := m.forward.Remove(key)
	if existed {
		m.reverse.Remove(oldVal)
	}
	return oldVal, existed
}

// RemoveValue deletes the entry for the given value from both directions.
// Returns the previous key and true if the value existed.
func (m *{{.MapName}}BiMap) RemoveValue(value {{.ValType}}) ({{.KeyType}}, bool) {
	oldKey, existed := m.reverse.Remove(value)
	if existed {
		m.forward.Remove(oldKey)
	}
	return oldKey, existed
}

// ContainsKey returns true if the map contains the given key.
func (m *{{.MapName}}BiMap) ContainsKey(key {{.KeyType}}) bool {
	return m.forward.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *{{.MapName}}BiMap) ContainsValue(value {{.ValType}}) bool {
	return m.reverse.ContainsKey(value)
}

// Len returns the number of elements. Use m.Len() == 0 to test for emptiness.
func (m *{{.MapName}}BiMap) Len() int {
	return m.forward.Len()
}

// Clear removes all entries from both directions.
func (m *{{.MapName}}BiMap) Clear() {
	m.forward.Clear()
	m.reverse.Clear()
}

// ForEach calls the given function for each key-value pair.
func (m *{{.MapName}}BiMap) ForEach(f func({{.KeyType}}, {{.ValType}})) {
	m.forward.ForEach(f)
}

// Keys returns an iter.Seq that yields all keys.
func (m *{{.MapName}}BiMap) Keys() iter.Seq[{{.KeyType}}] {
	return m.forward.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *{{.MapName}}BiMap) Values() iter.Seq[{{.ValType}}] {
	return m.forward.Values()
}

// Inverse returns a new {{.RevName}}BiMap with keys and values swapped.
func (m *{{.MapName}}BiMap) Inverse() *{{.RevName}}BiMap {
	result := New{{.RevName}}BiMap()
	m.forward.ForEach(func(k {{.KeyType}}, v {{.ValType}}) {
		result.Put(v, k)
	})
	return result
}

// String returns a string representation of the bi-map.
func (m *{{.MapName}}BiMap) String() string {
	if m.forward.Len() == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	m.forward.ForEach(func(k {{.KeyType}}, v {{.ValType}}) {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v: %v", k, v)
		first = false
	})
	sb.WriteString("}")
	return sb.String()
}

// Equals returns true if the other bi-map has the same key-value pairs.
func (m *{{.MapName}}BiMap) Equals(other *{{.MapName}}BiMap) bool {
	return m.forward.Equals(other.forward)
}
`

const objectKeyHashMapTmpl = genHeader + `package hashmap

import (
	"fmt"
	"iter"
	"strings"
)

const (
	{{.EntryStem}}DefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
)

// {{.MapName}} is an open-addressing hash map with generic comparable keys and {{.PrimType}} values.
// The value type is specialized to avoid boxing overhead.
type {{.MapName}}[K comparable] struct {
	keys     []K
	values   []{{.PrimType}}
	occupied []bool
	size     int
}

// New{{.MapName}} creates a new empty {{.MapName}} with default capacity.
func New{{.MapName}}[K comparable]() *{{.MapName}}[K] {
	return New{{.MapName}}WithCapacity[K]({{.EntryStem}}DefaultCapacity)
}

// New{{.MapName}}WithCapacity creates a new empty {{.MapName}} with the given initial capacity.
func New{{.MapName}}WithCapacity[K comparable](capacity int) *{{.MapName}}[K] {
	cap := nextPowerOfTwo{{.MapName}}(capacity)
	return &{{.MapName}}[K]{
		keys:     make([]K, cap),
		values:   make([]{{.PrimType}}, cap),
		occupied: make([]bool, cap),
		size:     0,
	}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *{{.MapName}}[K]) Put(key K, value {{.PrimType}}) ({{.PrimType}}, bool) {
	if m.needsResize() {
		m.resize()
	}
	cap := len(m.keys)
	mask := cap - 1
	idx := int(hashComparable(key)) & mask

	for {
		if !m.occupied[idx] {
			m.keys[idx] = key
			m.values[idx] = value
			m.occupied[idx] = true
			m.size++
			return {{.PrimZero}}, false
		}
		if m.keys[idx] == key {
			old := m.values[idx]
			m.values[idx] = value
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// Get returns the value for the given key and true if found, or the zero value and false if not.
func (m *{{.MapName}}[K]) Get(key K) ({{.PrimType}}, bool) {
	cap := len(m.keys)
	if cap == 0 {
		return {{.PrimZero}}, false
	}
	mask := cap - 1
	idx := int(hashComparable(key)) & mask

	for {
		if !m.occupied[idx] {
			return {{.PrimZero}}, false
		}
		if m.keys[idx] == key {
			return m.values[idx], true
		}
		idx = (idx + 1) & mask
	}
}

// GetOrDefault returns the value for the given key if present, or the default value otherwise.
func (m *{{.MapName}}[K]) GetOrDefault(key K, defaultValue {{.PrimType}}) {{.PrimType}} {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// Remove deletes the entry for the given key. Returns the previous value and true if the key existed.
func (m *{{.MapName}}[K]) Remove(key K) ({{.PrimType}}, bool) {
	cap := len(m.keys)
	if cap == 0 {
		return {{.PrimZero}}, false
	}
	mask := cap - 1
	idx := int(hashComparable(key)) & mask

	for {
		if !m.occupied[idx] {
			return {{.PrimZero}}, false
		}
		if m.keys[idx] == key {
			old := m.values[idx]
			m.occupied[idx] = false
			var zeroK K
			m.keys[idx] = zeroK
			m.values[idx] = {{.PrimZero}}
			m.size--
			m.rehashFrom{{.MapName}}(idx, mask)
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// ContainsKey returns true if the map contains the given key.
func (m *{{.MapName}}[K]) ContainsKey(key K) bool {
	_, ok := m.Get(key)
	return ok
}

// Len returns the number of elements. Use m.Len() == 0 to test for emptiness.
func (m *{{.MapName}}[K]) Len() int {
	return m.size
}

// Clear removes all entries from the map.
func (m *{{.MapName}}[K]) Clear() {
	var zeroK K
	for i := range m.occupied {
		m.occupied[i] = false
		m.keys[i] = zeroK
		m.values[i] = {{.PrimZero}}
	}
	m.size = 0
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *{{.MapName}}[K]) All() iter.Seq2[K, {{.PrimType}}] {
	return func(yield func(K, {{.PrimType}}) bool) {
		for i := range m.occupied {
			if m.occupied[i] {
				if !yield(m.keys[i], m.values[i]) {
					return
				}
			}
		}
	}
}

// Keys returns an iter.Seq that yields all keys.
func (m *{{.MapName}}[K]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for i := range m.occupied {
			if m.occupied[i] {
				if !yield(m.keys[i]) {
					return
				}
			}
		}
	}
}

// Values returns an iter.Seq that yields all values.
func (m *{{.MapName}}[K]) Values() iter.Seq[{{.PrimType}}] {
	return func(yield func({{.PrimType}}) bool) {
		for i := range m.occupied {
			if m.occupied[i] {
				if !yield(m.values[i]) {
					return
				}
			}
		}
	}
}

// ForEach calls the given function for each key-value pair.
func (m *{{.MapName}}[K]) ForEach(f func(K, {{.PrimType}})) {
	for i := range m.occupied {
		if m.occupied[i] {
			f(m.keys[i], m.values[i])
		}
	}
}

// Select returns a new map containing only entries that satisfy the predicate.
func (m *{{.MapName}}[K]) Select(predicate func(K, {{.PrimType}}) bool) *{{.MapName}}[K] {
	result := New{{.MapName}}[K]()
	for i := range m.occupied {
		if m.occupied[i] && predicate(m.keys[i], m.values[i]) {
			result.Put(m.keys[i], m.values[i])
		}
	}
	return result
}

// Reject returns a new map containing only entries that do not satisfy the predicate.
func (m *{{.MapName}}[K]) Reject(predicate func(K, {{.PrimType}}) bool) *{{.MapName}}[K] {
	result := New{{.MapName}}[K]()
	for i := range m.occupied {
		if m.occupied[i] && !predicate(m.keys[i], m.values[i]) {
			result.Put(m.keys[i], m.values[i])
		}
	}
	return result
}

// String returns a string representation of the map.
func (m *{{.MapName}}[K]) String() string {
	if m.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for i := range m.occupied {
		if m.occupied[i] {
			if !first {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%v: %v", m.keys[i], m.values[i])
			first = false
		}
	}
	sb.WriteString("}")
	return sb.String()
}

func (m *{{.MapName}}[K]) needsResize() bool {
	return (m.size+1)*4 >= len(m.keys)*3 // 0.75 load factor, integer math
}

func (m *{{.MapName}}[K]) resize() {
	oldKeys := m.keys
	oldValues := m.values
	oldOccupied := m.occupied
	newCap := len(oldKeys) * 2
	if newCap == 0 {
		newCap = {{.EntryStem}}DefaultCapacity
	}
	m.keys = make([]K, newCap)
	m.values = make([]{{.PrimType}}, newCap)
	m.occupied = make([]bool, newCap)
	m.size = 0

	for i := range oldOccupied {
		if oldOccupied[i] {
			m.Put(oldKeys[i], oldValues[i])
		}
	}
}

func (m *{{.MapName}}[K]) rehashFrom{{.MapName}}(deleted int, mask int) {
	idx := (deleted + 1) & mask
	for m.occupied[idx] {
		ideal := int(hashComparable(m.keys[idx])) & mask
		if (idx-ideal+len(m.keys))&mask > (idx-deleted+len(m.keys))&mask {
		} else {
			m.keys[deleted] = m.keys[idx]
			m.values[deleted] = m.values[idx]
			m.occupied[deleted] = true
			m.occupied[idx] = false
			var zeroK K
			m.keys[idx] = zeroK
			m.values[idx] = {{.PrimZero}}
			deleted = idx
		}
		idx = (idx + 1) & mask
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

const immutableObjectKeyHashMapTmpl = genHeader + `package hashmap

import (
	"iter"
)

// Immutable{{.MapName}} is an immutable view of an {{.MapName}}.
type Immutable{{.MapName}}[K comparable] struct {
	delegate *{{.MapName}}[K]
}

// NewImmutable{{.MapName}} creates an immutable object-{{.PrimType}} map by copying entries from a mutable map.
func NewImmutable{{.MapName}}From[K comparable](m *{{.MapName}}[K]) *Immutable{{.MapName}}[K] {
	copy := New{{.MapName}}WithCapacity[K](m.Len() * 2)
	m.ForEach(func(k K, v {{.PrimType}}) {
		copy.Put(k, v)
	})
	return &Immutable{{.MapName}}[K]{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *Immutable{{.MapName}}[K]) Get(key K) ({{.PrimType}}, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *Immutable{{.MapName}}[K]) GetOrDefault(key K, defaultValue {{.PrimType}}) {{.PrimType}} {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *Immutable{{.MapName}}[K]) ContainsKey(key K) bool {
	return m.delegate.ContainsKey(key)
}

// Len returns the number of elements. Use m.Len() == 0 to test for emptiness.
func (m *Immutable{{.MapName}}[K]) Len() int {
	return m.delegate.Len()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *Immutable{{.MapName}}[K]) All() iter.Seq2[K, {{.PrimType}}] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *Immutable{{.MapName}}[K]) Keys() iter.Seq[K] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *Immutable{{.MapName}}[K]) Values() iter.Seq[{{.PrimType}}] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *Immutable{{.MapName}}[K]) ForEach(f func(K, {{.PrimType}})) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries satisfying the predicate.
func (m *Immutable{{.MapName}}[K]) Select(predicate func(K, {{.PrimType}}) bool) *Immutable{{.MapName}}[K] {
	return &Immutable{{.MapName}}[K]{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries not satisfying the predicate.
func (m *Immutable{{.MapName}}[K]) Reject(predicate func(K, {{.PrimType}}) bool) *Immutable{{.MapName}}[K] {
	return &Immutable{{.MapName}}[K]{delegate: m.delegate.Reject(predicate)}
}

// String returns a string representation.
func (m *Immutable{{.MapName}}[K]) String() string {
	return m.delegate.String()
}

// ToMutable returns a mutable copy of this map.
func (m *Immutable{{.MapName}}[K]) ToMutable() *{{.MapName}}[K] {
	copy := New{{.MapName}}WithCapacity[K](m.Len() * 2)
	m.ForEach(func(k K, v {{.PrimType}}) {
		copy.Put(k, v)
	})
	return copy
}
`

const objectValueHashMapTmpl = genHeader + `package hashmap

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
)

// {{.MapName}} is an open-addressing hash map with {{.PrimType}} keys and generic values.
// The key type is specialized to avoid boxing overhead.
type {{.MapName}}[V any] struct {
	keys     []{{.PrimType}}
	values   []V
	occupied []bool
	size     int
}

// New{{.MapName}} creates a new empty {{.MapName}} with default capacity.
func New{{.MapName}}[V any]() *{{.MapName}}[V] {
	return New{{.MapName}}WithCapacity[V]({{.EntryStem}}DefaultCapacity)
}

// New{{.MapName}}WithCapacity creates a new empty {{.MapName}} with the given initial capacity.
func New{{.MapName}}WithCapacity[V any](capacity int) *{{.MapName}}[V] {
	cap := nextPowerOfTwo{{.MapName}}(capacity)
	return &{{.MapName}}[V]{
		keys:     make([]{{.PrimType}}, cap),
		values:   make([]V, cap),
		occupied: make([]bool, cap),
		size:     0,
	}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *{{.MapName}}[V]) Put(key {{.PrimType}}, value V) (V, bool) {
	if m.needsResize() {
		m.resize()
	}
	cap := len(m.keys)
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask

	for {
		if !m.occupied[idx] {
			m.keys[idx] = key
			m.values[idx] = value
			m.occupied[idx] = true
			m.size++
			var zero V
			return zero, false
		}
		if {{if .KeyIsFloat}}{{.KeyBitsFn}}(m.keys[idx]) == {{.KeyBitsFn}}(key){{else}}m.keys[idx] == key{{end}} {
			old := m.values[idx]
			m.values[idx] = value
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// Get returns the value for the given key and true if found, or the zero value and false if not.
func (m *{{.MapName}}[V]) Get(key {{.PrimType}}) (V, bool) {
	cap := len(m.keys)
	if cap == 0 {
		var zero V
		return zero, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask

	for {
		if !m.occupied[idx] {
			var zero V
			return zero, false
		}
		if {{if .KeyIsFloat}}{{.KeyBitsFn}}(m.keys[idx]) == {{.KeyBitsFn}}(key){{else}}m.keys[idx] == key{{end}} {
			return m.values[idx], true
		}
		idx = (idx + 1) & mask
	}
}

// GetOrDefault returns the value for the given key if present, or the default value otherwise.
func (m *{{.MapName}}[V]) GetOrDefault(key {{.PrimType}}, defaultValue V) V {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// Remove deletes the entry for the given key. Returns the previous value and true if the key existed.
func (m *{{.MapName}}[V]) Remove(key {{.PrimType}}) (V, bool) {
	cap := len(m.keys)
	if cap == 0 {
		var zero V
		return zero, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask

	for {
		if !m.occupied[idx] {
			var zero V
			return zero, false
		}
		if {{if .KeyIsFloat}}{{.KeyBitsFn}}(m.keys[idx]) == {{.KeyBitsFn}}(key){{else}}m.keys[idx] == key{{end}} {
			old := m.values[idx]
			m.occupied[idx] = false
			m.keys[idx] = {{.PrimZero}}
			var zeroV V
			m.values[idx] = zeroV
			m.size--
			m.rehashFrom(idx, mask)
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// ContainsKey returns true if the map contains the given key.
func (m *{{.MapName}}[V]) ContainsKey(key {{.PrimType}}) bool {
	_, ok := m.Get(key)
	return ok
}

// Len returns the number of elements. Use m.Len() == 0 to test for emptiness.
func (m *{{.MapName}}[V]) Len() int {
	return m.size
}

// Clear removes all entries from the map.
func (m *{{.MapName}}[V]) Clear() {
	var zeroV V
	for i := range m.occupied {
		m.occupied[i] = false
		m.keys[i] = {{.PrimZero}}
		m.values[i] = zeroV
	}
	m.size = 0
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *{{.MapName}}[V]) All() iter.Seq2[{{.PrimType}}, V] {
	return func(yield func({{.PrimType}}, V) bool) {
		for i := range m.occupied {
			if m.occupied[i] {
				if !yield(m.keys[i], m.values[i]) {
					return
				}
			}
		}
	}
}

// Keys returns an iter.Seq that yields all keys.
func (m *{{.MapName}}[V]) Keys() iter.Seq[{{.PrimType}}] {
	return func(yield func({{.PrimType}}) bool) {
		for i := range m.occupied {
			if m.occupied[i] {
				if !yield(m.keys[i]) {
					return
				}
			}
		}
	}
}

// Values returns an iter.Seq that yields all values.
func (m *{{.MapName}}[V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for i := range m.occupied {
			if m.occupied[i] {
				if !yield(m.values[i]) {
					return
				}
			}
		}
	}
}

// ForEach calls the given function for each key-value pair.
func (m *{{.MapName}}[V]) ForEach(f func({{.PrimType}}, V)) {
	for i := range m.occupied {
		if m.occupied[i] {
			f(m.keys[i], m.values[i])
		}
	}
}

// Select returns a new map containing only entries that satisfy the predicate.
func (m *{{.MapName}}[V]) Select(predicate func({{.PrimType}}, V) bool) *{{.MapName}}[V] {
	result := New{{.MapName}}[V]()
	for i := range m.occupied {
		if m.occupied[i] && predicate(m.keys[i], m.values[i]) {
			result.Put(m.keys[i], m.values[i])
		}
	}
	return result
}

// Reject returns a new map containing only entries that do not satisfy the predicate.
func (m *{{.MapName}}[V]) Reject(predicate func({{.PrimType}}, V) bool) *{{.MapName}}[V] {
	result := New{{.MapName}}[V]()
	for i := range m.occupied {
		if m.occupied[i] && !predicate(m.keys[i], m.values[i]) {
			result.Put(m.keys[i], m.values[i])
		}
	}
	return result
}

// String returns a string representation of the map.
func (m *{{.MapName}}[V]) String() string {
	if m.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for i := range m.occupied {
		if m.occupied[i] {
			if !first {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%v: %v", m.keys[i], m.values[i])
			first = false
		}
	}
	sb.WriteString("}")
	return sb.String()
}

func (m *{{.MapName}}[V]) hashKey(key {{.PrimType}}) uint64 {
	h := {{.KeyHashExpr}} * 0x9E3779B97F4A7C15
	return h ^ (h >> 32)
}

func (m *{{.MapName}}[V]) needsResize() bool {
	return (m.size+1)*4 >= len(m.keys)*3 // 0.75 load factor, integer math
}

func (m *{{.MapName}}[V]) resize() {
	oldKeys := m.keys
	oldValues := m.values
	oldOccupied := m.occupied
	newCap := len(oldKeys) * 2
	if newCap == 0 {
		newCap = {{.EntryStem}}DefaultCapacity
	}
	m.keys = make([]{{.PrimType}}, newCap)
	m.values = make([]V, newCap)
	m.occupied = make([]bool, newCap)
	m.size = 0

	for i := range oldOccupied {
		if oldOccupied[i] {
			m.Put(oldKeys[i], oldValues[i])
		}
	}
}

func (m *{{.MapName}}[V]) rehashFrom(deleted int, mask int) {
	idx := (deleted + 1) & mask
	for m.occupied[idx] {
		ideal := int(m.hashKey(m.keys[idx])) & mask
		if (idx-ideal+len(m.keys))&mask > (idx-deleted+len(m.keys))&mask {
		} else {
			m.keys[deleted] = m.keys[idx]
			m.values[deleted] = m.values[idx]
			m.occupied[deleted] = true
			m.occupied[idx] = false
			m.keys[idx] = {{.PrimZero}}
			var zeroV V
			m.values[idx] = zeroV
			deleted = idx
		}
		idx = (idx + 1) & mask
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

const immutableObjectValueHashMapTmpl = genHeader + `package hashmap

import (
	"iter"
)

// Immutable{{.MapName}} is an immutable view of a {{.MapName}}.
type Immutable{{.MapName}}[V any] struct {
	delegate *{{.MapName}}[V]
}

// NewImmutable{{.MapName}}From creates an immutable {{.PrimType}}-object map by copying entries from a mutable map.
func NewImmutable{{.MapName}}From[V any](m *{{.MapName}}[V]) *Immutable{{.MapName}}[V] {
	copy := New{{.MapName}}WithCapacity[V](m.Len() * 2)
	m.ForEach(func(k {{.PrimType}}, v V) {
		copy.Put(k, v)
	})
	return &Immutable{{.MapName}}[V]{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *Immutable{{.MapName}}[V]) Get(key {{.PrimType}}) (V, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *Immutable{{.MapName}}[V]) GetOrDefault(key {{.PrimType}}, defaultValue V) V {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *Immutable{{.MapName}}[V]) ContainsKey(key {{.PrimType}}) bool {
	return m.delegate.ContainsKey(key)
}

// Len returns the number of elements. Use m.Len() == 0 to test for emptiness.
func (m *Immutable{{.MapName}}[V]) Len() int {
	return m.delegate.Len()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *Immutable{{.MapName}}[V]) All() iter.Seq2[{{.PrimType}}, V] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *Immutable{{.MapName}}[V]) Keys() iter.Seq[{{.PrimType}}] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *Immutable{{.MapName}}[V]) Values() iter.Seq[V] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *Immutable{{.MapName}}[V]) ForEach(f func({{.PrimType}}, V)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries satisfying the predicate.
func (m *Immutable{{.MapName}}[V]) Select(predicate func({{.PrimType}}, V) bool) *Immutable{{.MapName}}[V] {
	return &Immutable{{.MapName}}[V]{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries not satisfying the predicate.
func (m *Immutable{{.MapName}}[V]) Reject(predicate func({{.PrimType}}, V) bool) *Immutable{{.MapName}}[V] {
	return &Immutable{{.MapName}}[V]{delegate: m.delegate.Reject(predicate)}
}

// String returns a string representation.
func (m *Immutable{{.MapName}}[V]) String() string {
	return m.delegate.String()
}

// ToMutable returns a mutable copy of this map.
func (m *Immutable{{.MapName}}[V]) ToMutable() *{{.MapName}}[V] {
	copy := New{{.MapName}}WithCapacity[V](m.Len() * 2)
	m.ForEach(func(k {{.PrimType}}, v V) {
		copy.Put(k, v)
	})
	return copy
}
`
