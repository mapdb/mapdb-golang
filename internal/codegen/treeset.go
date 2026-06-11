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
// phase-3 correctness fix). Note that the Floor/Ceiling/RangeValues equality
// and bound checks use raw == / < / >= even for floats — that is reproduced
// exactly, the cmpFloat helper governs only the strict directional ordering.
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
	{{.SnakeName}}TreeSetNodeRed   = false
	{{.SnakeName}}TreeSetNodeBlack = true
)

type {{.SnakeName}}TreeSetNode struct {
	key    {{.GoType}}
	left   *{{.SnakeName}}TreeSetNode
	right  *{{.SnakeName}}TreeSetNode
	parent *{{.SnakeName}}TreeSetNode
	color  bool
}

// {{.Name}}TreeSet is a sorted set of {{.GoType}} values, backed by a red-black tree.
type {{.Name}}TreeSet struct {
	root *{{.SnakeName}}TreeSetNode
	size int
}

// New{{.Name}}TreeSet creates a new empty sorted set.
func New{{.Name}}TreeSet() *{{.Name}}TreeSet {
	return &{{.Name}}TreeSet{}
}

// {{.Name}}TreeSetOf creates a new sorted set from the given values.
func {{.Name}}TreeSetOf(values ...{{.GoType}}) *{{.Name}}TreeSet {
	s := New{{.Name}}TreeSet()
	for _, v := range values {
		s.Add(v)
	}
	return s
}

// Add inserts a value. Returns true if added (not already present).
func (s *{{.Name}}TreeSet) Add(value {{.GoType}}) bool {
	if s.root == nil {
		s.root = &{{.SnakeName}}TreeSetNode{key: value, color: {{.SnakeName}}TreeSetNodeBlack}
		s.size++
		return true
	}
	node := s.root
	for {
		if {{if .IsFloat}}{{.CmpFn}}(value, node.key) < 0{{else}}value < node.key{{end}} {
			if node.left == nil {
				node.left = &{{.SnakeName}}TreeSetNode{key: value, parent: node, color: {{.SnakeName}}TreeSetNodeRed}
				s.fixAfterInsert(node.left)
				s.size++
				return true
			}
			node = node.left
		} else if {{if .IsFloat}}{{.CmpFn}}(value, node.key) > 0{{else}}value > node.key{{end}} {
			if node.right == nil {
				node.right = &{{.SnakeName}}TreeSetNode{key: value, parent: node, color: {{.SnakeName}}TreeSetNodeRed}
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
func (s *{{.Name}}TreeSet) Remove(value {{.GoType}}) bool {
	node := s.findNode(value)
	if node == nil {
		return false
	}
	s.deleteNode(node)
	s.size--
	return true
}

// Contains returns true if the set contains the value.
func (s *{{.Name}}TreeSet) Contains(value {{.GoType}}) bool {
	return s.findNode(value) != nil
}

// Size returns the number of elements.
func (s *{{.Name}}TreeSet) Size() int { return s.size }

// IsEmpty returns true if the set is empty.
func (s *{{.Name}}TreeSet) IsEmpty() bool { return s.size == 0 }

// Clear removes all elements.
func (s *{{.Name}}TreeSet) Clear() { s.root = nil; s.size = 0 }

// Min returns the smallest element, or zero and false if empty.
func (s *{{.Name}}TreeSet) Min() ({{.GoType}}, bool) {
	if s.root == nil {
		return {{.Zero}}, false
	}
	return s.minNode(s.root).key, true
}

// Max returns the largest element, or zero and false if empty.
func (s *{{.Name}}TreeSet) Max() ({{.GoType}}, bool) {
	if s.root == nil {
		return {{.Zero}}, false
	}
	return s.maxNode(s.root).key, true
}

// Floor returns the largest element <= value, or zero and false.
func (s *{{.Name}}TreeSet) Floor(value {{.GoType}}) ({{.GoType}}, bool) {
	var result *{{.SnakeName}}TreeSetNode
	node := s.root
	for node != nil {
		if value == node.key {
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
func (s *{{.Name}}TreeSet) Ceiling(value {{.GoType}}) ({{.GoType}}, bool) {
	var result *{{.SnakeName}}TreeSetNode
	node := s.root
	for node != nil {
		if value == node.key {
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
func (s *{{.Name}}TreeSet) All() iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
		var inorder func(node *{{.SnakeName}}TreeSetNode) bool
		inorder = func(node *{{.SnakeName}}TreeSetNode) bool {
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
func (s *{{.Name}}TreeSet) RangeValues(from, to {{.GoType}}) iter.Seq[{{.GoType}}] {
	return func(yield func({{.GoType}}) bool) {
		for v := range s.All() {
			if v < from {
				continue
			}
			if v >= to {
				return
			}
			if !yield(v) {
				return
			}
		}
	}
}

// ForEach calls the function for each element in ascending order.
func (s *{{.Name}}TreeSet) ForEach(f func({{.GoType}})) {
	for v := range s.All() {
		f(v)
	}
}

// Select returns a new sorted set with elements satisfying the predicate.
func (s *{{.Name}}TreeSet) Select(predicate func({{.GoType}}) bool) *{{.Name}}TreeSet {
	result := New{{.Name}}TreeSet()
	for v := range s.All() {
		if predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Reject returns a new sorted set with elements NOT satisfying the predicate.
func (s *{{.Name}}TreeSet) Reject(predicate func({{.GoType}}) bool) *{{.Name}}TreeSet {
	result := New{{.Name}}TreeSet()
	for v := range s.All() {
		if !predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Detect returns the first element satisfying the predicate, or (zero, false) if none.
func (s *{{.Name}}TreeSet) Detect(predicate func({{.GoType}}) bool) ({{.GoType}}, bool) {
	for v := range s.All() {
		if predicate(v) {
			return v, true
		}
	}
	var zero {{.GoType}}
	return zero, false
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *{{.Name}}TreeSet) AnySatisfy(predicate func({{.GoType}}) bool) bool {
	for v := range s.All() {
		if predicate(v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all elements satisfy the predicate.
func (s *{{.Name}}TreeSet) AllSatisfy(predicate func({{.GoType}}) bool) bool {
	for v := range s.All() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate.
func (s *{{.Name}}TreeSet) NoneSatisfy(predicate func({{.GoType}}) bool) bool {
	for v := range s.All() {
		if predicate(v) {
			return false
		}
	}
	return true
}

// Count returns the number of elements satisfying the predicate.
func (s *{{.Name}}TreeSet) Count(predicate func({{.GoType}}) bool) int {
	c := 0
	for v := range s.All() {
		if predicate(v) {
			c++
		}
	}
	return c
}

// Union returns a new sorted set with elements from both sets.
func (s *{{.Name}}TreeSet) Union(other *{{.Name}}TreeSet) *{{.Name}}TreeSet {
	result := New{{.Name}}TreeSet()
	for v := range s.All() {
		result.Add(v)
	}
	for v := range other.All() {
		result.Add(v)
	}
	return result
}

// Intersect returns a new sorted set with elements in both sets.
func (s *{{.Name}}TreeSet) Intersect(other *{{.Name}}TreeSet) *{{.Name}}TreeSet {
	result := New{{.Name}}TreeSet()
	for v := range s.All() {
		if other.Contains(v) {
			result.Add(v)
		}
	}
	return result
}

// Difference returns a new sorted set with elements in this but not other.
func (s *{{.Name}}TreeSet) Difference(other *{{.Name}}TreeSet) *{{.Name}}TreeSet {
	result := New{{.Name}}TreeSet()
	for v := range s.All() {
		if !other.Contains(v) {
			result.Add(v)
		}
	}
	return result
}

// ToSlice returns elements as a sorted slice.
func (s *{{.Name}}TreeSet) ToSlice() []{{.GoType}} {
	result := make([]{{.GoType}}, 0, s.size)
	for v := range s.All() {
		result = append(result, v)
	}
	return result
}

// With returns the set after adding the value (fluent API).
func (s *{{.Name}}TreeSet) With(value {{.GoType}}) *{{.Name}}TreeSet { s.Add(value); return s }

// Without returns the set after removing the value (fluent API).
func (s *{{.Name}}TreeSet) Without(value {{.GoType}}) *{{.Name}}TreeSet { s.Remove(value); return s }

// String returns a string representation in sorted order.
func (s *{{.Name}}TreeSet) String() string {
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

func (s *{{.Name}}TreeSet) findNode(key {{.GoType}}) *{{.SnakeName}}TreeSetNode {
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
func (s *{{.Name}}TreeSet) minNode(node *{{.SnakeName}}TreeSetNode) *{{.SnakeName}}TreeSetNode {
	for node.left != nil {
		node = node.left
	}
	return node
}
func (s *{{.Name}}TreeSet) maxNode(node *{{.SnakeName}}TreeSetNode) *{{.SnakeName}}TreeSetNode {
	for node.right != nil {
		node = node.right
	}
	return node
}

func (s *{{.Name}}TreeSet) rotateLeft(x *{{.SnakeName}}TreeSetNode) {
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
func (s *{{.Name}}TreeSet) rotateRight(x *{{.SnakeName}}TreeSetNode) {
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

func (s *{{.Name}}TreeSet) fixAfterInsert(z *{{.SnakeName}}TreeSetNode) {
	for z.parent != nil && z.parent.color == {{.SnakeName}}TreeSetNodeRed {
		if z.parent == z.parent.parent.left {
			y := z.parent.parent.right
			if y != nil && y.color == {{.SnakeName}}TreeSetNodeRed {
				z.parent.color = {{.SnakeName}}TreeSetNodeBlack
				y.color = {{.SnakeName}}TreeSetNodeBlack
				z.parent.parent.color = {{.SnakeName}}TreeSetNodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.right {
					z = z.parent
					s.rotateLeft(z)
				}
				z.parent.color = {{.SnakeName}}TreeSetNodeBlack
				z.parent.parent.color = {{.SnakeName}}TreeSetNodeRed
				s.rotateRight(z.parent.parent)
			}
		} else {
			y := z.parent.parent.left
			if y != nil && y.color == {{.SnakeName}}TreeSetNodeRed {
				z.parent.color = {{.SnakeName}}TreeSetNodeBlack
				y.color = {{.SnakeName}}TreeSetNodeBlack
				z.parent.parent.color = {{.SnakeName}}TreeSetNodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.left {
					z = z.parent
					s.rotateRight(z)
				}
				z.parent.color = {{.SnakeName}}TreeSetNodeBlack
				z.parent.parent.color = {{.SnakeName}}TreeSetNodeRed
				s.rotateLeft(z.parent.parent)
			}
		}
	}
	s.root.color = {{.SnakeName}}TreeSetNodeBlack
}

func (s *{{.Name}}TreeSet) deleteNode(z *{{.SnakeName}}TreeSetNode) {
	if z.left != nil && z.right != nil {
		succ := s.minNode(z.right)
		z.key = succ.key
		z = succ
	}
	var child *{{.SnakeName}}TreeSetNode
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
		if z.color == {{.SnakeName}}TreeSetNodeBlack {
			s.fixAfterDelete(child)
		}
	} else if z.parent == nil {
		s.root = nil
	} else {
		if z.color == {{.SnakeName}}TreeSetNodeBlack {
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

func (s *{{.Name}}TreeSet) fixAfterDelete(x *{{.SnakeName}}TreeSetNode) {
	for x != s.root && x.color == {{.SnakeName}}TreeSetNodeBlack {
		if x == x.parent.left {
			w := x.parent.right
			if w == nil {
				x = x.parent
				continue
			}
			if w.color == {{.SnakeName}}TreeSetNodeRed {
				w.color = {{.SnakeName}}TreeSetNodeBlack
				x.parent.color = {{.SnakeName}}TreeSetNodeRed
				s.rotateLeft(x.parent)
				w = x.parent.right
			}
			if w == nil {
				x = x.parent
				continue
			}
			lb := w.left == nil || w.left.color == {{.SnakeName}}TreeSetNodeBlack
			rb := w.right == nil || w.right.color == {{.SnakeName}}TreeSetNodeBlack
			if lb && rb {
				w.color = {{.SnakeName}}TreeSetNodeRed
				x = x.parent
			} else {
				if rb {
					if w.left != nil {
						w.left.color = {{.SnakeName}}TreeSetNodeBlack
					}
					w.color = {{.SnakeName}}TreeSetNodeRed
					s.rotateRight(w)
					w = x.parent.right
				}
				w.color = x.parent.color
				x.parent.color = {{.SnakeName}}TreeSetNodeBlack
				if w.right != nil {
					w.right.color = {{.SnakeName}}TreeSetNodeBlack
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
			if w.color == {{.SnakeName}}TreeSetNodeRed {
				w.color = {{.SnakeName}}TreeSetNodeBlack
				x.parent.color = {{.SnakeName}}TreeSetNodeRed
				s.rotateRight(x.parent)
				w = x.parent.left
			}
			if w == nil {
				x = x.parent
				continue
			}
			lb := w.left == nil || w.left.color == {{.SnakeName}}TreeSetNodeBlack
			rb := w.right == nil || w.right.color == {{.SnakeName}}TreeSetNodeBlack
			if lb && rb {
				w.color = {{.SnakeName}}TreeSetNodeRed
				x = x.parent
			} else {
				if lb {
					if w.right != nil {
						w.right.color = {{.SnakeName}}TreeSetNodeBlack
					}
					w.color = {{.SnakeName}}TreeSetNodeRed
					s.rotateLeft(w)
					w = x.parent.left
				}
				w.color = x.parent.color
				x.parent.color = {{.SnakeName}}TreeSetNodeBlack
				if w.left != nil {
					w.left.color = {{.SnakeName}}TreeSetNodeBlack
				}
				s.rotateRight(x.parent)
				x = s.root
			}
		}
	}
	x.color = {{.SnakeName}}TreeSetNodeBlack
}
`
