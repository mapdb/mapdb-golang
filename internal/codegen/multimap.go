package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

// mmData is the per key/value view the multimap templates iterate over.
//
// The multimap family maps each key to a collection of values. There are two
// variants — list multimap (appends, keeps duplicates) and set multimap (dedups
// values) — BASE only (no immutable/synchronized/object wrappers). The backing
// store is a Go BUILTIN map, NOT the production OpenHashMap.
//
// TWO independent float axes drive the only type-dependent logic:
//
// KEY axis (KeyIsFloat): integer/char keys are used directly as the builtin map
// key — `data map[<KeyType>][]<ValType>`. FLOAT keys are STRUCTURALLY different:
// Go builtin maps cannot safely key on floats (NaN is unreachable, ±0 collapse),
// so float-keyed multimaps key the data map on the IEEE bit pattern
// (`data map[<KeyBitsType>][]<ValType>`, KeyBitsType = uintN) and carry a side
// map `keys map[<KeyBitsType>]<KeyType>` to recover the original float key during
// iteration. Every key access goes through `math.Float{32,64}bits(key)`.
//
// VALUE axis (ValueIsFloat): float values compare by bit pattern instead of ==.
// This affects Equals and ContainsKeyValue in BOTH templates, and additionally
// the Put dedup in the SET template (the list template appends in Put with no
// dedup scan, but still compares values in ContainsKeyValue/Equals).
//
// Imports: "math" is needed iff (KeyIsFloat || ValueIsFloat).
type mmData struct {
	KeyName  string // Int32, Float32, Char (key identifier stem)
	KeyType  string // int32, float32, uint16 (Go key type)
	KeySnake string // int32, float32, char (key file-name stem)

	ValName  string // Int32, Float32, Char (value identifier stem)
	ValType  string // int32, float32, uint16 (Go value type)
	ValSnake string // int32, float32, char (value file-name stem)

	// KeyIsFloat selects the two-map (data + keys) float-key structure and gates
	// the bit-pattern key accesses.
	KeyIsFloat  bool
	KeyBitsFn   string // math.Float32bits / math.Float64bits (float keys only)
	KeyBitsType string // uint32 / uint64 (the builtin-map key type, float keys only)
	KeyIntType  string // int32 / int64 (signed bit type for the float total order)
	KeyBitShift int    // 31 / 63 (sign-bit position for the float total order)

	// ValueIsFloat selects bit-pattern value equality (Equals always; set Put /
	// ContainsKeyValue dedup additionally).
	ValueIsFloat bool
	ValBitsFn    string // math.Float32bits / math.Float64bits (float values only)

	// NeedsMath drives the import block.
	NeedsMath bool

	// MapName is the combined identifier stem, e.g. Int32Float32; MapSnake is the
	// file-name stem, e.g. int32_float32.
	MapName  string // Int32Float32
	MapSnake string // int32_float32
}

// genMultimap writes the 7×7 = 49 prim×prim list multimap sources
// (*_list_multimap.go) and the 49 prim×prim set multimap sources
// (*_set_multimap.go), 98 files total, into the current working directory.
// Invoked from multimap/ via go:generate. The hand-written shared generic
// multimap.go is NOT touched.
func genMultimap() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	list := parse("mm-list", listMultimapTmpl)
	set := parse("mm-set", setMultimapTmpl)
	keyCmp := parse("mm-keycmp", multimapKeyCmpTmpl)

	write := func(name string, tmpl *template.Template, data mmData) error {
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

	floatBitsFn := func(p Primitive) string {
		if p.ByteSize == 8 {
			return "math.Float64bits"
		}
		return "math.Float32bits"
	}

	prims := Primitives()
	for _, k := range prims {
		for _, v := range prims {
			data := mmData{
				KeyName:  k.Name,
				KeyType:  k.GoType,
				KeySnake: k.SnakeName,

				ValName:  v.Name,
				ValType:  v.GoType,
				ValSnake: v.SnakeName,

				KeyIsFloat:   k.IsFloating,
				ValueIsFloat: v.IsFloating,

				NeedsMath: k.IsFloating || v.IsFloating,

				MapName:  k.Name + v.Name,
				MapSnake: k.SnakeName + "_" + v.SnakeName,
			}
			if k.IsFloating {
				data.KeyBitsFn = floatBitsFn(k)
				data.KeyBitsType = fmt.Sprintf("uint%d", k.ByteSize*8)
			}
			if v.IsFloating {
				data.ValBitsFn = floatBitsFn(v)
			}

			if err := write(data.MapSnake+"_list_multimap.go", list, data); err != nil {
				return err
			}
			if err := write(data.MapSnake+"_set_multimap.go", set, data); err != nil {
				return err
			}
		}
		// One key comparator per key primitive (depends only on the key type),
		// used by the FromSorted bulk-load validators.
		kd := mmData{
			KeyName:    k.Name,
			KeyType:    k.GoType,
			KeySnake:   k.SnakeName,
			KeyIsFloat: k.IsFloating,
			NeedsMath:  k.IsFloating,
		}
		if k.IsFloating {
			kd.KeyBitsFn = floatBitsFn(k)
			kd.KeyBitsType = fmt.Sprintf("uint%d", k.ByteSize*8)
			kd.KeyIntType = fmt.Sprintf("int%d", k.ByteSize*8)
			kd.KeyBitShift = k.ByteSize*8 - 1
		}
		if err := write(k.SnakeName+"_key_cmp.go", keyCmp, kd); err != nil {
			return err
		}
	}

	return nil
}

const multimapKeyCmpTmpl = genHeader + `package multimap

{{- if .KeyIsFloat}}

import "math"
{{- end}}

// cmpKey{{.KeyName}} is the three-way ordering for {{.KeyType}} multimap keys used
// by the FromSorted bulk-load validators.
{{- if .KeyIsFloat}} For float keys it is the IEEE-754 total order (matching the
// tree families), so NaN sorts to the top and ±0 stay distinct.
{{- end}}
func cmpKey{{.KeyName}}(a, b {{.KeyType}}) int {
{{- if .KeyIsFloat}}
	ai := {{.KeyIntType}}({{.KeyBitsFn}}(a))
	bi := {{.KeyIntType}}({{.KeyBitsFn}}(b))
	ai ^= {{.KeyIntType}}({{.KeyBitsType}}(ai>>{{.KeyBitShift}}) >> 1)
	bi ^= {{.KeyIntType}}({{.KeyBitsType}}(bi>>{{.KeyBitShift}}) >> 1)
	switch {
	case ai < bi:
		return -1
	case ai > bi:
		return 1
	default:
		return 0
	}
{{- else}}
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
{{- end}}
}
`

const listMultimapTmpl = genHeader + `package multimap

import (
{{- if .NeedsMath}}
	"math"
{{- end}}
	"fmt"
	"strings"

	"github.com/mapdb/mapdb-golang/pump"
)

// {{.MapName}}List is a list multimap from {{.KeyType}} keys to {{.ValType}} values.
// Each key maps to a slice of values, preserving insertion order per key.
type {{.MapName}}List struct {
	data map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}}
{{- if .KeyIsFloat}}
	keys map[{{.KeyBitsType}}]{{.KeyType}}
{{- end}}
	size int
}

// New{{.MapName}}List creates a new empty {{.MapName}}List.
func New{{.MapName}}List() *{{.MapName}}List {
	return &{{.MapName}}List{
		data: make(map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}}),
{{- if .KeyIsFloat}}
		keys: make(map[{{.KeyBitsType}}]{{.KeyType}}),
{{- end}}
		size: 0,
	}
}

// {{.MapName}}ListBulkLoad builds a {{.MapName}}List from keys/values in a single
// pass, presizing the backing map for the input. keys[i] and values[i] form one
// pair (a length mismatch panics). The input need not be sorted; values are
// appended in input order, exactly as repeated Put. Duplicate keys are the normal
// grouping case and the duplicate policy does not apply (a list multimap keeps
// every value).
func {{.MapName}}ListBulkLoad(keys []{{.KeyType}}, values []{{.ValType}}) *{{.MapName}}List {
	if len(keys) != len(values) {
		panic("mapdb: {{.MapName}}ListBulkLoad: len(keys) != len(values)")
	}
	m := &{{.MapName}}List{
		data: make(map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}}, len(keys)),
{{- if .KeyIsFloat}}
		keys: make(map[{{.KeyBitsType}}]{{.KeyType}}, len(keys)),
{{- end}}
	}
	for i := range keys {
		m.Put(keys[i], values[i])
	}
	return m
}

// New{{.MapName}}ListFromSortedKeys builds a {{.MapName}}List from input grouped
// by ascending key: all values for a key are contiguous, and keys appear in
// ascending order (the IEEE-754 total order for float keys). It validates the key
// monotonicity in one pass and assigns each key's value slice directly, preserving
// value order within a key run. keys[i] and values[i] form one pair (a length
// mismatch panics). Out-of-order or interleaved keys return pump.ErrNotSorted.
// The result is observably identical to the same pairs inserted with Put.
func New{{.MapName}}ListFromSortedKeys(keys []{{.KeyType}}, values []{{.ValType}}) (*{{.MapName}}List, error) {
	if len(keys) != len(values) {
		panic("mapdb: New{{.MapName}}ListFromSortedKeys: len(keys) != len(values)")
	}
	m := &{{.MapName}}List{
		data: make(map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}}),
{{- if .KeyIsFloat}}
		keys: make(map[{{.KeyBitsType}}]{{.KeyType}}),
{{- end}}
	}
	i := 0
	for i < len(keys) {
		key := keys[i]
		if i > 0 && cmpKey{{.KeyName}}(key, keys[i-1]) <= 0 {
			return nil, pump.ErrNotSorted
		}
		j := i
		run := []{{.ValType}}{}
		for j < len(keys) && cmpKey{{.KeyName}}(keys[j], key) == 0 {
			run = append(run, values[j])
			j++
		}
{{- if .KeyIsFloat}}
		kb := {{.KeyBitsFn}}(key)
		m.data[kb] = run
		m.keys[kb] = key
{{- else}}
		m.data[key] = run
{{- end}}
		m.size += len(run)
		i = j
	}
	return m, nil
}

// New{{.MapName}}ListFromSortedKeyValues builds a {{.MapName}}List from input
// sorted by ascending key and, within each key, ascending value. It validates
// both key monotonicity and per-key value monotonicity (using the value type's
// own comparator — the IEEE-754 total order for float values) in one pass.
// Unlike set multimaps, list multimaps preserve equal adjacent values exactly.
// keys[i] and values[i] form one pair (a length mismatch panics).
// Out-of-order keys, or values that descend within a key run, return
// pump.ErrNotSorted before any partial collection is built. If your values are
// not sorted within each key, use {{.MapName}}ListBulkLoad instead.
func New{{.MapName}}ListFromSortedKeyValues(keys []{{.KeyType}}, values []{{.ValType}}) (*{{.MapName}}List, error) {
	if len(keys) != len(values) {
		panic("mapdb: New{{.MapName}}ListFromSortedKeyValues: len(keys) != len(values)")
	}
	m := &{{.MapName}}List{
		data: make(map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}}),
{{- if .KeyIsFloat}}
		keys: make(map[{{.KeyBitsType}}]{{.KeyType}}),
{{- end}}
	}
	i := 0
	for i < len(keys) {
		key := keys[i]
		if i > 0 && cmpKey{{.KeyName}}(key, keys[i-1]) <= 0 {
			return nil, pump.ErrNotSorted
		}
		j := i
		run := []{{.ValType}}{}
		for j < len(keys) && cmpKey{{.KeyName}}(keys[j], key) == 0 {
			v := values[j]
			if len(run) > 0 && cmpKey{{.ValName}}(run[len(run)-1], v) > 0 {
				return nil, pump.ErrNotSorted // value descends within key run
			}
			run = append(run, v)
			j++
		}
{{- if .KeyIsFloat}}
		kb := {{.KeyBitsFn}}(key)
		m.data[kb] = run
		m.keys[kb] = key
{{- else}}
		m.data[key] = run
{{- end}}
		m.size += len(run)
		i = j
	}
	return m, nil
}

// Put adds a value to the list for the given key.
func (m *{{.MapName}}List) Put(key {{.KeyType}}, value {{.ValType}}) {
	if m.data == nil {
		m.data = make(map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}})
{{- if .KeyIsFloat}}
		m.keys = make(map[{{.KeyBitsType}}]{{.KeyType}})
{{- end}}
	}
{{- if .KeyIsFloat}}
	kb := {{.KeyBitsFn}}(key)
	m.data[kb] = append(m.data[kb], value)
	m.keys[kb] = key
{{- else}}
	m.data[key] = append(m.data[key], value)
{{- end}}
	m.size++
}

// Get returns a copy of the values for the given key. Returns nil if the key is absent.
func (m *{{.MapName}}List) Get(key {{.KeyType}}) []{{.ValType}} {
	return m.GetAll(key)
}

// GetAll returns a copy of the values for the given key.
func (m *{{.MapName}}List) GetAll(key {{.KeyType}}) []{{.ValType}} {
	vals := m.data[{{if .KeyIsFloat}}{{.KeyBitsFn}}(key){{else}}key{{end}}]
	if vals == nil {
		return nil
	}
	result := make([]{{.ValType}}, len(vals))
	copy(result, vals)
	return result
}

// RemoveAll removes all values for the given key and returns them.
func (m *{{.MapName}}List) RemoveAll(key {{.KeyType}}) []{{.ValType}} {
{{- if .KeyIsFloat}}
	kb := {{.KeyBitsFn}}(key)
	vals, ok := m.data[kb]
	if !ok {
		return nil
	}
	delete(m.data, kb)
	delete(m.keys, kb)
{{- else}}
	vals, ok := m.data[key]
	if !ok {
		return nil
	}
	delete(m.data, key)
{{- end}}
	m.size -= len(vals)
	return vals
}

// ContainsKey returns true if the multimap contains the given key.
func (m *{{.MapName}}List) ContainsKey(key {{.KeyType}}) bool {
	_, ok := m.data[{{if .KeyIsFloat}}{{.KeyBitsFn}}(key){{else}}key{{end}}]
	return ok
}

// ContainsKeyValue returns true if the multimap contains the given key-value pair.
func (m *{{.MapName}}List) ContainsKeyValue(key {{.KeyType}}, value {{.ValType}}) bool {
	vals, ok := m.data[{{if .KeyIsFloat}}{{.KeyBitsFn}}(key){{else}}key{{end}}]
	if !ok {
		return false
	}
	for _, v := range vals {
		if {{if .ValueIsFloat}}{{.ValBitsFn}}(v) == {{.ValBitsFn}}(value){{else}}v == value{{end}} {
			return true
		}
	}
	return false
}

// KeysCount returns the number of distinct keys.
func (m *{{.MapName}}List) KeysCount() int {
	return len(m.data)
}

// Len returns the number of elements. Use m.Len() == 0 to test for emptiness.
func (m *{{.MapName}}List) Len() int {
	return m.size
}

// Clear removes all entries from the multimap.
func (m *{{.MapName}}List) Clear() {
	m.data = make(map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}})
{{- if .KeyIsFloat}}
	m.keys = make(map[{{.KeyBitsType}}]{{.KeyType}})
{{- end}}
	m.size = 0
}

// ForEach calls the given function for each key-value pair.
func (m *{{.MapName}}List) ForEach(f func({{.KeyType}}, {{.ValType}})) {
	for {{if .KeyIsFloat}}kb{{else}}key{{end}}, vals := range m.data {
{{- if .KeyIsFloat}}
		key := m.keys[kb]
{{- end}}
		for _, val := range vals {
			f(key, val)
		}
	}
}

// ForEachKeyValues calls the given function for each key with a copy of its values.
func (m *{{.MapName}}List) ForEachKeyValues(f func({{.KeyType}}, []{{.ValType}})) {
	for {{if .KeyIsFloat}}kb{{else}}key{{end}}, vals := range m.data {
{{- if .KeyIsFloat}}
		key := m.keys[kb]
{{- end}}
		copied := make([]{{.ValType}}, len(vals))
		copy(copied, vals)
		f(key, copied)
	}
}

// Keys returns a slice of all distinct keys.
func (m *{{.MapName}}List) Keys() []{{.KeyType}} {
	result := make([]{{.KeyType}}, 0, len(m.data))
{{- if .KeyIsFloat}}
	for _, key := range m.keys {
		result = append(result, key)
	}
{{- else}}
	for key := range m.data {
		result = append(result, key)
	}
{{- end}}
	return result
}

// Values returns a slice of all values across all keys.
func (m *{{.MapName}}List) Values() []{{.ValType}} {
	result := make([]{{.ValType}}, 0, m.size)
	for _, vals := range m.data {
		result = append(result, vals...)
	}
	return result
}

// Select returns a new multimap containing only key-value pairs that satisfy the predicate.
func (m *{{.MapName}}List) Select(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}}List {
	result := New{{.MapName}}List()
	for {{if .KeyIsFloat}}kb{{else}}key{{end}}, vals := range m.data {
{{- if .KeyIsFloat}}
		key := m.keys[kb]
{{- end}}
		for _, val := range vals {
			if predicate(key, val) {
				result.Put(key, val)
			}
		}
	}
	return result
}

// Reject returns a new multimap containing only key-value pairs that do not satisfy the predicate.
func (m *{{.MapName}}List) Reject(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}}List {
	result := New{{.MapName}}List()
	for {{if .KeyIsFloat}}kb{{else}}key{{end}}, vals := range m.data {
{{- if .KeyIsFloat}}
		key := m.keys[kb]
{{- end}}
		for _, val := range vals {
			if !predicate(key, val) {
				result.Put(key, val)
			}
		}
	}
	return result
}

// String returns a string representation of the multimap.
func (m *{{.MapName}}List) String() string {
	if m.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for {{if .KeyIsFloat}}kb{{else}}key{{end}}, vals := range m.data {
{{- if .KeyIsFloat}}
		key := m.keys[kb]
{{- end}}
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v=[", key)
		for i, val := range vals {
			if i > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%v", val)
		}
		sb.WriteString("]")
		first = false
	}
	sb.WriteString("}")
	return sb.String()
}

// Equals returns true if the other multimap has the same key-value pairs in the same order per key.
func (m *{{.MapName}}List) Equals(other *{{.MapName}}List) bool {
	if m.size != other.size {
		return false
	}
	if len(m.data) != len(other.data) {
		return false
	}
	for {{if .KeyIsFloat}}kb{{else}}key{{end}}, vals := range m.data {
		otherVals, ok := other.data[{{if .KeyIsFloat}}kb{{else}}key{{end}}]
		if !ok || len(vals) != len(otherVals) {
			return false
		}
		for i, val := range vals {
			if !({{if .ValueIsFloat}}{{.ValBitsFn}}(val) == {{.ValBitsFn}}(otherVals[i]){{else}}val == otherVals[i]{{end}}) {
				return false
			}
		}
	}
	return true
}

// KeysToSlice returns all distinct keys as a slice.
func (m *{{.MapName}}List) KeysToSlice() []{{.KeyType}} {
	return m.Keys()
}

// ValuesToSlice returns all values as a slice.
func (m *{{.MapName}}List) ValuesToSlice() []{{.ValType}} {
	return m.Values()
}

// PutReturning adds a key-value pair and returns the multimap (fluent API).
func (m *{{.MapName}}List) PutReturning(key {{.KeyType}}, value {{.ValType}}) *{{.MapName}}List {
	m.Put(key, value)
	return m
}

// RemoveKeyReturning removes all values for the key and returns the multimap (fluent API).
func (m *{{.MapName}}List) RemoveKeyReturning(key {{.KeyType}}) *{{.MapName}}List {
	m.RemoveAll(key)
	return m
}
`

const setMultimapTmpl = genHeader + `package multimap

import (
{{- if .NeedsMath}}
	"math"
{{- end}}
	"fmt"
	"strings"

	"github.com/mapdb/mapdb-golang/pump"
)

// {{.MapName}}Set is a set multimap from {{.KeyType}} keys to {{.ValType}} values.
// Each key maps to a set of unique values (duplicates on Put are silently dropped).
type {{.MapName}}Set struct {
	data map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}}
{{- if .KeyIsFloat}}
	keys map[{{.KeyBitsType}}]{{.KeyType}}
{{- end}}
	size int
}

// New{{.MapName}}Set creates a new empty {{.MapName}}Set.
func New{{.MapName}}Set() *{{.MapName}}Set {
	return &{{.MapName}}Set{
		data: make(map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}}),
{{- if .KeyIsFloat}}
		keys: make(map[{{.KeyBitsType}}]{{.KeyType}}),
{{- end}}
		size: 0,
	}
}

// {{.MapName}}SetBulkLoad builds a {{.MapName}}Set from keys/values in a single
// pass, presizing the backing map for the input. keys[i] and values[i] form one
// pair (a length mismatch panics). The input need not be sorted; per-key value
// duplicates are dropped exactly as repeated Put. Duplicate keys are the normal
// grouping case, so the duplicate policy does not apply.
func {{.MapName}}SetBulkLoad(keys []{{.KeyType}}, values []{{.ValType}}) *{{.MapName}}Set {
	if len(keys) != len(values) {
		panic("mapdb: {{.MapName}}SetBulkLoad: len(keys) != len(values)")
	}
	m := &{{.MapName}}Set{
		data: make(map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}}, len(keys)),
{{- if .KeyIsFloat}}
		keys: make(map[{{.KeyBitsType}}]{{.KeyType}}, len(keys)),
{{- end}}
	}
	for i := range keys {
		m.Put(keys[i], values[i])
	}
	return m
}

// New{{.MapName}}SetFromSortedKeyValues builds a {{.MapName}}Set from input sorted
// by ascending key and, within each key, ascending value. It validates both key
// monotonicity AND per-key value monotonicity (using the value type's own
// comparator — the IEEE-754 total order for float values) in one pass, deduping
// equal values per key (the sorted equivalent of the linear-scan dedupe Put
// performs). keys[i] and values[i] form one pair (a length mismatch panics).
// Out-of-order keys, or values that descend within a key run, return
// pump.ErrNotSorted before any partial collection is built. The result is
// observably identical to the same pairs inserted with Put; if your values are
// not sorted within each key, use {{.MapName}}SetBulkLoad instead.
func New{{.MapName}}SetFromSortedKeyValues(keys []{{.KeyType}}, values []{{.ValType}}) (*{{.MapName}}Set, error) {
	if len(keys) != len(values) {
		panic("mapdb: New{{.MapName}}SetFromSortedKeyValues: len(keys) != len(values)")
	}
	m := &{{.MapName}}Set{
		data: make(map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}}),
{{- if .KeyIsFloat}}
		keys: make(map[{{.KeyBitsType}}]{{.KeyType}}),
{{- end}}
	}
	i := 0
	for i < len(keys) {
		key := keys[i]
		if i > 0 && cmpKey{{.KeyName}}(key, keys[i-1]) <= 0 {
			return nil, pump.ErrNotSorted
		}
		j := i
		run := []{{.ValType}}{}
		for j < len(keys) && cmpKey{{.KeyName}}(keys[j], key) == 0 {
			v := values[j]
			if len(run) > 0 {
				c := cmpKey{{.ValName}}(run[len(run)-1], v)
				if c > 0 {
					return nil, pump.ErrNotSorted // value descends within key run
				}
				if c == 0 {
					j++
					continue // adjacent duplicate value (input is sorted, so equals are adjacent)
				}
			}
			run = append(run, v)
			j++
		}
{{- if .KeyIsFloat}}
		kb := {{.KeyBitsFn}}(key)
		m.data[kb] = run
		m.keys[kb] = key
{{- else}}
		m.data[key] = run
{{- end}}
		m.size += len(run)
		i = j
	}
	return m, nil
}

// Put adds a value to the set for the given key. Idempotent: a duplicate
// value for the same key is silently dropped.
func (m *{{.MapName}}Set) Put(key {{.KeyType}}, value {{.ValType}}) {
	if m.data == nil {
		m.data = make(map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}})
{{- if .KeyIsFloat}}
		m.keys = make(map[{{.KeyBitsType}}]{{.KeyType}})
{{- end}}
	}
{{- if .KeyIsFloat}}
	kb := {{.KeyBitsFn}}(key)
	for _, existing := range m.data[kb] {
		if {{if .ValueIsFloat}}{{.ValBitsFn}}(existing) == {{.ValBitsFn}}(value){{else}}existing == value{{end}} {
			return
		}
	}
	m.data[kb] = append(m.data[kb], value)
	m.keys[kb] = key
{{- else}}
	for _, existing := range m.data[key] {
		if {{if .ValueIsFloat}}{{.ValBitsFn}}(existing) == {{.ValBitsFn}}(value){{else}}existing == value{{end}} {
			return
		}
	}
	m.data[key] = append(m.data[key], value)
{{- end}}
	m.size++
}

// Get returns a copy of the values for the given key. Returns nil if the key is absent.
func (m *{{.MapName}}Set) Get(key {{.KeyType}}) []{{.ValType}} {
	return m.GetAll(key)
}

// GetAll returns a copy of the values for the given key.
func (m *{{.MapName}}Set) GetAll(key {{.KeyType}}) []{{.ValType}} {
	vals := m.data[{{if .KeyIsFloat}}{{.KeyBitsFn}}(key){{else}}key{{end}}]
	if vals == nil {
		return nil
	}
	result := make([]{{.ValType}}, len(vals))
	copy(result, vals)
	return result
}

// RemoveAll removes all values for the given key and returns them.
func (m *{{.MapName}}Set) RemoveAll(key {{.KeyType}}) []{{.ValType}} {
{{- if .KeyIsFloat}}
	kb := {{.KeyBitsFn}}(key)
	vals, ok := m.data[kb]
	if !ok {
		return nil
	}
	delete(m.data, kb)
	delete(m.keys, kb)
{{- else}}
	vals, ok := m.data[key]
	if !ok {
		return nil
	}
	delete(m.data, key)
{{- end}}
	m.size -= len(vals)
	return vals
}

// ContainsKey returns true if the multimap contains the given key.
func (m *{{.MapName}}Set) ContainsKey(key {{.KeyType}}) bool {
	_, ok := m.data[{{if .KeyIsFloat}}{{.KeyBitsFn}}(key){{else}}key{{end}}]
	return ok
}

// ContainsKeyValue returns true if the multimap contains the given key-value pair.
func (m *{{.MapName}}Set) ContainsKeyValue(key {{.KeyType}}, value {{.ValType}}) bool {
	vals, ok := m.data[{{if .KeyIsFloat}}{{.KeyBitsFn}}(key){{else}}key{{end}}]
	if !ok {
		return false
	}
	for _, v := range vals {
		if {{if .ValueIsFloat}}{{.ValBitsFn}}(v) == {{.ValBitsFn}}(value){{else}}v == value{{end}} {
			return true
		}
	}
	return false
}

// KeysCount returns the number of distinct keys.
func (m *{{.MapName}}Set) KeysCount() int {
	return len(m.data)
}

// Len returns the number of elements. Use m.Len() == 0 to test for emptiness.
func (m *{{.MapName}}Set) Len() int {
	return m.size
}

// Clear removes all entries from the multimap.
func (m *{{.MapName}}Set) Clear() {
	m.data = make(map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}})
{{- if .KeyIsFloat}}
	m.keys = make(map[{{.KeyBitsType}}]{{.KeyType}})
{{- end}}
	m.size = 0
}

// ForEach calls the given function for each key-value pair.
func (m *{{.MapName}}Set) ForEach(f func({{.KeyType}}, {{.ValType}})) {
	for {{if .KeyIsFloat}}kb{{else}}key{{end}}, vals := range m.data {
{{- if .KeyIsFloat}}
		key := m.keys[kb]
{{- end}}
		for _, val := range vals {
			f(key, val)
		}
	}
}

// ForEachKeyValues calls the given function for each key with a copy of its values.
func (m *{{.MapName}}Set) ForEachKeyValues(f func({{.KeyType}}, []{{.ValType}})) {
	for {{if .KeyIsFloat}}kb{{else}}key{{end}}, vals := range m.data {
{{- if .KeyIsFloat}}
		key := m.keys[kb]
{{- end}}
		copied := make([]{{.ValType}}, len(vals))
		copy(copied, vals)
		f(key, copied)
	}
}

// Keys returns a slice of all distinct keys.
func (m *{{.MapName}}Set) Keys() []{{.KeyType}} {
	result := make([]{{.KeyType}}, 0, len(m.data))
{{- if .KeyIsFloat}}
	for _, key := range m.keys {
		result = append(result, key)
	}
{{- else}}
	for key := range m.data {
		result = append(result, key)
	}
{{- end}}
	return result
}

// Values returns a slice of all values across all keys.
func (m *{{.MapName}}Set) Values() []{{.ValType}} {
	result := make([]{{.ValType}}, 0, m.size)
	for _, vals := range m.data {
		result = append(result, vals...)
	}
	return result
}

// Select returns a new multimap containing only key-value pairs that satisfy the predicate.
func (m *{{.MapName}}Set) Select(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}}Set {
	result := New{{.MapName}}Set()
	for {{if .KeyIsFloat}}kb{{else}}key{{end}}, vals := range m.data {
{{- if .KeyIsFloat}}
		key := m.keys[kb]
{{- end}}
		for _, val := range vals {
			if predicate(key, val) {
				result.Put(key, val)
			}
		}
	}
	return result
}

// Reject returns a new multimap containing only key-value pairs that do not satisfy the predicate.
func (m *{{.MapName}}Set) Reject(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}}Set {
	result := New{{.MapName}}Set()
	for {{if .KeyIsFloat}}kb{{else}}key{{end}}, vals := range m.data {
{{- if .KeyIsFloat}}
		key := m.keys[kb]
{{- end}}
		for _, val := range vals {
			if !predicate(key, val) {
				result.Put(key, val)
			}
		}
	}
	return result
}

// String returns a string representation of the multimap.
func (m *{{.MapName}}Set) String() string {
	if m.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for {{if .KeyIsFloat}}kb{{else}}key{{end}}, vals := range m.data {
{{- if .KeyIsFloat}}
		key := m.keys[kb]
{{- end}}
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v=[", key)
		for i, val := range vals {
			if i > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%v", val)
		}
		sb.WriteString("]")
		first = false
	}
	sb.WriteString("}")
	return sb.String()
}

// Equals returns true if the other multimap has the same key-value pairs in the same order per key.
func (m *{{.MapName}}Set) Equals(other *{{.MapName}}Set) bool {
	if m.size != other.size {
		return false
	}
	if len(m.data) != len(other.data) {
		return false
	}
	for {{if .KeyIsFloat}}kb{{else}}key{{end}}, vals := range m.data {
		otherVals, ok := other.data[{{if .KeyIsFloat}}kb{{else}}key{{end}}]
		if !ok || len(vals) != len(otherVals) {
			return false
		}
		for i, val := range vals {
			if !({{if .ValueIsFloat}}{{.ValBitsFn}}(val) == {{.ValBitsFn}}(otherVals[i]){{else}}val == otherVals[i]{{end}}) {
				return false
			}
		}
	}
	return true
}

// KeysToSlice returns all distinct keys as a slice.
func (m *{{.MapName}}Set) KeysToSlice() []{{.KeyType}} {
	return m.Keys()
}

// ValuesToSlice returns all values as a slice.
func (m *{{.MapName}}Set) ValuesToSlice() []{{.ValType}} {
	return m.Values()
}

// PutReturning adds a key-value pair and returns the multimap (fluent API).
func (m *{{.MapName}}Set) PutReturning(key {{.KeyType}}, value {{.ValType}}) *{{.MapName}}Set {
	m.Put(key, value)
	return m
}

// RemoveKeyReturning removes all values for the key and returns the multimap (fluent API).
func (m *{{.MapName}}Set) RemoveKeyReturning(key {{.KeyType}}) *{{.MapName}}Set {
	m.RemoveAll(key)
	return m
}
`
