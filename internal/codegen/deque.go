package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

// dqData is the per-primitive view the deque templates iterate over.
//
// The deque family is a ring-buffer (power-of-two capacity, modulo
// indexing) double-ended queue. It has no immutable variant, no hashing,
// and no shared cmp_float.go. The only type-dependent pieces are the
// GoType, the snake name in file names/identifiers, the zero literal used
// in the Remove/Peek error returns and slot wiping, and the float
// bit-pattern equality used by Contains/Equals (integers use ==).
type dqData struct {
	Name      string // Int32, Float32, Char (identifier stem)
	GoType    string // int32, float32, uint16 (Go element type)
	SnakeName string // int32, float32, char (file-name stem)
	Zero      string // zero literal for this element type ("0" or "0.0")
	BitsFunc  string // "math.Float32bits"/"math.Float64bits" for floats, "" otherwise
}

// EqExpr returns the per-element equality expression used by Contains
// (a == b form). Floats compare by bit pattern; everything else uses ==.
func (d dqData) EqExpr(a, b string) string {
	if d.BitsFunc != "" {
		return fmt.Sprintf("%s(%s) == %s(%s)", d.BitsFunc, a, d.BitsFunc, b)
	}
	return fmt.Sprintf("%s == %s", a, b)
}

// NeExpr returns the per-element inequality expression (a != b form). Floats
// compare by bit pattern; everything else uses != (byte-identical to the
// pre-fix int/char output).
func (d dqData) NeExpr(a, b string) string {
	if d.BitsFunc != "" {
		return fmt.Sprintf("%s(%s) != %s(%s)", d.BitsFunc, a, d.BitsFunc, b)
	}
	return fmt.Sprintf("%s != %s", a, b)
}

// IsFloat reports whether this primitive uses bit-pattern equality (and
// therefore needs the math import).
func (d dqData) IsFloat() bool { return d.BitsFunc != "" }

// genDeque writes the per-primitive ring-buffer deque sources (base and
// synchronized variants) into the current working directory. Invoked from
// deque/ via go:generate. There is no immutable variant, no shared
// cmp_float.go, and no hashing.
func genDeque() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	base := parse("dq-base", arrayDequeTmpl)
	synchronized := parse("dq-sync", synchronizedArrayDequeTmpl)

	write := func(name string, tmpl *template.Template, data dqData) error {
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
		data := dqData{
			Name:      p.Name,
			GoType:    p.GoType,
			SnakeName: p.SnakeName,
			Zero:      "0",
		}
		if p.IsFloating {
			data.Zero = "0.0"
			if p.ByteSize == 4 {
				data.BitsFunc = "math.Float32bits"
			} else {
				data.BitsFunc = "math.Float64bits"
			}
		}

		if err := write(p.SnakeName+"_array_deque.go", base, data); err != nil {
			return err
		}
		if err := write("synchronized_"+p.SnakeName+"_array_deque.go", synchronized, data); err != nil {
			return err
		}
	}

	return nil
}

const arrayDequeTmpl = genHeader + `package deque

import (
	"fmt"
	"iter"
{{- if .IsFloat}}
	"math"
{{- end}}
	"strings"

	"github.com/mapdb/mapdb-golang/internal/segment"
)

// {{.Name}} is a double-ended queue of {{.GoType}} values, backed by a
// power-of-two ring buffer. AddFirst, AddLast, RemoveFirst, RemoveLast,
// PeekFirst and PeekLast are all O(1) amortised.
//
// The public API is intentionally identical to the slice-backed deque
// that preceded it: callers that iterate via ToSlice or ForEach see
// elements in logical front-to-back order regardless of where head
// happens to sit in the underlying buffer.
type {{.Name}} struct {
	items []{{.GoType}} // len == capacity, always a power of two; indexed modulo cap
	head  int // index of the front element (0 when empty)
	size  int // number of logical elements
}

// initialDequeCap is the smallest power-of-two capacity allocated for a
// freshly constructed deque. It matches the previous slice-backed
// implementation's starting capacity so behaviour around the first few
// reallocations stays comparable.
const initial{{.Name}}DequeCap = 16

// ceilPow2 rounds n up to the next power of two, with a floor of
// initial{{.Name}}DequeCap. Used when sizing the buffer to fit a
// caller-supplied slice in {{.Name}}Of.
func ceilPow2{{.Name}}Deque(n int) int {
	cap := initial{{.Name}}DequeCap
	for cap < n {
		cap <<= 1
	}
	return cap
}

// New{{.Name}} creates a new empty {{.Name}}.
func New{{.Name}}() *{{.Name}} {
	return &{{.Name}}{items: make([]{{.GoType}}, initial{{.Name}}DequeCap)}
}

// New{{.Name}}WithCapacity creates a new empty {{.Name}} whose backing ring
// buffer is sized (rounded up to a power of two) to hold capacity elements
// without growing. It is the deque's bulk-load convenience: presize once, then
// push the prepared values in a single O(n) pass. A negative capacity panics.
func New{{.Name}}WithCapacity(capacity int) *{{.Name}} {
	if capacity < 0 {
		panic("mapdb: New{{.Name}}WithCapacity: negative capacity")
	}
	return &{{.Name}}{items: make([]{{.GoType}}, ceilPow2{{.Name}}Deque(capacity))}
}

// {{.Name}}Of creates a new {{.Name}} from the given values in
// front-to-back order.
func {{.Name}}Of(values ...{{.GoType}}) *{{.Name}} {
	d := &{{.Name}}{
		items: make([]{{.GoType}}, ceilPow2{{.Name}}Deque(len(values))),
		size:  len(values),
	}
	copy(d.items, values)
	return d
}

// grow doubles the backing buffer and repacks elements so that head is at 0.
// Called lazily when size would exceed capacity.
func (d *{{.Name}}) grow() {
	newCap := len(d.items) * 2
	if newCap == 0 {
		newCap = initial{{.Name}}DequeCap
	}
	next := make([]{{.GoType}}, newCap)
	// Copy tail segment (head..end), then wrap segment (0..head) so that
	// logical order is preserved and head resets to 0.
	n := copy(next, d.items[d.head:])
	copy(next[n:], d.items[:d.head])
	d.items = next
	d.head = 0
}

// AddFirst prepends a value to the front of the deque. O(1) amortised.
func (d *{{.Name}}) AddFirst(value {{.GoType}}) {
	if d.size == len(d.items) {
		d.grow()
	}
	mask := len(d.items) - 1
	d.head = (d.head - 1) & mask
	d.items[d.head] = value
	d.size++
}

// AddLast appends a value to the back of the deque. O(1) amortised.
func (d *{{.Name}}) AddLast(value {{.GoType}}) {
	if d.size == len(d.items) {
		d.grow()
	}
	mask := len(d.items) - 1
	d.items[(d.head+d.size)&mask] = value
	d.size++
}

// RemoveFirst removes and returns the front element. The bool is false if empty. O(1).
func (d *{{.Name}}) RemoveFirst() ({{.GoType}}, bool) {
	if d.size == 0 {
		return {{.Zero}}, false
	}
	mask := len(d.items) - 1
	v := d.items[d.head]
	d.items[d.head] = {{.Zero}} // let GC reclaim references if {{.GoType}} ever carries them
	d.head = (d.head + 1) & mask
	d.size--
	return v, true
}

// RemoveLast removes and returns the back element. The bool is false if empty. O(1).
func (d *{{.Name}}) RemoveLast() ({{.GoType}}, bool) {
	if d.size == 0 {
		return {{.Zero}}, false
	}
	mask := len(d.items) - 1
	d.size--
	idx := (d.head + d.size) & mask
	v := d.items[idx]
	d.items[idx] = {{.Zero}}
	return v, true
}

// PeekFirst returns the front element without removing it. The bool is false if empty.
func (d *{{.Name}}) PeekFirst() ({{.GoType}}, bool) {
	if d.size == 0 {
		return {{.Zero}}, false
	}
	return d.items[d.head], true
}

// PeekLast returns the back element without removing it. The bool is false if empty.
func (d *{{.Name}}) PeekLast() ({{.GoType}}, bool) {
	if d.size == 0 {
		return {{.Zero}}, false
	}
	mask := len(d.items) - 1
	return d.items[(d.head+d.size-1)&mask], true
}

// Len returns the number of elements in the deque. Use d.Len() == 0 to test
// for emptiness.
func (d *{{.Name}}) Len() int { return d.size }

// Clear removes all elements. The backing buffer is retained.
func (d *{{.Name}}) Clear() {
	// Wipe slots so retained references are released. Cheap for value types.
	mask := len(d.items) - 1
	for i := 0; i < d.size; i++ {
		d.items[(d.head+i)&mask] = {{.Zero}}
	}
	d.head = 0
	d.size = 0
}

// Contains returns true if the deque contains the given value.
func (d *{{.Name}}) Contains(value {{.GoType}}) bool {
	mask := len(d.items) - 1
	for i := 0; i < d.size; i++ {
		v := d.items[(d.head+i)&mask]
		if {{.EqExpr "v" "value"}} {
			return true
		}
	}
	return false
}

// ForEach applies the function to each element from front to back.
func (d *{{.Name}}) ForEach(f func({{.GoType}})) {
	mask := len(d.items) - 1
	for i := 0; i < d.size; i++ {
		f(d.items[(d.head+i)&mask])
	}
}

// at returns the logical i-th element (0 = front), mapping through the ring.
// The caller must ensure 0 <= i < d.size (so len(d.items) > 0 and the mask is
// valid).
func (d *{{.Name}}) at(i int) {{.GoType}} {
	return d.items[(d.head+i)&(len(d.items)-1)]
}

// All returns an iter.Seq over the elements in logical front-to-back order,
// regardless of where head sits in the ring.
func (d *{{.Name}}) All() iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
		for i := 0; i < d.size; i++ {
			if !yield(d.at(i)) {
				return
			}
		}
	}
}

// Segments cuts the deque into up to n balanced, contiguous, non-overlapping
// views over the LOGICAL front-to-back order (k = min(n, Len), or 1 when empty)
// whose concatenation reproduces All in order, so a *{{.Name}} satisfies
// par.Segmenter[{{.GoType}}] and feeds par.From directly. Each view maps its
// logical index range through the ring on the fly — the physical wrap is
// invisible. The views are live over the deque: mutating it while a view is
// consumed is undefined behavior.
func (d *{{.Name}}) Segments(n int) []iter.Seq[{{.GoType}}] {
	return segment.SplitIndex(d.size, n, d.at)
}

// AnySatisfy returns true if any element satisfies the predicate.
func (d *{{.Name}}) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	mask := len(d.items) - 1
	for i := 0; i < d.size; i++ {
		if predicate(d.items[(d.head+i)&mask]) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if every element satisfies the predicate.
func (d *{{.Name}}) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	mask := len(d.items) - 1
	for i := 0; i < d.size; i++ {
		if !predicate(d.items[(d.head+i)&mask]) {
			return false
		}
	}
	return true
}

// ToSlice returns a copy of the elements in front-to-back order.
func (d *{{.Name}}) ToSlice() []{{.GoType}} {
	out := make([]{{.GoType}}, d.size)
	if d.size == 0 {
		return out
	}
	cap := len(d.items)
	tail := cap - d.head
	if d.size <= tail {
		copy(out, d.items[d.head:d.head+d.size])
	} else {
		n := copy(out, d.items[d.head:])
		copy(out[n:], d.items[:d.size-n])
	}
	return out
}

// Equals returns true if the other deque has the same elements in the same order.
func (d *{{.Name}}) Equals(other *{{.Name}}) bool {
	if d.size != other.size {
		return false
	}
	dMask := len(d.items) - 1
	oMask := len(other.items) - 1
	for i := 0; i < d.size; i++ {
		a := d.items[(d.head+i)&dMask]
		b := other.items[(other.head+i)&oMask]
		if !({{.EqExpr "a" "b"}}) {
			return false
		}
	}
	return true
}

// String returns a string representation in front-to-back order.
func (d *{{.Name}}) String() string {
	if d.size == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteString("[")
	mask := len(d.items) - 1
	for i := 0; i < d.size; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v", d.items[(d.head+i)&mask])
	}
	sb.WriteString("]")
	return sb.String()
}
`

const synchronizedArrayDequeTmpl = genHeader + `package deque

import (
	"iter"
{{- if .IsFloat}}
	"math"
{{- end}}
	"sync"

	"github.com/mapdb/mapdb-golang/internal/segment"
)

// Synchronized{{.Name}} is a thread-safe wrapper around {{.Name}}.
type Synchronized{{.Name}} struct {
	delegate *{{.Name}}
	mu       sync.RWMutex
}

// NewSynchronized{{.Name}} creates a new thread-safe empty deque.
func NewSynchronized{{.Name}}() *Synchronized{{.Name}} {
	return &Synchronized{{.Name}}{delegate: New{{.Name}}()}
}

func (d *Synchronized{{.Name}}) AddFirst(value {{.GoType}}) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.delegate.AddFirst(value)
}

func (d *Synchronized{{.Name}}) AddLast(value {{.GoType}}) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.delegate.AddLast(value)
}

func (d *Synchronized{{.Name}}) RemoveFirst() ({{.GoType}}, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.delegate.RemoveFirst()
}

func (d *Synchronized{{.Name}}) RemoveLast() ({{.GoType}}, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.delegate.RemoveLast()
}

func (d *Synchronized{{.Name}}) PeekFirst() ({{.GoType}}, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.delegate.PeekFirst()
}

func (d *Synchronized{{.Name}}) PeekLast() ({{.GoType}}, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.delegate.PeekLast()
}

// Len returns the number of elements. Use d.Len() == 0 to test for emptiness.
func (d *Synchronized{{.Name}}) Len() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.delegate.Len()
}

func (d *Synchronized{{.Name}}) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.delegate.Clear()
}

func (d *Synchronized{{.Name}}) Contains(value {{.GoType}}) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.delegate.Contains(value)
}

func (d *Synchronized{{.Name}}) ForEach(f func({{.GoType}})) {
	d.mu.RLock()
	snapshot := d.delegate.ToSlice()
	d.mu.RUnlock()
	for _, v := range snapshot {
		f(v)
	}
}

func (d *Synchronized{{.Name}}) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	d.mu.RLock()
	snapshot := d.delegate.ToSlice()
	d.mu.RUnlock()
	for _, v := range snapshot {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (d *Synchronized{{.Name}}) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	d.mu.RLock()
	snapshot := d.delegate.ToSlice()
	d.mu.RUnlock()
	for _, v := range snapshot {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (d *Synchronized{{.Name}}) ToSlice() []{{.GoType}} {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.delegate.ToSlice()
}

// All returns an iter.Seq over a logical-order snapshot. The snapshot is taken
// once under RLock; iteration is lock-free.
func (d *Synchronized{{.Name}}) All() iter.Seq[{{.GoType}}] {
	snapshot := d.ToSlice()
	return func(yield func({{.GoType}}) bool) {
		for _, v := range snapshot {
			if !yield(v) {
				return
			}
		}
	}
}

// Segments cuts a logical-order snapshot into up to n balanced, contiguous,
// non-overlapping views covering it exactly once, satisfying par.Segmenter[{{.GoType}}].
// The snapshot (ToSlice, already unwrapped to logical order) is taken once under
// RLock; the views iterate it lock-free.
func (d *Synchronized{{.Name}}) Segments(n int) []iter.Seq[{{.GoType}}] {
	return segment.Split(d.ToSlice(), n)
}

func (d *Synchronized{{.Name}}) Equals(other *Synchronized{{.Name}}) bool {
	d.mu.RLock()
	thisSlice := d.delegate.ToSlice()
	d.mu.RUnlock()
	other.mu.RLock()
	otherSlice := other.delegate.ToSlice()
	other.mu.RUnlock()
	if len(thisSlice) != len(otherSlice) {
		return false
	}
	for i, v := range thisSlice {
		if {{.NeExpr "otherSlice[i]" "v"}} {
			return false
		}
	}
	return true
}

func (d *Synchronized{{.Name}}) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.delegate.String()
}
`
