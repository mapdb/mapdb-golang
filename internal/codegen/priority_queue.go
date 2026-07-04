package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

// pqData is the per-primitive view the priorityqueue templates iterate over.
//
// The priorityqueue family is a binary min-heap. It has a BASE variant and a
// SYNCHRONIZED wrapper (no immutable variant). The synchronized wrapper is
// fully type-independent apart from the element type. The base file has three
// type-dependent pieces, all of which are reproduced exactly:
//   - the zero literal used in the Pop/Peek empty-queue error returns ("0" for
//     integer/char elements, "0.0" for floats);
//   - the Contains equality: integer/char elements compare with raw ==, whereas
//     floats compare bit patterns via math.FloatNNbits so that a queued NaN can
//     be found (raw == is always false for NaN);
//   - the less(a, b) ordering: integer/char keys order with raw <, whereas
//     floats order via the shared IEEE total-order helper cmpFloat32/64 (the
//     phase-3 correctness fix) so that NaN sinks to the bottom of the min-heap.
//
// Float base files additionally import "math" for the bit-pattern equality.
type pqData struct {
	Recv      string // receiver variable for shared fragments (always "q")
	Name      string // Int32, Float32, Char (identifier stem)
	GoType    string // int32, float32, uint16 (Go element type)
	SnakeName string // int32, float32, char (file-name stem)
	Zero      string // zero literal for this element type ("0" or "0.0")
	IsFloat   bool
	CmpFn     string // cmpFloat32 / cmpFloat64 (floats only)
	BitsFn    string // math.Float32bits / math.Float64bits (floats only)
}

// genPriorityQueue writes the per-primitive binary min-heap priority queue
// sources (base and synchronized variants) plus the shared cmp_float.go into
// the current working directory. Invoked from priorityqueue/ via go:generate.
func genPriorityQueue() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	base := parse("pq-base", priorityQueueTmpl)
	synchronized := parse("pq-sync", synchronizedPriorityQueueTmpl)

	write := func(name string, tmpl *template.Template, data pqData) error {
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
		data := pqData{
			Recv:      "q",
			Name:      p.Name,
			GoType:    p.GoType,
			SnakeName: p.SnakeName,
			Zero:      "0",
			IsFloat:   p.IsFloating,
		}
		if p.IsFloating {
			data.Zero = "0.0"
			data.CmpFn = "cmpFloat32"
			data.BitsFn = "math.Float32bits"
			if p.ByteSize == 8 {
				data.CmpFn = "cmpFloat64"
				data.BitsFn = "math.Float64bits"
			}
		}

		if err := write(p.SnakeName+"_priority_queue.go", base, data); err != nil {
			return err
		}
		if err := write("synchronized_"+p.SnakeName+"_priority_queue.go", synchronized, data); err != nil {
			return err
		}
	}

	return genCmpFloat("priorityqueue")
}

const priorityQueueTmpl = genHeader + `package priorityqueue

import (
	"fmt"
	"iter"
{{if .IsFloat}}	"math"
{{end}}	"strings"

	"github.com/mapdb/mapdb-golang/internal/segment"
)

// {{.Name}} is a min-heap priority queue of {{.GoType}} values.
// O(log n) Push/Pop, O(1) Peek.
type {{.Name}} struct {
	items []{{.GoType}}
}

// New{{.Name}} creates a new empty {{.Name}}.
func New{{.Name}}() *{{.Name}} {
	return &{{.Name}}{items: make([]{{.GoType}}, 0, 16)}
}

// {{.Name}}Of creates a new {{.Name}} and heapifies the given values in O(n).
func {{.Name}}Of(values ...{{.GoType}}) *{{.Name}} {
	q := &{{.Name}}{items: make([]{{.GoType}}, len(values))}
	copy(q.items, values)
	if len(q.items) > 1 {
		for i := len(q.items)/2 - 1; i >= 0; i-- {
			q.siftDown(i)
		}
	}
	return q
}

// Push adds a value to the heap. O(log n).
func (q *{{.Name}}) Push(value {{.GoType}}) {
	q.items = append(q.items, value)
	q.siftUp(len(q.items) - 1)
}

// Pop removes and returns the smallest element. The bool is false if the
// queue is empty. O(log n).
func (q *{{.Name}}) Pop() ({{.GoType}}, bool) {
	if len(q.items) == 0 {
		return {{.Zero}}, false
	}
	top := q.items[0]
	last := len(q.items) - 1
	q.items[0] = q.items[last]
	q.items = q.items[:last]
	if len(q.items) > 0 {
		q.siftDown(0)
	}
	return top, true
}

// Peek returns the smallest element without removing it. The bool is false
// if the queue is empty.
func (q *{{.Name}}) Peek() ({{.GoType}}, bool) {
	if len(q.items) == 0 {
		return {{.Zero}}, false
	}
	return q.items[0], true
}

// Len returns the number of elements in the queue. Use q.Len() == 0 to test
// for emptiness.
func (q *{{.Name}}) Len() int { return len(q.items) }


// Clear removes all elements.
func (q *{{.Name}}) Clear() { q.items = q.items[:0] }

// Contains returns true if the queue contains the given value. O(n).
{{template "contains_slice" .}}

// ToSlice returns a copy of the internal heap array (NOT sorted).
func (q *{{.Name}}) ToSlice() []{{.GoType}} {
	out := make([]{{.GoType}}, len(q.items))
	copy(out, q.items)
	return out
}

// All returns an iter.Seq over the elements in internal heap-array order — the
// same order as ToSlice and String, NOT priority order (use DrainSorted for an
// ascending, queue-consuming walk). O(n), non-destructive.
func (q *{{.Name}}) All() iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
		for _, v := range q.items {
			if !yield(v) {
				return
			}
		}
	}
}

{{template "segments_slice" .}}

// DrainSorted pops all elements in ascending order, consuming the queue.
func (q *{{.Name}}) DrainSorted() []{{.GoType}} {
	out := make([]{{.GoType}}, 0, len(q.items))
	for len(q.items) > 0 {
		v, _ := q.Pop()
		out = append(out, v)
	}
	return out
}

// String returns a string representation in heap-array order.
func (q *{{.Name}}) String() string {
	if len(q.items) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteString("[")
	for i, v := range q.items {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v", v)
	}
	sb.WriteString("]")
	return sb.String()
}

func (q *{{.Name}}) less(a, b int) bool {
{{if .IsFloat}}	// Bit-tiebreak comparator: NaN compares as greatest, so it sinks to the
	// bottom of the min-heap and drains last instead of first (raw ` + "`<`" + ` returns
	// false for any NaN comparison, corrupting heap order).
	return {{.CmpFn}}(q.items[a], q.items[b]) < 0
{{else}}	return q.items[a] < q.items[b]
{{end}}}

func (q *{{.Name}}) siftUp(start int) {
	i := start
	for i > 0 {
		parent := (i - 1) / 2
		if q.less(i, parent) {
			q.items[i], q.items[parent] = q.items[parent], q.items[i]
			i = parent
		} else {
			break
		}
	}
}

func (q *{{.Name}}) siftDown(start int) {
	i := start
	n := len(q.items)
	for {
		left := 2*i + 1
		if left >= n {
			break
		}
		right := left + 1
		best := left
		if right < n && q.less(right, left) {
			best = right
		}
		if q.less(best, i) {
			q.items[best], q.items[i] = q.items[i], q.items[best]
			i = best
		} else {
			break
		}
	}
}
`

const synchronizedPriorityQueueTmpl = genHeader + `package priorityqueue

import (
	"iter"
	"sync"

	"github.com/mapdb/mapdb-golang/internal/segment"
)

// Synchronized{{.Name}} is a thread-safe wrapper around {{.Name}}.
type Synchronized{{.Name}} struct {
	delegate *{{.Name}}
	mu       sync.RWMutex
}

// NewSynchronized{{.Name}} creates a new thread-safe empty priority queue.
func NewSynchronized{{.Name}}() *Synchronized{{.Name}} {
	return &Synchronized{{.Name}}{delegate: New{{.Name}}()}
}

func (q *Synchronized{{.Name}}) Push(value {{.GoType}}) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.delegate.Push(value)
}

func (q *Synchronized{{.Name}}) Pop() ({{.GoType}}, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.delegate.Pop()
}

func (q *Synchronized{{.Name}}) Peek() ({{.GoType}}, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.delegate.Peek()
}

func (q *Synchronized{{.Name}}) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.delegate.Len()
}

func (q *Synchronized{{.Name}}) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.delegate.Clear()
}

func (q *Synchronized{{.Name}}) Contains(value {{.GoType}}) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.delegate.Contains(value)
}

func (q *Synchronized{{.Name}}) ToSlice() []{{.GoType}} {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.delegate.ToSlice()
}

// All returns an iter.Seq over a point-in-time snapshot in heap-array order (see
// {{.Name}}.All). The snapshot is taken once under RLock; iteration is lock-free.
func (q *Synchronized{{.Name}}) All() iter.Seq[{{.GoType}}] {
	snapshot := q.ToSlice()
	return func(yield func({{.GoType}}) bool) {
		for _, v := range snapshot {
			if !yield(v) {
				return
			}
		}
	}
}

// Segments cuts a heap-array snapshot into up to n balanced, contiguous,
// non-overlapping views covering it exactly once, satisfying par.Segmenter[{{.GoType}}].
// The snapshot (ToSlice) is taken once under RLock; the views iterate it lock-free.
func (q *Synchronized{{.Name}}) Segments(n int) []iter.Seq[{{.GoType}}] {
	return segment.Split(q.ToSlice(), n)
}

func (q *Synchronized{{.Name}}) DrainSorted() []{{.GoType}} {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.delegate.DrainSorted()
}

func (q *Synchronized{{.Name}}) String() string {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.delegate.String()
}
`
