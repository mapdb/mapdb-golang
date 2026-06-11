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
// The stack family is a plain LIFO slice-backed stack: it has no hashing,
// no float bit-pattern equality (Contains/Equals use ==), and no
// Sum/Min/Max reductions. The only type-dependent piece is the zero
// literal used in the Pop/Peek/PeekAt error returns and Detect.
type stData struct {
	Name      string // Int32, Float32, Char (identifier stem)
	GoType    string // int32, float32, uint16 (Go element type)
	SnakeName string // int32, float32, char (file-name stem)
	Zero      string // zero literal for this element type ("0" or "0.0")
}

// genStack writes the per-primitive array stack sources (base, immutable,
// synchronized variants) into the current working directory. Invoked from
// stack/ via go:generate. There is no shared cmp_float.go and no hashing.
func genStack() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	base := template.Must(template.New("st-base").Parse(arrayStackTmpl))
	immutable := template.Must(template.New("st-immutable").Parse(immutableArrayStackTmpl))
	synchronized := template.Must(template.New("st-sync").Parse(synchronizedArrayStackTmpl))

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
		}
		if p.IsFloating {
			data.Zero = "0.0"
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
	"strings"
)

// {{.Name}}ArrayStack is a LIFO (last-in, first-out) stack backed by a {{.GoType}} slice.
type {{.Name}}ArrayStack struct {
	items []{{.GoType}}
}

// New{{.Name}}ArrayStack creates a new empty {{.Name}}ArrayStack.
func New{{.Name}}ArrayStack() *{{.Name}}ArrayStack {
	return &{{.Name}}ArrayStack{
		items: make([]{{.GoType}}, 0, 16),
	}
}

// {{.Name}}ArrayStackOf creates a new {{.Name}}ArrayStack from the given values.
// The last value becomes the top of the stack.
func {{.Name}}ArrayStackOf(values ...{{.GoType}}) *{{.Name}}ArrayStack {
	s := &{{.Name}}ArrayStack{
		items: make([]{{.GoType}}, len(values)),
	}
	copy(s.items, values)
	return s
}

// Push adds a value to the top of the stack.
func (s *{{.Name}}ArrayStack) Push(value {{.GoType}}) {
	s.items = append(s.items, value)
}

// Pop removes and returns the top value, or an error if the stack is empty.
func (s *{{.Name}}ArrayStack) Pop() ({{.GoType}}, error) {
	if len(s.items) == 0 {
		return {{.Zero}}, fmt.Errorf("{{.Name}}ArrayStack: Pop on empty stack")
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return top, nil
}

// Peek returns the top value without removing it, or an error if the stack is empty.
func (s *{{.Name}}ArrayStack) Peek() ({{.GoType}}, error) {
	if len(s.items) == 0 {
		return {{.Zero}}, fmt.Errorf("{{.Name}}ArrayStack: Peek on empty stack")
	}
	return s.items[len(s.items)-1], nil
}

// PeekAt returns the element at the given distance from the top (0 = top),
// or an error if the index is out of bounds.
func (s *{{.Name}}ArrayStack) PeekAt(index int) ({{.GoType}}, error) {
	if index < 0 || index >= len(s.items) {
		return {{.Zero}}, fmt.Errorf("{{.Name}}ArrayStack: PeekAt index out of bounds: %d (size %d)", index, len(s.items))
	}
	return s.items[len(s.items)-1-index], nil
}

// Size returns the number of elements in the stack.
func (s *{{.Name}}ArrayStack) Size() int {
	return len(s.items)
}

// Len returns the number of elements. It is an alias for Size, matching
// Go convention (sort.Interface, container/list, bytes.Buffer).
func (s *{{.Name}}ArrayStack) Len() int { return s.Size() }

// IsEmpty returns true if the stack contains no elements.
func (s *{{.Name}}ArrayStack) IsEmpty() bool {
	return len(s.items) == 0
}

// Clear removes all elements from the stack.
func (s *{{.Name}}ArrayStack) Clear() {
	s.items = s.items[:0]
}

// Contains returns true if the stack contains the given value.
func (s *{{.Name}}ArrayStack) Contains(value {{.GoType}}) bool {
	for _, v := range s.items {
		if v == value {
			return true
		}
	}
	return false
}

// All returns an iter.Seq that yields elements from top to bottom.
func (s *{{.Name}}ArrayStack) All() iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
		for i := len(s.items) - 1; i >= 0; i-- {
			if !yield(s.items[i]) {
				return
			}
		}
	}
}

// ForEach calls the given function for each element from top to bottom.
func (s *{{.Name}}ArrayStack) ForEach(f func({{.GoType}})) {
	for i := len(s.items) - 1; i >= 0; i-- {
		f(s.items[i])
	}
}

// Select returns a new stack containing only elements that satisfy the predicate.
// Order is preserved (top of result corresponds to top of original that passed).
func (s *{{.Name}}ArrayStack) Select(predicate func({{.GoType}}) bool) *{{.Name}}ArrayStack {
	result := New{{.Name}}ArrayStack()
	for _, v := range s.items {
		if predicate(v) {
			result.Push(v)
		}
	}
	return result
}

// Reject returns a new stack containing only elements that do not satisfy the predicate.
func (s *{{.Name}}ArrayStack) Reject(predicate func({{.GoType}}) bool) *{{.Name}}ArrayStack {
	result := New{{.Name}}ArrayStack()
	for _, v := range s.items {
		if !predicate(v) {
			result.Push(v)
		}
	}
	return result
}

// Detect returns the first element from the top that satisfies the predicate, or zero and false.
func (s *{{.Name}}ArrayStack) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	for i := len(s.items) - 1; i >= 0; i-- {
		if predicate(s.items[i]) {
			return s.items[i], true
		}
	}
	return {{.Zero}}, false
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *{{.Name}}ArrayStack) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range s.items {
		if predicate(v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *{{.Name}}ArrayStack) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range s.items {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *{{.Name}}ArrayStack) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range s.items {
		if predicate(v) {
			return false
		}
	}
	return true
}

// Count returns the number of elements that satisfy the predicate.
func (s *{{.Name}}ArrayStack) Count(predicate func({{.GoType}}) bool) int {
	count := 0
	for _, v := range s.items {
		if predicate(v) {
			count++
		}
	}
	return count
}

// InjectInto performs a left fold from bottom to top.
func (s *{{.Name}}ArrayStack) InjectInto(initial {{.GoType}}, f func({{.GoType}}, {{.GoType}}) {{.GoType}}) {{.GoType}} {
	result := initial
	for _, v := range s.items {
		result = f(result, v)
	}
	return result
}

// ToSlice returns all elements as a slice (top element first).
func (s *{{.Name}}ArrayStack) ToSlice() []{{.GoType}} {
	result := make([]{{.GoType}}, len(s.items))
	for i, j := len(s.items)-1, 0; i >= 0; i, j = i-1, j+1 {
		result[j] = s.items[i]
	}
	return result
}

// ToList returns the elements as a slice in stack order (bottom first, for internal use).
func (s *{{.Name}}ArrayStack) toList() []{{.GoType}} {
	result := make([]{{.GoType}}, len(s.items))
	copy(result, s.items)
	return result
}

// With returns the stack after pushing the value (fluent API).
func (s *{{.Name}}ArrayStack) With(value {{.GoType}}) *{{.Name}}ArrayStack {
	s.Push(value)
	return s
}

// WithAll returns the stack after pushing all values (fluent API).
func (s *{{.Name}}ArrayStack) WithAll(values ...{{.GoType}}) *{{.Name}}ArrayStack {
	for _, v := range values {
		s.Push(v)
	}
	return s
}

// ToImmutable returns an immutable copy of this stack.
func (s *{{.Name}}ArrayStack) ToImmutable() *Immutable{{.Name}}ArrayStack {
	return Immutable{{.Name}}ArrayStackFrom(s)
}

// String returns a string representation of the stack (top element first).
func (s *{{.Name}}ArrayStack) String() string {
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
func (s *{{.Name}}ArrayStack) Equals(other *{{.Name}}ArrayStack) bool {
	if len(s.items) != len(other.items) {
		return false
	}
	for i := range s.items {
		if s.items[i] != other.items[i] {
			return false
		}
	}
	return true
}
`

const immutableArrayStackTmpl = genHeader + `package stack

import (
	"fmt"
	"iter"
)

// Immutable{{.Name}}ArrayStack is an immutable LIFO stack of {{.GoType}} values.
type Immutable{{.Name}}ArrayStack struct {
	delegate *{{.Name}}ArrayStack
}

// NewImmutable{{.Name}}ArrayStack creates an immutable stack from the given values.
// The last value becomes the top of the stack.
func NewImmutable{{.Name}}ArrayStack(values ...{{.GoType}}) *Immutable{{.Name}}ArrayStack {
	return &Immutable{{.Name}}ArrayStack{delegate: {{.Name}}ArrayStackOf(values...)}
}

// Immutable{{.Name}}ArrayStackFrom creates an immutable copy of a mutable stack.
func Immutable{{.Name}}ArrayStackFrom(s *{{.Name}}ArrayStack) *Immutable{{.Name}}ArrayStack {
	copy := &{{.Name}}ArrayStack{items: make([]{{.GoType}}, len(s.items))}
	for i := range s.items {
		copy.items[i] = s.items[i]
	}
	return &Immutable{{.Name}}ArrayStack{delegate: copy}
}

// Peek returns the top value without removing it, or an error if the stack is empty.
func (s *Immutable{{.Name}}ArrayStack) Peek() ({{.GoType}}, error) {
	return s.delegate.Peek()
}

// PeekAt returns the element at the given distance from the top,
// or an error if the index is out of bounds.
func (s *Immutable{{.Name}}ArrayStack) PeekAt(index int) ({{.GoType}}, error) {
	return s.delegate.PeekAt(index)
}

// Size returns the number of elements.
func (s *Immutable{{.Name}}ArrayStack) Size() int {
	return s.delegate.Size()
}

// Len returns the number of elements. It is an alias for Size, matching
// Go convention (sort.Interface, container/list, bytes.Buffer).
func (s *Immutable{{.Name}}ArrayStack) Len() int { return s.Size() }

// IsEmpty returns true if the stack contains no elements.
func (s *Immutable{{.Name}}ArrayStack) IsEmpty() bool {
	return s.delegate.IsEmpty()
}

// Contains returns true if the stack contains the given value.
func (s *Immutable{{.Name}}ArrayStack) Contains(value {{.GoType}}) bool {
	return s.delegate.Contains(value)
}

// All returns an iter.Seq that yields elements from top to bottom.
func (s *Immutable{{.Name}}ArrayStack) All() iter.Seq[{{.GoType}}] {
	return s.delegate.All()
}

// ForEach calls the given function for each element from top to bottom.
func (s *Immutable{{.Name}}ArrayStack) ForEach(f func({{.GoType}})) {
	s.delegate.ForEach(f)
}

// Select returns a new immutable stack with elements satisfying the predicate.
func (s *Immutable{{.Name}}ArrayStack) Select(predicate func({{.GoType}}) bool) *Immutable{{.Name}}ArrayStack {
	return &Immutable{{.Name}}ArrayStack{delegate: s.delegate.Select(predicate)}
}

// Reject returns a new immutable stack with elements not satisfying the predicate.
func (s *Immutable{{.Name}}ArrayStack) Reject(predicate func({{.GoType}}) bool) *Immutable{{.Name}}ArrayStack {
	return &Immutable{{.Name}}ArrayStack{delegate: s.delegate.Reject(predicate)}
}

// Detect returns the first element from top satisfying the predicate, or zero and false.
func (s *Immutable{{.Name}}ArrayStack) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	return s.delegate.Detect(predicate)
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *Immutable{{.Name}}ArrayStack) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	return s.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *Immutable{{.Name}}ArrayStack) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	return s.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *Immutable{{.Name}}ArrayStack) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	return s.delegate.NoneSatisfy(predicate)
}

// Count returns the number of elements satisfying the predicate.
func (s *Immutable{{.Name}}ArrayStack) Count(predicate func({{.GoType}}) bool) int {
	return s.delegate.Count(predicate)
}

// ToSlice returns all elements as a slice (top first).
func (s *Immutable{{.Name}}ArrayStack) ToSlice() []{{.GoType}} {
	return s.delegate.ToSlice()
}

// String returns a string representation.
func (s *Immutable{{.Name}}ArrayStack) String() string {
	return s.delegate.String()
}

// Equals returns true if the other immutable stack has the same elements.
func (s *Immutable{{.Name}}ArrayStack) Equals(other *Immutable{{.Name}}ArrayStack) bool {
	return s.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this stack.
func (s *Immutable{{.Name}}ArrayStack) ToMutable() *{{.Name}}ArrayStack {
	copy := &{{.Name}}ArrayStack{items: make([]{{.GoType}}, len(s.delegate.items))}
	for i := range s.delegate.items {
		copy.items[i] = s.delegate.items[i]
	}
	return copy
}

// Push returns a NEW immutable stack with the value pushed on top.
// The original stack is not modified.
func (s *Immutable{{.Name}}ArrayStack) Push(value {{.GoType}}) *Immutable{{.Name}}ArrayStack {
	newItems := make([]{{.GoType}}, len(s.delegate.items)+1)
	copy(newItems, s.delegate.items)
	newItems[len(s.delegate.items)] = value
	return &Immutable{{.Name}}ArrayStack{delegate: &{{.Name}}ArrayStack{items: newItems}}
}

// Pop returns a NEW immutable stack with the top element removed, and the removed value.
// The original stack is not modified. Returns an error if the stack is empty.
func (s *Immutable{{.Name}}ArrayStack) Pop() (*Immutable{{.Name}}ArrayStack, {{.GoType}}, error) {
	if s.delegate.IsEmpty() {
		return nil, {{.Zero}}, fmt.Errorf("Immutable{{.Name}}ArrayStack: Pop on empty stack")
	}
	top := s.delegate.items[len(s.delegate.items)-1]
	newItems := make([]{{.GoType}}, len(s.delegate.items)-1)
	copy(newItems, s.delegate.items[:len(s.delegate.items)-1])
	return &Immutable{{.Name}}ArrayStack{delegate: &{{.Name}}ArrayStack{items: newItems}}, top, nil
}
`

const synchronizedArrayStackTmpl = genHeader + `package stack

import (
	"iter"
	"sync"
	"unsafe"
)

// Synchronized{{.Name}}ArrayStack is a thread-safe wrapper around {{.Name}}ArrayStack.
//
// Read methods hold an RLock; writes hold a Lock. Callback methods
// (ForEach/Select/…) snapshot under RLock and run the callback
// unlocked so it can safely re-enter the wrapper.
type Synchronized{{.Name}}ArrayStack struct {
	delegate *{{.Name}}ArrayStack
	mu       sync.RWMutex
}

// NewSynchronized{{.Name}}ArrayStack creates a new thread-safe empty stack.
func NewSynchronized{{.Name}}ArrayStack() *Synchronized{{.Name}}ArrayStack {
	return &Synchronized{{.Name}}ArrayStack{delegate: New{{.Name}}ArrayStack()}
}

// NewSynchronized{{.Name}}ArrayStackFrom wraps an existing stack. The
// wrapper takes ownership — do not mutate the delegate directly.
func NewSynchronized{{.Name}}ArrayStackFrom(s *{{.Name}}ArrayStack) *Synchronized{{.Name}}ArrayStack {
	return &Synchronized{{.Name}}ArrayStack{delegate: s}
}

// snapshot copies the stack contents under RLock. The returned slice
// is ordered the same way the delegate's ToSlice would order it.
func (s *Synchronized{{.Name}}ArrayStack) snapshot() []{{.GoType}} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToSlice()
}

// ── writes ────────────────────────────────────────────────────────────

func (s *Synchronized{{.Name}}ArrayStack) Push(value {{.GoType}}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.Push(value)
}

func (s *Synchronized{{.Name}}ArrayStack) Pop() ({{.GoType}}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.delegate.Pop()
}

func (s *Synchronized{{.Name}}ArrayStack) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.Clear()
}

// ── simple reads ──────────────────────────────────────────────────────

func (s *Synchronized{{.Name}}ArrayStack) Peek() ({{.GoType}}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Peek()
}

func (s *Synchronized{{.Name}}ArrayStack) PeekAt(index int) ({{.GoType}}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.PeekAt(index)
}

func (s *Synchronized{{.Name}}ArrayStack) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Size()
}

// Len returns the number of elements. It is an alias for Size, matching
// Go convention (sort.Interface, container/list, bytes.Buffer).
func (s *Synchronized{{.Name}}ArrayStack) Len() int { return s.Size() }

func (s *Synchronized{{.Name}}ArrayStack) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.IsEmpty()
}

func (s *Synchronized{{.Name}}ArrayStack) Contains(value {{.GoType}}) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Contains(value)
}

func (s *Synchronized{{.Name}}ArrayStack) ToSlice() []{{.GoType}} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToSlice()
}

func (s *Synchronized{{.Name}}ArrayStack) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.String()
}

// ── iteration ────────────────────────────────────────────────────────

func (s *Synchronized{{.Name}}ArrayStack) All() iter.Seq[{{.GoType}}] {
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

func (s *Synchronized{{.Name}}ArrayStack) ForEach(f func({{.GoType}})) {
	for _, v := range s.snapshot() {
		f(v)
	}
}

func (s *Synchronized{{.Name}}ArrayStack) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (s *Synchronized{{.Name}}ArrayStack) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range s.snapshot() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (s *Synchronized{{.Name}}ArrayStack) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return false
		}
	}
	return true
}

func (s *Synchronized{{.Name}}ArrayStack) Count(predicate func({{.GoType}}) bool) int {
	n := 0
	for _, v := range s.snapshot() {
		if predicate(v) {
			n++
		}
	}
	return n
}

func (s *Synchronized{{.Name}}ArrayStack) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return v, true
		}
	}
	var zero {{.GoType}}
	return zero, false
}

func (s *Synchronized{{.Name}}ArrayStack) InjectInto(initial {{.GoType}}, f func({{.GoType}}, {{.GoType}}) {{.GoType}}) {{.GoType}} {
	acc := initial
	for _, v := range s.snapshot() {
		acc = f(acc, v)
	}
	return acc
}

// ── functional that return a new stack ───────────────────────────────

func (s *Synchronized{{.Name}}ArrayStack) Select(predicate func({{.GoType}}) bool) *{{.Name}}ArrayStack {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Select(predicate)
}

func (s *Synchronized{{.Name}}ArrayStack) Reject(predicate func({{.GoType}}) bool) *{{.Name}}ArrayStack {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Reject(predicate)
}

// ── fluent mutators ───────────────────────────────────────────────────

func (s *Synchronized{{.Name}}ArrayStack) With(value {{.GoType}}) *Synchronized{{.Name}}ArrayStack {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.With(value)
	return s
}

func (s *Synchronized{{.Name}}ArrayStack) WithAll(values ...{{.GoType}}) *Synchronized{{.Name}}ArrayStack {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.WithAll(values...)
	return s
}

// ── conversions & equals ──────────────────────────────────────────────

func (s *Synchronized{{.Name}}ArrayStack) ToImmutable() *Immutable{{.Name}}ArrayStack {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToImmutable()
}

// Equals compares by contents. Locks are acquired in pointer-address
// order to prevent A.Equals(B) / B.Equals(A) deadlocks.
func (s *Synchronized{{.Name}}ArrayStack) Equals(other *Synchronized{{.Name}}ArrayStack) bool {
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
