package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

// bagData is the per-primitive view the bag templates iterate over.
//
// The bag family is a multiset. It has FOUR variant groups per primitive:
//
//   - {type}_hash_bag.go         — base HASH bag (unordered, map-backed)
//   - immutable_{type}_hash_bag.go
//   - synchronized_{type}_hash_bag.go
//   - {type}_tree_bag.go         — base TREE bag (sorted slice, binary search)
//
// There is NO immutable/synchronized TREE bag (the tree bag is base-only).
//
// The HASH bag is STRUCTURALLY DIFFERENT for floats vs ints/char:
//
//   - int/char hash bag: counts is map[GoType]int — the value IS the map key,
//     the map value is the count.
//   - float hash bag: counts is map[BitsType]<snake>BagEntry, where BitsType is
//     the IEEE-754 bit-pattern integer (uint32/uint64). The map key is
//     math.FloatNbits(value) so NaN keys are findable and ±0 stay distinct; the
//     entry struct (value, count) recovers the original float. Every method
//     differs accordingly (keyed by k := <BitsFn>(value), iterate over entries).
//
// The TREE bag differs only in the search comparison (raw < / > for int/char,
// cmpFloatNN for floats) and the Zero literal used in Min/Max/Detect.
type bagData struct {
	Name      string // Int32, Float32, Char (identifier stem)
	GoType    string // int32, float32, uint16 (Go element type)
	SnakeName string // int32, float32, char (file-name stem)
	Zero      string // zero literal for this element type ("0" or "0.0")

	IsFloat bool
	// BitsType is the IEEE-754 bit-pattern integer type used as the hash-bag
	// map key (uint32 / uint64). Floats only.
	BitsType string
	// BitsFn is the float bit-pattern function (math.Float32bits /
	// math.Float64bits) used to derive the map key. Floats only.
	BitsFn string
	// CmpFn is the float total-order helper (cmpFloat32 / cmpFloat64) used by
	// the tree bag's binary search. Floats only.
	CmpFn string
}

// genBag writes the per-primitive bag sources into the current working
// directory: base hash bag + immutable + synchronized wrappers, base tree bag,
// and the shared cmp_float.go. Invoked from bag/ via go:generate.
func genBag() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	hash := template.Must(template.New("bag-hash").Parse(hashBagTmpl))
	immutable := template.Must(template.New("bag-immutable").Parse(immutableHashBagTmpl))
	synchronized := template.Must(template.New("bag-sync").Parse(synchronizedHashBagTmpl))
	tree := template.Must(template.New("bag-tree").Parse(treeBagTmpl))

	write := func(name string, tmpl *template.Template, data bagData) error {
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

	for _, p := range Primitives() {
		data := bagData{
			Name:      p.Name,
			GoType:    p.GoType,
			SnakeName: p.SnakeName,
			Zero:      "0",
			IsFloat:   p.IsFloating,
		}
		if p.IsFloating {
			data.Zero = "0.0"
			if p.ByteSize == 4 {
				data.BitsType = "uint32"
				data.BitsFn = "math.Float32bits"
				data.CmpFn = "cmpFloat32"
			} else {
				data.BitsType = "uint64"
				data.BitsFn = "math.Float64bits"
				data.CmpFn = "cmpFloat64"
			}
		}

		if err := write(data.SnakeName+"_hash_bag.go", hash, data); err != nil {
			return err
		}
		if err := write("immutable_"+data.SnakeName+"_hash_bag.go", immutable, data); err != nil {
			return err
		}
		if err := write("synchronized_"+data.SnakeName+"_hash_bag.go", synchronized, data); err != nil {
			return err
		}
		if err := write(data.SnakeName+"_tree_bag.go", tree, data); err != nil {
			return err
		}
	}

	return genCmpFloat("bag")
}

const hashBagTmpl = genHeader + `package bag

import (
	"fmt"
	"iter"
{{- if .IsFloat}}
	"math"
{{- end}}
	"strings"
)
{{if .IsFloat}}
// {{.SnakeName}}BagEntry stores the original {{.GoType}} value alongside its count.
// The backing map is keyed by the IEEE-754 bit pattern ({{.BitsFn}}) so
// that NaN keys are findable and +0.0 / -0.0 remain distinct; Go map equality
// on raw {{.GoType}} would make every NaN a fresh unreachable entry and collapse
// signed zero.
type {{.SnakeName}}BagEntry struct {
	value {{.GoType}}
	count int
}

// {{.Name}}HashBag is a bag (multiset) that counts occurrences of {{.GoType}} values.
// Backed by a map from value bit pattern to (value, count).
type {{.Name}}HashBag struct {
	counts map[{{.BitsType}}]{{.SnakeName}}BagEntry
	size   int // total count including duplicates
}
{{else}}
// {{.Name}}HashBag is a bag (multiset) that counts occurrences of {{.GoType}} values.
// Backed by a map from value to count.
type {{.Name}}HashBag struct {
	counts map[{{.GoType}}]int
	size   int // total count including duplicates
}
{{end}}
// New{{.Name}}HashBag creates a new empty {{.Name}}HashBag.
func New{{.Name}}HashBag() *{{.Name}}HashBag {
	return &{{.Name}}HashBag{
		counts: make(map[{{if .IsFloat}}{{.BitsType}}]{{.SnakeName}}BagEntry{{else}}{{.GoType}}]int{{end}}),
		size:   0,
	}
}

// {{.Name}}HashBagOf creates a new {{.Name}}HashBag from the given values.
func {{.Name}}HashBagOf(values ...{{.GoType}}) *{{.Name}}HashBag {
	b := New{{.Name}}HashBag()
	for _, v := range values {
		b.Add(v)
	}
	return b
}

// Add adds one occurrence of the value.
func (b *{{.Name}}HashBag) Add(value {{.GoType}}) {
{{- if .IsFloat}}
	k := {{.BitsFn}}(value)
	e := b.counts[k]
	e.value = value
	e.count++
	b.counts[k] = e
	b.size++
{{- else}}
	b.counts[value]++
	b.size++
{{- end}}
}

// AddOccurrences adds the given number of occurrences of the value.
// Returns the new count for this value. Panics if occurrences is negative.
func (b *{{.Name}}HashBag) AddOccurrences(value {{.GoType}}, occurrences int) int {
	if occurrences < 0 {
		panic(fmt.Sprintf("{{.Name}}HashBag: cannot add negative occurrences: %d", occurrences))
	}
{{- if .IsFloat}}
	k := {{.BitsFn}}(value)
	if occurrences == 0 {
		return b.counts[k].count
	}
	e := b.counts[k]
	e.value = value
	e.count += occurrences
	b.counts[k] = e
	b.size += occurrences
	return e.count
{{- else}}
	if occurrences == 0 {
		return b.counts[value]
	}
	b.counts[value] += occurrences
	b.size += occurrences
	return b.counts[value]
{{- end}}
}

// Remove removes one occurrence of the value. Returns true if the value was present.
func (b *{{.Name}}HashBag) Remove(value {{.GoType}}) bool {
{{- if .IsFloat}}
	k := {{.BitsFn}}(value)
	e, ok := b.counts[k]
	if !ok || e.count <= 0 {
		return false
	}
	if e.count == 1 {
		delete(b.counts, k)
	} else {
		e.count--
		b.counts[k] = e
	}
	b.size--
	return true
{{- else}}
	count, ok := b.counts[value]
	if !ok || count <= 0 {
		return false
	}
	if count == 1 {
		delete(b.counts, value)
	} else {
		b.counts[value] = count - 1
	}
	b.size--
	return true
{{- end}}
}

// RemoveOccurrences removes the given number of occurrences. Returns true if any were removed.
func (b *{{.Name}}HashBag) RemoveOccurrences(value {{.GoType}}, occurrences int) bool {
	if occurrences <= 0 {
		return false
	}
{{- if .IsFloat}}
	k := {{.BitsFn}}(value)
	e, ok := b.counts[k]
	if !ok || e.count <= 0 {
		return false
	}
	if occurrences >= e.count {
		delete(b.counts, k)
		b.size -= e.count
	} else {
		e.count -= occurrences
		b.counts[k] = e
		b.size -= occurrences
	}
	return true
{{- else}}
	count, ok := b.counts[value]
	if !ok || count <= 0 {
		return false
	}
	if occurrences >= count {
		delete(b.counts, value)
		b.size -= count
	} else {
		b.counts[value] = count - occurrences
		b.size -= occurrences
	}
	return true
{{- end}}
}

// RemoveAll removes all occurrences of the value. Returns the previous count.
func (b *{{.Name}}HashBag) RemoveAll(value {{.GoType}}) int {
{{- if .IsFloat}}
	k := {{.BitsFn}}(value)
	e, ok := b.counts[k]
	if !ok {
		return 0
	}
	delete(b.counts, k)
	b.size -= e.count
	return e.count
{{- else}}
	count, ok := b.counts[value]
	if !ok {
		return 0
	}
	delete(b.counts, value)
	b.size -= count
	return count
{{- end}}
}

// OccurrencesOf returns the number of occurrences of the value.
func (b *{{.Name}}HashBag) OccurrencesOf(value {{.GoType}}) int {
	return b.counts[{{if .IsFloat}}{{.BitsFn}}(value)].count{{else}}value]{{end}}
}

// Contains returns true if the bag contains at least one occurrence of the value.
func (b *{{.Name}}HashBag) Contains(value {{.GoType}}) bool {
	return b.counts[{{if .IsFloat}}{{.BitsFn}}(value)].count{{else}}value]{{end}} > 0
}

// Size returns the total number of elements including duplicates.
func (b *{{.Name}}HashBag) Size() int {
	return b.size
}

// SizeDistinct returns the number of distinct elements.
func (b *{{.Name}}HashBag) SizeDistinct() int {
	return len(b.counts)
}

// IsEmpty returns true if the bag contains no elements.
func (b *{{.Name}}HashBag) IsEmpty() bool {
	return b.size == 0
}

// Clear removes all elements from the bag.
func (b *{{.Name}}HashBag) Clear() {
	b.counts = make(map[{{if .IsFloat}}{{.BitsType}}]{{.SnakeName}}BagEntry{{else}}{{.GoType}}]int{{end}})
	b.size = 0
}

// All returns an iter.Seq that yields each element once per occurrence.
func (b *{{.Name}}HashBag) All() iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
{{- if .IsFloat}}
		for _, e := range b.counts {
			for i := 0; i < e.count; i++ {
				if !yield(e.value) {
					return
				}
			}
		}
{{- else}}
		for value, count := range b.counts {
			for i := 0; i < count; i++ {
				if !yield(value) {
					return
				}
			}
		}
{{- end}}
	}
}

// AllDistinct returns an iter.Seq that yields each distinct element once.
func (b *{{.Name}}HashBag) AllDistinct() iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
{{- if .IsFloat}}
		for _, e := range b.counts {
			if !yield(e.value) {
				return
			}
		}
{{- else}}
		for value := range b.counts {
			if !yield(value) {
				return
			}
		}
{{- end}}
	}
}

// AllWithOccurrences returns an iter.Seq2 that yields (value, count) pairs.
func (b *{{.Name}}HashBag) AllWithOccurrences() iter.Seq2[{{.GoType}}, int] {
	return func(yield func({{.GoType}}, int) bool) {
{{- if .IsFloat}}
		for _, e := range b.counts {
			if !yield(e.value, e.count) {
				return
			}
		}
{{- else}}
		for value, count := range b.counts {
			if !yield(value, count) {
				return
			}
		}
{{- end}}
	}
}

// ForEach calls the given function for each element (once per occurrence).
func (b *{{.Name}}HashBag) ForEach(f func({{.GoType}})) {
{{- if .IsFloat}}
	for _, e := range b.counts {
		for i := 0; i < e.count; i++ {
			f(e.value)
		}
	}
{{- else}}
	for value, count := range b.counts {
		for i := 0; i < count; i++ {
			f(value)
		}
	}
{{- end}}
}

// ForEachWithOccurrences calls the given function with each distinct element and its count.
func (b *{{.Name}}HashBag) ForEachWithOccurrences(f func({{.GoType}}, int)) {
{{- if .IsFloat}}
	for _, e := range b.counts {
		f(e.value, e.count)
	}
{{- else}}
	for value, count := range b.counts {
		f(value, count)
	}
{{- end}}
}

// Select returns a new bag containing only elements that satisfy the predicate.
func (b *{{.Name}}HashBag) Select(predicate func({{.GoType}}) bool) *{{.Name}}HashBag {
	result := New{{.Name}}HashBag()
{{- if .IsFloat}}
	for _, e := range b.counts {
		if predicate(e.value) {
			result.AddOccurrences(e.value, e.count)
		}
	}
{{- else}}
	for value, count := range b.counts {
		if predicate(value) {
			result.AddOccurrences(value, count)
		}
	}
{{- end}}
	return result
}

// Reject returns a new bag containing only elements that do not satisfy the predicate.
func (b *{{.Name}}HashBag) Reject(predicate func({{.GoType}}) bool) *{{.Name}}HashBag {
	result := New{{.Name}}HashBag()
{{- if .IsFloat}}
	for _, e := range b.counts {
		if !predicate(e.value) {
			result.AddOccurrences(e.value, e.count)
		}
	}
{{- else}}
	for value, count := range b.counts {
		if !predicate(value) {
			result.AddOccurrences(value, count)
		}
	}
{{- end}}
	return result
}

// Detect returns the first distinct element that satisfies the predicate, or zero value and false.
func (b *{{.Name}}HashBag) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
{{- if .IsFloat}}
	for _, e := range b.counts {
		if predicate(e.value) {
			return e.value, true
		}
	}
{{- else}}
	for value := range b.counts {
		if predicate(value) {
			return value, true
		}
	}
{{- end}}
	return {{.Zero}}, false
}

// AnySatisfy returns true if any element satisfies the predicate.
func (b *{{.Name}}HashBag) AnySatisfy(predicate func({{.GoType}}) bool) bool {
{{- if .IsFloat}}
	for _, e := range b.counts {
		if predicate(e.value) {
			return true
		}
	}
{{- else}}
	for value := range b.counts {
		if predicate(value) {
			return true
		}
	}
{{- end}}
	return false
}

// AllSatisfy returns true if all distinct elements satisfy the predicate.
func (b *{{.Name}}HashBag) AllSatisfy(predicate func({{.GoType}}) bool) bool {
{{- if .IsFloat}}
	for _, e := range b.counts {
		if !predicate(e.value) {
			return false
		}
	}
{{- else}}
	for value := range b.counts {
		if !predicate(value) {
			return false
		}
	}
{{- end}}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (b *{{.Name}}HashBag) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
{{- if .IsFloat}}
	for _, e := range b.counts {
		if predicate(e.value) {
			return false
		}
	}
{{- else}}
	for value := range b.counts {
		if predicate(value) {
			return false
		}
	}
{{- end}}
	return true
}

// TopOccurrences returns the n elements with the highest occurrence counts.
func (b *{{.Name}}HashBag) TopOccurrences(n int) []struct {
	Value {{.GoType}}
	Count int
} {
	type pair struct {
		Value {{.GoType}}
		Count int
	}
	pairs := make([]pair, 0, len(b.counts))
{{- if .IsFloat}}
	for _, e := range b.counts {
		pairs = append(pairs, pair{e.value, e.count})
	}
{{- else}}
	for value, count := range b.counts {
		pairs = append(pairs, pair{value, count})
	}
{{- end}}
	// Simple selection sort for top-n (good enough for typical use)
	for i := 0; i < n && i < len(pairs); i++ {
		maxIdx := i
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].Count > pairs[maxIdx].Count {
				maxIdx = j
			}
		}
		pairs[i], pairs[maxIdx] = pairs[maxIdx], pairs[i]
	}
	if n > len(pairs) {
		n = len(pairs)
	}
	result := make([]struct {
		Value {{.GoType}}
		Count int
	}, n)
	for i := 0; i < n; i++ {
		result[i].Value = pairs[i].Value
		result[i].Count = pairs[i].Count
	}
	return result
}

// ToSlice returns all elements as a slice (elements repeated per occurrence count).
func (b *{{.Name}}HashBag) ToSlice() []{{.GoType}} {
	result := make([]{{.GoType}}, 0, b.size)
{{- if .IsFloat}}
	for _, e := range b.counts {
		for i := 0; i < e.count; i++ {
			result = append(result, e.value)
		}
	}
{{- else}}
	for value, count := range b.counts {
		for i := 0; i < count; i++ {
			result = append(result, value)
		}
	}
{{- end}}
	return result
}

// With returns the bag after adding one occurrence of the value (fluent API).
func (b *{{.Name}}HashBag) With(value {{.GoType}}) *{{.Name}}HashBag {
	b.Add(value)
	return b
}

// Without returns the bag after removing all occurrences of the value (fluent API).
func (b *{{.Name}}HashBag) Without(value {{.GoType}}) *{{.Name}}HashBag {
	b.RemoveAll(value)
	return b
}

// String returns a string representation of the bag.
func (b *{{.Name}}HashBag) String() string {
	if b.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
{{- if .IsFloat}}
	for _, e := range b.counts {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v×%d", e.value, e.count)
		first = false
	}
{{- else}}
	for value, count := range b.counts {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v×%d", value, count)
		first = false
	}
{{- end}}
	sb.WriteString("}")
	return sb.String()
}

// WithAll returns the bag after adding all values (fluent API).
func (b *{{.Name}}HashBag) WithAll(values ...{{.GoType}}) *{{.Name}}HashBag {
	for _, v := range values {
		b.Add(v)
	}
	return b
}

// WithoutAll removes all occurrences of the given values.
func (b *{{.Name}}HashBag) WithoutAll(values ...{{.GoType}}) *{{.Name}}HashBag {
	for _, v := range values {
		b.RemoveAll(v)
	}
	return b
}

// ToImmutable returns an immutable copy of this bag.
func (b *{{.Name}}HashBag) ToImmutable() *Immutable{{.Name}}HashBag {
	return Immutable{{.Name}}HashBagFrom(b)
}

// Equals returns true if the other bag has the same elements with the same counts.
func (b *{{.Name}}HashBag) Equals(other *{{.Name}}HashBag) bool {
	if b.size != other.size || len(b.counts) != len(other.counts) {
		return false
	}
{{- if .IsFloat}}
	for k, e := range b.counts {
		if other.counts[k].count != e.count {
			return false
		}
	}
{{- else}}
	for value, count := range b.counts {
		if other.counts[value] != count {
			return false
		}
	}
{{- end}}
	return true
}
`

const immutableHashBagTmpl = genHeader + `package bag

import (
	"iter"
)

// Immutable{{.Name}}HashBag is an immutable view of a {{.Name}}HashBag.
type Immutable{{.Name}}HashBag struct {
	delegate *{{.Name}}HashBag
}

// NewImmutable{{.Name}}HashBag creates an immutable bag from the given values.
func NewImmutable{{.Name}}HashBag(values ...{{.GoType}}) *Immutable{{.Name}}HashBag {
	return &Immutable{{.Name}}HashBag{delegate: {{.Name}}HashBagOf(values...)}
}

// Immutable{{.Name}}HashBagFrom creates an immutable copy of a mutable bag.
func Immutable{{.Name}}HashBagFrom(b *{{.Name}}HashBag) *Immutable{{.Name}}HashBag {
	copy := New{{.Name}}HashBag()
	b.ForEachWithOccurrences(func(v {{.GoType}}, count int) {
		copy.AddOccurrences(v, count)
	})
	return &Immutable{{.Name}}HashBag{delegate: copy}
}

// OccurrencesOf returns the number of occurrences of the value.
func (b *Immutable{{.Name}}HashBag) OccurrencesOf(value {{.GoType}}) int {
	return b.delegate.OccurrencesOf(value)
}

// Contains returns true if the bag contains at least one occurrence.
func (b *Immutable{{.Name}}HashBag) Contains(value {{.GoType}}) bool {
	return b.delegate.Contains(value)
}

// Size returns the total count including duplicates.
func (b *Immutable{{.Name}}HashBag) Size() int {
	return b.delegate.Size()
}

// SizeDistinct returns the number of distinct elements.
func (b *Immutable{{.Name}}HashBag) SizeDistinct() int {
	return b.delegate.SizeDistinct()
}

// IsEmpty returns true if the bag contains no elements.
func (b *Immutable{{.Name}}HashBag) IsEmpty() bool {
	return b.delegate.IsEmpty()
}

// All returns an iter.Seq yielding each element once per occurrence.
func (b *Immutable{{.Name}}HashBag) All() iter.Seq[{{.GoType}}] {
	return b.delegate.All()
}

// AllDistinct returns an iter.Seq yielding each distinct element once.
func (b *Immutable{{.Name}}HashBag) AllDistinct() iter.Seq[{{.GoType}}] {
	return b.delegate.AllDistinct()
}

// AllWithOccurrences returns an iter.Seq2 yielding (value, count) pairs.
func (b *Immutable{{.Name}}HashBag) AllWithOccurrences() iter.Seq2[{{.GoType}}, int] {
	return b.delegate.AllWithOccurrences()
}

// ForEach calls the function for each element (once per occurrence).
func (b *Immutable{{.Name}}HashBag) ForEach(f func({{.GoType}})) {
	b.delegate.ForEach(f)
}

// ForEachWithOccurrences calls the function with each distinct element and its count.
func (b *Immutable{{.Name}}HashBag) ForEachWithOccurrences(f func({{.GoType}}, int)) {
	b.delegate.ForEachWithOccurrences(f)
}

// Select returns a new immutable bag with elements satisfying the predicate.
func (b *Immutable{{.Name}}HashBag) Select(predicate func({{.GoType}}) bool) *Immutable{{.Name}}HashBag {
	return &Immutable{{.Name}}HashBag{delegate: b.delegate.Select(predicate)}
}

// Reject returns a new immutable bag with elements not satisfying the predicate.
func (b *Immutable{{.Name}}HashBag) Reject(predicate func({{.GoType}}) bool) *Immutable{{.Name}}HashBag {
	return &Immutable{{.Name}}HashBag{delegate: b.delegate.Reject(predicate)}
}

// Detect returns the first distinct element satisfying the predicate, or zero and false.
func (b *Immutable{{.Name}}HashBag) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	return b.delegate.Detect(predicate)
}

// AnySatisfy returns true if any element satisfies the predicate.
func (b *Immutable{{.Name}}HashBag) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	return b.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (b *Immutable{{.Name}}HashBag) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	return b.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (b *Immutable{{.Name}}HashBag) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	return b.delegate.NoneSatisfy(predicate)
}

// TopOccurrences returns the n elements with the highest counts.
func (b *Immutable{{.Name}}HashBag) TopOccurrences(n int) []struct {
	Value {{.GoType}}
	Count int
} {
	return b.delegate.TopOccurrences(n)
}

// ToSlice returns all elements as a slice (repeated per count).
func (b *Immutable{{.Name}}HashBag) ToSlice() []{{.GoType}} {
	return b.delegate.ToSlice()
}

// String returns a string representation.
func (b *Immutable{{.Name}}HashBag) String() string {
	return b.delegate.String()
}

// Equals returns true if the other immutable bag has the same elements and counts.
func (b *Immutable{{.Name}}HashBag) Equals(other *Immutable{{.Name}}HashBag) bool {
	return b.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this bag.
func (b *Immutable{{.Name}}HashBag) ToMutable() *{{.Name}}HashBag {
	copy := New{{.Name}}HashBag()
	b.ForEachWithOccurrences(func(v {{.GoType}}, count int) {
		copy.AddOccurrences(v, count)
	})
	return copy
}
`

const synchronizedHashBagTmpl = genHeader + `package bag

import (
	"iter"
	"sync"
	"unsafe"
)

// Synchronized{{.Name}}HashBag is a thread-safe wrapper around {{.Name}}HashBag.
//
// Read methods hold an RLock; writes hold a Lock. Functional methods
// (ForEach/Select/Reject/AnySatisfy/…) snapshot (value, count) pairs
// under RLock, release, and run the callback against the snapshot so
// the callback may safely re-enter the wrapper.
type Synchronized{{.Name}}HashBag struct {
	delegate *{{.Name}}HashBag
	mu       sync.RWMutex
}

// NewSynchronized{{.Name}}HashBag creates a new thread-safe empty bag.
func NewSynchronized{{.Name}}HashBag() *Synchronized{{.Name}}HashBag {
	return &Synchronized{{.Name}}HashBag{delegate: New{{.Name}}HashBag()}
}

// NewSynchronized{{.Name}}HashBagFrom wraps an existing bag. The
// wrapper takes ownership — do not mutate the delegate directly.
func NewSynchronized{{.Name}}HashBagFrom(b *{{.Name}}HashBag) *Synchronized{{.Name}}HashBag {
	return &Synchronized{{.Name}}HashBag{delegate: b}
}

// snapshotDistinct returns (values, counts) for every distinct element,
// held only briefly under RLock.
func (b *Synchronized{{.Name}}HashBag) snapshotDistinct() (values []{{.GoType}}, counts []int) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for v, c := range b.delegate.AllWithOccurrences() {
		values = append(values, v)
		counts = append(counts, c)
	}
	return
}

// ── writes ────────────────────────────────────────────────────────────

func (b *Synchronized{{.Name}}HashBag) Add(value {{.GoType}}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.Add(value)
}

func (b *Synchronized{{.Name}}HashBag) AddOccurrences(value {{.GoType}}, occurrences int) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegate.AddOccurrences(value, occurrences)
}

func (b *Synchronized{{.Name}}HashBag) Remove(value {{.GoType}}) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegate.Remove(value)
}

func (b *Synchronized{{.Name}}HashBag) RemoveOccurrences(value {{.GoType}}, occurrences int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegate.RemoveOccurrences(value, occurrences)
}

func (b *Synchronized{{.Name}}HashBag) RemoveAll(value {{.GoType}}) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegate.RemoveAll(value)
}

func (b *Synchronized{{.Name}}HashBag) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.Clear()
}

// ── simple reads ──────────────────────────────────────────────────────

func (b *Synchronized{{.Name}}HashBag) OccurrencesOf(value {{.GoType}}) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.OccurrencesOf(value)
}

func (b *Synchronized{{.Name}}HashBag) Contains(value {{.GoType}}) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.Contains(value)
}

func (b *Synchronized{{.Name}}HashBag) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.Size()
}

func (b *Synchronized{{.Name}}HashBag) SizeDistinct() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.SizeDistinct()
}

func (b *Synchronized{{.Name}}HashBag) IsEmpty() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.IsEmpty()
}

func (b *Synchronized{{.Name}}HashBag) ToSlice() []{{.GoType}} {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.ToSlice()
}

func (b *Synchronized{{.Name}}HashBag) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.String()
}

// ── iteration (snapshot-based) ────────────────────────────────────────

// All yields every occurrence (multiplicity preserved).
func (b *Synchronized{{.Name}}HashBag) All() iter.Seq[{{.GoType}}] {
	values, counts := b.snapshotDistinct()
	return func(yield func({{.GoType}}) bool) {
		for i, v := range values {
			for j := 0; j < counts[i]; j++ {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// AllDistinct yields each distinct value exactly once.
func (b *Synchronized{{.Name}}HashBag) AllDistinct() iter.Seq[{{.GoType}}] {
	values, _ := b.snapshotDistinct()
	return func(yield func({{.GoType}}) bool) {
		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}

// AllWithOccurrences yields (value, count) pairs for each distinct value.
func (b *Synchronized{{.Name}}HashBag) AllWithOccurrences() iter.Seq2[{{.GoType}}, int] {
	values, counts := b.snapshotDistinct()
	return func(yield func({{.GoType}}, int) bool) {
		for i, v := range values {
			if !yield(v, counts[i]) {
				return
			}
		}
	}
}

// ── functional over snapshot ──────────────────────────────────────────

func (b *Synchronized{{.Name}}HashBag) ForEach(f func({{.GoType}})) {
	values, counts := b.snapshotDistinct()
	for i, v := range values {
		for j := 0; j < counts[i]; j++ {
			f(v)
		}
	}
}

func (b *Synchronized{{.Name}}HashBag) ForEachWithOccurrences(f func({{.GoType}}, int)) {
	values, counts := b.snapshotDistinct()
	for i, v := range values {
		f(v, counts[i])
	}
}

func (b *Synchronized{{.Name}}HashBag) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	values, _ := b.snapshotDistinct()
	for _, v := range values {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (b *Synchronized{{.Name}}HashBag) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	values, _ := b.snapshotDistinct()
	for _, v := range values {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (b *Synchronized{{.Name}}HashBag) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	values, _ := b.snapshotDistinct()
	for _, v := range values {
		if predicate(v) {
			return false
		}
	}
	return true
}

func (b *Synchronized{{.Name}}HashBag) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	values, _ := b.snapshotDistinct()
	for _, v := range values {
		if predicate(v) {
			return v, true
		}
	}
	var zero {{.GoType}}
	return zero, false
}

// ── functional that return new bags ──────────────────────────────────

func (b *Synchronized{{.Name}}HashBag) Select(predicate func({{.GoType}}) bool) *{{.Name}}HashBag {
	values, counts := b.snapshotDistinct()
	result := New{{.Name}}HashBag()
	for i, v := range values {
		if predicate(v) {
			result.AddOccurrences(v, counts[i])
		}
	}
	return result
}

func (b *Synchronized{{.Name}}HashBag) Reject(predicate func({{.GoType}}) bool) *{{.Name}}HashBag {
	values, counts := b.snapshotDistinct()
	result := New{{.Name}}HashBag()
	for i, v := range values {
		if !predicate(v) {
			result.AddOccurrences(v, counts[i])
		}
	}
	return result
}

// TopOccurrences returns the n most frequent elements. Returns the
// exact same shape as the underlying bag.
func (b *Synchronized{{.Name}}HashBag) TopOccurrences(n int) []struct {
	Value {{.GoType}}
	Count int
} {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.TopOccurrences(n)
}

// ── fluent mutators ───────────────────────────────────────────────────

func (b *Synchronized{{.Name}}HashBag) With(value {{.GoType}}) *Synchronized{{.Name}}HashBag {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.With(value)
	return b
}

func (b *Synchronized{{.Name}}HashBag) WithAll(values ...{{.GoType}}) *Synchronized{{.Name}}HashBag {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.WithAll(values...)
	return b
}

func (b *Synchronized{{.Name}}HashBag) Without(value {{.GoType}}) *Synchronized{{.Name}}HashBag {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.Without(value)
	return b
}

func (b *Synchronized{{.Name}}HashBag) WithoutAll(values ...{{.GoType}}) *Synchronized{{.Name}}HashBag {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.WithoutAll(values...)
	return b
}

// ── conversions & equals ──────────────────────────────────────────────

func (b *Synchronized{{.Name}}HashBag) ToImmutable() *Immutable{{.Name}}HashBag {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.ToImmutable()
}

// Equals compares by contents. Locks acquired in pointer-address
// order to prevent A.Equals(B) / B.Equals(A) deadlocks.
func (b *Synchronized{{.Name}}HashBag) Equals(other *Synchronized{{.Name}}HashBag) bool {
	if b == other {
		b.mu.RLock()
		defer b.mu.RUnlock()
		return b.delegate.Equals(other.delegate)
	}
	first, second := b, other
	if uintptr(unsafe.Pointer(b)) > uintptr(unsafe.Pointer(other)) {
		first, second = other, b
	}
	first.mu.RLock()
	defer first.mu.RUnlock()
	second.mu.RLock()
	defer second.mu.RUnlock()
	return b.delegate.Equals(other.delegate)
}
`

const treeBagTmpl = genHeader + `package bag

import (
	"fmt"
	"iter"
	"sort"
	"strings"
)

// {{.Name}}TreeBagEntry holds a value and its occurrence count in a {{.Name}}TreeBag.
type {{.Name}}TreeBagEntry struct {
	value {{.GoType}}
	count int
}

// {{.Name}}TreeBag is a sorted bag (multiset) that counts occurrences of {{.GoType}} values.
// Backed by a sorted slice of entries with binary search for O(log n) lookup.
type {{.Name}}TreeBag struct {
	entries []{{.Name}}TreeBagEntry
	size    int // total count including duplicates
}

// New{{.Name}}TreeBag creates a new empty {{.Name}}TreeBag.
func New{{.Name}}TreeBag() *{{.Name}}TreeBag {
	return &{{.Name}}TreeBag{
		entries: nil,
		size:    0,
	}
}

// {{.Name}}TreeBagOf creates a new {{.Name}}TreeBag from the given values.
func {{.Name}}TreeBagOf(values ...{{.GoType}}) *{{.Name}}TreeBag {
	b := New{{.Name}}TreeBag()
	for _, v := range values {
		b.Add(v)
	}
	return b
}

// search returns the index where value is or would be inserted.
// Returns (index, found).
func (b *{{.Name}}TreeBag) search(value {{.GoType}}) (int, bool) {
	lo, hi := 0, len(b.entries)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if {{if .IsFloat}}{{.CmpFn}}(b.entries[mid].value, value) < 0{{else}}b.entries[mid].value < value{{end}} {
			lo = mid + 1
		} else if {{if .IsFloat}}{{.CmpFn}}(b.entries[mid].value, value) > 0{{else}}b.entries[mid].value > value{{end}} {
			hi = mid
		} else {
			return mid, true
		}
	}
	return lo, false
}

// Add adds one occurrence of the value.
func (b *{{.Name}}TreeBag) Add(value {{.GoType}}) {
	idx, found := b.search(value)
	if found {
		b.entries[idx].count++
		b.size++
		return
	}
	// Insert at idx to keep sorted order
	b.entries = append(b.entries, {{.Name}}TreeBagEntry{})
	copy(b.entries[idx+1:], b.entries[idx:])
	b.entries[idx] = {{.Name}}TreeBagEntry{value: value, count: 1}
	b.size++
}

// AddOccurrences adds the given number of occurrences of the value.
// Returns the new count for this value. Panics if occurrences is negative.
func (b *{{.Name}}TreeBag) AddOccurrences(value {{.GoType}}, occurrences int) int {
	if occurrences < 0 {
		panic(fmt.Sprintf("{{.Name}}TreeBag: cannot add negative occurrences: %d", occurrences))
	}
	if occurrences == 0 {
		idx, found := b.search(value)
		if found {
			return b.entries[idx].count
		}
		return 0
	}
	idx, found := b.search(value)
	if found {
		b.entries[idx].count += occurrences
		b.size += occurrences
		return b.entries[idx].count
	}
	b.entries = append(b.entries, {{.Name}}TreeBagEntry{})
	copy(b.entries[idx+1:], b.entries[idx:])
	b.entries[idx] = {{.Name}}TreeBagEntry{value: value, count: occurrences}
	b.size += occurrences
	return occurrences
}

// Remove removes one occurrence of the value. Returns true if the value was present.
func (b *{{.Name}}TreeBag) Remove(value {{.GoType}}) bool {
	idx, found := b.search(value)
	if !found {
		return false
	}
	if b.entries[idx].count == 1 {
		b.entries = append(b.entries[:idx], b.entries[idx+1:]...)
	} else {
		b.entries[idx].count--
	}
	b.size--
	return true
}

// RemoveOccurrences removes the given number of occurrences. Returns true if any were removed.
func (b *{{.Name}}TreeBag) RemoveOccurrences(value {{.GoType}}, occurrences int) bool {
	if occurrences <= 0 {
		return false
	}
	idx, found := b.search(value)
	if !found {
		return false
	}
	if occurrences >= b.entries[idx].count {
		b.size -= b.entries[idx].count
		b.entries = append(b.entries[:idx], b.entries[idx+1:]...)
	} else {
		b.entries[idx].count -= occurrences
		b.size -= occurrences
	}
	return true
}

// RemoveAll removes all occurrences of the value. Returns the previous count.
func (b *{{.Name}}TreeBag) RemoveAll(value {{.GoType}}) int {
	idx, found := b.search(value)
	if !found {
		return 0
	}
	count := b.entries[idx].count
	b.entries = append(b.entries[:idx], b.entries[idx+1:]...)
	b.size -= count
	return count
}

// OccurrencesOf returns the number of occurrences of the value.
func (b *{{.Name}}TreeBag) OccurrencesOf(value {{.GoType}}) int {
	idx, found := b.search(value)
	if !found {
		return 0
	}
	return b.entries[idx].count
}

// Contains returns true if the bag contains at least one occurrence of the value.
func (b *{{.Name}}TreeBag) Contains(value {{.GoType}}) bool {
	_, found := b.search(value)
	return found
}

// Size returns the total number of elements including duplicates.
func (b *{{.Name}}TreeBag) Size() int {
	return b.size
}

// SizeDistinct returns the number of distinct elements.
func (b *{{.Name}}TreeBag) SizeDistinct() int {
	return len(b.entries)
}

// IsEmpty returns true if the bag contains no elements.
func (b *{{.Name}}TreeBag) IsEmpty() bool {
	return b.size == 0
}

// Clear removes all elements from the bag.
func (b *{{.Name}}TreeBag) Clear() {
	b.entries = nil
	b.size = 0
}

// Min returns the smallest element, or zero value and false if empty.
func (b *{{.Name}}TreeBag) Min() ({{.GoType}}, bool) {
	if len(b.entries) == 0 {
		return {{.Zero}}, false
	}
	return b.entries[0].value, true
}

// Max returns the largest element, or zero value and false if empty.
func (b *{{.Name}}TreeBag) Max() ({{.GoType}}, bool) {
	if len(b.entries) == 0 {
		return {{.Zero}}, false
	}
	return b.entries[len(b.entries)-1].value, true
}

// All returns an iter.Seq that yields each element once per occurrence, in sorted order.
func (b *{{.Name}}TreeBag) All() iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
		for _, entry := range b.entries {
			for i := 0; i < entry.count; i++ {
				if !yield(entry.value) {
					return
				}
			}
		}
	}
}

// AllDistinct returns an iter.Seq that yields each distinct element once, in sorted order.
func (b *{{.Name}}TreeBag) AllDistinct() iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
		for _, entry := range b.entries {
			if !yield(entry.value) {
				return
			}
		}
	}
}

// AllWithOccurrences returns an iter.Seq2 that yields (value, count) pairs in sorted order.
func (b *{{.Name}}TreeBag) AllWithOccurrences() iter.Seq2[{{.GoType}}, int] {
	return func(yield func({{.GoType}}, int) bool) {
		for _, entry := range b.entries {
			if !yield(entry.value, entry.count) {
				return
			}
		}
	}
}

// ForEach calls the given function for each element (once per occurrence), in sorted order.
func (b *{{.Name}}TreeBag) ForEach(f func({{.GoType}})) {
	for _, entry := range b.entries {
		for i := 0; i < entry.count; i++ {
			f(entry.value)
		}
	}
}

// ForEachWithOccurrences calls the given function with each distinct element and its count, in sorted order.
func (b *{{.Name}}TreeBag) ForEachWithOccurrences(f func({{.GoType}}, int)) {
	for _, entry := range b.entries {
		f(entry.value, entry.count)
	}
}

// Select returns a new tree bag containing only elements that satisfy the predicate.
func (b *{{.Name}}TreeBag) Select(predicate func({{.GoType}}) bool) *{{.Name}}TreeBag {
	result := New{{.Name}}TreeBag()
	for _, entry := range b.entries {
		if predicate(entry.value) {
			result.AddOccurrences(entry.value, entry.count)
		}
	}
	return result
}

// Reject returns a new tree bag containing only elements that do not satisfy the predicate.
func (b *{{.Name}}TreeBag) Reject(predicate func({{.GoType}}) bool) *{{.Name}}TreeBag {
	result := New{{.Name}}TreeBag()
	for _, entry := range b.entries {
		if !predicate(entry.value) {
			result.AddOccurrences(entry.value, entry.count)
		}
	}
	return result
}

// Detect returns the first distinct element (in sorted order) that satisfies the predicate, or zero value and false.
func (b *{{.Name}}TreeBag) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	for _, entry := range b.entries {
		if predicate(entry.value) {
			return entry.value, true
		}
	}
	return {{.Zero}}, false
}

// AnySatisfy returns true if any distinct element satisfies the predicate.
func (b *{{.Name}}TreeBag) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for _, entry := range b.entries {
		if predicate(entry.value) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all distinct elements satisfy the predicate.
func (b *{{.Name}}TreeBag) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, entry := range b.entries {
		if !predicate(entry.value) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no distinct element satisfies the predicate.
func (b *{{.Name}}TreeBag) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, entry := range b.entries {
		if predicate(entry.value) {
			return false
		}
	}
	return true
}

// TopOccurrences returns the n elements with the highest occurrence counts.
func (b *{{.Name}}TreeBag) TopOccurrences(n int) []struct {
	Value {{.GoType}}
	Count int
} {
	// Copy entries and sort by count descending
	sorted := make([]{{.Name}}TreeBagEntry, len(b.entries))
	copy(sorted, b.entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})
	if n > len(sorted) {
		n = len(sorted)
	}
	result := make([]struct {
		Value {{.GoType}}
		Count int
	}, n)
	for i := 0; i < n; i++ {
		result[i].Value = sorted[i].value
		result[i].Count = sorted[i].count
	}
	return result
}

// ToSlice returns all elements as a slice (elements repeated per occurrence count), in sorted order.
func (b *{{.Name}}TreeBag) ToSlice() []{{.GoType}} {
	result := make([]{{.GoType}}, 0, b.size)
	for _, entry := range b.entries {
		for i := 0; i < entry.count; i++ {
			result = append(result, entry.value)
		}
	}
	return result
}

// ToSortedSlice returns all distinct elements as a sorted slice.
func (b *{{.Name}}TreeBag) ToSortedSlice() []{{.GoType}} {
	result := make([]{{.GoType}}, 0, len(b.entries))
	for _, entry := range b.entries {
		result = append(result, entry.value)
	}
	return result
}

// With returns the bag after adding one occurrence of the value (fluent API).
func (b *{{.Name}}TreeBag) With(value {{.GoType}}) *{{.Name}}TreeBag {
	b.Add(value)
	return b
}

// Without returns the bag after removing all occurrences of the value (fluent API).
func (b *{{.Name}}TreeBag) Without(value {{.GoType}}) *{{.Name}}TreeBag {
	b.RemoveAll(value)
	return b
}

// WithAll returns the bag after adding all values (fluent API).
func (b *{{.Name}}TreeBag) WithAll(values ...{{.GoType}}) *{{.Name}}TreeBag {
	for _, v := range values {
		b.Add(v)
	}
	return b
}

// WithoutAll removes all occurrences of the given values (fluent API).
func (b *{{.Name}}TreeBag) WithoutAll(values ...{{.GoType}}) *{{.Name}}TreeBag {
	for _, v := range values {
		b.RemoveAll(v)
	}
	return b
}

// String returns a string representation of the bag in sorted order.
func (b *{{.Name}}TreeBag) String() string {
	if b.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for _, entry := range b.entries {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v\u00d7%d", entry.value, entry.count)
		first = false
	}
	sb.WriteString("}")
	return sb.String()
}

// Equals returns true if the other tree bag has the same elements with the same counts.
func (b *{{.Name}}TreeBag) Equals(other *{{.Name}}TreeBag) bool {
	if b.size != other.size || len(b.entries) != len(other.entries) {
		return false
	}
	for i, entry := range b.entries {
		if entry.value != other.entries[i].value || entry.count != other.entries[i].count {
			return false
		}
	}
	return true
}
`
