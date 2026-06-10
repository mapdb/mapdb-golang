package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

// alData is the per-primitive view the arraylist templates iterate over.
type alData struct {
	Name     string // Int32, Float32, Char (identifier stem)
	GoType   string // int32, float32, uint16
	SnakeName string
	Zero      string // "0" or "0.0" (zero literal for this element type)
	IsFloat   bool
	BitsFn    string // math.Float32bits / math.Float64bits (floats only)
	BitsType  string // uint32 / uint64 (float bit-pattern map key)
	CmpFn     string // cmpFloat32 / cmpFloat64 (floats only)
	SumType   string // int64 (integers/char) or the element GoType (floats)
	SumDoc    string // trailing words on the Sum doc comment
}

// genArrayList writes the per-primitive array list sources (base, immutable,
// synchronized variants) plus the shared cmp_float.go into the current working
// directory. Invoked from arraylist/ via go:generate.
func genArrayList() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	base := template.Must(template.New("al-base").Parse(arrayListTmpl))
	immutable := template.Must(template.New("al-immutable").Parse(immutableArrayListTmpl))
	synchronized := template.Must(template.New("al-sync").Parse(synchronizedArrayListTmpl))

	write := func(name string, tmpl *template.Template, data alData) error {
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
		data := alData{
			Name:      p.Name,
			GoType:    p.GoType,
			SnakeName: p.SnakeName,
			IsFloat:   p.IsFloating,
		}
		if p.IsFloating {
			data.Zero = "0.0"
			data.SumType = p.GoType
			data.SumDoc = "."
			bits := "Float32bits"
			data.BitsType = "uint32"
			data.CmpFn = "cmpFloat32"
			if p.ByteSize == 8 {
				bits = "Float64bits"
				data.BitsType = "uint64"
				data.CmpFn = "cmpFloat64"
			}
			data.BitsFn = "math." + bits
		} else {
			data.Zero = "0"
			data.SumType = "int64"
			data.SumDoc = " as int64 to avoid overflow."
		}

		if err := write(p.SnakeName+"_array_list.go", base, data); err != nil {
			return err
		}
		if err := write("immutable_"+p.SnakeName+"_array_list.go", immutable, data); err != nil {
			return err
		}
		if err := write("synchronized_"+p.SnakeName+"_array_list.go", synchronized, data); err != nil {
			return err
		}
	}

	return genCmpFloat("arraylist")
}

const arrayListTmpl = genHeader + `package arraylist

import (
	"fmt"
	"iter"
{{- if .IsFloat}}
	"math"
{{- end}}
	"sort"
	"strings"
)

// {{.Name}}ArrayList is a resizable array-backed list of {{.GoType}} values.
// Length is always len(l.items); there is no separate size counter.
type {{.Name}}ArrayList struct {
	items []{{.GoType}}
}

// New{{.Name}}ArrayList creates a new empty {{.Name}}ArrayList.
func New{{.Name}}ArrayList() *{{.Name}}ArrayList {
	return &{{.Name}}ArrayList{items: make([]{{.GoType}}, 0, 16)}
}

// New{{.Name}}ArrayListWithCapacity creates a new empty {{.Name}}ArrayList with the given initial capacity.
func New{{.Name}}ArrayListWithCapacity(capacity int) *{{.Name}}ArrayList {
	return &{{.Name}}ArrayList{items: make([]{{.GoType}}, 0, capacity)}
}

// {{.Name}}ArrayListOf creates a new {{.Name}}ArrayList from the given values.
func {{.Name}}ArrayListOf(values ...{{.GoType}}) *{{.Name}}ArrayList {
	l := &{{.Name}}ArrayList{items: make([]{{.GoType}}, len(values))}
	copy(l.items, values)
	return l
}

// Add appends a value to the end of the list.
func (l *{{.Name}}ArrayList) Add(value {{.GoType}}) {
	l.items = append(l.items, value)
}

// AddAll appends all values to the end of the list.
func (l *{{.Name}}ArrayList) AddAll(values ...{{.GoType}}) {
	l.items = append(l.items, values...)
}

// Get returns the value at the given index, or an error if the index is out of bounds.
func (l *{{.Name}}ArrayList) Get(index int) ({{.GoType}}, error) {
	if index < 0 || index >= len(l.items) {
		return {{.Zero}}, fmt.Errorf("{{.Name}}ArrayList: index out of bounds: %d (size %d)", index, len(l.items))
	}
	return l.items[index], nil
}

// Set sets the value at the given index, returning the previous value.
// Returns an error if the index is out of bounds.
func (l *{{.Name}}ArrayList) Set(index int, value {{.GoType}}) ({{.GoType}}, error) {
	if index < 0 || index >= len(l.items) {
		return {{.Zero}}, fmt.Errorf("{{.Name}}ArrayList: index out of bounds: %d (size %d)", index, len(l.items))
	}
	old := l.items[index]
	l.items[index] = value
	return old, nil
}

// RemoveAtIndex removes the value at the given index and returns it.
// Returns an error if the index is out of bounds.
func (l *{{.Name}}ArrayList) RemoveAtIndex(index int) ({{.GoType}}, error) {
	if index < 0 || index >= len(l.items) {
		return {{.Zero}}, fmt.Errorf("{{.Name}}ArrayList: index out of bounds: %d (size %d)", index, len(l.items))
	}
	old := l.items[index]
	copy(l.items[index:], l.items[index+1:])
	l.items = l.items[:len(l.items)-1]
	return old, nil
}

// Remove removes the first occurrence of the value. Returns true if found and removed.
func (l *{{.Name}}ArrayList) Remove(value {{.GoType}}) bool {
	for i, v := range l.items {
		if {{if .IsFloat}}{{.BitsFn}}(v) == {{.BitsFn}}(value){{else}}v == value{{end}} {
			_, _ = l.RemoveAtIndex(i)
			return true
		}
	}
	return false
}

// Contains returns true if the list contains the given value.
func (l *{{.Name}}ArrayList) Contains(value {{.GoType}}) bool {
	for _, v := range l.items {
		if {{if .IsFloat}}{{.BitsFn}}(v) == {{.BitsFn}}(value){{else}}v == value{{end}} {
			return true
		}
	}
	return false
}

// IndexOf returns the index of the first occurrence of the value, or -1 if not found.
func (l *{{.Name}}ArrayList) IndexOf(value {{.GoType}}) int {
	for i, v := range l.items {
		if {{if .IsFloat}}{{.BitsFn}}(v) == {{.BitsFn}}(value){{else}}v == value{{end}} {
			return i
		}
	}
	return -1
}

// Size returns the number of elements in the list.
func (l *{{.Name}}ArrayList) Size() int { return len(l.items) }

// IsEmpty returns true if the list contains no elements.
func (l *{{.Name}}ArrayList) IsEmpty() bool { return len(l.items) == 0 }

// Clear removes all elements from the list.
func (l *{{.Name}}ArrayList) Clear() { l.items = l.items[:0] }

// All returns an iter.Seq that yields all elements in order.
func (l *{{.Name}}ArrayList) All() iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
		for _, v := range l.items {
			if !yield(v) {
				return
			}
		}
	}
}

// AllWithIndex returns an iter.Seq2 that yields (index, value) pairs.
func (l *{{.Name}}ArrayList) AllWithIndex() iter.Seq2[int, {{.GoType}}] {
	return func(yield func(int, {{.GoType}}) bool) {
		for i, v := range l.items {
			if !yield(i, v) {
				return
			}
		}
	}
}

// ForEach calls the given function for each element.
func (l *{{.Name}}ArrayList) ForEach(f func({{.GoType}})) {
	for _, v := range l.items {
		f(v)
	}
}

// ForEachWithIndex calls the given function with each element and its index.
func (l *{{.Name}}ArrayList) ForEachWithIndex(f func({{.GoType}}, int)) {
	for i, v := range l.items {
		f(v, i)
	}
}

// Select returns a new list containing only elements that satisfy the predicate.
func (l *{{.Name}}ArrayList) Select(predicate func({{.GoType}}) bool) *{{.Name}}ArrayList {
	result := New{{.Name}}ArrayList()
	for _, v := range l.items {
		if predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Reject returns a new list containing only elements that do not satisfy the predicate.
func (l *{{.Name}}ArrayList) Reject(predicate func({{.GoType}}) bool) *{{.Name}}ArrayList {
	result := New{{.Name}}ArrayList()
	for _, v := range l.items {
		if !predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Detect returns the first element that satisfies the predicate, or the zero value and false.
func (l *{{.Name}}ArrayList) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	for _, v := range l.items {
		if predicate(v) {
			return v, true
		}
	}
	return {{.Zero}}, false
}

// AnySatisfy returns true if any element satisfies the predicate.
func (l *{{.Name}}ArrayList) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range l.items {
		if predicate(v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (l *{{.Name}}ArrayList) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range l.items {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (l *{{.Name}}ArrayList) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range l.items {
		if predicate(v) {
			return false
		}
	}
	return true
}

// Count returns the number of elements that satisfy the predicate.
func (l *{{.Name}}ArrayList) Count(predicate func({{.GoType}}) bool) int {
	count := 0
	for _, v := range l.items {
		if predicate(v) {
			count++
		}
	}
	return count
}

// InjectInto performs a left fold over the list.
func (l *{{.Name}}ArrayList) InjectInto(initial {{.GoType}}, f func({{.GoType}}, {{.GoType}}) {{.GoType}}) {{.GoType}} {
	result := initial
	for _, v := range l.items {
		result = f(result, v)
	}
	return result
}

// Sum returns the sum of all elements{{.SumDoc}}
func (l *{{.Name}}ArrayList) Sum() {{.SumType}} {
	var sum {{.SumType}}
	for _, v := range l.items {
		sum += {{if .IsFloat}}v{{else}}int64(v){{end}}
	}
	return sum
}

// Min returns the minimum element, or the zero value and false if empty.
func (l *{{.Name}}ArrayList) Min() ({{.GoType}}, bool) {
	if len(l.items) == 0 {
		return {{.Zero}}, false
	}
	min := l.items[0]
	for _, v := range l.items[1:] {
		if {{if .IsFloat}}{{.CmpFn}}(v, min) < 0{{else}}v < min{{end}} {
			min = v
		}
	}
	return min, true
}

// Max returns the maximum element, or the zero value and false if empty.
func (l *{{.Name}}ArrayList) Max() ({{.GoType}}, bool) {
	if len(l.items) == 0 {
		return {{.Zero}}, false
	}
	max := l.items[0]
	for _, v := range l.items[1:] {
		if {{if .IsFloat}}{{.CmpFn}}(v, max) > 0{{else}}v > max{{end}} {
			max = v
		}
	}
	return max, true
}

// Sort sorts the list in ascending order.
func (l *{{.Name}}ArrayList) Sort() {
	sort.Slice(l.items, func(i, j int) bool {
		return {{if .IsFloat}}{{.CmpFn}}(l.items[i], l.items[j]) < 0{{else}}l.items[i] < l.items[j]{{end}}
	})
}

// SortWithComparator sorts the list using the given comparison function.
func (l *{{.Name}}ArrayList) SortWithComparator(less func({{.GoType}}, {{.GoType}}) bool) {
	sort.Slice(l.items, func(i, j int) bool {
		return less(l.items[i], l.items[j])
	})
}

// BinarySearch searches for a value in a sorted list. Returns the index and true if found.
func (l *{{.Name}}ArrayList) BinarySearch(value {{.GoType}}) (int, bool) {
	lo, hi := 0, len(l.items)-1
	for lo <= hi {
		mid := lo + (hi-lo)/2
		if {{if .IsFloat}}{{.BitsFn}}(l.items[mid]) == {{.BitsFn}}(value){{else}}l.items[mid] == value{{end}} {
			return mid, true
		}
		if l.items[mid] < value {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return lo, false
}

// Reversed returns a new list with elements in reverse order.
func (l *{{.Name}}ArrayList) Reversed() *{{.Name}}ArrayList {
	n := len(l.items)
	result := New{{.Name}}ArrayListWithCapacity(n)
	for i := n - 1; i >= 0; i-- {
		result.Add(l.items[i])
	}
	return result
}

// Distinct returns a new list with duplicate elements removed (preserving first occurrence order).
func (l *{{.Name}}ArrayList) Distinct() *{{.Name}}ArrayList {
{{- if .IsFloat}}
	// Key by bit pattern so NaN dedupes against itself and -0 stays distinct
	// from +0 (a plain map[{{.GoType}}] would never match NaN and would collapse
	// the two zeroes together).
	seen := make(map[{{.BitsType}}]struct{})
	result := New{{.Name}}ArrayList()
	for _, v := range l.items {
		bits := {{.BitsFn}}(v)
		if _, ok := seen[bits]; !ok {
			seen[bits] = struct{}{}
			result.Add(v)
		}
	}
	return result
{{- else}}
	seen := make(map[{{.GoType}}]struct{})
	result := New{{.Name}}ArrayList()
	for _, v := range l.items {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result.Add(v)
		}
	}
	return result
{{- end}}
}

// ToSlice returns a copy of the list elements as a slice.
func (l *{{.Name}}ArrayList) ToSlice() []{{.GoType}} {
	result := make([]{{.GoType}}, len(l.items))
	copy(result, l.items)
	return result
}

// With returns the list after adding the value (fluent API).
func (l *{{.Name}}ArrayList) With(value {{.GoType}}) *{{.Name}}ArrayList {
	l.Add(value)
	return l
}

// Without returns the list after removing the first occurrence of value (fluent API).
func (l *{{.Name}}ArrayList) Without(value {{.GoType}}) *{{.Name}}ArrayList {
	l.Remove(value)
	return l
}

// String returns a string representation of the list.
func (l *{{.Name}}ArrayList) String() string {
	if len(l.items) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteString("[")
	for i, v := range l.items {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v", v)
	}
	sb.WriteString("]")
	return sb.String()
}

// Equals returns true if the other list has the same elements in the same order.
func (l *{{.Name}}ArrayList) Equals(other *{{.Name}}ArrayList) bool {
	if len(l.items) != len(other.items) {
		return false
	}
	for i, v := range l.items {
		if !({{if .IsFloat}}{{.BitsFn}}(v) == {{.BitsFn}}(other.items[i]){{else}}v == other.items[i]{{end}}) {
			return false
		}
	}
	return true
}

// WithAll returns the list after adding all values (fluent API).
func (l *{{.Name}}ArrayList) WithAll(values ...{{.GoType}}) *{{.Name}}ArrayList {
	l.AddAll(values...)
	return l
}

// WithoutAll removes every occurrence of any of the given values.
// Compacts in place — keeps the existing backing storage and avoids
// the temporary-list allocation the previous implementation made.
func (l *{{.Name}}ArrayList) WithoutAll(values ...{{.GoType}}) *{{.Name}}ArrayList {
	if len(values) == 0 || len(l.items) == 0 {
		return l
	}
{{- if .IsFloat}}
	// Key by bit pattern so NaN values are actually removed and -0 is not
	// conflated with +0 (a plain map[{{.GoType}}] never matches NaN and treats
	// the two zeroes as equal).
	remove := make(map[{{.BitsType}}]struct{}, len(values))
	for _, v := range values {
		remove[{{.BitsFn}}(v)] = struct{}{}
	}
{{- else}}
	remove := make(map[{{.GoType}}]struct{}, len(values))
	for _, v := range values {
		remove[v] = struct{}{}
	}
{{- end}}
	// Two-index compaction: write is the cursor into the kept portion,
	// read iterates every original element.
	write := 0
	for _, v := range l.items {
		if _, skip := remove[{{if .IsFloat}}{{.BitsFn}}(v){{else}}v{{end}}]; skip {
			continue
		}
		l.items[write] = v
		write++
	}
	// Zero out the tail so GC can reclaim any references the slots held
	// (cheap no-op for plain numeric {{.GoType}} but keeps the template honest
	// for future value-carrying specialisations).
	for i := write; i < len(l.items); i++ {
		l.items[i] = {{.Zero}}
	}
	l.items = l.items[:write]
	return l
}

// ToImmutable returns an immutable copy of this list.
func (l *{{.Name}}ArrayList) ToImmutable() *Immutable{{.Name}}ArrayList {
	return Immutable{{.Name}}ArrayListFrom(l)
}
`

const immutableArrayListTmpl = genHeader + `package arraylist

import (
	"iter"
)

// Immutable{{.Name}}ArrayList is an immutable view of a {{.Name}}ArrayList.
type Immutable{{.Name}}ArrayList struct {
	delegate *{{.Name}}ArrayList
}

// NewImmutable{{.Name}}ArrayList creates an immutable list from the given values.
func NewImmutable{{.Name}}ArrayList(values ...{{.GoType}}) *Immutable{{.Name}}ArrayList {
	return &Immutable{{.Name}}ArrayList{delegate: {{.Name}}ArrayListOf(values...)}
}

// Immutable{{.Name}}ArrayListFrom creates an immutable copy of a mutable list.
func Immutable{{.Name}}ArrayListFrom(l *{{.Name}}ArrayList) *Immutable{{.Name}}ArrayList {
	copy := {{.Name}}ArrayListOf(l.ToSlice()...)
	return &Immutable{{.Name}}ArrayList{delegate: copy}
}

// Get returns the value at the given index, or an error if the index is out of bounds.
func (l *Immutable{{.Name}}ArrayList) Get(index int) ({{.GoType}}, error) {
	return l.delegate.Get(index)
}

// Size returns the number of elements.
func (l *Immutable{{.Name}}ArrayList) Size() int {
	return l.delegate.Size()
}

// IsEmpty returns true if the list contains no elements.
func (l *Immutable{{.Name}}ArrayList) IsEmpty() bool {
	return l.delegate.IsEmpty()
}

// Contains returns true if the list contains the given value.
func (l *Immutable{{.Name}}ArrayList) Contains(value {{.GoType}}) bool {
	return l.delegate.Contains(value)
}

// IndexOf returns the index of the first occurrence, or -1.
func (l *Immutable{{.Name}}ArrayList) IndexOf(value {{.GoType}}) int {
	return l.delegate.IndexOf(value)
}

// All returns an iter.Seq that yields all elements in order.
func (l *Immutable{{.Name}}ArrayList) All() iter.Seq[{{.GoType}}] {
	return l.delegate.All()
}

// AllWithIndex returns an iter.Seq2 that yields (index, value) pairs.
func (l *Immutable{{.Name}}ArrayList) AllWithIndex() iter.Seq2[int, {{.GoType}}] {
	return l.delegate.AllWithIndex()
}

// ForEach calls the given function for each element.
func (l *Immutable{{.Name}}ArrayList) ForEach(f func({{.GoType}})) {
	l.delegate.ForEach(f)
}

// Select returns a new immutable list with elements satisfying the predicate.
func (l *Immutable{{.Name}}ArrayList) Select(predicate func({{.GoType}}) bool) *Immutable{{.Name}}ArrayList {
	return &Immutable{{.Name}}ArrayList{delegate: l.delegate.Select(predicate)}
}

// Reject returns a new immutable list with elements not satisfying the predicate.
func (l *Immutable{{.Name}}ArrayList) Reject(predicate func({{.GoType}}) bool) *Immutable{{.Name}}ArrayList {
	return &Immutable{{.Name}}ArrayList{delegate: l.delegate.Reject(predicate)}
}

// Detect returns the first element satisfying the predicate, or zero and false.
func (l *Immutable{{.Name}}ArrayList) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	return l.delegate.Detect(predicate)
}

// AnySatisfy returns true if any element satisfies the predicate.
func (l *Immutable{{.Name}}ArrayList) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	return l.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (l *Immutable{{.Name}}ArrayList) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	return l.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (l *Immutable{{.Name}}ArrayList) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	return l.delegate.NoneSatisfy(predicate)
}

// Count returns the number of elements satisfying the predicate.
func (l *Immutable{{.Name}}ArrayList) Count(predicate func({{.GoType}}) bool) int {
	return l.delegate.Count(predicate)
}

// Reversed returns a new immutable list in reverse order.
func (l *Immutable{{.Name}}ArrayList) Reversed() *Immutable{{.Name}}ArrayList {
	return &Immutable{{.Name}}ArrayList{delegate: l.delegate.Reversed()}
}

// Distinct returns a new immutable list with duplicates removed.
func (l *Immutable{{.Name}}ArrayList) Distinct() *Immutable{{.Name}}ArrayList {
	return &Immutable{{.Name}}ArrayList{delegate: l.delegate.Distinct()}
}

// ToSlice returns a copy of all elements as a slice.
func (l *Immutable{{.Name}}ArrayList) ToSlice() []{{.GoType}} {
	return l.delegate.ToSlice()
}

// String returns a string representation.
func (l *Immutable{{.Name}}ArrayList) String() string {
	return l.delegate.String()
}

// Equals returns true if the other immutable list has the same elements in order.
func (l *Immutable{{.Name}}ArrayList) Equals(other *Immutable{{.Name}}ArrayList) bool {
	return l.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this list.
func (l *Immutable{{.Name}}ArrayList) ToMutable() *{{.Name}}ArrayList {
	return {{.Name}}ArrayListOf(l.ToSlice()...)
}
`

const synchronizedArrayListTmpl = genHeader + `package arraylist

import (
	"iter"
	"sync"
	"unsafe"
)

// Synchronized{{.Name}}ArrayList is a thread-safe wrapper around {{.Name}}ArrayList.
//
// Read methods hold an RLock; writes hold a Lock. Methods that take
// a caller-supplied function (Select, ForEach, InjectInto, …) snapshot
// the backing slice under RLock and release it before invoking the
// callback, so the callback is free to call back into the wrapper
// without deadlocking.
//
// Methods that return a fresh collection (Select, Reject, Distinct,
// Reversed) return an unwrapped *{{.Name}}ArrayList: the caller owns it and
// is free to add their own synchronisation if they need it.
type Synchronized{{.Name}}ArrayList struct {
	delegate *{{.Name}}ArrayList
	mu       sync.RWMutex
}

// NewSynchronized{{.Name}}ArrayList creates a new thread-safe empty list.
func NewSynchronized{{.Name}}ArrayList() *Synchronized{{.Name}}ArrayList {
	return &Synchronized{{.Name}}ArrayList{delegate: New{{.Name}}ArrayList()}
}

// NewSynchronized{{.Name}}ArrayListWithCapacity creates a new thread-safe
// empty list with the given initial capacity.
func NewSynchronized{{.Name}}ArrayListWithCapacity(capacity int) *Synchronized{{.Name}}ArrayList {
	return &Synchronized{{.Name}}ArrayList{delegate: New{{.Name}}ArrayListWithCapacity(capacity)}
}

// NewSynchronized{{.Name}}ArrayListFrom wraps an existing list. The
// wrapper takes ownership of the delegate — callers must not continue
// to mutate it directly without locking.
func NewSynchronized{{.Name}}ArrayListFrom(l *{{.Name}}ArrayList) *Synchronized{{.Name}}ArrayList {
	return &Synchronized{{.Name}}ArrayList{delegate: l}
}

// Synchronized{{.Name}}ArrayListOf creates a new thread-safe list
// containing the given values in order.
func Synchronized{{.Name}}ArrayListOf(values ...{{.GoType}}) *Synchronized{{.Name}}ArrayList {
	return &Synchronized{{.Name}}ArrayList{delegate: {{.Name}}ArrayListOf(values...)}
}

// snapshot returns a defensive copy of the backing slice taken under
// RLock. Callers iterate the snapshot without holding the lock.
func (l *Synchronized{{.Name}}ArrayList) snapshot() []{{.GoType}} {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.ToSlice()
}

// ── simple writes ─────────────────────────────────────────────────────

func (l *Synchronized{{.Name}}ArrayList) Add(value {{.GoType}}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.Add(value)
}

func (l *Synchronized{{.Name}}ArrayList) AddAll(values ...{{.GoType}}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.AddAll(values...)
}

func (l *Synchronized{{.Name}}ArrayList) Set(index int, value {{.GoType}}) ({{.GoType}}, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.delegate.Set(index, value)
}

func (l *Synchronized{{.Name}}ArrayList) RemoveAtIndex(index int) ({{.GoType}}, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.delegate.RemoveAtIndex(index)
}

func (l *Synchronized{{.Name}}ArrayList) Remove(value {{.GoType}}) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.delegate.Remove(value)
}

func (l *Synchronized{{.Name}}ArrayList) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.Clear()
}

// Sort sorts the backing list in place. Holds the write lock for the
// duration; do not call back into this wrapper from a custom comparator.
func (l *Synchronized{{.Name}}ArrayList) Sort() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.Sort()
}

// SortWithComparator sorts using the given less function, under the
// write lock. The comparator must not call back into this wrapper.
func (l *Synchronized{{.Name}}ArrayList) SortWithComparator(less func({{.GoType}}, {{.GoType}}) bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.SortWithComparator(less)
}

// ── simple reads ──────────────────────────────────────────────────────

func (l *Synchronized{{.Name}}ArrayList) Get(index int) ({{.GoType}}, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.Get(index)
}

func (l *Synchronized{{.Name}}ArrayList) Contains(value {{.GoType}}) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.Contains(value)
}

func (l *Synchronized{{.Name}}ArrayList) IndexOf(value {{.GoType}}) int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.IndexOf(value)
}

func (l *Synchronized{{.Name}}ArrayList) Size() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.Size()
}

func (l *Synchronized{{.Name}}ArrayList) IsEmpty() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.IsEmpty()
}

func (l *Synchronized{{.Name}}ArrayList) ToSlice() []{{.GoType}} {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.ToSlice()
}

func (l *Synchronized{{.Name}}ArrayList) String() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.String()
}

func (l *Synchronized{{.Name}}ArrayList) Sum() {{.SumType}} {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.Sum()
}

func (l *Synchronized{{.Name}}ArrayList) Min() ({{.GoType}}, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.Min()
}

func (l *Synchronized{{.Name}}ArrayList) Max() ({{.GoType}}, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.Max()
}

// BinarySearch requires the delegate to be sorted. Callers must
// ensure that (e.g. by calling Sort() beforehand, both happening
// before any concurrent Add).
func (l *Synchronized{{.Name}}ArrayList) BinarySearch(value {{.GoType}}) (int, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.BinarySearch(value)
}

// Equals returns true if the other list has the same elements in the
// same order. Both wrappers are locked under RLock to prevent torn
// reads; locks are acquired in pointer-address order so two goroutines
// calling A.Equals(B) and B.Equals(A) concurrently cannot deadlock.
func (l *Synchronized{{.Name}}ArrayList) Equals(other *Synchronized{{.Name}}ArrayList) bool {
	if l == other {
		l.mu.RLock()
		defer l.mu.RUnlock()
		return l.delegate.Equals(other.delegate)
	}
	first, second := l, other
	if uintptr(unsafe.Pointer(l)) > uintptr(unsafe.Pointer(other)) {
		first, second = other, l
	}
	first.mu.RLock()
	defer first.mu.RUnlock()
	second.mu.RLock()
	defer second.mu.RUnlock()
	return l.delegate.Equals(other.delegate)
}

// ── iteration (snapshot-based) ────────────────────────────────────────

// All returns an iter.Seq over a snapshot. Iteration is lock-free.
func (l *Synchronized{{.Name}}ArrayList) All() iter.Seq[{{.GoType}}] {
	snapshot := l.snapshot()
	return func(yield func({{.GoType}}) bool) {
		for _, v := range snapshot {
			if !yield(v) {
				return
			}
		}
	}
}

// AllWithIndex returns an iter.Seq2 over a snapshot. Iteration is lock-free.
func (l *Synchronized{{.Name}}ArrayList) AllWithIndex() iter.Seq2[int, {{.GoType}}] {
	snapshot := l.snapshot()
	return func(yield func(int, {{.GoType}}) bool) {
		for i, v := range snapshot {
			if !yield(i, v) {
				return
			}
		}
	}
}

// ── functional over snapshot ──────────────────────────────────────────

func (l *Synchronized{{.Name}}ArrayList) ForEach(f func({{.GoType}})) {
	for _, v := range l.snapshot() {
		f(v)
	}
}

func (l *Synchronized{{.Name}}ArrayList) ForEachWithIndex(f func({{.GoType}}, int)) {
	for i, v := range l.snapshot() {
		f(v, i)
	}
}

func (l *Synchronized{{.Name}}ArrayList) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range l.snapshot() {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (l *Synchronized{{.Name}}ArrayList) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range l.snapshot() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (l *Synchronized{{.Name}}ArrayList) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range l.snapshot() {
		if predicate(v) {
			return false
		}
	}
	return true
}

func (l *Synchronized{{.Name}}ArrayList) Count(predicate func({{.GoType}}) bool) int {
	n := 0
	for _, v := range l.snapshot() {
		if predicate(v) {
			n++
		}
	}
	return n
}

func (l *Synchronized{{.Name}}ArrayList) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	for _, v := range l.snapshot() {
		if predicate(v) {
			return v, true
		}
	}
	var zero {{.GoType}}
	return zero, false
}

func (l *Synchronized{{.Name}}ArrayList) InjectInto(initial {{.GoType}}, f func({{.GoType}}, {{.GoType}}) {{.GoType}}) {{.GoType}} {
	acc := initial
	for _, v := range l.snapshot() {
		acc = f(acc, v)
	}
	return acc
}

// ── functional that return a new list ─────────────────────────────────

// Select returns a new (unsynchronized) list of elements satisfying the predicate.
func (l *Synchronized{{.Name}}ArrayList) Select(predicate func({{.GoType}}) bool) *{{.Name}}ArrayList {
	snapshot := l.snapshot()
	result := New{{.Name}}ArrayList()
	for _, v := range snapshot {
		if predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Reject returns a new (unsynchronized) list of elements not satisfying the predicate.
func (l *Synchronized{{.Name}}ArrayList) Reject(predicate func({{.GoType}}) bool) *{{.Name}}ArrayList {
	snapshot := l.snapshot()
	result := New{{.Name}}ArrayList()
	for _, v := range snapshot {
		if !predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Distinct returns a new (unsynchronized) list with duplicates removed,
// order preserved.
func (l *Synchronized{{.Name}}ArrayList) Distinct() *{{.Name}}ArrayList {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.Distinct()
}

// Reversed returns a new (unsynchronized) list in reverse order.
func (l *Synchronized{{.Name}}ArrayList) Reversed() *{{.Name}}ArrayList {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.Reversed()
}

// ── fluent mutators ───────────────────────────────────────────────────
// All return the wrapper so chained calls stay thread-safe.

func (l *Synchronized{{.Name}}ArrayList) With(value {{.GoType}}) *Synchronized{{.Name}}ArrayList {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.With(value)
	return l
}

func (l *Synchronized{{.Name}}ArrayList) WithAll(values ...{{.GoType}}) *Synchronized{{.Name}}ArrayList {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.WithAll(values...)
	return l
}

func (l *Synchronized{{.Name}}ArrayList) Without(value {{.GoType}}) *Synchronized{{.Name}}ArrayList {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.Without(value)
	return l
}

func (l *Synchronized{{.Name}}ArrayList) WithoutAll(values ...{{.GoType}}) *Synchronized{{.Name}}ArrayList {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.WithoutAll(values...)
	return l
}

// ── conversions ───────────────────────────────────────────────────────

// ToImmutable returns an immutable copy of the underlying list taken
// while holding the read lock. The returned value is independent of
// this wrapper.
func (l *Synchronized{{.Name}}ArrayList) ToImmutable() *Immutable{{.Name}}ArrayList {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.ToImmutable()
}
`
