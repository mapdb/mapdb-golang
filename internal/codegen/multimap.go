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

	list := template.Must(template.New("mm-list").Parse(listMultimapTmpl))
	set := template.Must(template.New("mm-set").Parse(setMultimapTmpl))

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
	}

	return nil
}

const listMultimapTmpl = genHeader + `package multimap

import (
{{- if .NeedsMath}}
	"math"
{{- end}}
	"fmt"
	"strings"
)

// {{.MapName}}ListMultimap is a list multimap from {{.KeyType}} keys to {{.ValType}} values.
// Each key maps to a slice of values, preserving insertion order per key.
type {{.MapName}}ListMultimap struct {
	data map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}}
{{- if .KeyIsFloat}}
	keys map[{{.KeyBitsType}}]{{.KeyType}}
{{- end}}
	size int
}

// New{{.MapName}}ListMultimap creates a new empty {{.MapName}}ListMultimap.
func New{{.MapName}}ListMultimap() *{{.MapName}}ListMultimap {
	return &{{.MapName}}ListMultimap{
		data: make(map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}}),
{{- if .KeyIsFloat}}
		keys: make(map[{{.KeyBitsType}}]{{.KeyType}}),
{{- end}}
		size: 0,
	}
}

// Put adds a value to the list for the given key.
func (m *{{.MapName}}ListMultimap) Put(key {{.KeyType}}, value {{.ValType}}) {
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
func (m *{{.MapName}}ListMultimap) Get(key {{.KeyType}}) []{{.ValType}} {
	return m.GetAll(key)
}

// GetAll returns a copy of the values for the given key.
func (m *{{.MapName}}ListMultimap) GetAll(key {{.KeyType}}) []{{.ValType}} {
	vals := m.data[{{if .KeyIsFloat}}{{.KeyBitsFn}}(key){{else}}key{{end}}]
	if vals == nil {
		return nil
	}
	result := make([]{{.ValType}}, len(vals))
	copy(result, vals)
	return result
}

// RemoveAll removes all values for the given key and returns them.
func (m *{{.MapName}}ListMultimap) RemoveAll(key {{.KeyType}}) []{{.ValType}} {
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
func (m *{{.MapName}}ListMultimap) ContainsKey(key {{.KeyType}}) bool {
	_, ok := m.data[{{if .KeyIsFloat}}{{.KeyBitsFn}}(key){{else}}key{{end}}]
	return ok
}

// ContainsKeyValue returns true if the multimap contains the given key-value pair.
func (m *{{.MapName}}ListMultimap) ContainsKeyValue(key {{.KeyType}}, value {{.ValType}}) bool {
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
func (m *{{.MapName}}ListMultimap) KeysCount() int {
	return len(m.data)
}

// Size returns the total number of values across all keys.
func (m *{{.MapName}}ListMultimap) Size() int {
	return m.size
}

// Len returns the number of elements. It is an alias for Size, matching
// Go convention (sort.Interface, container/list, bytes.Buffer).
func (m *{{.MapName}}ListMultimap) Len() int { return m.Size() }

// IsEmpty returns true if the multimap contains no values.
func (m *{{.MapName}}ListMultimap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the multimap.
func (m *{{.MapName}}ListMultimap) Clear() {
	m.data = make(map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}})
{{- if .KeyIsFloat}}
	m.keys = make(map[{{.KeyBitsType}}]{{.KeyType}})
{{- end}}
	m.size = 0
}

// ForEach calls the given function for each key-value pair.
func (m *{{.MapName}}ListMultimap) ForEach(f func({{.KeyType}}, {{.ValType}})) {
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
func (m *{{.MapName}}ListMultimap) ForEachKeyValues(f func({{.KeyType}}, []{{.ValType}})) {
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
func (m *{{.MapName}}ListMultimap) Keys() []{{.KeyType}} {
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
func (m *{{.MapName}}ListMultimap) Values() []{{.ValType}} {
	result := make([]{{.ValType}}, 0, m.size)
	for _, vals := range m.data {
		result = append(result, vals...)
	}
	return result
}

// Select returns a new multimap containing only key-value pairs that satisfy the predicate.
func (m *{{.MapName}}ListMultimap) Select(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}}ListMultimap {
	result := New{{.MapName}}ListMultimap()
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
func (m *{{.MapName}}ListMultimap) Reject(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}}ListMultimap {
	result := New{{.MapName}}ListMultimap()
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
func (m *{{.MapName}}ListMultimap) String() string {
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
func (m *{{.MapName}}ListMultimap) Equals(other *{{.MapName}}ListMultimap) bool {
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
func (m *{{.MapName}}ListMultimap) KeysToSlice() []{{.KeyType}} {
	return m.Keys()
}

// ValuesToSlice returns all values as a slice.
func (m *{{.MapName}}ListMultimap) ValuesToSlice() []{{.ValType}} {
	return m.Values()
}

// WithKeyValue adds a key-value pair and returns the multimap (fluent API).
func (m *{{.MapName}}ListMultimap) WithKeyValue(key {{.KeyType}}, value {{.ValType}}) *{{.MapName}}ListMultimap {
	m.Put(key, value)
	return m
}

// WithoutKey removes all values for the key and returns the multimap (fluent API).
func (m *{{.MapName}}ListMultimap) WithoutKey(key {{.KeyType}}) *{{.MapName}}ListMultimap {
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
)

// {{.MapName}}SetMultimap is a set multimap from {{.KeyType}} keys to {{.ValType}} values.
// Each key maps to a set of unique values (duplicates on Put are silently dropped).
type {{.MapName}}SetMultimap struct {
	data map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}}
{{- if .KeyIsFloat}}
	keys map[{{.KeyBitsType}}]{{.KeyType}}
{{- end}}
	size int
}

// New{{.MapName}}SetMultimap creates a new empty {{.MapName}}SetMultimap.
func New{{.MapName}}SetMultimap() *{{.MapName}}SetMultimap {
	return &{{.MapName}}SetMultimap{
		data: make(map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}}),
{{- if .KeyIsFloat}}
		keys: make(map[{{.KeyBitsType}}]{{.KeyType}}),
{{- end}}
		size: 0,
	}
}

// Put adds a value to the set for the given key. Idempotent: a duplicate
// value for the same key is silently dropped.
func (m *{{.MapName}}SetMultimap) Put(key {{.KeyType}}, value {{.ValType}}) {
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
func (m *{{.MapName}}SetMultimap) Get(key {{.KeyType}}) []{{.ValType}} {
	return m.GetAll(key)
}

// GetAll returns a copy of the values for the given key.
func (m *{{.MapName}}SetMultimap) GetAll(key {{.KeyType}}) []{{.ValType}} {
	vals := m.data[{{if .KeyIsFloat}}{{.KeyBitsFn}}(key){{else}}key{{end}}]
	if vals == nil {
		return nil
	}
	result := make([]{{.ValType}}, len(vals))
	copy(result, vals)
	return result
}

// RemoveAll removes all values for the given key and returns them.
func (m *{{.MapName}}SetMultimap) RemoveAll(key {{.KeyType}}) []{{.ValType}} {
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
func (m *{{.MapName}}SetMultimap) ContainsKey(key {{.KeyType}}) bool {
	_, ok := m.data[{{if .KeyIsFloat}}{{.KeyBitsFn}}(key){{else}}key{{end}}]
	return ok
}

// ContainsKeyValue returns true if the multimap contains the given key-value pair.
func (m *{{.MapName}}SetMultimap) ContainsKeyValue(key {{.KeyType}}, value {{.ValType}}) bool {
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
func (m *{{.MapName}}SetMultimap) KeysCount() int {
	return len(m.data)
}

// Size returns the total number of values across all keys.
func (m *{{.MapName}}SetMultimap) Size() int {
	return m.size
}

// Len returns the number of elements. It is an alias for Size, matching
// Go convention (sort.Interface, container/list, bytes.Buffer).
func (m *{{.MapName}}SetMultimap) Len() int { return m.Size() }

// IsEmpty returns true if the multimap contains no values.
func (m *{{.MapName}}SetMultimap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the multimap.
func (m *{{.MapName}}SetMultimap) Clear() {
	m.data = make(map[{{if .KeyIsFloat}}{{.KeyBitsType}}{{else}}{{.KeyType}}{{end}}][]{{.ValType}})
{{- if .KeyIsFloat}}
	m.keys = make(map[{{.KeyBitsType}}]{{.KeyType}})
{{- end}}
	m.size = 0
}

// ForEach calls the given function for each key-value pair.
func (m *{{.MapName}}SetMultimap) ForEach(f func({{.KeyType}}, {{.ValType}})) {
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
func (m *{{.MapName}}SetMultimap) ForEachKeyValues(f func({{.KeyType}}, []{{.ValType}})) {
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
func (m *{{.MapName}}SetMultimap) Keys() []{{.KeyType}} {
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
func (m *{{.MapName}}SetMultimap) Values() []{{.ValType}} {
	result := make([]{{.ValType}}, 0, m.size)
	for _, vals := range m.data {
		result = append(result, vals...)
	}
	return result
}

// Select returns a new multimap containing only key-value pairs that satisfy the predicate.
func (m *{{.MapName}}SetMultimap) Select(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}}SetMultimap {
	result := New{{.MapName}}SetMultimap()
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
func (m *{{.MapName}}SetMultimap) Reject(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}}SetMultimap {
	result := New{{.MapName}}SetMultimap()
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
func (m *{{.MapName}}SetMultimap) String() string {
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
func (m *{{.MapName}}SetMultimap) Equals(other *{{.MapName}}SetMultimap) bool {
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
func (m *{{.MapName}}SetMultimap) KeysToSlice() []{{.KeyType}} {
	return m.Keys()
}

// ValuesToSlice returns all values as a slice.
func (m *{{.MapName}}SetMultimap) ValuesToSlice() []{{.ValType}} {
	return m.Values()
}

// WithKeyValue adds a key-value pair and returns the multimap (fluent API).
func (m *{{.MapName}}SetMultimap) WithKeyValue(key {{.KeyType}}, value {{.ValType}}) *{{.MapName}}SetMultimap {
	m.Put(key, value)
	return m
}

// WithoutKey removes all values for the key and returns the multimap (fluent API).
func (m *{{.MapName}}SetMultimap) WithoutKey(key {{.KeyType}}) *{{.MapName}}SetMultimap {
	m.RemoveAll(key)
	return m
}
`
