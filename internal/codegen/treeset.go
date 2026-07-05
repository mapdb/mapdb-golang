package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

// tsData is the per-primitive view the treeset template iterates over.
//
// The treeset family is a red-black tree set. It has a BASE variant only
// (no immutable, no synchronized wrappers). The only type-dependent pieces
// are the zero literal used in the Min/Max/Floor/Ceiling "empty" returns and
// the ordering comparison: integer/char keys order with raw < / >, whereas
// float keys order via the shared IEEE total-order helper cmpFloat32/64 (the
// phase-3 correctness fix). For float keys the Floor/Ceiling exact-match
// short-circuits and the RangeValues [from, to) bounds also route through
// cmpFloat32/64 so navigation is a true total order (NaN sorts to the top,
// ±0 are distinguished); int/char keys keep raw == / < / >=.
type tsData struct {
	Name      string // Int32, Float32, Char (identifier stem)
	GoType    string // int32, float32, uint16 (Go element type)
	SnakeName string // int32, float32, char (file-name stem)
	Zero      string // zero literal for this element type ("0" or "0.0")
	IsFloat   bool
	CmpFn     string // cmpFloat32 / cmpFloat64 (floats only)
}

// genTreeSet writes the per-primitive red-black tree set sources (base variant
// only) plus the shared cmp_float.go into the current working directory.
// Invoked from treeset/ via go:generate.
func genTreeSet() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	base := parse("ts-base", treeSetTmpl)

	write := func(name string, tmpl *template.Template, data tsData) error {
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
		data := tsData{
			Name:      p.Name,
			GoType:    p.GoType,
			SnakeName: p.SnakeName,
			Zero:      "0",
			IsFloat:   p.IsFloating,
		}
		if p.IsFloating {
			data.Zero = "0.0"
			data.CmpFn = "cmpFloat32"
			if p.ByteSize == 8 {
				data.CmpFn = "cmpFloat64"
			}
		}

		if err := write(p.SnakeName+"_tree_set.go", base, data); err != nil {
			return err
		}
	}

	if err := genCmpFloat("treeset"); err != nil {
		return err
	}

	// Stamp the conformance laws (todo 14 §4). All() and ToSlice() both yield
	// ascending sorted order, so law 1 is order-sensitive.
	return genConformanceForPrimitives("treeset", true)
}

const treeSetTmpl = genHeader + `package treeset

import (
	"fmt"
	"iter"
	"strings"

	"github.com/mapdb/mapdb-golang/internal/segment"
	"github.com/mapdb/mapdb-golang/pump"
)

const (
	{{.SnakeName}}NodeRed   = false
	{{.SnakeName}}NodeBlack = true
)

type {{.SnakeName}}Node struct {
	key    {{.GoType}}
	left   *{{.SnakeName}}Node
	right  *{{.SnakeName}}Node
	parent *{{.SnakeName}}Node
	color  bool
	// size is the number of nodes in the subtree rooted here (this node plus
	// both children's subtrees). Maintained in O(1) on every structural change
	// -- insert, remove, and all rotations -- so order-statistic Rank/Select run
	// in O(log n). Invariant after any operation: size == 1 + size(left) + size(right).
	size int
}

// {{.SnakeName}}NodeSize returns the subtree size of a node link (0 if nil).
func {{.SnakeName}}NodeSize(n *{{.SnakeName}}Node) int {
	if n == nil {
		return 0
	}
	return n.size
}

// {{.SnakeName}}NodeFixSize recomputes a node's cached subtree size from its
// children. Called after any rotation or child relink so the augmentation stays
// consistent.
func {{.SnakeName}}NodeFixSize(n *{{.SnakeName}}Node) {
	n.size = 1 + {{.SnakeName}}NodeSize(n.left) + {{.SnakeName}}NodeSize(n.right)
}

// {{.Name}} is a sorted set of {{.GoType}} values, backed by a red-black tree.
type {{.Name}} struct {
	root *{{.SnakeName}}Node
	size int
}

// New{{.Name}} creates a new empty sorted set.
func New{{.Name}}() *{{.Name}} {
	return &{{.Name}}{}
}

// {{.Name}}Of creates a new sorted set from the given values.
func {{.Name}}Of(values ...{{.GoType}}) *{{.Name}} {
	s := New{{.Name}}()
	for _, v := range values {
		s.Add(v)
	}
	return s
}

// Add inserts a value. Returns true if added (not already present).
func (s *{{.Name}}) Add(value {{.GoType}}) bool {
	if s.root == nil {
		s.root = &{{.SnakeName}}Node{key: value, color: {{.SnakeName}}NodeBlack, size: 1}
		s.size++
		return true
	}
	node := s.root
	for {
		if {{if .IsFloat}}{{.CmpFn}}(value, node.key) < 0{{else}}value < node.key{{end}} {
			if node.left == nil {
				node.left = &{{.SnakeName}}Node{key: value, parent: node, color: {{.SnakeName}}NodeRed, size: 1}
				s.incSizeToRoot(node)
				s.fixAfterInsert(node.left)
				s.size++
				return true
			}
			node = node.left
		} else if {{if .IsFloat}}{{.CmpFn}}(value, node.key) > 0{{else}}value > node.key{{end}} {
			if node.right == nil {
				node.right = &{{.SnakeName}}Node{key: value, parent: node, color: {{.SnakeName}}NodeRed, size: 1}
				s.incSizeToRoot(node)
				s.fixAfterInsert(node.right)
				s.size++
				return true
			}
			node = node.right
		} else {
			return false // already exists
		}
	}
}

// Remove removes a value. Returns true if found and removed.
func (s *{{.Name}}) Remove(value {{.GoType}}) bool {
	node := s.findNode(value)
	if node == nil {
		return false
	}
	s.deleteNode(node)
	s.size--
	return true
}

// Contains returns true if the set contains the value.
func (s *{{.Name}}) Contains(value {{.GoType}}) bool {
	return s.findNode(value) != nil
}

// Len returns the number of elements. Use s.Len() == 0 to test for emptiness.
func (s *{{.Name}}) Len() int { return s.size }

// Clear removes all elements.
func (s *{{.Name}}) Clear() { s.root = nil; s.size = 0 }

// Min returns the smallest element, or zero and false if empty.
func (s *{{.Name}}) Min() ({{.GoType}}, bool) {
	if s.root == nil {
		return {{.Zero}}, false
	}
	return s.minNode(s.root).key, true
}

// Max returns the largest element, or zero and false if empty.
func (s *{{.Name}}) Max() ({{.GoType}}, bool) {
	if s.root == nil {
		return {{.Zero}}, false
	}
	return s.maxNode(s.root).key, true
}

// First is an alias of Min — the smallest element, or zero and false if empty.
func (s *{{.Name}}) First() ({{.GoType}}, bool) { return s.Min() }

// Last is an alias of Max — the largest element, or zero and false if empty.
func (s *{{.Name}}) Last() ({{.GoType}}, bool) { return s.Max() }

// PollFirst removes and returns the smallest element, or zero and false if empty.
// Does not trap on an empty set.
func (s *{{.Name}}) PollFirst() ({{.GoType}}, bool) {
	v, ok := s.Min()
	if !ok {
		return {{.Zero}}, false
	}
	s.Remove(v)
	return v, true
}

// PollLast removes and returns the largest element, or zero and false if empty.
// Does not trap on an empty set.
func (s *{{.Name}}) PollLast() ({{.GoType}}, bool) {
	v, ok := s.Max()
	if !ok {
		return {{.Zero}}, false
	}
	s.Remove(v)
	return v, true
}

// Higher returns the smallest element strictly > value, or zero and false.
// Unlike Ceiling, never returns value itself.
func (s *{{.Name}}) Higher(value {{.GoType}}) ({{.GoType}}, bool) {
	var result *{{.SnakeName}}Node
	node := s.root
	for node != nil {
		if {{if .IsFloat}}{{.CmpFn}}(value, node.key) < 0{{else}}value < node.key{{end}} {
			result = node
			node = node.left
		} else {
			node = node.right
		}
	}
	if result == nil {
		return {{.Zero}}, false
	}
	return result.key, true
}

// Lower returns the largest element strictly < value, or zero and false.
// Unlike Floor, never returns value itself.
func (s *{{.Name}}) Lower(value {{.GoType}}) ({{.GoType}}, bool) {
	var result *{{.SnakeName}}Node
	node := s.root
	for node != nil {
		if {{if .IsFloat}}{{.CmpFn}}(value, node.key) > 0{{else}}value > node.key{{end}} {
			result = node
			node = node.right
		} else {
			node = node.left
		}
	}
	if result == nil {
		return {{.Zero}}, false
	}
	return result.key, true
}

// DescendingElements returns an iter.Seq that yields elements in descending order.
func (s *{{.Name}}) DescendingElements() iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
		var reverse func(node *{{.SnakeName}}Node) bool
		reverse = func(node *{{.SnakeName}}Node) bool {
			if node == nil {
				return true
			}
			if !reverse(node.right) {
				return false
			}
			if !yield(node.key) {
				return false
			}
			return reverse(node.left)
		}
		reverse(s.root)
	}
}

// Floor returns the largest element <= value, or zero and false.
func (s *{{.Name}}) Floor(value {{.GoType}}) ({{.GoType}}, bool) {
	var result *{{.SnakeName}}Node
	node := s.root
	for node != nil {
		if {{if .IsFloat}}{{.CmpFn}}(value, node.key) == 0{{else}}value == node.key{{end}} {
			return node.key, true
		}
		if {{if .IsFloat}}{{.CmpFn}}(value, node.key) > 0{{else}}value > node.key{{end}} {
			result = node
			node = node.right
		} else {
			node = node.left
		}
	}
	if result == nil {
		return {{.Zero}}, false
	}
	return result.key, true
}

// Ceiling returns the smallest element >= value, or zero and false.
func (s *{{.Name}}) Ceiling(value {{.GoType}}) ({{.GoType}}, bool) {
	var result *{{.SnakeName}}Node
	node := s.root
	for node != nil {
		if {{if .IsFloat}}{{.CmpFn}}(value, node.key) == 0{{else}}value == node.key{{end}} {
			return node.key, true
		}
		if {{if .IsFloat}}{{.CmpFn}}(value, node.key) < 0{{else}}value < node.key{{end}} {
			result = node
			node = node.left
		} else {
			node = node.right
		}
	}
	if result == nil {
		return {{.Zero}}, false
	}
	return result.key, true
}

// All returns an iter.Seq that yields elements in ascending order.
func (s *{{.Name}}) All() iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
		var inorder func(node *{{.SnakeName}}Node) bool
		inorder = func(node *{{.SnakeName}}Node) bool {
			if node == nil {
				return true
			}
			if !inorder(node.left) {
				return false
			}
			if !yield(node.key) {
				return false
			}
			return inorder(node.right)
		}
		inorder(s.root)
	}
}

// Segments cuts the set into up to n balanced, contiguous, non-overlapping views
// over the SORTED order (k = min(n, Len), or 1 when empty) whose concatenation
// reproduces All in ascending order — so a *{{.Name}} satisfies
// par.Segmenter[{{.GoType}}] and feeds par.From with ORDERED segments, each
// covering a contiguous rank range.
//
// Boundaries come from the per-node subtree-size augmentation: each view walks
// only its rank range in-order, pruning whole subtrees that fall outside it, so
// the split is O(Len + n·height) total, not O(Len·log Len). The views are live
// over the tree, and the rank ranges are fixed from Len at the moment of this
// call: mutating the set any time after Segments returns and before the returned
// views are exhausted or discarded is undefined behavior (a resize shifts every
// rank, so stale ranges would drop or double elements).
func (s *{{.Name}}) Segments(n int) []iter.Seq[{{.GoType}}] {
	ranges := segment.SplitRanges(s.size, n)
	segs := make([]iter.Seq[{{.GoType}}], len(ranges))
	for i := range ranges {
		lo, hi := ranges[i][0], ranges[i][1] // per-iteration copies for this view's closure
		segs[i] = func(yield func({{.GoType}}) bool) {
			// In-order walk yielding elements whose 0-based rank is in [lo, hi).
			// base is the rank of the first element in node's subtree; the node's
			// own rank is base + size(node.left). Subtrees entirely outside
			// [lo, hi) are pruned via the size augmentation.
			var walk func(node *{{.SnakeName}}Node, base int) bool
			walk = func(node *{{.SnakeName}}Node, base int) bool {
				if node == nil {
					return true
				}
				rank := base + {{.SnakeName}}NodeSize(node.left)
				if lo < rank { // left subtree holds ranks [base, rank)
					if !walk(node.left, base) {
						return false
					}
				}
				if rank >= lo && rank < hi {
					if !yield(node.key) {
						return false
					}
				}
				if hi > rank+1 { // right subtree holds ranks [rank+1, base+size)
					if !walk(node.right, rank+1) {
						return false
					}
				}
				return true
			}
			walk(s.root, 0)
		}
	}
	return segs
}

// RangeValues returns an iter.Seq that yields elements in [from, to).
func (s *{{.Name}}) RangeValues(from, to {{.GoType}}) iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
		for v := range s.All() {
			if {{if .IsFloat}}{{.CmpFn}}(v, from) < 0{{else}}v < from{{end}} {
				continue
			}
			if {{if .IsFloat}}{{.CmpFn}}(v, to) >= 0{{else}}v >= to{{end}} {
				return
			}
			if !yield(v) {
				return
			}
		}
	}
}

// ForEach calls the function for each element in ascending order.
func (s *{{.Name}}) ForEach(f func({{.GoType}})) {
	for v := range s.All() {
		f(v)
	}
}

// SelectWhere returns a new sorted set with elements satisfying the predicate.
//
// Named SelectWhere (not Select) so the bare Select name is reserved for the
// order-statistic Select (i-th smallest by 0-based rank), per
// spec/features/rank-select.md.
func (s *{{.Name}}) SelectWhere(predicate func({{.GoType}}) bool) *{{.Name}} {
	result := New{{.Name}}()
	for v := range s.All() {
		if predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Reject returns a new sorted set with elements NOT satisfying the predicate.
func (s *{{.Name}}) Reject(predicate func({{.GoType}}) bool) *{{.Name}} {
	result := New{{.Name}}()
	for v := range s.All() {
		if !predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Detect returns the first element satisfying the predicate, or (zero, false) if none.
func (s *{{.Name}}) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	for v := range s.All() {
		if predicate(v) {
			return v, true
		}
	}
	var zero {{.GoType}}
	return zero, false
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *{{.Name}}) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for v := range s.All() {
		if predicate(v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *{{.Name}}) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for v := range s.All() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *{{.Name}}) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for v := range s.All() {
		if predicate(v) {
			return false
		}
	}
	return true
}

// Count returns the number of elements satisfying the predicate.
func (s *{{.Name}}) Count(predicate func({{.GoType}}) bool) int {
	c := 0
	for v := range s.All() {
		if predicate(v) {
			c++
		}
	}
	return c
}

// Union returns a new sorted set with elements from both sets.
func (s *{{.Name}}) Union(other *{{.Name}}) *{{.Name}} {
	result := New{{.Name}}()
	for v := range s.All() {
		result.Add(v)
	}
	for v := range other.All() {
		result.Add(v)
	}
	return result
}

// Intersect returns a new sorted set with elements in both sets.
func (s *{{.Name}}) Intersect(other *{{.Name}}) *{{.Name}} {
	result := New{{.Name}}()
	for v := range s.All() {
		if other.Contains(v) {
			result.Add(v)
		}
	}
	return result
}

// Difference returns a new sorted set with elements in this but not other.
func (s *{{.Name}}) Difference(other *{{.Name}}) *{{.Name}} {
	result := New{{.Name}}()
	for v := range s.All() {
		if !other.Contains(v) {
			result.Add(v)
		}
	}
	return result
}

// ToSlice returns elements as a sorted slice.
func (s *{{.Name}}) ToSlice() []{{.GoType}} {
	result := make([]{{.GoType}}, 0, s.size)
	for v := range s.All() {
		result = append(result, v)
	}
	return result
}

// AddReturning adds the value to the set and returns the receiver (mutating, fluent).
func (s *{{.Name}}) AddReturning(value {{.GoType}}) *{{.Name}} { s.Add(value); return s }

// RemoveReturning removes the value from the set and returns the receiver (mutating, fluent).
func (s *{{.Name}}) RemoveReturning(value {{.GoType}}) *{{.Name}} { s.Remove(value); return s }

// --- Order statistics (rank / select) ---
//
// Backed by the per-node subtree-size augmentation; both run in O(log n) on the
// balanced tree. Comparisons use the same ordering as insertion and in-order
// traversal{{if .IsFloat}} (the IEEE total order via {{.CmpFn}} for float elements, so NaN sorts
// to the top and ±0 are distinguished){{end}}.

// Rank returns the number of elements strictly less than value under the set's
// ordering -- the 0-based lower-bound index value occupies (if present) or would
// occupy (if absent). Defined for present and absent values alike; the result is
// in 0..=Len() (Len() for any value greater than the maximum). Pure query.
func (s *{{.Name}}) Rank(value {{.GoType}}) int {
	rank := 0
	node := s.root
	for node != nil {
		if {{if .IsFloat}}{{.CmpFn}}(value, node.key) < 0{{else}}value < node.key{{end}} {
			node = node.left
		} else if {{if .IsFloat}}{{.CmpFn}}(value, node.key) > 0{{else}}value > node.key{{end}} {
			rank += 1 + {{.SnakeName}}NodeSize(node.left)
			node = node.right
		} else {
			return rank + {{.SnakeName}}NodeSize(node.left)
		}
	}
	return rank
}

// Select returns the i-th smallest element (0-based) and true, or the zero value
// and false if i >= Len() or i < 0. Out-of-range indices (including on an empty
// set and negative i) return absence and do not trap. Round-trips with Rank:
// Select(Rank(x)) == (x, true) for present x, and Rank(x') == i for the x'
// returned by Select(i) at every 0 <= i < Len().
//
// This is the order-statistic select; the predicate-based filtering convenience
// is SelectWhere (per spec/features/rank-select.md, the bare Select name is the
// order-statistic select).
func (s *{{.Name}}) Select(i int) ({{.GoType}}, bool) {
	if i < 0 {
		return {{.Zero}}, false
	}
	node := s.root
	for node != nil {
		left := {{.SnakeName}}NodeSize(node.left)
		if i < left {
			node = node.left
		} else if i == left {
			return node.key, true
		} else {
			i -= left + 1
			node = node.right
		}
	}
	return {{.Zero}}, false
}

// String returns a string representation in sorted order.
func (s *{{.Name}}) String() string {
	if s.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for v := range s.All() {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v", v)
		first = false
	}
	sb.WriteString("}")
	return sb.String()
}

// --- Data pump (bulk import) ---

// New{{.Name}}FromSorted builds a {{.Name}} from presorted, ascending values in
// a single O(n) pass, skipping the per-insert rebalancing of Add. values must be
// in ascending order according to the set's own comparator (the IEEE-754 total
// order for float values).
//
// On an out-of-order value it returns pump.ErrNotSorted. On a duplicate it
// returns pump.ErrDuplicateKey unless policy is
// pump.IgnoreDuplicates, in which case the duplicate is skipped. A failed
// build returns a nil set, never a half-built one. The result is observably
// identical to the same values inserted one-by-one with Add, and is a valid
// red-black tree so later Add/Remove preserve the invariant.
func New{{.Name}}FromSorted(values []{{.GoType}}, policy pump.DuplicatePolicy) (*{{.Name}}, error) {
	dv, err := dedup{{.Name}}Sorted(values, policy)
	if err != nil {
		return nil, err
	}
	s := New{{.Name}}()
	s.root = s.build{{.SnakeName}}(dv, 0, len(dv)-1, 0, pump.RedBlackRedLevel(len(dv)), nil)
	s.size = len(dv)
	return s, nil
}

// dedup{{.Name}}Sorted validates ascending order and applies the duplicate
// policy, returning a compacted value slice.
func dedup{{.Name}}Sorted(values []{{.GoType}}, policy pump.DuplicatePolicy) ([]{{.GoType}}, error) {
	if len(values) == 0 {
		return values, nil
	}
	out := make([]{{.GoType}}, 0, len(values))
	out = append(out, values[0])
	for i := 1; i < len(values); i++ {
		cmp := {{if .IsFloat}}{{.CmpFn}}(values[i], values[i-1]){{else}}cmp{{.Name}}(values[i], values[i-1]){{end}}
		if cmp < 0 {
			return nil, pump.ErrNotSorted
		}
		if cmp == 0 {
			if policy == pump.IgnoreDuplicates {
				continue
			}
			return nil, pump.ErrDuplicateKey
		}
		out = append(out, values[i])
	}
	return out, nil
}

{{if not .IsFloat}}
// cmp{{.Name}} is the three-way ordering used by the bulk-load validator for
// integer/char values (float values use the IEEE total-order helper instead).
func cmp{{.Name}}(a, b {{.GoType}}) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
{{end}}

// build{{.SnakeName}} recursively builds a perfectly balanced subtree over
// [lo, hi], colouring nodes on redLevel red and all others black (classic JDK
// buildFromSorted).
func (s *{{.Name}}) build{{.SnakeName}}(values []{{.GoType}}, lo, hi, level, redLevel int, parent *{{.SnakeName}}Node) *{{.SnakeName}}Node {
	if lo > hi {
		return nil
	}
	mid := (lo + hi) / 2
	node := &{{.SnakeName}}Node{key: values[mid], parent: parent, color: {{.SnakeName}}NodeBlack}
	if level == redLevel {
		node.color = {{.SnakeName}}NodeRed
	}
	node.left = s.build{{.SnakeName}}(values, lo, mid-1, level+1, redLevel, node)
	node.right = s.build{{.SnakeName}}(values, mid+1, hi, level+1, redLevel, node)
	// Set the subtree-size augmentation bottom-up so Rank/Select work after a
	// bulk load exactly as they do after one-by-one Add. Children are already
	// built above, so their sizes are final here.
	node.size = 1 + {{.SnakeName}}NodeSize(node.left) + {{.SnakeName}}NodeSize(node.right)
	return node
}

// {{.Name}}Sink is a streaming builder for a {{.Name}}: callers Add ascending
// values, then Build the finished set. It is a thin wrapper over
// New{{.Name}}FromSorted. After an error or after Build the sink is poisoned and
// further Add/Build calls panic.
type {{.Name}}Sink struct {
	values []{{.GoType}}
	policy pump.DuplicatePolicy
	done   bool
}

// New{{.Name}}Sink creates a streaming sink with the given duplicate policy.
func New{{.Name}}Sink(policy pump.DuplicatePolicy) *{{.Name}}Sink {
	return &{{.Name}}Sink{policy: policy}
}

// Add appends one value. Values must be supplied in ascending order; order and
// duplicate violations are reported by Build. Calling Add after Build panics.
func (s *{{.Name}}Sink) Add(value {{.GoType}}) {
	if s.done {
		panic("mapdb: Add on a finished {{.Name}}Sink")
	}
	s.values = append(s.values, value)
}

// Build finishes the sink and returns the set. The sink is poisoned afterwards
// (a second Build panics).
func (s *{{.Name}}Sink) Build() (*{{.Name}}, error) {
	if s.done {
		panic("mapdb: Build on a finished {{.Name}}Sink")
	}
	s.done = true
	return New{{.Name}}FromSorted(s.values, s.policy)
}

// --- Red-black tree internals (same as TreeMap) ---

func (s *{{.Name}}) findNode(key {{.GoType}}) *{{.SnakeName}}Node {
	node := s.root
	for node != nil {
		if {{if .IsFloat}}{{.CmpFn}}(key, node.key) < 0{{else}}key < node.key{{end}} {
			node = node.left
		} else if {{if .IsFloat}}{{.CmpFn}}(key, node.key) > 0{{else}}key > node.key{{end}} {
			node = node.right
		} else {
			return node
		}
	}
	return nil
}
func (s *{{.Name}}) minNode(node *{{.SnakeName}}Node) *{{.SnakeName}}Node {
	for node.left != nil {
		node = node.left
	}
	return node
}
func (s *{{.Name}}) maxNode(node *{{.SnakeName}}Node) *{{.SnakeName}}Node {
	for node.right != nil {
		node = node.right
	}
	return node
}

func (s *{{.Name}}) rotateLeft(x *{{.SnakeName}}Node) {
	y := x.right
	x.right = y.left
	if y.left != nil {
		y.left.parent = x
	}
	y.parent = x.parent
	if x.parent == nil {
		s.root = y
	} else if x == x.parent.left {
		x.parent.left = y
	} else {
		x.parent.right = y
	}
	y.left = x
	x.parent = y
	// y took x's former position so it inherits x's old subtree size; recompute
	// bottom-up: the demoted x first (now y's left child), then the promoted y.
	{{.SnakeName}}NodeFixSize(x)
	{{.SnakeName}}NodeFixSize(y)
}
func (s *{{.Name}}) rotateRight(x *{{.SnakeName}}Node) {
	y := x.left
	x.left = y.right
	if y.right != nil {
		y.right.parent = x
	}
	y.parent = x.parent
	if x.parent == nil {
		s.root = y
	} else if x == x.parent.right {
		x.parent.right = y
	} else {
		x.parent.left = y
	}
	y.right = x
	x.parent = y
	// Symmetric to rotateLeft: recompute the demoted x, then the promoted y.
	{{.SnakeName}}NodeFixSize(x)
	{{.SnakeName}}NodeFixSize(y)
}

// incSizeToRoot walks from n up to the root, bumping each ancestor's cached
// subtree size by one after a new leaf was linked below n.
func (s *{{.Name}}) incSizeToRoot(n *{{.SnakeName}}Node) {
	for ; n != nil; n = n.parent {
		n.size++
	}
}

// fixSizeToRoot walks from n up to the root recomputing each node's cached
// subtree size from its children. Used after a delete splice (the rotations
// inside fixAfterDelete already maintain their own sizes).
func (s *{{.Name}}) fixSizeToRoot(n *{{.SnakeName}}Node) {
	for ; n != nil; n = n.parent {
		{{.SnakeName}}NodeFixSize(n)
	}
}

func (s *{{.Name}}) fixAfterInsert(z *{{.SnakeName}}Node) {
	for z.parent != nil && z.parent.color == {{.SnakeName}}NodeRed {
		if z.parent == z.parent.parent.left {
			y := z.parent.parent.right
			if y != nil && y.color == {{.SnakeName}}NodeRed {
				z.parent.color = {{.SnakeName}}NodeBlack
				y.color = {{.SnakeName}}NodeBlack
				z.parent.parent.color = {{.SnakeName}}NodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.right {
					z = z.parent
					s.rotateLeft(z)
				}
				z.parent.color = {{.SnakeName}}NodeBlack
				z.parent.parent.color = {{.SnakeName}}NodeRed
				s.rotateRight(z.parent.parent)
			}
		} else {
			y := z.parent.parent.left
			if y != nil && y.color == {{.SnakeName}}NodeRed {
				z.parent.color = {{.SnakeName}}NodeBlack
				y.color = {{.SnakeName}}NodeBlack
				z.parent.parent.color = {{.SnakeName}}NodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.left {
					z = z.parent
					s.rotateRight(z)
				}
				z.parent.color = {{.SnakeName}}NodeBlack
				z.parent.parent.color = {{.SnakeName}}NodeRed
				s.rotateLeft(z.parent.parent)
			}
		}
	}
	s.root.color = {{.SnakeName}}NodeBlack
}

func (s *{{.Name}}) deleteNode(z *{{.SnakeName}}Node) {
	if z.left != nil && z.right != nil {
		succ := s.minNode(z.right)
		z.key = succ.key
		z = succ
	}
	// z is now the node physically spliced out. fixSizeFrom is the lowest node
	// whose cached subtree size must be refreshed (the removed node's surviving
	// parent at unlink time); recomputing that path to the root once the structure
	// is final restores the invariant. Rotations inside fixAfterDelete maintain
	// their own sizes and everything below fixSizeFrom is left consistent.
	var fixSizeFrom *{{.SnakeName}}Node
	var child *{{.SnakeName}}Node
	if z.left != nil {
		child = z.left
	} else {
		child = z.right
	}
	if child != nil {
		child.parent = z.parent
		if z.parent == nil {
			s.root = child
		} else if z == z.parent.left {
			z.parent.left = child
		} else {
			z.parent.right = child
		}
		fixSizeFrom = child
		if z.color == {{.SnakeName}}NodeBlack {
			s.fixAfterDelete(child)
		}
	} else if z.parent == nil {
		s.root = nil
	} else {
		if z.color == {{.SnakeName}}NodeBlack {
			s.fixAfterDelete(z)
		}
		// fixAfterDelete may have rotated z to a new parent; read it now.
		fixSizeFrom = z.parent
		if z.parent != nil {
			if z == z.parent.left {
				z.parent.left = nil
			} else {
				z.parent.right = nil
			}
		}
	}
	s.fixSizeToRoot(fixSizeFrom)
}

func (s *{{.Name}}) fixAfterDelete(x *{{.SnakeName}}Node) {
	for x != s.root && x.color == {{.SnakeName}}NodeBlack {
		if x == x.parent.left {
			w := x.parent.right
			if w == nil {
				x = x.parent
				continue
			}
			if w.color == {{.SnakeName}}NodeRed {
				w.color = {{.SnakeName}}NodeBlack
				x.parent.color = {{.SnakeName}}NodeRed
				s.rotateLeft(x.parent)
				w = x.parent.right
			}
			if w == nil {
				x = x.parent
				continue
			}
			lb := w.left == nil || w.left.color == {{.SnakeName}}NodeBlack
			rb := w.right == nil || w.right.color == {{.SnakeName}}NodeBlack
			if lb && rb {
				w.color = {{.SnakeName}}NodeRed
				x = x.parent
			} else {
				if rb {
					if w.left != nil {
						w.left.color = {{.SnakeName}}NodeBlack
					}
					w.color = {{.SnakeName}}NodeRed
					s.rotateRight(w)
					w = x.parent.right
				}
				w.color = x.parent.color
				x.parent.color = {{.SnakeName}}NodeBlack
				if w.right != nil {
					w.right.color = {{.SnakeName}}NodeBlack
				}
				s.rotateLeft(x.parent)
				x = s.root
			}
		} else {
			w := x.parent.left
			if w == nil {
				x = x.parent
				continue
			}
			if w.color == {{.SnakeName}}NodeRed {
				w.color = {{.SnakeName}}NodeBlack
				x.parent.color = {{.SnakeName}}NodeRed
				s.rotateRight(x.parent)
				w = x.parent.left
			}
			if w == nil {
				x = x.parent
				continue
			}
			lb := w.left == nil || w.left.color == {{.SnakeName}}NodeBlack
			rb := w.right == nil || w.right.color == {{.SnakeName}}NodeBlack
			if lb && rb {
				w.color = {{.SnakeName}}NodeRed
				x = x.parent
			} else {
				if lb {
					if w.right != nil {
						w.right.color = {{.SnakeName}}NodeBlack
					}
					w.color = {{.SnakeName}}NodeRed
					s.rotateLeft(w)
					w = x.parent.left
				}
				w.color = x.parent.color
				x.parent.color = {{.SnakeName}}NodeBlack
				if w.left != nil {
					w.left.color = {{.SnakeName}}NodeBlack
				}
				s.rotateRight(x.parent)
				x = s.root
			}
		}
	}
	x.color = {{.SnakeName}}NodeBlack
}
`
