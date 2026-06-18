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

	base := template.Must(template.New("ts-base").Parse(treeSetTmpl))

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

	return genCmpFloat("treeset")
}

const treeSetTmpl = genHeader + `package treeset

import (
	"fmt"
	"iter"
	"strings"
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
		s.root = &{{.SnakeName}}Node{key: value, color: {{.SnakeName}}NodeBlack}
		s.size++
		return true
	}
	node := s.root
	for {
		if {{if .IsFloat}}{{.CmpFn}}(value, node.key) < 0{{else}}value < node.key{{end}} {
			if node.left == nil {
				node.left = &{{.SnakeName}}Node{key: value, parent: node, color: {{.SnakeName}}NodeRed}
				s.fixAfterInsert(node.left)
				s.size++
				return true
			}
			node = node.left
		} else if {{if .IsFloat}}{{.CmpFn}}(value, node.key) > 0{{else}}value > node.key{{end}} {
			if node.right == nil {
				node.right = &{{.SnakeName}}Node{key: value, parent: node, color: {{.SnakeName}}NodeRed}
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

// Select returns a new sorted set with elements satisfying the predicate.
func (s *{{.Name}}) Select(predicate func({{.GoType}}) bool) *{{.Name}} {
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
		if z.color == {{.SnakeName}}NodeBlack {
			s.fixAfterDelete(child)
		}
	} else if z.parent == nil {
		s.root = nil
	} else {
		if z.color == {{.SnakeName}}NodeBlack {
			s.fixAfterDelete(z)
		}
		if z.parent != nil {
			if z == z.parent.left {
				z.parent.left = nil
			} else {
				z.parent.right = nil
			}
		}
	}
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
