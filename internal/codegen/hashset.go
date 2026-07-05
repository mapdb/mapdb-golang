package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

// hsData is the per-element-type view the hashset templates iterate over.
type hsData struct {
	Name      string // Int32, Float32, Char, Bool (identifier stem)
	GoType    string // int32, float32, uint16, bool
	SnakeName string // int32, float32, char, bool (file-name stem)
	Zero      string // zero literal: "0", "0.0", "false"

	// IsFloat selects bit-pattern equality (math.FloatNbits) over ==.
	IsFloat bool
	// BitsFn is the float bit-pattern function for equality comparison
	// (math.Float32bits / math.Float64bits). Floats only.
	BitsFn string

	// NeedsMath drives the import block of the base file.
	NeedsMath bool

	// IsBool selects the boolean hash body (no golden-ratio mixing).
	IsBool bool
	// HashExpr is the inner operand of the golden-ratio multiply for the
	// non-bool hash body: h := <HashExpr> * 0x9E3779B97F4A7C15. Captured
	// per type because the integer/char/float reinterpretations differ
	// (and int32 alone double-casts through uint32).
	HashExpr string
}

// hashSetTypes returns the element types the hashset family is generated for.
// It deliberately does NOT reuse Primitives(): hashset additionally supports
// bool, and adding bool to the shared list would make arraylist/interval drift.
// Order is the canonical iteration order for code generation (fixed slice ⇒
// idempotent output).
func hashSetTypes() []hsData {
	return []hsData{
		{Name: "Int8", GoType: "int8", SnakeName: "int8", Zero: "0", HashExpr: "uint64(value)"},
		{Name: "Int16", GoType: "int16", SnakeName: "int16", Zero: "0", HashExpr: "uint64(value)"},
		{Name: "Int32", GoType: "int32", SnakeName: "int32", Zero: "0", HashExpr: "uint64(uint32(value))"},
		{Name: "Int64", GoType: "int64", SnakeName: "int64", Zero: "0", HashExpr: "uint64(value)"},
		{Name: "Char", GoType: "uint16", SnakeName: "char", Zero: "0", HashExpr: "uint64(value)"},
		{
			Name: "Float32", GoType: "float32", SnakeName: "float32", Zero: "0.0",
			IsFloat: true, BitsFn: "math.Float32bits", NeedsMath: true,
			HashExpr: "uint64(math.Float32bits(value))",
		},
		{
			Name: "Float64", GoType: "float64", SnakeName: "float64", Zero: "0.0",
			IsFloat: true, BitsFn: "math.Float64bits", NeedsMath: true,
			HashExpr: "math.Float64bits(value)",
		},
		{Name: "Bool", GoType: "bool", SnakeName: "bool", Zero: "false", IsBool: true},
	}
}

// genHashSet writes the per-element-type hash set sources (base, immutable,
// synchronized variants) into the current working directory. Invoked from
// hashset/ via go:generate.
func genHashSet() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	base := parse("hs-base", hashSetTmpl)
	immutable := parse("hs-immutable", immutableHashSetTmpl)
	synchronized := parse("hs-sync", synchronizedHashSetTmpl)

	write := func(name string, tmpl *template.Template, data hsData) error {
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

	for _, d := range hashSetTypes() {
		if err := write(d.SnakeName+"_hash_set.go", base, d); err != nil {
			return err
		}
		if err := write("immutable_"+d.SnakeName+"_hash_set.go", immutable, d); err != nil {
			return err
		}
		if err := write("synchronized_"+d.SnakeName+"_hash_set.go", synchronized, d); err != nil {
			return err
		}
	}

	// Stamp the conformance laws (todo 14 §4). A hash set has no iteration
	// order, so law 1 is checked as a multiset. Bool is excluded: its two-value
	// domain makes the shared numeric fixture degenerate.
	names := make([]string, 0, len(hashSetTypes()))
	goTypes := make([]string, 0, len(hashSetTypes()))
	for _, d := range hashSetTypes() {
		names = append(names, d.Name)
		goTypes = append(goTypes, d.GoType)
	}
	return genConformanceForOfTypes("hashset", false, names, goTypes, map[string]bool{"bool": true})
}

const hashSetTmpl = genHeader + `package hashset

import (
	"fmt"
	"iter"
{{- if .NeedsMath}}
	"math"
{{- end}}
	"strings"

	"github.com/mapdb/mapdb-golang/internal/bits"
	"github.com/mapdb/mapdb-golang/pump"
)

const (
	{{.SnakeName}}DefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
)

type {{.SnakeName}}Entry struct {
	key      {{.GoType}}
	occupied bool
}

// {{.Name}} is an open-addressing hash set for {{.GoType}} values.
type {{.Name}} struct {
	entries []{{.SnakeName}}Entry
	size    int
}

// New{{.Name}} creates a new empty {{.Name}}.
func New{{.Name}}() *{{.Name}} {
	return New{{.Name}}WithCapacity({{.SnakeName}}DefaultCapacity)
}

// New{{.Name}}WithCapacity creates a new empty {{.Name}} with the given initial capacity.
func New{{.Name}}WithCapacity(capacity int) *{{.Name}} {
	cap := bits.NextPowerOfTwo(capacity)
	return &{{.Name}}{
		entries: make([]{{.SnakeName}}Entry, cap),
		size:    0,
	}
}

// {{.Name}}Of creates a new {{.Name}} from the given values.
func {{.Name}}Of(values ...{{.GoType}}) *{{.Name}} {
	s := New{{.Name}}WithCapacity(len(values) * 2)
	for _, v := range values {
		s.Add(v)
	}
	return s
}

// {{.Name}}BulkLoad builds a {{.Name}} from values in a single pass, presizing
// the table once to fit len(values) at the 0.75 load factor. The input need not
// be sorted. On a duplicate value it returns pump.ErrDuplicateKey unless
// policy is pump.IgnoreDuplicates, in which case duplicates are skipped.
// The result is observably identical to the same values inserted one-by-one with
// Add. The size is a hint; use {{.Name}}BulkLoadExact for the zero-rehash
// guarantee.
func {{.Name}}BulkLoad(values []{{.GoType}}, policy pump.DuplicatePolicy) (*{{.Name}}, error) {
	s := &{{.Name}}{entries: make([]{{.SnakeName}}Entry, {{.Name}}bulkCap(len(values)))}
	for _, v := range values {
		if s.needsResize() {
			s.resize()
		}
		if _, err := s.bulkAdd(v, policy); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// {{.Name}}BulkLoadExact is like {{.Name}}BulkLoad but guarantees zero mid-load
// rehash: the table is sized for exactly n consumed values. It returns
// pump.ErrTooManyElements if the source yields more than n values, even when
// the extra values are duplicates skipped by pump.IgnoreDuplicates. n must be
// non-negative (negative panics).
func {{.Name}}BulkLoadExact(values []{{.GoType}}, n int, policy pump.DuplicatePolicy) (*{{.Name}}, error) {
	if n < 0 {
		panic("mapdb: {{.Name}}BulkLoadExact: negative n")
	}
	s := &{{.Name}}{entries: make([]{{.SnakeName}}Entry, {{.Name}}bulkCap(n))}
	if len(values) > n {
		return nil, pump.ErrTooManyElements
	}
	for _, v := range values {
		if _, err := s.bulkAdd(v, policy); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// bulkAdd inserts a single value via the ordinary probe without a resize check
// (callers guarantee capacity), applying the duplicate policy.
func (s *{{.Name}}) bulkAdd(value {{.GoType}}, policy pump.DuplicatePolicy) (bool, error) {
	mask := len(s.entries) - 1
	idx := int(s.hash(value)) & mask
	for {
		if !s.entries[idx].occupied {
			s.entries[idx].key = value
			s.entries[idx].occupied = true
			s.size++
			return false, nil
		}
		if {{if .IsFloat}}{{.BitsFn}}(s.entries[idx].key) == {{.BitsFn}}(value){{else}}s.entries[idx].key == value{{end}} {
			if policy == pump.IgnoreDuplicates {
				return true, nil
			}
			return true, pump.ErrDuplicateKey
		}
		idx = (idx + 1) & mask
	}
}

// {{.Name}}bulkCap returns the presized table capacity for n values that avoids
// any mid-load rehash, floored at the family default.
func {{.Name}}bulkCap(n int) int {
	c := pump.HashCapacityFor(n)
	if c < {{.SnakeName}}DefaultCapacity {
		return {{.SnakeName}}DefaultCapacity
	}
	return c
}

// Add inserts a value into the set. Returns true if the value was added (not already present).
func (s *{{.Name}}) Add(value {{.GoType}}) bool {
	if s.needsResize() {
		s.resize()
	}
	cap := len(s.entries)
	mask := cap - 1
	idx := int(s.hash(value)) & mask

	for {
		if !s.entries[idx].occupied {
			s.entries[idx].key = value
			s.entries[idx].occupied = true
			s.size++
			return true
		}
		if {{if .IsFloat}}{{.BitsFn}}(s.entries[idx].key) == {{.BitsFn}}(value){{else}}s.entries[idx].key == value{{end}} {
			return false
		}
		idx = (idx + 1) & mask
	}
}

// AddAll inserts all values into the set.
func (s *{{.Name}}) AddAll(values ...{{.GoType}}) {
	for _, v := range values {
		s.Add(v)
	}
}

// Remove removes a value from the set. Returns true if the value was found and removed.
func (s *{{.Name}}) Remove(value {{.GoType}}) bool {
	cap := len(s.entries)
	if cap == 0 {
		return false
	}
	mask := cap - 1
	idx := int(s.hash(value)) & mask

	for {
		if !s.entries[idx].occupied {
			return false
		}
		if {{if .IsFloat}}{{.BitsFn}}(s.entries[idx].key) == {{.BitsFn}}(value){{else}}s.entries[idx].key == value{{end}} {
			s.entries[idx] = {{.SnakeName}}Entry{}
			s.size--
			s.rehashFrom(idx, mask)
			return true
		}
		idx = (idx + 1) & mask
	}
}

// Contains returns true if the set contains the given value.
func (s *{{.Name}}) Contains(value {{.GoType}}) bool {
	cap := len(s.entries)
	if cap == 0 {
		return false
	}
	mask := cap - 1
	idx := int(s.hash(value)) & mask

	for {
		if !s.entries[idx].occupied {
			return false
		}
		if {{if .IsFloat}}{{.BitsFn}}(s.entries[idx].key) == {{.BitsFn}}(value){{else}}s.entries[idx].key == value{{end}} {
			return true
		}
		idx = (idx + 1) & mask
	}
}

// Len returns the number of elements. Use s.Len() == 0 to test for emptiness.
func (s *{{.Name}}) Len() int {
	return s.size
}

// Clear removes all elements from the set.
func (s *{{.Name}}) Clear() {
	for i := range s.entries {
		s.entries[i] = {{.SnakeName}}Entry{}
	}
	s.size = 0
}

// All returns an iter.Seq that yields all elements.
func (s *{{.Name}}) All() iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
		for i := range s.entries {
			if s.entries[i].occupied {
				if !yield(s.entries[i].key) {
					return
				}
			}
		}
	}
}

// ForEach calls the given function for each element.
func (s *{{.Name}}) ForEach(f func({{.GoType}})) {
	for i := range s.entries {
		if s.entries[i].occupied {
			f(s.entries[i].key)
		}
	}
}

// Select returns a new set containing only elements that satisfy the predicate.
func (s *{{.Name}}) Select(predicate func({{.GoType}}) bool) *{{.Name}} {
	result := New{{.Name}}()
	for i := range s.entries {
		if s.entries[i].occupied && predicate(s.entries[i].key) {
			result.Add(s.entries[i].key)
		}
	}
	return result
}

// Reject returns a new set containing only elements that do not satisfy the predicate.
func (s *{{.Name}}) Reject(predicate func({{.GoType}}) bool) *{{.Name}} {
	result := New{{.Name}}()
	for i := range s.entries {
		if s.entries[i].occupied && !predicate(s.entries[i].key) {
			result.Add(s.entries[i].key)
		}
	}
	return result
}

// Detect returns the first element that satisfies the predicate, or zero value and false.
func (s *{{.Name}}) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	for i := range s.entries {
		if s.entries[i].occupied && predicate(s.entries[i].key) {
			return s.entries[i].key, true
		}
	}
	return {{.Zero}}, false
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *{{.Name}}) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for i := range s.entries {
		if s.entries[i].occupied && predicate(s.entries[i].key) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *{{.Name}}) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for i := range s.entries {
		if s.entries[i].occupied && !predicate(s.entries[i].key) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *{{.Name}}) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for i := range s.entries {
		if s.entries[i].occupied && predicate(s.entries[i].key) {
			return false
		}
	}
	return true
}

// Union returns a new set containing all elements from both sets.
func (s *{{.Name}}) Union(other *{{.Name}}) *{{.Name}} {
	result := New{{.Name}}WithCapacity((s.size + other.size) * 2)
	for i := range s.entries {
		if s.entries[i].occupied {
			result.Add(s.entries[i].key)
		}
	}
	for i := range other.entries {
		if other.entries[i].occupied {
			result.Add(other.entries[i].key)
		}
	}
	return result
}

// Intersect returns a new set containing only elements present in both sets.
func (s *{{.Name}}) Intersect(other *{{.Name}}) *{{.Name}} {
	result := New{{.Name}}()
	smaller, larger := s, other
	if s.size > other.size {
		smaller, larger = other, s
	}
	for i := range smaller.entries {
		if smaller.entries[i].occupied && larger.Contains(smaller.entries[i].key) {
			result.Add(smaller.entries[i].key)
		}
	}
	return result
}

// Difference returns a new set containing elements in this set but not in the other.
func (s *{{.Name}}) Difference(other *{{.Name}}) *{{.Name}} {
	result := New{{.Name}}()
	for i := range s.entries {
		if s.entries[i].occupied && !other.Contains(s.entries[i].key) {
			result.Add(s.entries[i].key)
		}
	}
	return result
}

// SymmetricDifference returns a new set containing elements in either set but not both.
func (s *{{.Name}}) SymmetricDifference(other *{{.Name}}) *{{.Name}} {
	result := New{{.Name}}()
	for i := range s.entries {
		if s.entries[i].occupied && !other.Contains(s.entries[i].key) {
			result.Add(s.entries[i].key)
		}
	}
	for i := range other.entries {
		if other.entries[i].occupied && !s.Contains(other.entries[i].key) {
			result.Add(other.entries[i].key)
		}
	}
	return result
}

// ToSlice returns all elements as a slice.
func (s *{{.Name}}) ToSlice() []{{.GoType}} {
	result := make([]{{.GoType}}, 0, s.size)
	for i := range s.entries {
		if s.entries[i].occupied {
			result = append(result, s.entries[i].key)
		}
	}
	return result
}

// AddReturning adds the value to the set and returns the receiver (mutating, fluent).
func (s *{{.Name}}) AddReturning(value {{.GoType}}) *{{.Name}} {
	s.Add(value)
	return s
}

// RemoveReturning removes the value from the set and returns the receiver (mutating, fluent).
func (s *{{.Name}}) RemoveReturning(value {{.GoType}}) *{{.Name}} {
	s.Remove(value)
	return s
}

// AddAllReturning adds all values to the set and returns the receiver (mutating, fluent).
func (s *{{.Name}}) AddAllReturning(values ...{{.GoType}}) *{{.Name}} {
	s.AddAll(values...)
	return s
}

// RemoveAllReturning removes all given values from the set and returns the receiver (mutating, fluent).
func (s *{{.Name}}) RemoveAllReturning(values ...{{.GoType}}) *{{.Name}} {
	for _, v := range values {
		s.Remove(v)
	}
	return s
}

// ToImmutable returns an immutable copy of this set.
func (s *{{.Name}}) ToImmutable() *Immutable{{.Name}} {
	return Immutable{{.Name}}From(s)
}

// String returns a string representation of the set.
func (s *{{.Name}}) String() string {
	if s.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for i := range s.entries {
		if s.entries[i].occupied {
			if !first {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%v", s.entries[i].key)
			first = false
		}
	}
	sb.WriteString("}")
	return sb.String()
}

// Equals returns true if the other set has the same elements.
func (s *{{.Name}}) Equals(other *{{.Name}}) bool {
	if s.size != other.size {
		return false
	}
	for i := range s.entries {
		if s.entries[i].occupied && !other.Contains(s.entries[i].key) {
			return false
		}
	}
	return true
}

func (s *{{.Name}}) hash(value {{.GoType}}) uint64 {
{{- if .IsBool}}
	if value {
		return 1
	}
	return 0
{{- else}}
	h := {{.HashExpr}} * 0x9E3779B97F4A7C15
	return h ^ (h >> 32)
{{- end}}
}

func (s *{{.Name}}) needsResize() bool {
	return (s.size+1)*4 >= len(s.entries)*3 // 0.75 load factor, integer math
}

func (s *{{.Name}}) resize() {
	oldEntries := s.entries
	newCap := len(oldEntries) * 2
	if newCap == 0 {
		newCap = {{.SnakeName}}DefaultCapacity
	}
	s.entries = make([]{{.SnakeName}}Entry, newCap)
	s.size = 0

	for i := range oldEntries {
		if oldEntries[i].occupied {
			s.Add(oldEntries[i].key)
		}
	}
}

func (s *{{.Name}}) rehashFrom(deleted int, mask int) {
	c := len(s.entries)
	idx := (deleted + 1) & mask
	for s.entries[idx].occupied {
		ideal := int(s.hash(s.entries[idx].key)) & mask
		distCurrent := (idx - ideal + c) & mask
		distGap := (deleted - ideal + c) & mask
		if distCurrent > distGap {
			s.entries[deleted] = s.entries[idx]
			s.entries[idx] = {{.SnakeName}}Entry{}
			deleted = idx
		}
		idx = (idx + 1) & mask
		if idx == deleted {
			break
		}
	}
}

`

const immutableHashSetTmpl = genHeader + `package hashset

import (
	"iter"
)

// Immutable{{.Name}} is an immutable view of a {{.Name}}.
type Immutable{{.Name}} struct {
	delegate *{{.Name}}
}

// NewImmutable{{.Name}} creates an immutable set from the given values.
func NewImmutable{{.Name}}(values ...{{.GoType}}) *Immutable{{.Name}} {
	return &Immutable{{.Name}}{delegate: {{.Name}}Of(values...)}
}

// Immutable{{.Name}}From creates an immutable copy of a mutable set.
func Immutable{{.Name}}From(s *{{.Name}}) *Immutable{{.Name}} {
	copy := {{.Name}}Of(s.ToSlice()...)
	return &Immutable{{.Name}}{delegate: copy}
}

// Contains returns true if the set contains the given value.
func (s *Immutable{{.Name}}) Contains(value {{.GoType}}) bool {
	return s.delegate.Contains(value)
}

// Len returns the number of elements. Use s.Len() == 0 to test for emptiness.
func (s *Immutable{{.Name}}) Len() int {
	return s.delegate.Len()
}

// All returns an iter.Seq that yields all elements.
func (s *Immutable{{.Name}}) All() iter.Seq[{{.GoType}}] {
	return s.delegate.All()
}

// ForEach calls the given function for each element.
func (s *Immutable{{.Name}}) ForEach(f func({{.GoType}})) {
	s.delegate.ForEach(f)
}

// Select returns a new immutable set with elements satisfying the predicate.
func (s *Immutable{{.Name}}) Select(predicate func({{.GoType}}) bool) *Immutable{{.Name}} {
	return &Immutable{{.Name}}{delegate: s.delegate.Select(predicate)}
}

// Reject returns a new immutable set with elements not satisfying the predicate.
func (s *Immutable{{.Name}}) Reject(predicate func({{.GoType}}) bool) *Immutable{{.Name}} {
	return &Immutable{{.Name}}{delegate: s.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *Immutable{{.Name}}) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	return s.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *Immutable{{.Name}}) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	return s.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *Immutable{{.Name}}) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	return s.delegate.NoneSatisfy(predicate)
}

// Union returns a new immutable set with elements from both sets.
func (s *Immutable{{.Name}}) Union(other *Immutable{{.Name}}) *Immutable{{.Name}} {
	return &Immutable{{.Name}}{delegate: s.delegate.Union(other.delegate)}
}

// Intersect returns a new immutable set with elements in both sets.
func (s *Immutable{{.Name}}) Intersect(other *Immutable{{.Name}}) *Immutable{{.Name}} {
	return &Immutable{{.Name}}{delegate: s.delegate.Intersect(other.delegate)}
}

// Difference returns a new immutable set with elements in this but not other.
func (s *Immutable{{.Name}}) Difference(other *Immutable{{.Name}}) *Immutable{{.Name}} {
	return &Immutable{{.Name}}{delegate: s.delegate.Difference(other.delegate)}
}

// ToSlice returns all elements as a slice.
func (s *Immutable{{.Name}}) ToSlice() []{{.GoType}} {
	return s.delegate.ToSlice()
}

// String returns a string representation.
func (s *Immutable{{.Name}}) String() string {
	return s.delegate.String()
}

// Equals returns true if the other immutable set has the same elements.
func (s *Immutable{{.Name}}) Equals(other *Immutable{{.Name}}) bool {
	return s.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this set.
func (s *Immutable{{.Name}}) ToMutable() *{{.Name}} {
	return {{.Name}}Of(s.ToSlice()...)
}
`

const synchronizedHashSetTmpl = genHeader + `package hashset

import (
	"iter"
	"sync"
	"unsafe"
)

// Synchronized{{.Name}} is a thread-safe wrapper around {{.Name}}.
//
// Read methods hold an RLock; writes hold a Lock. Methods that take a
// caller-supplied function (Select, ForEach, AnySatisfy, …) snapshot
// the backing set under RLock and release it before invoking the
// callback, so the callback is free to call back into the wrapper
// without deadlocking.
//
// Methods that return a new set (Select, Reject, Union, Intersect,
// Difference, SymmetricDifference) return an unwrapped *{{.Name}};
// the caller owns it.
type Synchronized{{.Name}} struct {
	delegate *{{.Name}}
	mu       sync.RWMutex
}

// NewSynchronized{{.Name}} creates a new thread-safe empty set.
func NewSynchronized{{.Name}}() *Synchronized{{.Name}} {
	return &Synchronized{{.Name}}{delegate: New{{.Name}}()}
}

// NewSynchronized{{.Name}}From wraps an existing set. The
// wrapper takes ownership — callers must not mutate the delegate
// directly without locking.
func NewSynchronized{{.Name}}From(s *{{.Name}}) *Synchronized{{.Name}} {
	return &Synchronized{{.Name}}{delegate: s}
}

// Synchronized{{.Name}}Of constructs a synchronized set from values.
func Synchronized{{.Name}}Of(values ...{{.GoType}}) *Synchronized{{.Name}} {
	s := New{{.Name}}()
	for _, v := range values {
		s.Add(v)
	}
	return &Synchronized{{.Name}}{delegate: s}
}

// snapshot returns a defensive copy of the set's elements under RLock.
func (s *Synchronized{{.Name}}) snapshot() []{{.GoType}} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToSlice()
}

// ── writes ────────────────────────────────────────────────────────────

func (s *Synchronized{{.Name}}) Add(value {{.GoType}}) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.delegate.Add(value)
}

func (s *Synchronized{{.Name}}) AddAll(values ...{{.GoType}}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.AddAll(values...)
}

func (s *Synchronized{{.Name}}) Remove(value {{.GoType}}) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.delegate.Remove(value)
}

func (s *Synchronized{{.Name}}) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.Clear()
}

// ── simple reads ──────────────────────────────────────────────────────

func (s *Synchronized{{.Name}}) Contains(value {{.GoType}}) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Contains(value)
}

// Len returns the number of elements. Use s.Len() == 0 to test for emptiness.
func (s *Synchronized{{.Name}}) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Len()
}

func (s *Synchronized{{.Name}}) ToSlice() []{{.GoType}} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToSlice()
}

func (s *Synchronized{{.Name}}) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.String()
}

// ── iteration ────────────────────────────────────────────────────────

// All returns an iter.Seq over a snapshot. Iteration is lock-free.
func (s *Synchronized{{.Name}}) All() iter.Seq[{{.GoType}}] {
	snapshot := s.snapshot()
	return func(yield func({{.GoType}}) bool) {
		for _, v := range snapshot {
			if !yield(v) {
				return
			}
		}
	}
}

// ── functional over snapshot ──────────────────────────────────────────

func (s *Synchronized{{.Name}}) ForEach(f func({{.GoType}})) {
	for _, v := range s.snapshot() {
		f(v)
	}
}

func (s *Synchronized{{.Name}}) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (s *Synchronized{{.Name}}) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range s.snapshot() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (s *Synchronized{{.Name}}) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return false
		}
	}
	return true
}

func (s *Synchronized{{.Name}}) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return v, true
		}
	}
	var zero {{.GoType}}
	return zero, false
}

// ── functional that return a new set ─────────────────────────────────

func (s *Synchronized{{.Name}}) Select(predicate func({{.GoType}}) bool) *{{.Name}} {
	snapshot := s.snapshot()
	result := New{{.Name}}()
	for _, v := range snapshot {
		if predicate(v) {
			result.Add(v)
		}
	}
	return result
}

func (s *Synchronized{{.Name}}) Reject(predicate func({{.GoType}}) bool) *{{.Name}} {
	snapshot := s.snapshot()
	result := New{{.Name}}()
	for _, v := range snapshot {
		if !predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// ── set operations (two-lock, deadlock-safe) ──────────────────────────

// lockPair acquires two RLocks in pointer-address order and returns
// a release function. Guarantees no A.op(B) ⟷ B.op(A) deadlock.
func (s *Synchronized{{.Name}}) lockPair(other *Synchronized{{.Name}}) func() {
	if s == other {
		s.mu.RLock()
		return func() { s.mu.RUnlock() }
	}
	first, second := s, other
	if uintptr(unsafe.Pointer(s)) > uintptr(unsafe.Pointer(other)) {
		first, second = other, s
	}
	first.mu.RLock()
	second.mu.RLock()
	return func() { second.mu.RUnlock(); first.mu.RUnlock() }
}

func (s *Synchronized{{.Name}}) Union(other *Synchronized{{.Name}}) *{{.Name}} {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Union(other.delegate)
}

func (s *Synchronized{{.Name}}) Intersect(other *Synchronized{{.Name}}) *{{.Name}} {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Intersect(other.delegate)
}

func (s *Synchronized{{.Name}}) Difference(other *Synchronized{{.Name}}) *{{.Name}} {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Difference(other.delegate)
}

func (s *Synchronized{{.Name}}) SymmetricDifference(other *Synchronized{{.Name}}) *{{.Name}} {
	release := s.lockPair(other)
	defer release()
	return s.delegate.SymmetricDifference(other.delegate)
}

// ── fluent mutators ───────────────────────────────────────────────────

// AddReturning adds the value and returns the receiver (mutating, fluent).
func (s *Synchronized{{.Name}}) AddReturning(value {{.GoType}}) *Synchronized{{.Name}} {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.AddReturning(value)
	return s
}

// AddAllReturning adds all values and returns the receiver (mutating, fluent).
func (s *Synchronized{{.Name}}) AddAllReturning(values ...{{.GoType}}) *Synchronized{{.Name}} {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.AddAllReturning(values...)
	return s
}

// RemoveReturning removes the value and returns the receiver (mutating, fluent).
func (s *Synchronized{{.Name}}) RemoveReturning(value {{.GoType}}) *Synchronized{{.Name}} {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.RemoveReturning(value)
	return s
}

// RemoveAllReturning removes all given values and returns the receiver (mutating, fluent).
func (s *Synchronized{{.Name}}) RemoveAllReturning(values ...{{.GoType}}) *Synchronized{{.Name}} {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.RemoveAllReturning(values...)
	return s
}

// ── conversions ───────────────────────────────────────────────────────

func (s *Synchronized{{.Name}}) ToImmutable() *Immutable{{.Name}} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToImmutable()
}

// Equals compares by contents. Locks are acquired in pointer-address
// order to prevent deadlocks under concurrent A.Equals(B) / B.Equals(A).
func (s *Synchronized{{.Name}}) Equals(other *Synchronized{{.Name}}) bool {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Equals(other.delegate)
}
`
