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
	Name      string // Int32, Float32, Char (identifier stem)
	GoType    string // int32, float32, uint16
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

	base := parse("al-base", arrayListTmpl)
	immutable := parse("al-immutable", immutableArrayListTmpl)
	synchronized := parse("al-sync", synchronizedArrayListTmpl)

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
{{- if not .IsFloat}}
	"cmp"
{{- end}}
	"fmt"
	"iter"
{{- if .IsFloat}}
	"math"
{{- end}}
	"slices"
	"strings"
)

// {{.Name}} is a resizable array-backed list of {{.GoType}} values.
// Length is always len(l.items); there is no separate size counter.
//
// The zero value is an empty, ready-to-use list.
type {{.Name}} struct {
	items []{{.GoType}}
}

// New{{.Name}} creates a new empty {{.Name}}.
func New{{.Name}}() *{{.Name}} {
	return &{{.Name}}{items: make([]{{.GoType}}, 0, 16)}
}

// New{{.Name}}WithCapacity creates a new empty {{.Name}} with the given initial capacity.
func New{{.Name}}WithCapacity(capacity int) *{{.Name}} {
	return &{{.Name}}{items: make([]{{.GoType}}, 0, capacity)}
}

// {{.Name}}Of creates a new {{.Name}} from the given values.
func {{.Name}}Of(values ...{{.GoType}}) *{{.Name}} {
	l := &{{.Name}}{items: make([]{{.GoType}}, len(values))}
	copy(l.items, values)
	return l
}

// Add appends a value to the end of the list.
func (l *{{.Name}}) Add(value {{.GoType}}) {
	l.items = append(l.items, value)
}

// AddAll appends all values to the end of the list.
func (l *{{.Name}}) AddAll(values ...{{.GoType}}) {
	l.items = append(l.items, values...)
}

// Get returns the value at the given index. It panics if the index is out of
// bounds, matching the semantics of a native Go slice.
func (l *{{.Name}}) Get(index int) {{.GoType}} {
	if index < 0 || index >= len(l.items) {
		panic(fmt.Sprintf("arraylist.{{.Name}}: index out of range [%d] with length %d", index, len(l.items)))
	}
	return l.items[index]
}

// Set sets the value at the given index, returning the previous value. It
// panics if the index is out of bounds, matching the semantics of a native
// Go slice.
func (l *{{.Name}}) Set(index int, value {{.GoType}}) {{.GoType}} {
	if index < 0 || index >= len(l.items) {
		panic(fmt.Sprintf("arraylist.{{.Name}}: index out of range [%d] with length %d", index, len(l.items)))
	}
	old := l.items[index]
	l.items[index] = value
	return old
}

// RemoveAtIndex removes the value at the given index and returns it. It panics
// if the index is out of bounds, matching the semantics of a native Go slice.
func (l *{{.Name}}) RemoveAtIndex(index int) {{.GoType}} {
	if index < 0 || index >= len(l.items) {
		panic(fmt.Sprintf("arraylist.{{.Name}}: index out of range [%d] with length %d", index, len(l.items)))
	}
	old := l.items[index]
	copy(l.items[index:], l.items[index+1:])
	l.items = l.items[:len(l.items)-1]
	return old
}

// Remove removes the first occurrence of the value. Returns true if found and removed.
func (l *{{.Name}}) Remove(value {{.GoType}}) bool {
	for i, v := range l.items {
		if {{if .IsFloat}}{{.BitsFn}}(v) == {{.BitsFn}}(value){{else}}v == value{{end}} {
			l.RemoveAtIndex(i)
			return true
		}
	}
	return false
}

// Contains returns true if the list contains the given value.
func (l *{{.Name}}) Contains(value {{.GoType}}) bool {
	for _, v := range l.items {
		if {{if .IsFloat}}{{.BitsFn}}(v) == {{.BitsFn}}(value){{else}}v == value{{end}} {
			return true
		}
	}
	return false
}

// IndexOf returns the index of the first occurrence of the value, or -1 if not found.
func (l *{{.Name}}) IndexOf(value {{.GoType}}) int {
	for i, v := range l.items {
		if {{if .IsFloat}}{{.BitsFn}}(v) == {{.BitsFn}}(value){{else}}v == value{{end}} {
			return i
		}
	}
	return -1
}

// Len returns the number of elements in the list, matching the Go
// convention (sort.Interface, container/list, bytes.Buffer). Use
// l.Len() == 0 to test for emptiness.
func (l *{{.Name}}) Len() int { return len(l.items) }

// Clear removes all elements from the list.
func (l *{{.Name}}) Clear() { l.items = l.items[:0] }

// All returns an iter.Seq that yields all elements in order.
func (l *{{.Name}}) All() iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
		for _, v := range l.items {
			if !yield(v) {
				return
			}
		}
	}
}

// AllWithIndex returns an iter.Seq2 that yields (index, value) pairs.
func (l *{{.Name}}) AllWithIndex() iter.Seq2[int, {{.GoType}}] {
	return func(yield func(int, {{.GoType}}) bool) {
		for i, v := range l.items {
			if !yield(i, v) {
				return
			}
		}
	}
}

// ForEach calls the given function for each element.
func (l *{{.Name}}) ForEach(f func({{.GoType}})) {
	for _, v := range l.items {
		f(v)
	}
}

// ForEachWithIndex calls the given function with each element and its index.
func (l *{{.Name}}) ForEachWithIndex(f func({{.GoType}}, int)) {
	for i, v := range l.items {
		f(v, i)
	}
}

// Select returns a new list containing only elements that satisfy the predicate.
func (l *{{.Name}}) Select(predicate func({{.GoType}}) bool) *{{.Name}} {
	result := New{{.Name}}()
	for _, v := range l.items {
		if predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Reject returns a new list containing only elements that do not satisfy the predicate.
func (l *{{.Name}}) Reject(predicate func({{.GoType}}) bool) *{{.Name}} {
	result := New{{.Name}}()
	for _, v := range l.items {
		if !predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Detect returns the first element that satisfies the predicate, or the zero value and false.
func (l *{{.Name}}) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	for _, v := range l.items {
		if predicate(v) {
			return v, true
		}
	}
	return {{.Zero}}, false
}

// AnySatisfy returns true if any element satisfies the predicate.
func (l *{{.Name}}) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range l.items {
		if predicate(v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (l *{{.Name}}) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range l.items {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (l *{{.Name}}) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range l.items {
		if predicate(v) {
			return false
		}
	}
	return true
}

// Count returns the number of elements that satisfy the predicate.
func (l *{{.Name}}) Count(predicate func({{.GoType}}) bool) int {
	count := 0
	for _, v := range l.items {
		if predicate(v) {
			count++
		}
	}
	return count
}

// InjectInto performs a left fold over the list.
func (l *{{.Name}}) InjectInto(initial {{.GoType}}, f func({{.GoType}}, {{.GoType}}) {{.GoType}}) {{.GoType}} {
	result := initial
	for _, v := range l.items {
		result = f(result, v)
	}
	return result
}

// Sum returns the sum of all elements{{.SumDoc}}
func (l *{{.Name}}) Sum() {{.SumType}} {
	var sum {{.SumType}}
	for _, v := range l.items {
		sum += {{if .IsFloat}}v{{else}}int64(v){{end}}
	}
	return sum
}

// Min returns the minimum element, or the zero value and false if empty.
func (l *{{.Name}}) Min() ({{.GoType}}, bool) {
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
func (l *{{.Name}}) Max() ({{.GoType}}, bool) {
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
func (l *{{.Name}}) Sort() {
	slices.SortFunc(l.items, func(a, b {{.GoType}}) int {
		return {{if .IsFloat}}{{.CmpFn}}(a, b){{else}}cmp.Compare(a, b){{end}}
	})
}

// SortWithComparator sorts the list using the given less function.
func (l *{{.Name}}) SortWithComparator(less func({{.GoType}}, {{.GoType}}) bool) {
	slices.SortFunc(l.items, func(a, b {{.GoType}}) int {
		switch {
		case less(a, b):
			return -1
		case less(b, a):
			return 1
		default:
			return 0
		}
	})
}

// BinarySearch searches for a value in a sorted list. Returns the index and true if found.
func (l *{{.Name}}) BinarySearch(value {{.GoType}}) (int, bool) {
	lo, hi := 0, len(l.items)-1
	for lo <= hi {
		mid := lo + (hi-lo)/2
		if {{if .IsFloat}}{{.BitsFn}}(l.items[mid]) == {{.BitsFn}}(value){{else}}l.items[mid] == value{{end}} {
			return mid, true
		}
		if {{if .IsFloat}}{{.CmpFn}}(l.items[mid], value) < 0{{else}}l.items[mid] < value{{end}} {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return lo, false
}

// Reversed returns a new list with elements in reverse order.
func (l *{{.Name}}) Reversed() *{{.Name}} {
	n := len(l.items)
	result := New{{.Name}}WithCapacity(n)
	for i := n - 1; i >= 0; i-- {
		result.Add(l.items[i])
	}
	return result
}

// Distinct returns a new list with duplicate elements removed (preserving first occurrence order).
func (l *{{.Name}}) Distinct() *{{.Name}} {
{{- if .IsFloat}}
	// Key by bit pattern so NaN dedupes against itself and -0 stays distinct
	// from +0 (a plain map[{{.GoType}}] would never match NaN and would collapse
	// the two zeroes together).
	seen := make(map[{{.BitsType}}]struct{})
	result := New{{.Name}}()
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
	result := New{{.Name}}()
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
func (l *{{.Name}}) ToSlice() []{{.GoType}} {
	result := make([]{{.GoType}}, len(l.items))
	copy(result, l.items)
	return result
}

// With returns the list after adding the value (fluent API).
func (l *{{.Name}}) AddReturning(value {{.GoType}}) *{{.Name}} {
	l.Add(value)
	return l
}

// RemoveReturning removes the first occurrence of value and returns the receiver (mutating, fluent).
func (l *{{.Name}}) RemoveReturning(value {{.GoType}}) *{{.Name}} {
	l.Remove(value)
	return l
}

// String returns a string representation of the list.
func (l *{{.Name}}) String() string {
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
func (l *{{.Name}}) Equals(other *{{.Name}}) bool {
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

// AddAllReturning adds all values and returns the receiver (mutating, fluent).
func (l *{{.Name}}) AddAllReturning(values ...{{.GoType}}) *{{.Name}} {
	l.AddAll(values...)
	return l
}

// RemoveAllReturning removes every occurrence of any of the given values.
// Compacts in place — keeps the existing backing storage and avoids
// the temporary-list allocation the previous implementation made.
func (l *{{.Name}}) RemoveAllReturning(values ...{{.GoType}}) *{{.Name}} {
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
func (l *{{.Name}}) ToImmutable() *Immutable{{.Name}} {
	return Immutable{{.Name}}From(l)
}
`

const immutableArrayListTmpl = genHeader + `package arraylist

import (
	"iter"
)

// Immutable{{.Name}} is an immutable view of a {{.Name}}.
type Immutable{{.Name}} struct {
	delegate *{{.Name}}
}

// NewImmutable{{.Name}} creates an immutable list from the given values.
func NewImmutable{{.Name}}(values ...{{.GoType}}) *Immutable{{.Name}} {
	return &Immutable{{.Name}}{delegate: {{.Name}}Of(values...)}
}

// Immutable{{.Name}}From creates an immutable copy of a mutable list.
func Immutable{{.Name}}From(l *{{.Name}}) *Immutable{{.Name}} {
	copy := {{.Name}}Of(l.ToSlice()...)
	return &Immutable{{.Name}}{delegate: copy}
}

// Get returns the value at the given index. It panics if the index is out of
// bounds, matching the semantics of a native Go slice.
func (l *Immutable{{.Name}}) Get(index int) {{.GoType}} {
	return l.delegate.Get(index)
}

// Len returns the number of elements. Use l.Len() == 0 to test for
// emptiness.
func (l *Immutable{{.Name}}) Len() int { return l.delegate.Len() }

// Contains returns true if the list contains the given value.
func (l *Immutable{{.Name}}) Contains(value {{.GoType}}) bool {
	return l.delegate.Contains(value)
}

// IndexOf returns the index of the first occurrence, or -1.
func (l *Immutable{{.Name}}) IndexOf(value {{.GoType}}) int {
	return l.delegate.IndexOf(value)
}

// All returns an iter.Seq that yields all elements in order.
func (l *Immutable{{.Name}}) All() iter.Seq[{{.GoType}}] {
	return l.delegate.All()
}

// AllWithIndex returns an iter.Seq2 that yields (index, value) pairs.
func (l *Immutable{{.Name}}) AllWithIndex() iter.Seq2[int, {{.GoType}}] {
	return l.delegate.AllWithIndex()
}

// ForEach calls the given function for each element.
func (l *Immutable{{.Name}}) ForEach(f func({{.GoType}})) {
	l.delegate.ForEach(f)
}

// Select returns a new immutable list with elements satisfying the predicate.
func (l *Immutable{{.Name}}) Select(predicate func({{.GoType}}) bool) *Immutable{{.Name}} {
	return &Immutable{{.Name}}{delegate: l.delegate.Select(predicate)}
}

// Reject returns a new immutable list with elements not satisfying the predicate.
func (l *Immutable{{.Name}}) Reject(predicate func({{.GoType}}) bool) *Immutable{{.Name}} {
	return &Immutable{{.Name}}{delegate: l.delegate.Reject(predicate)}
}

// Detect returns the first element satisfying the predicate, or zero and false.
func (l *Immutable{{.Name}}) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	return l.delegate.Detect(predicate)
}

// AnySatisfy returns true if any element satisfies the predicate.
func (l *Immutable{{.Name}}) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	return l.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (l *Immutable{{.Name}}) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	return l.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (l *Immutable{{.Name}}) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	return l.delegate.NoneSatisfy(predicate)
}

// Count returns the number of elements satisfying the predicate.
func (l *Immutable{{.Name}}) Count(predicate func({{.GoType}}) bool) int {
	return l.delegate.Count(predicate)
}

// Reversed returns a new immutable list in reverse order.
func (l *Immutable{{.Name}}) Reversed() *Immutable{{.Name}} {
	return &Immutable{{.Name}}{delegate: l.delegate.Reversed()}
}

// Distinct returns a new immutable list with duplicates removed.
func (l *Immutable{{.Name}}) Distinct() *Immutable{{.Name}} {
	return &Immutable{{.Name}}{delegate: l.delegate.Distinct()}
}

// ToSlice returns a copy of all elements as a slice.
func (l *Immutable{{.Name}}) ToSlice() []{{.GoType}} {
	return l.delegate.ToSlice()
}

// String returns a string representation.
func (l *Immutable{{.Name}}) String() string {
	return l.delegate.String()
}

// Equals returns true if the other immutable list has the same elements in order.
func (l *Immutable{{.Name}}) Equals(other *Immutable{{.Name}}) bool {
	return l.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this list.
func (l *Immutable{{.Name}}) ToMutable() *{{.Name}} {
	return {{.Name}}Of(l.ToSlice()...)
}
`

const synchronizedArrayListTmpl = genHeader + `package arraylist

import (
	"iter"
	"sync"
	"unsafe"
)

// Synchronized{{.Name}} is a thread-safe wrapper around {{.Name}}.
//
// Read methods hold an RLock; writes hold a Lock. Methods that take
// a caller-supplied function (Select, ForEach, InjectInto, …) snapshot
// the backing slice under RLock and release it before invoking the
// callback, so the callback is free to call back into the wrapper
// without deadlocking.
//
// Methods that return a fresh collection (Select, Reject, Distinct,
// Reversed) return an unwrapped *{{.Name}}: the caller owns it and
// is free to add their own synchronisation if they need it.
type Synchronized{{.Name}} struct {
	delegate *{{.Name}}
	mu       sync.RWMutex
}

// NewSynchronized{{.Name}} creates a new thread-safe empty list.
func NewSynchronized{{.Name}}() *Synchronized{{.Name}} {
	return &Synchronized{{.Name}}{delegate: New{{.Name}}()}
}

// NewSynchronized{{.Name}}WithCapacity creates a new thread-safe
// empty list with the given initial capacity.
func NewSynchronized{{.Name}}WithCapacity(capacity int) *Synchronized{{.Name}} {
	return &Synchronized{{.Name}}{delegate: New{{.Name}}WithCapacity(capacity)}
}

// NewSynchronized{{.Name}}From wraps an existing list. The
// wrapper takes ownership of the delegate — callers must not continue
// to mutate it directly without locking.
func NewSynchronized{{.Name}}From(l *{{.Name}}) *Synchronized{{.Name}} {
	return &Synchronized{{.Name}}{delegate: l}
}

// Synchronized{{.Name}}Of creates a new thread-safe list
// containing the given values in order.
func Synchronized{{.Name}}Of(values ...{{.GoType}}) *Synchronized{{.Name}} {
	return &Synchronized{{.Name}}{delegate: {{.Name}}Of(values...)}
}

// snapshot returns a defensive copy of the backing slice taken under
// RLock. Callers iterate the snapshot without holding the lock.
func (l *Synchronized{{.Name}}) snapshot() []{{.GoType}} {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.ToSlice()
}

// ── simple writes ─────────────────────────────────────────────────────

func (l *Synchronized{{.Name}}) Add(value {{.GoType}}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.Add(value)
}

func (l *Synchronized{{.Name}}) AddAll(values ...{{.GoType}}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.AddAll(values...)
}

func (l *Synchronized{{.Name}}) Set(index int, value {{.GoType}}) {{.GoType}} {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.delegate.Set(index, value)
}

func (l *Synchronized{{.Name}}) RemoveAtIndex(index int) {{.GoType}} {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.delegate.RemoveAtIndex(index)
}

func (l *Synchronized{{.Name}}) Remove(value {{.GoType}}) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.delegate.Remove(value)
}

func (l *Synchronized{{.Name}}) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.Clear()
}

// Sort sorts the backing list in place. Holds the write lock for the
// duration; do not call back into this wrapper from a custom comparator.
func (l *Synchronized{{.Name}}) Sort() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.Sort()
}

// SortWithComparator sorts using the given less function, under the
// write lock. The comparator must not call back into this wrapper.
func (l *Synchronized{{.Name}}) SortWithComparator(less func({{.GoType}}, {{.GoType}}) bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.SortWithComparator(less)
}

// ── simple reads ──────────────────────────────────────────────────────

func (l *Synchronized{{.Name}}) Get(index int) {{.GoType}} {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.Get(index)
}

func (l *Synchronized{{.Name}}) Contains(value {{.GoType}}) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.Contains(value)
}

func (l *Synchronized{{.Name}}) IndexOf(value {{.GoType}}) int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.IndexOf(value)
}

func (l *Synchronized{{.Name}}) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.Len()
}

func (l *Synchronized{{.Name}}) ToSlice() []{{.GoType}} {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.ToSlice()
}

func (l *Synchronized{{.Name}}) String() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.String()
}

func (l *Synchronized{{.Name}}) Sum() {{.SumType}} {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.Sum()
}

func (l *Synchronized{{.Name}}) Min() ({{.GoType}}, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.Min()
}

func (l *Synchronized{{.Name}}) Max() ({{.GoType}}, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.Max()
}

// BinarySearch requires the delegate to be sorted. Callers must
// ensure that (e.g. by calling Sort() beforehand, both happening
// before any concurrent Add).
func (l *Synchronized{{.Name}}) BinarySearch(value {{.GoType}}) (int, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.BinarySearch(value)
}

// Equals returns true if the other list has the same elements in the
// same order. Both wrappers are locked under RLock to prevent torn
// reads; locks are acquired in pointer-address order so two goroutines
// calling A.Equals(B) and B.Equals(A) concurrently cannot deadlock.
func (l *Synchronized{{.Name}}) Equals(other *Synchronized{{.Name}}) bool {
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
func (l *Synchronized{{.Name}}) All() iter.Seq[{{.GoType}}] {
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
func (l *Synchronized{{.Name}}) AllWithIndex() iter.Seq2[int, {{.GoType}}] {
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

func (l *Synchronized{{.Name}}) ForEach(f func({{.GoType}})) {
	for _, v := range l.snapshot() {
		f(v)
	}
}

func (l *Synchronized{{.Name}}) ForEachWithIndex(f func({{.GoType}}, int)) {
	for i, v := range l.snapshot() {
		f(v, i)
	}
}

func (l *Synchronized{{.Name}}) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range l.snapshot() {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (l *Synchronized{{.Name}}) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range l.snapshot() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (l *Synchronized{{.Name}}) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range l.snapshot() {
		if predicate(v) {
			return false
		}
	}
	return true
}

func (l *Synchronized{{.Name}}) Count(predicate func({{.GoType}}) bool) int {
	n := 0
	for _, v := range l.snapshot() {
		if predicate(v) {
			n++
		}
	}
	return n
}

func (l *Synchronized{{.Name}}) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	for _, v := range l.snapshot() {
		if predicate(v) {
			return v, true
		}
	}
	var zero {{.GoType}}
	return zero, false
}

func (l *Synchronized{{.Name}}) InjectInto(initial {{.GoType}}, f func({{.GoType}}, {{.GoType}}) {{.GoType}}) {{.GoType}} {
	acc := initial
	for _, v := range l.snapshot() {
		acc = f(acc, v)
	}
	return acc
}

// ── functional that return a new list ─────────────────────────────────

// Select returns a new (unsynchronized) list of elements satisfying the predicate.
func (l *Synchronized{{.Name}}) Select(predicate func({{.GoType}}) bool) *{{.Name}} {
	snapshot := l.snapshot()
	result := New{{.Name}}()
	for _, v := range snapshot {
		if predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Reject returns a new (unsynchronized) list of elements not satisfying the predicate.
func (l *Synchronized{{.Name}}) Reject(predicate func({{.GoType}}) bool) *{{.Name}} {
	snapshot := l.snapshot()
	result := New{{.Name}}()
	for _, v := range snapshot {
		if !predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Distinct returns a new (unsynchronized) list with duplicates removed,
// order preserved.
func (l *Synchronized{{.Name}}) Distinct() *{{.Name}} {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.Distinct()
}

// Reversed returns a new (unsynchronized) list in reverse order.
func (l *Synchronized{{.Name}}) Reversed() *{{.Name}} {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.Reversed()
}

// ── fluent mutators ───────────────────────────────────────────────────
// All return the wrapper so chained calls stay thread-safe.

func (l *Synchronized{{.Name}}) AddReturning(value {{.GoType}}) *Synchronized{{.Name}} {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.AddReturning(value)
	return l
}

func (l *Synchronized{{.Name}}) AddAllReturning(values ...{{.GoType}}) *Synchronized{{.Name}} {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.AddAllReturning(values...)
	return l
}

func (l *Synchronized{{.Name}}) RemoveReturning(value {{.GoType}}) *Synchronized{{.Name}} {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.RemoveReturning(value)
	return l
}

func (l *Synchronized{{.Name}}) RemoveAllReturning(values ...{{.GoType}}) *Synchronized{{.Name}} {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.delegate.RemoveAllReturning(values...)
	return l
}

// ── conversions ───────────────────────────────────────────────────────

// ToImmutable returns an immutable copy of the underlying list taken
// while holding the read lock. The returned value is independent of
// this wrapper.
func (l *Synchronized{{.Name}}) ToImmutable() *Immutable{{.Name}} {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.delegate.ToImmutable()
}
`
