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
// hashing/equality and/or value equality. "unsafe" is needed iff KeyIsFloat
// (the float hashKey reinterpret). The synchronized wrapper always imports
// "unsafe" (pointer-address lock ordering in Equals), independent of types.
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
	// unsafe-based float hashKey body; it also (with ValueIsFloat) gates the
	// math import and (alone) gates the unsafe import.
	KeyIsFloat bool
	// KeyBitsFn is the float bit-pattern function for key equality
	// (math.Float32bits / math.Float64bits). Float keys only.
	KeyBitsFn string
	// KeyHashExpr is the inner operand of the golden-ratio multiply in
	// hashKey: h := <KeyHashExpr> * 0x9E3779B97F4A7C15. Captured per key type
	// because the integer/char/float reinterpretations differ (int32 alone
	// double-casts through uint32; floats reinterpret via unsafe).
	KeyHashExpr string

	// ValueIsFloat selects bit-pattern value equality in ContainsValue/Equals.
	ValueIsFloat bool
	// ValBitsFn is the float bit-pattern function for value equality. Float
	// values only.
	ValBitsFn string

	// NeedsMath / NeedsUnsafe drive the base file's import block.
	NeedsMath   bool
	NeedsUnsafe bool

	// MapName is the combined identifier stem, e.g. Int32Float32 (used for the
	// exported type Int32Float32HashMap); MapSnake is the file-name stem,
	// e.g. int32_float32.
	MapName  string // Int32Float32
	MapSnake string // int32_float32
	// EntryStem is the lower-camel struct stem, e.g. int32Float32 (used for
	// int32Float32HashMapEntry / the Entry type Int32Float32Entry uses MapName).
	EntryStem string // int32Float32
}

// keyHashExpr returns the per-key-type inner operand of the golden-ratio
// multiply in hashKey, reproduced verbatim from the hand-written originals.
func keyHashExpr(p Primitive) string {
	switch {
	case p.IsFloating && p.ByteSize == 4:
		return "uint64(*(*uint32)(unsafe.Pointer(&key)))"
	case p.IsFloating && p.ByteSize == 8:
		return "*(*uint64)(unsafe.Pointer(&key))"
	case p.GoType == "int32":
		return "uint64(uint32(key))"
	default:
		return "uint64(key)"
	}
}

// genHashMap writes the 7×7 = 49 prim×prim hash map sources for each of the
// base, immutable, and synchronized variants (147 files) into the current
// working directory. Invoked from hashmap/ via go:generate.
func genHashMap() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	base := template.Must(template.New("hm-base").Parse(hashMapTmpl))
	immutable := template.Must(template.New("hm-immutable").Parse(immutableHashMapTmpl))
	synchronized := template.Must(template.New("hm-sync").Parse(synchronizedHashMapTmpl))

	write := func(name string, tmpl *template.Template, data hmData) error {
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

				NeedsMath:   k.IsFloating || v.IsFloating,
				NeedsUnsafe: k.IsFloating,

				MapName:   k.Name + v.Name,
				MapSnake:  k.SnakeName + "_" + v.SnakeName,
				EntryStem: lowerFirst(k.Name) + v.Name,
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
{{- if .NeedsUnsafe}}
	"unsafe"
{{- end}}
)

const (
	{{.EntryStem}}HashMapDefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
)

// {{.EntryStem}}HashMapEntry holds a single slot in the hash map for cache locality.
type {{.EntryStem}}HashMapEntry struct {
	key      {{.KeyType}}
	value    {{.ValType}}
	occupied bool
}

// {{.MapName}}HashMap is an open-addressing hash map with {{.KeyType}} keys and {{.ValType}} values.
type {{.MapName}}HashMap struct {
	entries []{{.EntryStem}}HashMapEntry
	size    int
}

// New{{.MapName}}HashMap creates a new empty {{.MapName}}HashMap with default capacity.
func New{{.MapName}}HashMap() *{{.MapName}}HashMap {
	return New{{.MapName}}HashMapWithCapacity({{.EntryStem}}HashMapDefaultCapacity)
}

// New{{.MapName}}HashMapWithCapacity creates a new empty {{.MapName}}HashMap with the given initial capacity.
func New{{.MapName}}HashMapWithCapacity(capacity int) *{{.MapName}}HashMap {
	cap := nextPowerOfTwo{{.MapName}}HashMap(capacity)
	return &{{.MapName}}HashMap{
		entries: make([]{{.EntryStem}}HashMapEntry, cap),
		size:    0,
	}
}

// {{.MapName}}HashMapOf creates a new {{.MapName}}HashMap from key-value pairs.
func {{.MapName}}HashMapOf(pairs ...struct {
	Key   {{.KeyType}}
	Value {{.ValType}}
}) *{{.MapName}}HashMap {
	m := New{{.MapName}}HashMapWithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return m
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *{{.MapName}}HashMap) Put(key {{.KeyType}}, value {{.ValType}}) ({{.ValType}}, bool) {
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
func (m *{{.MapName}}HashMap) Get(key {{.KeyType}}) ({{.ValType}}, bool) {
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
func (m *{{.MapName}}HashMap) GetOrDefault(key {{.KeyType}}, defaultValue {{.ValType}}) {{.ValType}} {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// Remove deletes the entry for the given key. Returns the previous value and true if the key existed.
func (m *{{.MapName}}HashMap) Remove(key {{.KeyType}}) ({{.ValType}}, bool) {
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
func (m *{{.MapName}}HashMap) ContainsKey(key {{.KeyType}}) bool {
	_, ok := m.Get(key)
	return ok
}

// ContainsValue returns true if the map contains the given value.
func (m *{{.MapName}}HashMap) ContainsValue(value {{.ValType}}) bool {
	for i := range m.entries {
		if m.entries[i].occupied && {{if .ValueIsFloat}}{{.ValBitsFn}}(m.entries[i].value) == {{.ValBitsFn}}(value){{else}}m.entries[i].value == value{{end}} {
			return true
		}
	}
	return false
}

// Size returns the number of key-value pairs in the map.
func (m *{{.MapName}}HashMap) Size() int {
	return m.size
}

// IsEmpty returns true if the map contains no entries.
func (m *{{.MapName}}HashMap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the map.
func (m *{{.MapName}}HashMap) Clear() {
	for i := range m.entries {
		m.entries[i] = {{.EntryStem}}HashMapEntry{}
	}
	m.size = 0
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *{{.MapName}}HashMap) All() iter.Seq2[{{.KeyType}}, {{.ValType}}] {
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
func (m *{{.MapName}}HashMap) Keys() iter.Seq[{{.KeyType}}] {
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
func (m *{{.MapName}}HashMap) Values() iter.Seq[{{.ValType}}] {
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
func (m *{{.MapName}}HashMap) ForEach(f func({{.KeyType}}, {{.ValType}})) {
	for i := range m.entries {
		if m.entries[i].occupied {
			f(m.entries[i].key, m.entries[i].value)
		}
	}
}

// ForEachKey calls the given function for each key.
func (m *{{.MapName}}HashMap) ForEachKey(f func({{.KeyType}})) {
	for i := range m.entries {
		if m.entries[i].occupied {
			f(m.entries[i].key)
		}
	}
}

// ForEachValue calls the given function for each value.
func (m *{{.MapName}}HashMap) ForEachValue(f func({{.ValType}})) {
	for i := range m.entries {
		if m.entries[i].occupied {
			f(m.entries[i].value)
		}
	}
}

// Select returns a new map containing only the key-value pairs that satisfy the predicate.
func (m *{{.MapName}}HashMap) Select(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}}HashMap {
	result := New{{.MapName}}HashMap()
	for i := range m.entries {
		if m.entries[i].occupied && predicate(m.entries[i].key, m.entries[i].value) {
			result.Put(m.entries[i].key, m.entries[i].value)
		}
	}
	return result
}

// Reject returns a new map containing only the key-value pairs that do not satisfy the predicate.
func (m *{{.MapName}}HashMap) Reject(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}}HashMap {
	result := New{{.MapName}}HashMap()
	for i := range m.entries {
		if m.entries[i].occupied && !predicate(m.entries[i].key, m.entries[i].value) {
			result.Put(m.entries[i].key, m.entries[i].value)
		}
	}
	return result
}

// Detect returns the first key-value pair that satisfies the predicate, or zero values and false.
func (m *{{.MapName}}HashMap) Detect(predicate func({{.KeyType}}, {{.ValType}}) bool) ({{.KeyType}}, {{.ValType}}, bool) {
	for i := range m.entries {
		if m.entries[i].occupied && predicate(m.entries[i].key, m.entries[i].value) {
			return m.entries[i].key, m.entries[i].value, true
		}
	}
	return 0, 0, false
}

// AnySatisfy returns true if any key-value pair satisfies the predicate.
func (m *{{.MapName}}HashMap) AnySatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	for i := range m.entries {
		if m.entries[i].occupied && predicate(m.entries[i].key, m.entries[i].value) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all key-value pairs satisfy the predicate.
func (m *{{.MapName}}HashMap) AllSatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	for i := range m.entries {
		if m.entries[i].occupied && !predicate(m.entries[i].key, m.entries[i].value) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no key-value pair satisfies the predicate.
func (m *{{.MapName}}HashMap) NoneSatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	for i := range m.entries {
		if m.entries[i].occupied && predicate(m.entries[i].key, m.entries[i].value) {
			return false
		}
	}
	return true
}

// Count returns the number of key-value pairs that satisfy the predicate.
func (m *{{.MapName}}HashMap) Count(predicate func({{.KeyType}}, {{.ValType}}) bool) int {
	count := 0
	for i := range m.entries {
		if m.entries[i].occupied && predicate(m.entries[i].key, m.entries[i].value) {
			count++
		}
	}
	return count
}

// String returns a string representation of the map.
func (m *{{.MapName}}HashMap) String() string {
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
func (m *{{.MapName}}HashMap) Equals(other *{{.MapName}}HashMap) bool {
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
func (m *{{.MapName}}HashMap) KeysToSlice() []{{.KeyType}} {
	result := make([]{{.KeyType}}, 0, m.size)
	for i := range m.entries {
		if m.entries[i].occupied {
			result = append(result, m.entries[i].key)
		}
	}
	return result
}

// ValuesToSlice returns all values as a slice.
func (m *{{.MapName}}HashMap) ValuesToSlice() []{{.ValType}} {
	result := make([]{{.ValType}}, 0, m.size)
	for i := range m.entries {
		if m.entries[i].occupied {
			result = append(result, m.entries[i].value)
		}
	}
	return result
}

// ToImmutable returns an immutable copy of this map.
func (m *{{.MapName}}HashMap) ToImmutable() *Immutable{{.MapName}}HashMap {
	return Immutable{{.MapName}}HashMapFrom(m)
}

// InjectInto performs a left fold over all key-value pairs.
func (m *{{.MapName}}HashMap) InjectInto(initial {{.ValType}}, f func({{.ValType}}, {{.KeyType}}, {{.ValType}}) {{.ValType}}) {{.ValType}} {
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
func (m *{{.MapName}}HashMap) AddToValue(key {{.KeyType}}, amount {{.ValType}}) {{.ValType}} {
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
func (m *{{.MapName}}HashMap) UpdateValue(key {{.KeyType}}, initialValue {{.ValType}}, f func({{.ValType}}) {{.ValType}}) {{.ValType}} {
	if v, ok := m.Get(key); ok {
		newVal := f(v)
		m.Put(key, newVal)
		return newVal
	}
	newVal := f(initialValue)
	m.Put(key, newVal)
	return newVal
}

// WithKeyValue returns the map after putting the key-value pair (fluent API).
func (m *{{.MapName}}HashMap) WithKeyValue(key {{.KeyType}}, value {{.ValType}}) *{{.MapName}}HashMap {
	m.Put(key, value)
	return m
}

// WithoutKey returns the map after removing the key (fluent API).
func (m *{{.MapName}}HashMap) WithoutKey(key {{.KeyType}}) *{{.MapName}}HashMap {
	m.Remove(key)
	return m
}

// WithoutAllKeys removes all given keys (fluent API).
func (m *{{.MapName}}HashMap) WithoutAllKeys(keys []{{.KeyType}}) *{{.MapName}}HashMap {
	for _, k := range keys {
		m.Remove(k)
	}
	return m
}

// SumOfValues returns the sum of all values.
func (m *{{.MapName}}HashMap) SumOfValues() {{.ValType}} {
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
// Synchronized{{.MapName}}HashMap wrapper's Lock / RLock, or your own mutex) is required
// when multiple goroutines share the same underlying map. The name is
// modelled on Rust's std::collections::hash_map::Entry, not on Java's
// ConcurrentMap.compute; there is no internal locking, no CAS, and no
// atomicity guarantee across callback invocation.
func (m *{{.MapName}}HashMap) Entry(key {{.KeyType}}) {{.MapName}}Entry {
	return {{.MapName}}Entry{m: m, key: key}
}

// {{.MapName}}Entry provides in-place check-and-modify operations for a single
// key. Not thread-safe — see {{.MapName}}HashMap.Entry.
type {{.MapName}}Entry struct {
	m   *{{.MapName}}HashMap
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

func (m *{{.MapName}}HashMap) hashKey(key {{.KeyType}}) uint64 {
	return func() uint64 { h := {{.KeyHashExpr}} * 0x9E3779B97F4A7C15; return h ^ (h >> 32) }()
}

func (m *{{.MapName}}HashMap) needsResize() bool {
	return (m.size+1)*4 >= len(m.entries)*3 // 0.75 load factor, integer math
}

func (m *{{.MapName}}HashMap) resize() {
	oldEntries := m.entries
	newCap := len(oldEntries) * 2
	if newCap == 0 {
		newCap = {{.EntryStem}}HashMapDefaultCapacity
	}
	m.entries = make([]{{.EntryStem}}HashMapEntry, newCap)
	m.size = 0

	for i := range oldEntries {
		if oldEntries[i].occupied {
			m.Put(oldEntries[i].key, oldEntries[i].value)
		}
	}
}

// rehashFrom fixes the invariant after a deletion using backward-shift.
func (m *{{.MapName}}HashMap) rehashFrom(deleted int, mask int) {
	c := len(m.entries)
	idx := (deleted + 1) & mask
	for m.entries[idx].occupied {
		ideal := int(m.hashKey(m.entries[idx].key)) & mask
		distCurrent := (idx - ideal + c) & mask
		distGap := (deleted - ideal + c) & mask
		if distCurrent > distGap {
			m.entries[deleted] = m.entries[idx]
			m.entries[idx] = {{.EntryStem}}HashMapEntry{}
			deleted = idx
		}
		idx = (idx + 1) & mask
		if idx == deleted {
			break
		}
	}
}

func nextPowerOfTwo{{.MapName}}HashMap(n int) int {
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

// Immutable{{.MapName}}HashMap is an immutable view of a {{.MapName}}HashMap.
// It exposes only read operations. Any attempt to modify requires
// creating a mutable copy first via ToMutable().
type Immutable{{.MapName}}HashMap struct {
	delegate *{{.MapName}}HashMap
}

// NewImmutable{{.MapName}}HashMap creates an immutable map from key-value pairs.
func NewImmutable{{.MapName}}HashMap(pairs ...struct {
	Key   {{.KeyType}}
	Value {{.ValType}}
}) *Immutable{{.MapName}}HashMap {
	m := New{{.MapName}}HashMapWithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return &Immutable{{.MapName}}HashMap{delegate: m}
}

// Immutable{{.MapName}}HashMapFrom creates an immutable copy of a mutable map.
func Immutable{{.MapName}}HashMapFrom(m *{{.MapName}}HashMap) *Immutable{{.MapName}}HashMap {
	copy := New{{.MapName}}HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k {{.KeyType}}, v {{.ValType}}) {
		copy.Put(k, v)
	})
	return &Immutable{{.MapName}}HashMap{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *Immutable{{.MapName}}HashMap) Get(key {{.KeyType}}) ({{.ValType}}, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *Immutable{{.MapName}}HashMap) GetOrDefault(key {{.KeyType}}, defaultValue {{.ValType}}) {{.ValType}} {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *Immutable{{.MapName}}HashMap) ContainsKey(key {{.KeyType}}) bool {
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *Immutable{{.MapName}}HashMap) ContainsValue(value {{.ValType}}) bool {
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *Immutable{{.MapName}}HashMap) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *Immutable{{.MapName}}HashMap) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *Immutable{{.MapName}}HashMap) All() iter.Seq2[{{.KeyType}}, {{.ValType}}] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *Immutable{{.MapName}}HashMap) Keys() iter.Seq[{{.KeyType}}] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *Immutable{{.MapName}}HashMap) Values() iter.Seq[{{.ValType}}] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *Immutable{{.MapName}}HashMap) ForEach(f func({{.KeyType}}, {{.ValType}})) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries that satisfy the predicate.
func (m *Immutable{{.MapName}}HashMap) Select(predicate func({{.KeyType}}, {{.ValType}}) bool) *Immutable{{.MapName}}HashMap {
	return &Immutable{{.MapName}}HashMap{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries that do not satisfy the predicate.
func (m *Immutable{{.MapName}}HashMap) Reject(predicate func({{.KeyType}}, {{.ValType}}) bool) *Immutable{{.MapName}}HashMap {
	return &Immutable{{.MapName}}HashMap{delegate: m.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *Immutable{{.MapName}}HashMap) AnySatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	return m.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *Immutable{{.MapName}}HashMap) AllSatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	return m.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *Immutable{{.MapName}}HashMap) NoneSatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	return m.delegate.NoneSatisfy(predicate)
}

// KeysToSlice returns all keys as a slice.
func (m *Immutable{{.MapName}}HashMap) KeysToSlice() []{{.KeyType}} {
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns all values as a slice.
func (m *Immutable{{.MapName}}HashMap) ValuesToSlice() []{{.ValType}} {
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *Immutable{{.MapName}}HashMap) String() string {
	return m.delegate.String()
}

// Equals returns true if the other immutable map has the same entries.
func (m *Immutable{{.MapName}}HashMap) Equals(other *Immutable{{.MapName}}HashMap) bool {
	return m.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this map.
func (m *Immutable{{.MapName}}HashMap) ToMutable() *{{.MapName}}HashMap {
	copy := New{{.MapName}}HashMapWithCapacity(m.Size() * 2)
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

// Synchronized{{.MapName}}HashMap is a thread-safe wrapper around {{.MapName}}HashMap.
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
type Synchronized{{.MapName}}HashMap struct {
	delegate *{{.MapName}}HashMap
	mu       sync.RWMutex
}

// NewSynchronized{{.MapName}}HashMap wraps a mutable map with synchronization.
func NewSynchronized{{.MapName}}HashMap() *Synchronized{{.MapName}}HashMap {
	return &Synchronized{{.MapName}}HashMap{delegate: New{{.MapName}}HashMap()}
}

// NewSynchronized{{.MapName}}HashMapWithCapacity wraps a new map with the given initial capacity.
func NewSynchronized{{.MapName}}HashMapWithCapacity(capacity int) *Synchronized{{.MapName}}HashMap {
	return &Synchronized{{.MapName}}HashMap{delegate: New{{.MapName}}HashMapWithCapacity(capacity)}
}

// NewSynchronized{{.MapName}}HashMapFrom wraps an existing map with synchronization.
// The wrapper takes ownership — do not mutate the delegate directly.
func NewSynchronized{{.MapName}}HashMapFrom(m *{{.MapName}}HashMap) *Synchronized{{.MapName}}HashMap {
	return &Synchronized{{.MapName}}HashMap{delegate: m}
}

// snapshot returns (keys, values) slices in matching order, taken under RLock.
func (m *Synchronized{{.MapName}}HashMap) snapshot() (keys []{{.KeyType}}, values []{{.ValType}}) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.KeysToSlice(), m.delegate.ValuesToSlice()
}

// ── writes ────────────────────────────────────────────────────────────

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *Synchronized{{.MapName}}HashMap) Put(key {{.KeyType}}, value {{.ValType}}) ({{.ValType}}, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.Put(key, value)
}

// Remove deletes the entry for the given key. Returns the previous value and true if found.
func (m *Synchronized{{.MapName}}HashMap) Remove(key {{.KeyType}}) ({{.ValType}}, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.Remove(key)
}

// Clear removes all entries.
func (m *Synchronized{{.MapName}}HashMap) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.Clear()
}

// AddToValue increments the value for the given key by ` + "`amount`" + `,
// inserting it if absent. Holds the write lock; returns the new value.
func (m *Synchronized{{.MapName}}HashMap) AddToValue(key {{.KeyType}}, amount {{.ValType}}) {{.ValType}} {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.AddToValue(key, amount)
}

// UpdateValue applies f to the current (or initial) value under the
// write lock. The callback must not re-enter this wrapper — it will
// deadlock. Prefer Get + Put on caller side if re-entry is needed.
func (m *Synchronized{{.MapName}}HashMap) UpdateValue(key {{.KeyType}}, initial {{.ValType}}, f func({{.ValType}}) {{.ValType}}) {{.ValType}} {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.UpdateValue(key, initial, f)
}

// ── simple reads ──────────────────────────────────────────────────────

// Get returns the value for the given key and true if found.
func (m *Synchronized{{.MapName}}HashMap) Get(key {{.KeyType}}) ({{.ValType}}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *Synchronized{{.MapName}}HashMap) GetOrDefault(key {{.KeyType}}, defaultValue {{.ValType}}) {{.ValType}} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *Synchronized{{.MapName}}HashMap) ContainsKey(key {{.KeyType}}) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if any entry's value matches.
func (m *Synchronized{{.MapName}}HashMap) ContainsValue(value {{.ValType}}) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *Synchronized{{.MapName}}HashMap) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *Synchronized{{.MapName}}HashMap) IsEmpty() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.IsEmpty()
}

// SumOfValues returns the sum of all values, under RLock.
func (m *Synchronized{{.MapName}}HashMap) SumOfValues() {{.ValType}} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.SumOfValues()
}

// KeysToSlice returns a copy of all keys.
func (m *Synchronized{{.MapName}}HashMap) KeysToSlice() []{{.KeyType}} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns a copy of all values.
func (m *Synchronized{{.MapName}}HashMap) ValuesToSlice() []{{.ValType}} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *Synchronized{{.MapName}}HashMap) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.String()
}

// ── iteration (snapshot-based) ────────────────────────────────────────

// All returns an iter.Seq2 over a snapshot of all key-value pairs.
// Iteration is lock-free.
func (m *Synchronized{{.MapName}}HashMap) All() iter.Seq2[{{.KeyType}}, {{.ValType}}] {
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
func (m *Synchronized{{.MapName}}HashMap) Keys() iter.Seq[{{.KeyType}}] {
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
func (m *Synchronized{{.MapName}}HashMap) Values() iter.Seq[{{.ValType}}] {
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
func (m *Synchronized{{.MapName}}HashMap) ForEach(f func({{.KeyType}}, {{.ValType}})) {
	keys, values := m.snapshot()
	for i := range keys {
		f(keys[i], values[i])
	}
}

// ForEachKey iterates keys over a snapshot. Callback runs unlocked.
func (m *Synchronized{{.MapName}}HashMap) ForEachKey(f func({{.KeyType}})) {
	keys, _ := m.snapshot()
	for _, k := range keys {
		f(k)
	}
}

// ForEachValue iterates values over a snapshot. Callback runs unlocked.
func (m *Synchronized{{.MapName}}HashMap) ForEachValue(f func({{.ValType}})) {
	_, values := m.snapshot()
	for _, v := range values {
		f(v)
	}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *Synchronized{{.MapName}}HashMap) AnySatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if every entry satisfies the predicate.
func (m *Synchronized{{.MapName}}HashMap) AllSatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if !predicate(keys[i], values[i]) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *Synchronized{{.MapName}}HashMap) NoneSatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return false
		}
	}
	return true
}

// Count returns the number of entries satisfying the predicate.
func (m *Synchronized{{.MapName}}HashMap) Count(predicate func({{.KeyType}}, {{.ValType}}) bool) int {
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
func (m *Synchronized{{.MapName}}HashMap) Detect(predicate func({{.KeyType}}, {{.ValType}}) bool) ({{.KeyType}}, {{.ValType}}, bool) {
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
func (m *Synchronized{{.MapName}}HashMap) InjectInto(initial {{.ValType}}, f func({{.ValType}}, {{.KeyType}}, {{.ValType}}) {{.ValType}}) {{.ValType}} {
	keys, values := m.snapshot()
	acc := initial
	for i := range keys {
		acc = f(acc, keys[i], values[i])
	}
	return acc
}

// ── functional that return a new map ─────────────────────────────────

// Select returns a new (unsynchronized) map with entries satisfying predicate.
func (m *Synchronized{{.MapName}}HashMap) Select(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}}HashMap {
	keys, values := m.snapshot()
	result := New{{.MapName}}HashMap()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			result.Put(keys[i], values[i])
		}
	}
	return result
}

// Reject returns a new (unsynchronized) map with entries NOT satisfying predicate.
func (m *Synchronized{{.MapName}}HashMap) Reject(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}}HashMap {
	keys, values := m.snapshot()
	result := New{{.MapName}}HashMap()
	for i := range keys {
		if !predicate(keys[i], values[i]) {
			result.Put(keys[i], values[i])
		}
	}
	return result
}

// ── fluent mutators ───────────────────────────────────────────────────

func (m *Synchronized{{.MapName}}HashMap) WithKeyValue(key {{.KeyType}}, value {{.ValType}}) *Synchronized{{.MapName}}HashMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.WithKeyValue(key, value)
	return m
}

func (m *Synchronized{{.MapName}}HashMap) WithoutKey(key {{.KeyType}}) *Synchronized{{.MapName}}HashMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.WithoutKey(key)
	return m
}

// WithoutAllKeys is variadic for caller convenience; internally the
// slice is passed straight through since the underlying method already
// accepts a slice.
func (m *Synchronized{{.MapName}}HashMap) WithoutAllKeys(keys ...{{.KeyType}}) *Synchronized{{.MapName}}HashMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.WithoutAllKeys(keys)
	return m
}

// ── conversions & equals ──────────────────────────────────────────────

func (m *Synchronized{{.MapName}}HashMap) ToImmutable() *Immutable{{.MapName}}HashMap {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ToImmutable()
}

// Equals compares by contents. Locks acquired in pointer-address
// order to prevent A.Equals(B) / B.Equals(A) deadlocks.
func (m *Synchronized{{.MapName}}HashMap) Equals(other *Synchronized{{.MapName}}HashMap) bool {
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
