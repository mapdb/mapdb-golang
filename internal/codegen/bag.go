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
// The TREE bag differs in the search comparison (raw < / > for int/char,
// cmpFloatNN for floats), the Equals element comparison (raw == for int/char,
// bit-pattern BitsFn for floats so NaN entries match and ±0 stay distinct), and
// the Zero literal used in Min/Max/Detect.
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

	"github.com/mapdb/mapdb-golang/pump"
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

// Hash{{.Name}} is a bag (multiset) that counts occurrences of {{.GoType}} values.
// Backed by a map from value bit pattern to (value, count).
type Hash{{.Name}} struct {
	counts map[{{.BitsType}}]{{.SnakeName}}BagEntry
	size   int // total count including duplicates
}
{{else}}
// Hash{{.Name}} is a bag (multiset) that counts occurrences of {{.GoType}} values.
// Backed by a map from value to count.
type Hash{{.Name}} struct {
	counts map[{{.GoType}}]int
	size   int // total count including duplicates
}
{{end}}
// NewHash{{.Name}} creates a new empty Hash{{.Name}}.
func NewHash{{.Name}}() *Hash{{.Name}} {
	return &Hash{{.Name}}{
		counts: make(map[{{if .IsFloat}}{{.BitsType}}]{{.SnakeName}}BagEntry{{else}}{{.GoType}}]int{{end}}),
		size:   0,
	}
}

// Hash{{.Name}}Of creates a new Hash{{.Name}} from the given values.
func Hash{{.Name}}Of(values ...{{.GoType}}) *Hash{{.Name}} {
	b := NewHash{{.Name}}()
	for _, v := range values {
		b.Add(v)
	}
	return b
}

// Hash{{.Name}}BulkLoad builds a Hash{{.Name}} from values in a single pass,
// presizing the backing count map for len(values). Duplicate values are counted
// as occurrences; bag multiplicity is not a duplicate-key error.
func Hash{{.Name}}BulkLoad(values []{{.GoType}}) *Hash{{.Name}} {
	b := &Hash{{.Name}}{
		counts: make(map[{{if .IsFloat}}{{.BitsType}}]{{.SnakeName}}BagEntry{{else}}{{.GoType}}]int{{end}}, pump.HashCapacityFor(len(values))),
	}
	for _, v := range values {
		b.Add(v)
	}
	return b
}

// Hash{{.Name}}BulkLoadExact is like Hash{{.Name}}BulkLoad but presizes from an
// exact total occurrence count and fails if values contains more than n items.
func Hash{{.Name}}BulkLoadExact(values []{{.GoType}}, n int) (*Hash{{.Name}}, error) {
	if n < 0 {
		panic("mapdb: Hash{{.Name}}BulkLoadExact: negative n")
	}
	if len(values) > n {
		return nil, pump.ErrTooManyElements
	}
	b := &Hash{{.Name}}{
		counts: make(map[{{if .IsFloat}}{{.BitsType}}]{{.SnakeName}}BagEntry{{else}}{{.GoType}}]int{{end}}, pump.HashCapacityFor(n)),
	}
	for _, v := range values {
		b.Add(v)
	}
	return b, nil
}

// Add adds one occurrence of the value.
func (b *Hash{{.Name}}) Add(value {{.GoType}}) {
	if b.counts == nil {
		b.counts = make(map[{{if .IsFloat}}{{.BitsType}}]{{.SnakeName}}BagEntry{{else}}{{.GoType}}]int{{end}})
	}
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
func (b *Hash{{.Name}}) AddOccurrences(value {{.GoType}}, occurrences int) int {
	if occurrences < 0 {
		panic(fmt.Sprintf("Hash{{.Name}}: cannot add negative occurrences: %d", occurrences))
	}
{{- if .IsFloat}}
	k := {{.BitsFn}}(value)
	if occurrences == 0 {
		return b.counts[k].count
	}
	if b.counts == nil {
		b.counts = make(map[{{.BitsType}}]{{.SnakeName}}BagEntry)
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
	if b.counts == nil {
		b.counts = make(map[{{.GoType}}]int)
	}
	b.counts[value] += occurrences
	b.size += occurrences
	return b.counts[value]
{{- end}}
}

// Remove removes one occurrence of the value. Returns true if the value was present.
func (b *Hash{{.Name}}) Remove(value {{.GoType}}) bool {
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
func (b *Hash{{.Name}}) RemoveOccurrences(value {{.GoType}}, occurrences int) bool {
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
func (b *Hash{{.Name}}) RemoveAll(value {{.GoType}}) int {
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
func (b *Hash{{.Name}}) OccurrencesOf(value {{.GoType}}) int {
	return b.counts[{{if .IsFloat}}{{.BitsFn}}(value)].count{{else}}value]{{end}}
}

// Contains returns true if the bag contains at least one occurrence of the value.
func (b *Hash{{.Name}}) Contains(value {{.GoType}}) bool {
	return b.counts[{{if .IsFloat}}{{.BitsFn}}(value)].count{{else}}value]{{end}} > 0
}

// Len returns the total number of elements including duplicates.
func (b *Hash{{.Name}}) Len() int {
	return b.size
}

// Len returns the number of elements. It is an alias for Size, matching
// Go convention (sort.Interface, container/list, bytes.Buffer).

// SizeDistinct returns the number of distinct elements.
func (b *Hash{{.Name}}) SizeDistinct() int {
	return len(b.counts)
}


// Clear removes all elements from the bag.
func (b *Hash{{.Name}}) Clear() {
	b.counts = make(map[{{if .IsFloat}}{{.BitsType}}]{{.SnakeName}}BagEntry{{else}}{{.GoType}}]int{{end}})
	b.size = 0
}

// All returns an iter.Seq that yields each element once per occurrence.
func (b *Hash{{.Name}}) All() iter.Seq[{{.GoType}}] {
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
func (b *Hash{{.Name}}) AllDistinct() iter.Seq[{{.GoType}}] {
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
func (b *Hash{{.Name}}) AllWithOccurrences() iter.Seq2[{{.GoType}}, int] {
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
func (b *Hash{{.Name}}) ForEach(f func({{.GoType}})) {
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
func (b *Hash{{.Name}}) ForEachWithOccurrences(f func({{.GoType}}, int)) {
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
func (b *Hash{{.Name}}) Select(predicate func({{.GoType}}) bool) *Hash{{.Name}} {
	result := NewHash{{.Name}}()
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
func (b *Hash{{.Name}}) Reject(predicate func({{.GoType}}) bool) *Hash{{.Name}} {
	result := NewHash{{.Name}}()
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
func (b *Hash{{.Name}}) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
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
func (b *Hash{{.Name}}) AnySatisfy(predicate func({{.GoType}}) bool) bool {
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
func (b *Hash{{.Name}}) AllSatisfy(predicate func({{.GoType}}) bool) bool {
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
func (b *Hash{{.Name}}) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
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
func (b *Hash{{.Name}}) TopOccurrences(n int) []struct {
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
func (b *Hash{{.Name}}) ToSlice() []{{.GoType}} {
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
func (b *Hash{{.Name}}) AddReturning(value {{.GoType}}) *Hash{{.Name}} {
	b.Add(value)
	return b
}

// Without returns the bag after removing all occurrences of the value (fluent API).
func (b *Hash{{.Name}}) RemoveReturning(value {{.GoType}}) *Hash{{.Name}} {
	b.RemoveAll(value)
	return b
}

// String returns a string representation of the bag.
func (b *Hash{{.Name}}) String() string {
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
func (b *Hash{{.Name}}) AddAllReturning(values ...{{.GoType}}) *Hash{{.Name}} {
	for _, v := range values {
		b.Add(v)
	}
	return b
}

// WithoutAll removes all occurrences of the given values.
func (b *Hash{{.Name}}) RemoveAllReturning(values ...{{.GoType}}) *Hash{{.Name}} {
	for _, v := range values {
		b.RemoveAll(v)
	}
	return b
}

// ToImmutable returns an immutable copy of this bag.
func (b *Hash{{.Name}}) ToImmutable() *ImmutableHash{{.Name}} {
	return ImmutableHash{{.Name}}From(b)
}

// Equals returns true if the other bag has the same elements with the same counts.
func (b *Hash{{.Name}}) Equals(other *Hash{{.Name}}) bool {
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

// ImmutableHash{{.Name}} is an immutable view of a Hash{{.Name}}.
type ImmutableHash{{.Name}} struct {
	delegate *Hash{{.Name}}
}

// NewImmutableHash{{.Name}} creates an immutable bag from the given values.
func NewImmutableHash{{.Name}}(values ...{{.GoType}}) *ImmutableHash{{.Name}} {
	return &ImmutableHash{{.Name}}{delegate: Hash{{.Name}}Of(values...)}
}

// ImmutableHash{{.Name}}From creates an immutable copy of a mutable bag.
func ImmutableHash{{.Name}}From(b *Hash{{.Name}}) *ImmutableHash{{.Name}} {
	copy := NewHash{{.Name}}()
	b.ForEachWithOccurrences(func(v {{.GoType}}, count int) {
		copy.AddOccurrences(v, count)
	})
	return &ImmutableHash{{.Name}}{delegate: copy}
}

// OccurrencesOf returns the number of occurrences of the value.
func (b *ImmutableHash{{.Name}}) OccurrencesOf(value {{.GoType}}) int {
	return b.delegate.OccurrencesOf(value)
}

// Contains returns true if the bag contains at least one occurrence.
func (b *ImmutableHash{{.Name}}) Contains(value {{.GoType}}) bool {
	return b.delegate.Contains(value)
}

// Size returns the total count including duplicates.
func (b *ImmutableHash{{.Name}}) Len() int {
	return b.delegate.Len()
}

// Len returns the number of elements. It is an alias for Size, matching
// Go convention (sort.Interface, container/list, bytes.Buffer).

// SizeDistinct returns the number of distinct elements.
func (b *ImmutableHash{{.Name}}) SizeDistinct() int {
	return b.delegate.SizeDistinct()
}


// All returns an iter.Seq yielding each element once per occurrence.
func (b *ImmutableHash{{.Name}}) All() iter.Seq[{{.GoType}}] {
	return b.delegate.All()
}

// AllDistinct returns an iter.Seq yielding each distinct element once.
func (b *ImmutableHash{{.Name}}) AllDistinct() iter.Seq[{{.GoType}}] {
	return b.delegate.AllDistinct()
}

// AllWithOccurrences returns an iter.Seq2 yielding (value, count) pairs.
func (b *ImmutableHash{{.Name}}) AllWithOccurrences() iter.Seq2[{{.GoType}}, int] {
	return b.delegate.AllWithOccurrences()
}

// ForEach calls the function for each element (once per occurrence).
func (b *ImmutableHash{{.Name}}) ForEach(f func({{.GoType}})) {
	b.delegate.ForEach(f)
}

// ForEachWithOccurrences calls the function with each distinct element and its count.
func (b *ImmutableHash{{.Name}}) ForEachWithOccurrences(f func({{.GoType}}, int)) {
	b.delegate.ForEachWithOccurrences(f)
}

// Select returns a new immutable bag with elements satisfying the predicate.
func (b *ImmutableHash{{.Name}}) Select(predicate func({{.GoType}}) bool) *ImmutableHash{{.Name}} {
	return &ImmutableHash{{.Name}}{delegate: b.delegate.Select(predicate)}
}

// Reject returns a new immutable bag with elements not satisfying the predicate.
func (b *ImmutableHash{{.Name}}) Reject(predicate func({{.GoType}}) bool) *ImmutableHash{{.Name}} {
	return &ImmutableHash{{.Name}}{delegate: b.delegate.Reject(predicate)}
}

// Detect returns the first distinct element satisfying the predicate, or zero and false.
func (b *ImmutableHash{{.Name}}) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	return b.delegate.Detect(predicate)
}

// AnySatisfy returns true if any element satisfies the predicate.
func (b *ImmutableHash{{.Name}}) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	return b.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (b *ImmutableHash{{.Name}}) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	return b.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (b *ImmutableHash{{.Name}}) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	return b.delegate.NoneSatisfy(predicate)
}

// TopOccurrences returns the n elements with the highest counts.
func (b *ImmutableHash{{.Name}}) TopOccurrences(n int) []struct {
	Value {{.GoType}}
	Count int
} {
	return b.delegate.TopOccurrences(n)
}

// ToSlice returns all elements as a slice (repeated per count).
func (b *ImmutableHash{{.Name}}) ToSlice() []{{.GoType}} {
	return b.delegate.ToSlice()
}

// String returns a string representation.
func (b *ImmutableHash{{.Name}}) String() string {
	return b.delegate.String()
}

// Equals returns true if the other immutable bag has the same elements and counts.
func (b *ImmutableHash{{.Name}}) Equals(other *ImmutableHash{{.Name}}) bool {
	return b.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this bag.
func (b *ImmutableHash{{.Name}}) ToMutable() *Hash{{.Name}} {
	copy := NewHash{{.Name}}()
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

// SynchronizedHash{{.Name}} is a thread-safe wrapper around Hash{{.Name}}.
//
// Read methods hold an RLock; writes hold a Lock. Functional methods
// (ForEach/Select/Reject/AnySatisfy/…) snapshot (value, count) pairs
// under RLock, release, and run the callback against the snapshot so
// the callback may safely re-enter the wrapper.
type SynchronizedHash{{.Name}} struct {
	delegate *Hash{{.Name}}
	mu       sync.RWMutex
}

// NewSynchronizedHash{{.Name}} creates a new thread-safe empty bag.
func NewSynchronizedHash{{.Name}}() *SynchronizedHash{{.Name}} {
	return &SynchronizedHash{{.Name}}{delegate: NewHash{{.Name}}()}
}

// NewSynchronizedHash{{.Name}}From wraps an existing bag. The
// wrapper takes ownership — do not mutate the delegate directly.
func NewSynchronizedHash{{.Name}}From(b *Hash{{.Name}}) *SynchronizedHash{{.Name}} {
	return &SynchronizedHash{{.Name}}{delegate: b}
}

// snapshotDistinct returns (values, counts) for every distinct element,
// held only briefly under RLock.
func (b *SynchronizedHash{{.Name}}) snapshotDistinct() (values []{{.GoType}}, counts []int) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for v, c := range b.delegate.AllWithOccurrences() {
		values = append(values, v)
		counts = append(counts, c)
	}
	return
}

// ── writes ────────────────────────────────────────────────────────────

func (b *SynchronizedHash{{.Name}}) Add(value {{.GoType}}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.Add(value)
}

func (b *SynchronizedHash{{.Name}}) AddOccurrences(value {{.GoType}}, occurrences int) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegate.AddOccurrences(value, occurrences)
}

func (b *SynchronizedHash{{.Name}}) Remove(value {{.GoType}}) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegate.Remove(value)
}

func (b *SynchronizedHash{{.Name}}) RemoveOccurrences(value {{.GoType}}, occurrences int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegate.RemoveOccurrences(value, occurrences)
}

func (b *SynchronizedHash{{.Name}}) RemoveAll(value {{.GoType}}) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegate.RemoveAll(value)
}

func (b *SynchronizedHash{{.Name}}) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.Clear()
}

// ── simple reads ──────────────────────────────────────────────────────

func (b *SynchronizedHash{{.Name}}) OccurrencesOf(value {{.GoType}}) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.OccurrencesOf(value)
}

func (b *SynchronizedHash{{.Name}}) Contains(value {{.GoType}}) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.Contains(value)
}

func (b *SynchronizedHash{{.Name}}) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.Len()
}

// Len returns the number of elements. It is an alias for Size, matching
// Go convention (sort.Interface, container/list, bytes.Buffer).

func (b *SynchronizedHash{{.Name}}) SizeDistinct() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.SizeDistinct()
}


func (b *SynchronizedHash{{.Name}}) ToSlice() []{{.GoType}} {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.ToSlice()
}

func (b *SynchronizedHash{{.Name}}) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.String()
}

// ── iteration (snapshot-based) ────────────────────────────────────────

// All yields every occurrence (multiplicity preserved).
func (b *SynchronizedHash{{.Name}}) All() iter.Seq[{{.GoType}}] {
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
func (b *SynchronizedHash{{.Name}}) AllDistinct() iter.Seq[{{.GoType}}] {
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
func (b *SynchronizedHash{{.Name}}) AllWithOccurrences() iter.Seq2[{{.GoType}}, int] {
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

func (b *SynchronizedHash{{.Name}}) ForEach(f func({{.GoType}})) {
	values, counts := b.snapshotDistinct()
	for i, v := range values {
		for j := 0; j < counts[i]; j++ {
			f(v)
		}
	}
}

func (b *SynchronizedHash{{.Name}}) ForEachWithOccurrences(f func({{.GoType}}, int)) {
	values, counts := b.snapshotDistinct()
	for i, v := range values {
		f(v, counts[i])
	}
}

func (b *SynchronizedHash{{.Name}}) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	values, _ := b.snapshotDistinct()
	for _, v := range values {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (b *SynchronizedHash{{.Name}}) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	values, _ := b.snapshotDistinct()
	for _, v := range values {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (b *SynchronizedHash{{.Name}}) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	values, _ := b.snapshotDistinct()
	for _, v := range values {
		if predicate(v) {
			return false
		}
	}
	return true
}

func (b *SynchronizedHash{{.Name}}) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
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

func (b *SynchronizedHash{{.Name}}) Select(predicate func({{.GoType}}) bool) *Hash{{.Name}} {
	values, counts := b.snapshotDistinct()
	result := NewHash{{.Name}}()
	for i, v := range values {
		if predicate(v) {
			result.AddOccurrences(v, counts[i])
		}
	}
	return result
}

func (b *SynchronizedHash{{.Name}}) Reject(predicate func({{.GoType}}) bool) *Hash{{.Name}} {
	values, counts := b.snapshotDistinct()
	result := NewHash{{.Name}}()
	for i, v := range values {
		if !predicate(v) {
			result.AddOccurrences(v, counts[i])
		}
	}
	return result
}

// TopOccurrences returns the n most frequent elements. Returns the
// exact same shape as the underlying bag.
func (b *SynchronizedHash{{.Name}}) TopOccurrences(n int) []struct {
	Value {{.GoType}}
	Count int
} {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.TopOccurrences(n)
}

// ── fluent mutators ───────────────────────────────────────────────────

func (b *SynchronizedHash{{.Name}}) AddReturning(value {{.GoType}}) *SynchronizedHash{{.Name}} {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.AddReturning(value)
	return b
}

func (b *SynchronizedHash{{.Name}}) AddAllReturning(values ...{{.GoType}}) *SynchronizedHash{{.Name}} {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.AddAllReturning(values...)
	return b
}

func (b *SynchronizedHash{{.Name}}) RemoveReturning(value {{.GoType}}) *SynchronizedHash{{.Name}} {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.RemoveReturning(value)
	return b
}

func (b *SynchronizedHash{{.Name}}) RemoveAllReturning(values ...{{.GoType}}) *SynchronizedHash{{.Name}} {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.RemoveAllReturning(values...)
	return b
}

// ── conversions & equals ──────────────────────────────────────────────

func (b *SynchronizedHash{{.Name}}) ToImmutable() *ImmutableHash{{.Name}} {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.ToImmutable()
}

// Equals compares by contents. Locks acquired in pointer-address
// order to prevent A.Equals(B) / B.Equals(A) deadlocks.
func (b *SynchronizedHash{{.Name}}) Equals(other *SynchronizedHash{{.Name}}) bool {
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
	"cmp"
	"fmt"
	"iter"
{{- if .IsFloat}}
	"math"
{{- end}}
	"slices"
	"strings"

	"github.com/mapdb/mapdb-golang/pump"
)

// Tree{{.Name}}Entry holds a value and its occurrence count in a Tree{{.Name}}.
type Tree{{.Name}}Entry struct {
	value {{.GoType}}
	count int
}

// Tree{{.Name}} is a sorted bag (multiset) that counts occurrences of {{.GoType}} values.
// Backed by a sorted slice of entries with binary search for O(log n) lookup.
type Tree{{.Name}} struct {
	entries []Tree{{.Name}}Entry
	size    int // total count including duplicates
}

// NewTree{{.Name}} creates a new empty Tree{{.Name}}.
func NewTree{{.Name}}() *Tree{{.Name}} {
	return &Tree{{.Name}}{
		entries: nil,
		size:    0,
	}
}

// Tree{{.Name}}Of creates a new Tree{{.Name}} from the given values. It sorts a
// copy of the input once and coalesces equal runs into counted entries in a
// single pass — O(n log n) overall, versus the O(n²) of repeated Add (each Add
// shifts the sorted slice). The result is identical to repeated Add.
func Tree{{.Name}}Of(values ...{{.GoType}}) *Tree{{.Name}} {
	if len(values) == 0 {
		return NewTree{{.Name}}()
	}
	sorted := make([]{{.GoType}}, len(values))
	copy(sorted, values)
	slices.SortFunc(sorted, func(a, c {{.GoType}}) int {
		return {{if .IsFloat}}{{.CmpFn}}(a, c){{else}}cmp.Compare(a, c){{end}}
	})
	b := NewTree{{.Name}}()
	b.entries = coalesce{{.Name}}Sorted(sorted)
	b.size = len(sorted)
	return b
}

// NewTree{{.Name}}FromSorted builds a Tree{{.Name}} from presorted ascending
// values in a single O(n) pass, coalescing equal runs into counts. values must
// be in ascending order according to the bag's own comparator (the IEEE-754
// total order for floats); out-of-order input returns pump.ErrNotSorted.
//
// Duplicate values are the normal bag case and increment the count, so the
// duplicate policy does not apply here. Run-length coalescing is overflow-checked
// (each count fits in an int by construction since it cannot exceed len(values)).
// The result is observably identical to the same values added one-by-one.
func NewTree{{.Name}}FromSorted(values []{{.GoType}}) (*Tree{{.Name}}, error) {
	if len(values) == 0 {
		return NewTree{{.Name}}(), nil
	}
	for i := 1; i < len(values); i++ {
		if {{if .IsFloat}}{{.CmpFn}}(values[i], values[i-1]){{else}}cmp.Compare(values[i], values[i-1]){{end}} < 0 {
			return nil, pump.ErrNotSorted
		}
	}
	b := NewTree{{.Name}}()
	b.entries = coalesce{{.Name}}Sorted(values)
	b.size = len(values)
	return b, nil
}

// coalesce{{.Name}}Sorted compresses a sorted value slice into counted entries
// in one pass. The input must already be sorted by the bag's comparator.
func coalesce{{.Name}}Sorted(sorted []{{.GoType}}) []Tree{{.Name}}Entry {
	entries := make([]Tree{{.Name}}Entry, 0, len(sorted))
	entries = append(entries, Tree{{.Name}}Entry{value: sorted[0], count: 1})
	for i := 1; i < len(sorted); i++ {
		last := &entries[len(entries)-1]
		if {{if .IsFloat}}{{.CmpFn}}(sorted[i], last.value) == 0{{else}}sorted[i] == last.value{{end}} {
			last.count++
		} else {
			entries = append(entries, Tree{{.Name}}Entry{value: sorted[i], count: 1})
		}
	}
	return entries
}

// search returns the index where value is or would be inserted.
// Returns (index, found).
func (b *Tree{{.Name}}) search(value {{.GoType}}) (int, bool) {
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
func (b *Tree{{.Name}}) Add(value {{.GoType}}) {
	idx, found := b.search(value)
	if found {
		b.entries[idx].count++
		b.size++
		return
	}
	// Insert at idx to keep sorted order
	b.entries = append(b.entries, Tree{{.Name}}Entry{})
	copy(b.entries[idx+1:], b.entries[idx:])
	b.entries[idx] = Tree{{.Name}}Entry{value: value, count: 1}
	b.size++
}

// AddOccurrences adds the given number of occurrences of the value.
// Returns the new count for this value. Panics if occurrences is negative.
func (b *Tree{{.Name}}) AddOccurrences(value {{.GoType}}, occurrences int) int {
	if occurrences < 0 {
		panic(fmt.Sprintf("Tree{{.Name}}: cannot add negative occurrences: %d", occurrences))
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
	b.entries = append(b.entries, Tree{{.Name}}Entry{})
	copy(b.entries[idx+1:], b.entries[idx:])
	b.entries[idx] = Tree{{.Name}}Entry{value: value, count: occurrences}
	b.size += occurrences
	return occurrences
}

// Remove removes one occurrence of the value. Returns true if the value was present.
func (b *Tree{{.Name}}) Remove(value {{.GoType}}) bool {
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
func (b *Tree{{.Name}}) RemoveOccurrences(value {{.GoType}}, occurrences int) bool {
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
func (b *Tree{{.Name}}) RemoveAll(value {{.GoType}}) int {
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
func (b *Tree{{.Name}}) OccurrencesOf(value {{.GoType}}) int {
	idx, found := b.search(value)
	if !found {
		return 0
	}
	return b.entries[idx].count
}

// Contains returns true if the bag contains at least one occurrence of the value.
func (b *Tree{{.Name}}) Contains(value {{.GoType}}) bool {
	_, found := b.search(value)
	return found
}

// Len returns the total number of elements including duplicates.
func (b *Tree{{.Name}}) Len() int {
	return b.size
}

// Len returns the number of elements. It is an alias for Size, matching
// Go convention (sort.Interface, container/list, bytes.Buffer).

// SizeDistinct returns the number of distinct elements.
func (b *Tree{{.Name}}) SizeDistinct() int {
	return len(b.entries)
}


// Clear removes all elements from the bag.
func (b *Tree{{.Name}}) Clear() {
	b.entries = nil
	b.size = 0
}

// Min returns the smallest element, or zero value and false if empty.
func (b *Tree{{.Name}}) Min() ({{.GoType}}, bool) {
	if len(b.entries) == 0 {
		return {{.Zero}}, false
	}
	return b.entries[0].value, true
}

// Max returns the largest element, or zero value and false if empty.
func (b *Tree{{.Name}}) Max() ({{.GoType}}, bool) {
	if len(b.entries) == 0 {
		return {{.Zero}}, false
	}
	return b.entries[len(b.entries)-1].value, true
}

// All returns an iter.Seq that yields each element once per occurrence, in sorted order.
func (b *Tree{{.Name}}) All() iter.Seq[{{.GoType}}] {
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
func (b *Tree{{.Name}}) AllDistinct() iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
		for _, entry := range b.entries {
			if !yield(entry.value) {
				return
			}
		}
	}
}

// AllWithOccurrences returns an iter.Seq2 that yields (value, count) pairs in sorted order.
func (b *Tree{{.Name}}) AllWithOccurrences() iter.Seq2[{{.GoType}}, int] {
	return func(yield func({{.GoType}}, int) bool) {
		for _, entry := range b.entries {
			if !yield(entry.value, entry.count) {
				return
			}
		}
	}
}

// ForEach calls the given function for each element (once per occurrence), in sorted order.
func (b *Tree{{.Name}}) ForEach(f func({{.GoType}})) {
	for _, entry := range b.entries {
		for i := 0; i < entry.count; i++ {
			f(entry.value)
		}
	}
}

// ForEachWithOccurrences calls the given function with each distinct element and its count, in sorted order.
func (b *Tree{{.Name}}) ForEachWithOccurrences(f func({{.GoType}}, int)) {
	for _, entry := range b.entries {
		f(entry.value, entry.count)
	}
}

// Select returns a new tree bag containing only elements that satisfy the predicate.
func (b *Tree{{.Name}}) Select(predicate func({{.GoType}}) bool) *Tree{{.Name}} {
	result := NewTree{{.Name}}()
	for _, entry := range b.entries {
		if predicate(entry.value) {
			result.AddOccurrences(entry.value, entry.count)
		}
	}
	return result
}

// Reject returns a new tree bag containing only elements that do not satisfy the predicate.
func (b *Tree{{.Name}}) Reject(predicate func({{.GoType}}) bool) *Tree{{.Name}} {
	result := NewTree{{.Name}}()
	for _, entry := range b.entries {
		if !predicate(entry.value) {
			result.AddOccurrences(entry.value, entry.count)
		}
	}
	return result
}

// Detect returns the first distinct element (in sorted order) that satisfies the predicate, or zero value and false.
func (b *Tree{{.Name}}) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	for _, entry := range b.entries {
		if predicate(entry.value) {
			return entry.value, true
		}
	}
	return {{.Zero}}, false
}

// AnySatisfy returns true if any distinct element satisfies the predicate.
func (b *Tree{{.Name}}) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for _, entry := range b.entries {
		if predicate(entry.value) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all distinct elements satisfy the predicate.
func (b *Tree{{.Name}}) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, entry := range b.entries {
		if !predicate(entry.value) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no distinct element satisfies the predicate.
func (b *Tree{{.Name}}) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, entry := range b.entries {
		if predicate(entry.value) {
			return false
		}
	}
	return true
}

// TopOccurrences returns the n elements with the highest occurrence counts.
func (b *Tree{{.Name}}) TopOccurrences(n int) []struct {
	Value {{.GoType}}
	Count int
} {
	// Copy entries and sort by count descending
	sorted := make([]Tree{{.Name}}Entry, len(b.entries))
	copy(sorted, b.entries)
	slices.SortFunc(sorted, func(a, b Tree{{.Name}}Entry) int {
		return cmp.Compare(b.count, a.count)
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
func (b *Tree{{.Name}}) ToSlice() []{{.GoType}} {
	result := make([]{{.GoType}}, 0, b.size)
	for _, entry := range b.entries {
		for i := 0; i < entry.count; i++ {
			result = append(result, entry.value)
		}
	}
	return result
}

// ToSortedSlice returns all distinct elements as a sorted slice.
func (b *Tree{{.Name}}) ToSortedSlice() []{{.GoType}} {
	result := make([]{{.GoType}}, 0, len(b.entries))
	for _, entry := range b.entries {
		result = append(result, entry.value)
	}
	return result
}

// With returns the bag after adding one occurrence of the value (fluent API).
func (b *Tree{{.Name}}) AddReturning(value {{.GoType}}) *Tree{{.Name}} {
	b.Add(value)
	return b
}

// Without returns the bag after removing all occurrences of the value (fluent API).
func (b *Tree{{.Name}}) RemoveReturning(value {{.GoType}}) *Tree{{.Name}} {
	b.RemoveAll(value)
	return b
}

// WithAll returns the bag after adding all values (fluent API).
func (b *Tree{{.Name}}) AddAllReturning(values ...{{.GoType}}) *Tree{{.Name}} {
	for _, v := range values {
		b.Add(v)
	}
	return b
}

// WithoutAll removes all occurrences of the given values (fluent API).
func (b *Tree{{.Name}}) RemoveAllReturning(values ...{{.GoType}}) *Tree{{.Name}} {
	for _, v := range values {
		b.RemoveAll(v)
	}
	return b
}

// String returns a string representation of the bag in sorted order.
func (b *Tree{{.Name}}) String() string {
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
func (b *Tree{{.Name}}) Equals(other *Tree{{.Name}}) bool {
	if b.size != other.size || len(b.entries) != len(other.entries) {
		return false
	}
	for i, entry := range b.entries {
		if {{if .IsFloat}}{{.BitsFn}}(entry.value) != {{.BitsFn}}(other.entries[i].value){{else}}entry.value != other.entries[i].value{{end}} || entry.count != other.entries[i].count {
			return false
		}
	}
	return true
}
`
