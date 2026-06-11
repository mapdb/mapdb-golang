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

	// NeedsMath / NeedsUnsafe drive the import block of the base file.
	NeedsMath   bool
	NeedsUnsafe bool

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
			IsFloat: true, BitsFn: "math.Float32bits", NeedsMath: true, NeedsUnsafe: true,
			HashExpr: "uint64(*(*uint32)(unsafe.Pointer(&value)))",
		},
		{
			Name: "Float64", GoType: "float64", SnakeName: "float64", Zero: "0.0",
			IsFloat: true, BitsFn: "math.Float64bits", NeedsMath: true, NeedsUnsafe: true,
			HashExpr: "*(*uint64)(unsafe.Pointer(&value))",
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

	base := template.Must(template.New("hs-base").Parse(hashSetTmpl))
	immutable := template.Must(template.New("hs-immutable").Parse(immutableHashSetTmpl))
	synchronized := template.Must(template.New("hs-sync").Parse(synchronizedHashSetTmpl))

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

	return nil
}

const hashSetTmpl = genHeader + `package hashset

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
	{{.SnakeName}}HashSetDefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
)

type {{.SnakeName}}HashSetEntry struct {
	key      {{.GoType}}
	occupied bool
}

// {{.Name}}HashSet is an open-addressing hash set for {{.GoType}} values.
type {{.Name}}HashSet struct {
	entries []{{.SnakeName}}HashSetEntry
	size    int
}

// New{{.Name}}HashSet creates a new empty {{.Name}}HashSet.
func New{{.Name}}HashSet() *{{.Name}}HashSet {
	return New{{.Name}}HashSetWithCapacity({{.SnakeName}}HashSetDefaultCapacity)
}

// New{{.Name}}HashSetWithCapacity creates a new empty {{.Name}}HashSet with the given initial capacity.
func New{{.Name}}HashSetWithCapacity(capacity int) *{{.Name}}HashSet {
	cap := nextPowerOfTwo{{.Name}}HashSet(capacity)
	return &{{.Name}}HashSet{
		entries: make([]{{.SnakeName}}HashSetEntry, cap),
		size:    0,
	}
}

// {{.Name}}HashSetOf creates a new {{.Name}}HashSet from the given values.
func {{.Name}}HashSetOf(values ...{{.GoType}}) *{{.Name}}HashSet {
	s := New{{.Name}}HashSetWithCapacity(len(values) * 2)
	for _, v := range values {
		s.Add(v)
	}
	return s
}

// Add inserts a value into the set. Returns true if the value was added (not already present).
func (s *{{.Name}}HashSet) Add(value {{.GoType}}) bool {
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
func (s *{{.Name}}HashSet) AddAll(values ...{{.GoType}}) {
	for _, v := range values {
		s.Add(v)
	}
}

// Remove removes a value from the set. Returns true if the value was found and removed.
func (s *{{.Name}}HashSet) Remove(value {{.GoType}}) bool {
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
			s.entries[idx] = {{.SnakeName}}HashSetEntry{}
			s.size--
			s.rehashFrom(idx, mask)
			return true
		}
		idx = (idx + 1) & mask
	}
}

// Contains returns true if the set contains the given value.
func (s *{{.Name}}HashSet) Contains(value {{.GoType}}) bool {
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

// Size returns the number of elements in the set.
func (s *{{.Name}}HashSet) Size() int {
	return s.size
}

// IsEmpty returns true if the set contains no elements.
func (s *{{.Name}}HashSet) IsEmpty() bool {
	return s.size == 0
}

// Clear removes all elements from the set.
func (s *{{.Name}}HashSet) Clear() {
	for i := range s.entries {
		s.entries[i] = {{.SnakeName}}HashSetEntry{}
	}
	s.size = 0
}

// All returns an iter.Seq that yields all elements.
func (s *{{.Name}}HashSet) All() iter.Seq[{{.GoType}}] {
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
func (s *{{.Name}}HashSet) ForEach(f func({{.GoType}})) {
	for i := range s.entries {
		if s.entries[i].occupied {
			f(s.entries[i].key)
		}
	}
}

// Select returns a new set containing only elements that satisfy the predicate.
func (s *{{.Name}}HashSet) Select(predicate func({{.GoType}}) bool) *{{.Name}}HashSet {
	result := New{{.Name}}HashSet()
	for i := range s.entries {
		if s.entries[i].occupied && predicate(s.entries[i].key) {
			result.Add(s.entries[i].key)
		}
	}
	return result
}

// Reject returns a new set containing only elements that do not satisfy the predicate.
func (s *{{.Name}}HashSet) Reject(predicate func({{.GoType}}) bool) *{{.Name}}HashSet {
	result := New{{.Name}}HashSet()
	for i := range s.entries {
		if s.entries[i].occupied && !predicate(s.entries[i].key) {
			result.Add(s.entries[i].key)
		}
	}
	return result
}

// Detect returns the first element that satisfies the predicate, or zero value and false.
func (s *{{.Name}}HashSet) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	for i := range s.entries {
		if s.entries[i].occupied && predicate(s.entries[i].key) {
			return s.entries[i].key, true
		}
	}
	return {{.Zero}}, false
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *{{.Name}}HashSet) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for i := range s.entries {
		if s.entries[i].occupied && predicate(s.entries[i].key) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *{{.Name}}HashSet) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for i := range s.entries {
		if s.entries[i].occupied && !predicate(s.entries[i].key) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *{{.Name}}HashSet) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for i := range s.entries {
		if s.entries[i].occupied && predicate(s.entries[i].key) {
			return false
		}
	}
	return true
}

// Union returns a new set containing all elements from both sets.
func (s *{{.Name}}HashSet) Union(other *{{.Name}}HashSet) *{{.Name}}HashSet {
	result := New{{.Name}}HashSetWithCapacity((s.size + other.size) * 2)
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
func (s *{{.Name}}HashSet) Intersect(other *{{.Name}}HashSet) *{{.Name}}HashSet {
	result := New{{.Name}}HashSet()
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
func (s *{{.Name}}HashSet) Difference(other *{{.Name}}HashSet) *{{.Name}}HashSet {
	result := New{{.Name}}HashSet()
	for i := range s.entries {
		if s.entries[i].occupied && !other.Contains(s.entries[i].key) {
			result.Add(s.entries[i].key)
		}
	}
	return result
}

// SymmetricDifference returns a new set containing elements in either set but not both.
func (s *{{.Name}}HashSet) SymmetricDifference(other *{{.Name}}HashSet) *{{.Name}}HashSet {
	result := New{{.Name}}HashSet()
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
func (s *{{.Name}}HashSet) ToSlice() []{{.GoType}} {
	result := make([]{{.GoType}}, 0, s.size)
	for i := range s.entries {
		if s.entries[i].occupied {
			result = append(result, s.entries[i].key)
		}
	}
	return result
}

// With returns the set after adding the value (fluent API).
func (s *{{.Name}}HashSet) With(value {{.GoType}}) *{{.Name}}HashSet {
	s.Add(value)
	return s
}

// Without returns the set after removing the value (fluent API).
func (s *{{.Name}}HashSet) Without(value {{.GoType}}) *{{.Name}}HashSet {
	s.Remove(value)
	return s
}

// WithAll returns the set after adding all values (fluent API).
func (s *{{.Name}}HashSet) WithAll(values ...{{.GoType}}) *{{.Name}}HashSet {
	s.AddAll(values...)
	return s
}

// WithoutAll returns the set after removing all given values (fluent API).
func (s *{{.Name}}HashSet) WithoutAll(values ...{{.GoType}}) *{{.Name}}HashSet {
	for _, v := range values {
		s.Remove(v)
	}
	return s
}

// ToImmutable returns an immutable copy of this set.
func (s *{{.Name}}HashSet) ToImmutable() *Immutable{{.Name}}HashSet {
	return Immutable{{.Name}}HashSetFrom(s)
}

// String returns a string representation of the set.
func (s *{{.Name}}HashSet) String() string {
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
func (s *{{.Name}}HashSet) Equals(other *{{.Name}}HashSet) bool {
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

func (s *{{.Name}}HashSet) hash(value {{.GoType}}) uint64 {
{{- if .IsBool}}
	return func() uint64 {
		if value {
			return 1
		} else {
			return 0
		}
	}()
{{- else}}
	return func() uint64 { h := {{.HashExpr}} * 0x9E3779B97F4A7C15; return h ^ (h >> 32) }()
{{- end}}
}

func (s *{{.Name}}HashSet) needsResize() bool {
	return (s.size+1)*4 >= len(s.entries)*3 // 0.75 load factor, integer math
}

func (s *{{.Name}}HashSet) resize() {
	oldEntries := s.entries
	newCap := len(oldEntries) * 2
	if newCap == 0 {
		newCap = {{.SnakeName}}HashSetDefaultCapacity
	}
	s.entries = make([]{{.SnakeName}}HashSetEntry, newCap)
	s.size = 0

	for i := range oldEntries {
		if oldEntries[i].occupied {
			s.Add(oldEntries[i].key)
		}
	}
}

func (s *{{.Name}}HashSet) rehashFrom(deleted int, mask int) {
	c := len(s.entries)
	idx := (deleted + 1) & mask
	for s.entries[idx].occupied {
		ideal := int(s.hash(s.entries[idx].key)) & mask
		distCurrent := (idx - ideal + c) & mask
		distGap := (deleted - ideal + c) & mask
		if distCurrent > distGap {
			s.entries[deleted] = s.entries[idx]
			s.entries[idx] = {{.SnakeName}}HashSetEntry{}
			deleted = idx
		}
		idx = (idx + 1) & mask
		if idx == deleted {
			break
		}
	}
}

func nextPowerOfTwo{{.Name}}HashSet(n int) int {
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

const immutableHashSetTmpl = genHeader + `package hashset

import (
	"iter"
)

// Immutable{{.Name}}HashSet is an immutable view of a {{.Name}}HashSet.
type Immutable{{.Name}}HashSet struct {
	delegate *{{.Name}}HashSet
}

// NewImmutable{{.Name}}HashSet creates an immutable set from the given values.
func NewImmutable{{.Name}}HashSet(values ...{{.GoType}}) *Immutable{{.Name}}HashSet {
	return &Immutable{{.Name}}HashSet{delegate: {{.Name}}HashSetOf(values...)}
}

// Immutable{{.Name}}HashSetFrom creates an immutable copy of a mutable set.
func Immutable{{.Name}}HashSetFrom(s *{{.Name}}HashSet) *Immutable{{.Name}}HashSet {
	copy := {{.Name}}HashSetOf(s.ToSlice()...)
	return &Immutable{{.Name}}HashSet{delegate: copy}
}

// Contains returns true if the set contains the given value.
func (s *Immutable{{.Name}}HashSet) Contains(value {{.GoType}}) bool {
	return s.delegate.Contains(value)
}

// Size returns the number of elements.
func (s *Immutable{{.Name}}HashSet) Size() int {
	return s.delegate.Size()
}

// IsEmpty returns true if the set contains no elements.
func (s *Immutable{{.Name}}HashSet) IsEmpty() bool {
	return s.delegate.IsEmpty()
}

// All returns an iter.Seq that yields all elements.
func (s *Immutable{{.Name}}HashSet) All() iter.Seq[{{.GoType}}] {
	return s.delegate.All()
}

// ForEach calls the given function for each element.
func (s *Immutable{{.Name}}HashSet) ForEach(f func({{.GoType}})) {
	s.delegate.ForEach(f)
}

// Select returns a new immutable set with elements satisfying the predicate.
func (s *Immutable{{.Name}}HashSet) Select(predicate func({{.GoType}}) bool) *Immutable{{.Name}}HashSet {
	return &Immutable{{.Name}}HashSet{delegate: s.delegate.Select(predicate)}
}

// Reject returns a new immutable set with elements not satisfying the predicate.
func (s *Immutable{{.Name}}HashSet) Reject(predicate func({{.GoType}}) bool) *Immutable{{.Name}}HashSet {
	return &Immutable{{.Name}}HashSet{delegate: s.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *Immutable{{.Name}}HashSet) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	return s.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *Immutable{{.Name}}HashSet) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	return s.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *Immutable{{.Name}}HashSet) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	return s.delegate.NoneSatisfy(predicate)
}

// Union returns a new immutable set with elements from both sets.
func (s *Immutable{{.Name}}HashSet) Union(other *Immutable{{.Name}}HashSet) *Immutable{{.Name}}HashSet {
	return &Immutable{{.Name}}HashSet{delegate: s.delegate.Union(other.delegate)}
}

// Intersect returns a new immutable set with elements in both sets.
func (s *Immutable{{.Name}}HashSet) Intersect(other *Immutable{{.Name}}HashSet) *Immutable{{.Name}}HashSet {
	return &Immutable{{.Name}}HashSet{delegate: s.delegate.Intersect(other.delegate)}
}

// Difference returns a new immutable set with elements in this but not other.
func (s *Immutable{{.Name}}HashSet) Difference(other *Immutable{{.Name}}HashSet) *Immutable{{.Name}}HashSet {
	return &Immutable{{.Name}}HashSet{delegate: s.delegate.Difference(other.delegate)}
}

// ToSlice returns all elements as a slice.
func (s *Immutable{{.Name}}HashSet) ToSlice() []{{.GoType}} {
	return s.delegate.ToSlice()
}

// String returns a string representation.
func (s *Immutable{{.Name}}HashSet) String() string {
	return s.delegate.String()
}

// Equals returns true if the other immutable set has the same elements.
func (s *Immutable{{.Name}}HashSet) Equals(other *Immutable{{.Name}}HashSet) bool {
	return s.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this set.
func (s *Immutable{{.Name}}HashSet) ToMutable() *{{.Name}}HashSet {
	return {{.Name}}HashSetOf(s.ToSlice()...)
}
`

const synchronizedHashSetTmpl = genHeader + `package hashset

import (
	"iter"
	"sync"
	"unsafe"
)

// Synchronized{{.Name}}HashSet is a thread-safe wrapper around {{.Name}}HashSet.
//
// Read methods hold an RLock; writes hold a Lock. Methods that take a
// caller-supplied function (Select, ForEach, AnySatisfy, …) snapshot
// the backing set under RLock and release it before invoking the
// callback, so the callback is free to call back into the wrapper
// without deadlocking.
//
// Methods that return a new set (Select, Reject, Union, Intersect,
// Difference, SymmetricDifference) return an unwrapped *{{.Name}}HashSet;
// the caller owns it.
type Synchronized{{.Name}}HashSet struct {
	delegate *{{.Name}}HashSet
	mu       sync.RWMutex
}

// NewSynchronized{{.Name}}HashSet creates a new thread-safe empty set.
func NewSynchronized{{.Name}}HashSet() *Synchronized{{.Name}}HashSet {
	return &Synchronized{{.Name}}HashSet{delegate: New{{.Name}}HashSet()}
}

// NewSynchronized{{.Name}}HashSetFrom wraps an existing set. The
// wrapper takes ownership — callers must not mutate the delegate
// directly without locking.
func NewSynchronized{{.Name}}HashSetFrom(s *{{.Name}}HashSet) *Synchronized{{.Name}}HashSet {
	return &Synchronized{{.Name}}HashSet{delegate: s}
}

// Synchronized{{.Name}}HashSetOf constructs a synchronized set from values.
func Synchronized{{.Name}}HashSetOf(values ...{{.GoType}}) *Synchronized{{.Name}}HashSet {
	s := New{{.Name}}HashSet()
	for _, v := range values {
		s.Add(v)
	}
	return &Synchronized{{.Name}}HashSet{delegate: s}
}

// snapshot returns a defensive copy of the set's elements under RLock.
func (s *Synchronized{{.Name}}HashSet) snapshot() []{{.GoType}} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToSlice()
}

// ── writes ────────────────────────────────────────────────────────────

func (s *Synchronized{{.Name}}HashSet) Add(value {{.GoType}}) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.delegate.Add(value)
}

func (s *Synchronized{{.Name}}HashSet) AddAll(values ...{{.GoType}}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.AddAll(values...)
}

func (s *Synchronized{{.Name}}HashSet) Remove(value {{.GoType}}) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.delegate.Remove(value)
}

func (s *Synchronized{{.Name}}HashSet) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.Clear()
}

// ── simple reads ──────────────────────────────────────────────────────

func (s *Synchronized{{.Name}}HashSet) Contains(value {{.GoType}}) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Contains(value)
}

func (s *Synchronized{{.Name}}HashSet) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Size()
}

func (s *Synchronized{{.Name}}HashSet) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.IsEmpty()
}

func (s *Synchronized{{.Name}}HashSet) ToSlice() []{{.GoType}} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToSlice()
}

func (s *Synchronized{{.Name}}HashSet) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.String()
}

// ── iteration ────────────────────────────────────────────────────────

// All returns an iter.Seq over a snapshot. Iteration is lock-free.
func (s *Synchronized{{.Name}}HashSet) All() iter.Seq[{{.GoType}}] {
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

func (s *Synchronized{{.Name}}HashSet) ForEach(f func({{.GoType}})) {
	for _, v := range s.snapshot() {
		f(v)
	}
}

func (s *Synchronized{{.Name}}HashSet) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (s *Synchronized{{.Name}}HashSet) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range s.snapshot() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (s *Synchronized{{.Name}}HashSet) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return false
		}
	}
	return true
}

func (s *Synchronized{{.Name}}HashSet) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return v, true
		}
	}
	var zero {{.GoType}}
	return zero, false
}

// ── functional that return a new set ─────────────────────────────────

func (s *Synchronized{{.Name}}HashSet) Select(predicate func({{.GoType}}) bool) *{{.Name}}HashSet {
	snapshot := s.snapshot()
	result := New{{.Name}}HashSet()
	for _, v := range snapshot {
		if predicate(v) {
			result.Add(v)
		}
	}
	return result
}

func (s *Synchronized{{.Name}}HashSet) Reject(predicate func({{.GoType}}) bool) *{{.Name}}HashSet {
	snapshot := s.snapshot()
	result := New{{.Name}}HashSet()
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
func (s *Synchronized{{.Name}}HashSet) lockPair(other *Synchronized{{.Name}}HashSet) func() {
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

func (s *Synchronized{{.Name}}HashSet) Union(other *Synchronized{{.Name}}HashSet) *{{.Name}}HashSet {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Union(other.delegate)
}

func (s *Synchronized{{.Name}}HashSet) Intersect(other *Synchronized{{.Name}}HashSet) *{{.Name}}HashSet {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Intersect(other.delegate)
}

func (s *Synchronized{{.Name}}HashSet) Difference(other *Synchronized{{.Name}}HashSet) *{{.Name}}HashSet {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Difference(other.delegate)
}

func (s *Synchronized{{.Name}}HashSet) SymmetricDifference(other *Synchronized{{.Name}}HashSet) *{{.Name}}HashSet {
	release := s.lockPair(other)
	defer release()
	return s.delegate.SymmetricDifference(other.delegate)
}

// ── fluent mutators ───────────────────────────────────────────────────

func (s *Synchronized{{.Name}}HashSet) With(value {{.GoType}}) *Synchronized{{.Name}}HashSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.With(value)
	return s
}

func (s *Synchronized{{.Name}}HashSet) WithAll(values ...{{.GoType}}) *Synchronized{{.Name}}HashSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.WithAll(values...)
	return s
}

func (s *Synchronized{{.Name}}HashSet) Without(value {{.GoType}}) *Synchronized{{.Name}}HashSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.Without(value)
	return s
}

func (s *Synchronized{{.Name}}HashSet) WithoutAll(values ...{{.GoType}}) *Synchronized{{.Name}}HashSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.WithoutAll(values...)
	return s
}

// ── conversions ───────────────────────────────────────────────────────

func (s *Synchronized{{.Name}}HashSet) ToImmutable() *Immutable{{.Name}}HashSet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToImmutable()
}

// Equals compares by contents. Locks are acquired in pointer-address
// order to prevent deadlocks under concurrent A.Equals(B) / B.Equals(A).
func (s *Synchronized{{.Name}}HashSet) Equals(other *Synchronized{{.Name}}HashSet) bool {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Equals(other.delegate)
}
`
