package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

// stData is the per-primitive view the stack templates iterate over.
//
// The stack family is a plain LIFO slice-backed stack: it has no hashing
// and no Sum/Min/Max reductions. Membership/equality (Contains/Equals) on
// float values uses IEEE-754 bit-pattern comparison (so Contains(NaN) works
// and ±0 stay distinct); int/char values use raw ==. The type-dependent
// pieces are the zero literal used in the Pop/Peek/PeekAt error returns and
// Detect, plus the float bit-pattern equality function.
type stData struct {
	Name      string // Int32, Float32, Char (identifier stem)
	GoType    string // int32, float32, uint16 (Go element type)
	SnakeName string // int32, float32, char (file-name stem)
	Zero      string // zero literal for this element type ("0" or "0.0")
	IsFloat   bool
	BitsFn    string // math.Float32bits / math.Float64bits (floats only)
}

// genStack writes the per-primitive array stack sources (base, immutable,
// synchronized variants) into the current working directory. Invoked from
// stack/ via go:generate. There is no shared cmp_float.go and no hashing.
func genStack() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	base := parse("st-base", arrayStackTmpl)
	immutable := parse("st-immutable", immutableArrayStackTmpl)
	synchronized := parse("st-sync", synchronizedArrayStackTmpl)

	write := func(name string, tmpl *template.Template, data stData) error {
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
		data := stData{
			Name:      p.Name,
			GoType:    p.GoType,
			SnakeName: p.SnakeName,
			Zero:      "0",
			IsFloat:   p.IsFloating,
		}
		if p.IsFloating {
			data.Zero = "0.0"
			data.BitsFn = "math.Float32bits"
			if p.ByteSize == 8 {
				data.BitsFn = "math.Float64bits"
			}
		}

		if err := write(p.SnakeName+"_array_stack.go", base, data); err != nil {
			return err
		}
		if err := write("immutable_"+p.SnakeName+"_array_stack.go", immutable, data); err != nil {
			return err
		}
		if err := write("synchronized_"+p.SnakeName+"_array_stack.go", synchronized, data); err != nil {
			return err
		}
	}

	return nil
}

const arrayStackTmpl = genHeader + `package stack

import (
	"fmt"
	"iter"
{{- if .IsFloat}}
	"math"
{{- end}}
	"strings"
)

// {{.Name}} is a LIFO (last-in, first-out) stack backed by a {{.GoType}} slice.
type {{.Name}} struct {
	items []{{.GoType}}
}

// New{{.Name}} creates a new empty {{.Name}}.
func New{{.Name}}() *{{.Name}} {
	return &{{.Name}}{
		items: make([]{{.GoType}}, 0, 16),
	}
}

// New{{.Name}}WithCapacity creates a new empty {{.Name}} that can hold capacity
// values before its backing slice grows. It is the stack's bulk-load convenience:
// presize once, then Push the prepared values in a single O(n) pass with no
// intermediate reallocation. A negative capacity panics.
func New{{.Name}}WithCapacity(capacity int) *{{.Name}} {
	if capacity < 0 {
		panic("mapdb: New{{.Name}}WithCapacity: negative capacity")
	}
	return &{{.Name}}{
		items: make([]{{.GoType}}, 0, capacity),
	}
}

// {{.Name}}Of creates a new {{.Name}} from the given values.
// The last value becomes the top of the stack.
func {{.Name}}Of(values ...{{.GoType}}) *{{.Name}} {
	s := &{{.Name}}{
		items: make([]{{.GoType}}, len(values)),
	}
	copy(s.items, values)
	return s
}

// Push adds a value to the top of the stack.
func (s *{{.Name}}) Push(value {{.GoType}}) {
	s.items = append(s.items, value)
}

// Pop removes and returns the top value. The bool is false if the stack is empty.
func (s *{{.Name}}) Pop() ({{.GoType}}, bool) {
	if len(s.items) == 0 {
		return {{.Zero}}, false
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return top, true
}

// Peek returns the top value without removing it. The bool is false if the stack is empty.
func (s *{{.Name}}) Peek() ({{.GoType}}, bool) {
	if len(s.items) == 0 {
		return {{.Zero}}, false
	}
	return s.items[len(s.items)-1], true
}

// PeekAt returns the element at the given distance from the top (0 = top).
// It panics if the index is out of range, like a native Go slice.
func (s *{{.Name}}) PeekAt(index int) {{.GoType}} {
	if index < 0 || index >= len(s.items) {
		panic(fmt.Sprintf("stack.{{.Name}}: index out of range [%d] with length %d", index, len(s.items)))
	}
	return s.items[len(s.items)-1-index]
}

// Len returns the number of elements in the stack. Use s.Len() == 0 to test
// for emptiness.
func (s *{{.Name}}) Len() int {
	return len(s.items)
}

// Clear removes all elements from the stack.
func (s *{{.Name}}) Clear() {
	s.items = s.items[:0]
}

// Contains returns true if the stack contains the given value.
func (s *{{.Name}}) Contains(value {{.GoType}}) bool {
	for _, v := range s.items {
		if {{if .IsFloat}}{{.BitsFn}}(v) == {{.BitsFn}}(value){{else}}v == value{{end}} {
			return true
		}
	}
	return false
}

// All returns an iter.Seq that yields elements from top to bottom.
func (s *{{.Name}}) All() iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
		for i := len(s.items) - 1; i >= 0; i-- {
			if !yield(s.items[i]) {
				return
			}
		}
	}
}

// ForEach calls the given function for each element from top to bottom.
func (s *{{.Name}}) ForEach(f func({{.GoType}})) {
	for i := len(s.items) - 1; i >= 0; i-- {
		f(s.items[i])
	}
}

// Select returns a new stack containing only elements that satisfy the predicate.
// Order is preserved (top of result corresponds to top of original that passed).
func (s *{{.Name}}) Select(predicate func({{.GoType}}) bool) *{{.Name}} {
	result := New{{.Name}}()
	for _, v := range s.items {
		if predicate(v) {
			result.Push(v)
		}
	}
	return result
}

// Reject returns a new stack containing only elements that do not satisfy the predicate.
func (s *{{.Name}}) Reject(predicate func({{.GoType}}) bool) *{{.Name}} {
	result := New{{.Name}}()
	for _, v := range s.items {
		if !predicate(v) {
			result.Push(v)
		}
	}
	return result
}

// Detect returns the first element from the top that satisfies the predicate, or zero and false.
func (s *{{.Name}}) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	for i := len(s.items) - 1; i >= 0; i-- {
		if predicate(s.items[i]) {
			return s.items[i], true
		}
	}
	return {{.Zero}}, false
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *{{.Name}}) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range s.items {
		if predicate(v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *{{.Name}}) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range s.items {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *{{.Name}}) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range s.items {
		if predicate(v) {
			return false
		}
	}
	return true
}

// Count returns the number of elements that satisfy the predicate.
func (s *{{.Name}}) Count(predicate func({{.GoType}}) bool) int {
	count := 0
	for _, v := range s.items {
		if predicate(v) {
			count++
		}
	}
	return count
}

// InjectInto performs a left fold from bottom to top.
func (s *{{.Name}}) InjectInto(initial {{.GoType}}, f func({{.GoType}}, {{.GoType}}) {{.GoType}}) {{.GoType}} {
	result := initial
	for _, v := range s.items {
		result = f(result, v)
	}
	return result
}

// ToSlice returns all elements as a slice (top element first).
func (s *{{.Name}}) ToSlice() []{{.GoType}} {
	result := make([]{{.GoType}}, len(s.items))
	for i, j := len(s.items)-1, 0; i >= 0; i, j = i-1, j+1 {
		result[j] = s.items[i]
	}
	return result
}

// ToList returns the elements as a slice in stack order (bottom first, for internal use).
func (s *{{.Name}}) toList() []{{.GoType}} {
	result := make([]{{.GoType}}, len(s.items))
	copy(result, s.items)
	return result
}

// AddReturning pushes the value and returns the receiver (mutating, fluent).
func (s *{{.Name}}) AddReturning(value {{.GoType}}) *{{.Name}} {
	s.Push(value)
	return s
}

// AddAllReturning pushes all values and returns the receiver (mutating, fluent).
func (s *{{.Name}}) AddAllReturning(values ...{{.GoType}}) *{{.Name}} {
	for _, v := range values {
		s.Push(v)
	}
	return s
}

// ToImmutable returns an immutable copy of this stack.
func (s *{{.Name}}) ToImmutable() *Immutable{{.Name}} {
	return Immutable{{.Name}}From(s)
}

// String returns a string representation of the stack (top element first).
func (s *{{.Name}}) String() string {
	if len(s.items) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteString("[")
	for i := len(s.items) - 1; i >= 0; i-- {
		if i < len(s.items)-1 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v", s.items[i])
	}
	sb.WriteString("]")
	return sb.String()
}

// Equals returns true if the other stack has the same elements in the same order.
func (s *{{.Name}}) Equals(other *{{.Name}}) bool {
	if len(s.items) != len(other.items) {
		return false
	}
	for i := range s.items {
		if {{if .IsFloat}}{{.BitsFn}}(s.items[i]) != {{.BitsFn}}(other.items[i]){{else}}s.items[i] != other.items[i]{{end}} {
			return false
		}
	}
	return true
}
`

const immutableArrayStackTmpl = genHeader + `package stack

import (
	"iter"
)

// Immutable{{.Name}} is an immutable LIFO stack of {{.GoType}} values.
type Immutable{{.Name}} struct {
	delegate *{{.Name}}
}

// NewImmutable{{.Name}} creates an immutable stack from the given values.
// The last value becomes the top of the stack.
func NewImmutable{{.Name}}(values ...{{.GoType}}) *Immutable{{.Name}} {
	return &Immutable{{.Name}}{delegate: {{.Name}}Of(values...)}
}

// Immutable{{.Name}}From creates an immutable copy of a mutable stack.
func Immutable{{.Name}}From(s *{{.Name}}) *Immutable{{.Name}} {
	copy := &{{.Name}}{items: make([]{{.GoType}}, len(s.items))}
	for i := range s.items {
		copy.items[i] = s.items[i]
	}
	return &Immutable{{.Name}}{delegate: copy}
}

// Peek returns the top value without removing it. The bool is false if the stack is empty.
func (s *Immutable{{.Name}}) Peek() ({{.GoType}}, bool) {
	return s.delegate.Peek()
}

// PeekAt returns the element at the given distance from the top.
// It panics if the index is out of range, like a native Go slice.
func (s *Immutable{{.Name}}) PeekAt(index int) {{.GoType}} {
	return s.delegate.PeekAt(index)
}

// Len returns the number of elements. Use s.Len() == 0 to test for emptiness.
func (s *Immutable{{.Name}}) Len() int { return s.delegate.Len() }

// Contains returns true if the stack contains the given value.
func (s *Immutable{{.Name}}) Contains(value {{.GoType}}) bool {
	return s.delegate.Contains(value)
}

// All returns an iter.Seq that yields elements from top to bottom.
func (s *Immutable{{.Name}}) All() iter.Seq[{{.GoType}}] {
	return s.delegate.All()
}

// ForEach calls the given function for each element from top to bottom.
func (s *Immutable{{.Name}}) ForEach(f func({{.GoType}})) {
	s.delegate.ForEach(f)
}

// Select returns a new immutable stack with elements satisfying the predicate.
func (s *Immutable{{.Name}}) Select(predicate func({{.GoType}}) bool) *Immutable{{.Name}} {
	return &Immutable{{.Name}}{delegate: s.delegate.Select(predicate)}
}

// Reject returns a new immutable stack with elements not satisfying the predicate.
func (s *Immutable{{.Name}}) Reject(predicate func({{.GoType}}) bool) *Immutable{{.Name}} {
	return &Immutable{{.Name}}{delegate: s.delegate.Reject(predicate)}
}

// Detect returns the first element from top satisfying the predicate, or zero and false.
func (s *Immutable{{.Name}}) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	return s.delegate.Detect(predicate)
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

// Count returns the number of elements satisfying the predicate.
func (s *Immutable{{.Name}}) Count(predicate func({{.GoType}}) bool) int {
	return s.delegate.Count(predicate)
}

// ToSlice returns all elements as a slice (top first).
func (s *Immutable{{.Name}}) ToSlice() []{{.GoType}} {
	return s.delegate.ToSlice()
}

// String returns a string representation.
func (s *Immutable{{.Name}}) String() string {
	return s.delegate.String()
}

// Equals returns true if the other immutable stack has the same elements.
func (s *Immutable{{.Name}}) Equals(other *Immutable{{.Name}}) bool {
	return s.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this stack.
func (s *Immutable{{.Name}}) ToMutable() *{{.Name}} {
	copy := &{{.Name}}{items: make([]{{.GoType}}, len(s.delegate.items))}
	for i := range s.delegate.items {
		copy.items[i] = s.delegate.items[i]
	}
	return copy
}

// Push returns a NEW immutable stack with the value pushed on top.
// The original stack is not modified.
func (s *Immutable{{.Name}}) Push(value {{.GoType}}) *Immutable{{.Name}} {
	newItems := make([]{{.GoType}}, len(s.delegate.items)+1)
	copy(newItems, s.delegate.items)
	newItems[len(s.delegate.items)] = value
	return &Immutable{{.Name}}{delegate: &{{.Name}}{items: newItems}}
}

// Pop returns a NEW immutable stack with the top element removed, and the removed value.
// The original stack is not modified. The bool is false if the stack is empty.
func (s *Immutable{{.Name}}) Pop() (*Immutable{{.Name}}, {{.GoType}}, bool) {
	if s.delegate.Len() == 0 {
		return nil, {{.Zero}}, false
	}
	top := s.delegate.items[len(s.delegate.items)-1]
	newItems := make([]{{.GoType}}, len(s.delegate.items)-1)
	copy(newItems, s.delegate.items[:len(s.delegate.items)-1])
	return &Immutable{{.Name}}{delegate: &{{.Name}}{items: newItems}}, top, true
}
`

const synchronizedArrayStackTmpl = genHeader + `package stack

import (
	"iter"
	"sync"
	"unsafe"
)

// Synchronized{{.Name}} is a thread-safe wrapper around {{.Name}}.
//
// Read methods hold an RLock; writes hold a Lock. Callback methods
// (ForEach/Select/…) snapshot under RLock and run the callback
// unlocked so it can safely re-enter the wrapper.
type Synchronized{{.Name}} struct {
	delegate *{{.Name}}
	mu       sync.RWMutex
}

// NewSynchronized{{.Name}} creates a new thread-safe empty stack.
func NewSynchronized{{.Name}}() *Synchronized{{.Name}} {
	return &Synchronized{{.Name}}{delegate: New{{.Name}}()}
}

// NewSynchronized{{.Name}}From wraps an existing stack. The
// wrapper takes ownership — do not mutate the delegate directly.
func NewSynchronized{{.Name}}From(s *{{.Name}}) *Synchronized{{.Name}} {
	return &Synchronized{{.Name}}{delegate: s}
}

// snapshot copies the stack contents under RLock. The returned slice
// is ordered the same way the delegate's ToSlice would order it.
func (s *Synchronized{{.Name}}) snapshot() []{{.GoType}} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToSlice()
}

// ── writes ────────────────────────────────────────────────────────────

func (s *Synchronized{{.Name}}) Push(value {{.GoType}}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.Push(value)
}

func (s *Synchronized{{.Name}}) Pop() ({{.GoType}}, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.delegate.Pop()
}

func (s *Synchronized{{.Name}}) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.Clear()
}

// ── simple reads ──────────────────────────────────────────────────────

func (s *Synchronized{{.Name}}) Peek() ({{.GoType}}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Peek()
}

func (s *Synchronized{{.Name}}) PeekAt(index int) {{.GoType}} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.PeekAt(index)
}

// Len returns the number of elements. Use s.Len() == 0 to test for emptiness.
func (s *Synchronized{{.Name}}) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Len()
}

func (s *Synchronized{{.Name}}) Contains(value {{.GoType}}) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Contains(value)
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

func (s *Synchronized{{.Name}}) Count(predicate func({{.GoType}}) bool) int {
	n := 0
	for _, v := range s.snapshot() {
		if predicate(v) {
			n++
		}
	}
	return n
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

func (s *Synchronized{{.Name}}) InjectInto(initial {{.GoType}}, f func({{.GoType}}, {{.GoType}}) {{.GoType}}) {{.GoType}} {
	acc := initial
	for _, v := range s.snapshot() {
		acc = f(acc, v)
	}
	return acc
}

// ── functional that return a new stack ───────────────────────────────

func (s *Synchronized{{.Name}}) Select(predicate func({{.GoType}}) bool) *{{.Name}} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Select(predicate)
}

func (s *Synchronized{{.Name}}) Reject(predicate func({{.GoType}}) bool) *{{.Name}} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Reject(predicate)
}

// ── fluent mutators ───────────────────────────────────────────────────

func (s *Synchronized{{.Name}}) AddReturning(value {{.GoType}}) *Synchronized{{.Name}} {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.AddReturning(value)
	return s
}

func (s *Synchronized{{.Name}}) AddAllReturning(values ...{{.GoType}}) *Synchronized{{.Name}} {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.AddAllReturning(values...)
	return s
}

// ── conversions & equals ──────────────────────────────────────────────

func (s *Synchronized{{.Name}}) ToImmutable() *Immutable{{.Name}} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToImmutable()
}

// Equals compares by contents. Locks are acquired in pointer-address
// order to prevent A.Equals(B) / B.Equals(A) deadlocks.
func (s *Synchronized{{.Name}}) Equals(other *Synchronized{{.Name}}) bool {
	if s == other {
		s.mu.RLock()
		defer s.mu.RUnlock()
		return s.delegate.Equals(other.delegate)
	}
	first, second := s, other
	if uintptr(unsafe.Pointer(s)) > uintptr(unsafe.Pointer(other)) {
		first, second = other, s
	}
	first.mu.RLock()
	defer first.mu.RUnlock()
	second.mu.RLock()
	defer second.mu.RUnlock()
	return s.delegate.Equals(other.delegate)
}
`
