package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

// tmData is the per key/value view the treemap template iterates over.
//
// The treemap family is a red-black tree sorted map and the first K×V (two
// type-parameter) family in the generator. It has a BASE variant only (no
// immutable, no synchronized wrappers; no object/bool element types).
//
// The KEY type drives all the interesting logic: integer/char keys order with
// raw < / >, whereas float keys order via the shared IEEE total-order helper
// cmpFloat32/64 (the phase-3 correctness fix). As with treeset, for float keys
// the Floor/Ceiling exact-match short-circuits and the RangeKeys/HeadMap/TailMap
// [from, to) bound checks also route through cmpFloat32/64 so key navigation is
// a true total order (NaN sorts to the top, ±0 are distinguished); int/char keys
// keep raw == / < / >=. The cmpFloat helper already governs the strict
// directional ordering used during tree traversal (Put/findNode/Floor/Ceiling/
// Higher/Lower).
//
// The VALUE type is a carried payload only: it appears in method signatures,
// the node's value field, and return tuples. treemap never compares values, so
// the value axis is a pure type-name + zero-literal substitution. Note that
// the KEY zero literal in tuple returns is always "0" (the hand-written
// originals never use 0.0 in the key position), whereas the VALUE zero literal
// follows the value type ("0.0" for floats).
type tmData struct {
	KeyName    string // Int32, Float32, Char (key identifier stem)
	KeyType    string // int32, float32, uint16 (Go key type)
	KeySnake   string // int32, float32, char (key file-name stem)
	KeyIsFloat bool
	CmpFn      string // cmpFloat32 / cmpFloat64 (float keys only)

	ValName  string // Int32, Float32, Char (value identifier stem)
	ValType  string // int32, float32, uint16 (Go value type)
	ValSnake string // int32, float32, char (value file-name stem)
	ValZero  string // value zero literal ("0" or "0.0")

	// MapName is the combined identifier stem, e.g. Int32Float32 (used for the
	// exported type Int32Float32TreeMap); NodeName is the lower-camel node stem,
	// e.g. int32Float32 (used for int32Float32TreeNode).
	MapName  string // Int32Float32
	NodeName string // int32Float32
}

// genTreeMap writes the per key/value red-black tree map sources (base variant
// only, 7×7 = 49 files) plus the shared cmp_float.go into the current working
// directory. Invoked from treemap/ via go:generate.
func genTreeMap() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	base := template.Must(template.New("tm-base").Parse(treeMapTmpl))

	write := func(name string, data tmData) error {
		var buf bytes.Buffer
		if err := base.Execute(&buf, data); err != nil {
			return fmt.Errorf("execute %s: %w", name, err)
		}
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return fmt.Errorf("format %s: %w\n---\n%s", name, err, buf.String())
		}
		out := filepath.Join(cwd, name)
		return os.WriteFile(out, formatted, 0o644)
	}

	prims := Primitives()
	for _, k := range prims {
		for _, v := range prims {
			data := tmData{
				KeyName:    k.Name,
				KeyType:    k.GoType,
				KeySnake:   k.SnakeName,
				KeyIsFloat: k.IsFloating,

				ValName:  v.Name,
				ValType:  v.GoType,
				ValSnake: v.SnakeName,
				ValZero:  "0",

				MapName:  k.Name + v.Name,
				NodeName: lowerFirst(k.Name) + v.Name,
			}
			if k.IsFloating {
				data.CmpFn = "cmpFloat32"
				if k.ByteSize == 8 {
					data.CmpFn = "cmpFloat64"
				}
			}
			if v.IsFloating {
				data.ValZero = "0.0"
			}

			name := k.SnakeName + "_" + v.SnakeName + "_tree_map.go"
			if err := write(name, data); err != nil {
				return err
			}
		}
	}

	return genCmpFloat("treemap")
}

// lowerFirst lowercases the first rune of s (Int32 -> int32, Char -> char).
// The primitive Name stems are ASCII, so a byte-level lowercase is sufficient.
func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	c := s[0]
	if c >= 'A' && c <= 'Z' {
		return string(c+('a'-'A')) + s[1:]
	}
	return s
}

const treeMapTmpl = genHeader + `package treemap

import (
	"fmt"
	"iter"
	"strings"
)

const (
	{{.NodeName}}TreeNodeRed   = false
	{{.NodeName}}TreeNodeBlack = true
)

type {{.NodeName}}TreeNode struct {
	key    {{.KeyType}}
	value  {{.ValType}}
	left   *{{.NodeName}}TreeNode
	right  *{{.NodeName}}TreeNode
	parent *{{.NodeName}}TreeNode
	color  bool
}

// {{.MapName}} is a sorted map with {{.KeyType}} keys and {{.ValType}} values, backed by a red-black tree.
// Keys are maintained in ascending order.
type {{.MapName}} struct {
	root *{{.NodeName}}TreeNode
	size int
}

// New{{.MapName}} creates a new empty sorted map.
func New{{.MapName}}() *{{.MapName}} {
	return &{{.MapName}}{}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *{{.MapName}}) Put(key {{.KeyType}}, value {{.ValType}}) ({{.ValType}}, bool) {
	if m.root == nil {
		m.root = &{{.NodeName}}TreeNode{key: key, value: value, color: {{.NodeName}}TreeNodeBlack}
		m.size++
		return {{.ValZero}}, false
	}
	node := m.root
	for {
		if {{if .KeyIsFloat}}{{.CmpFn}}(key, node.key) < 0{{else}}key < node.key{{end}} {
			if node.left == nil {
				node.left = &{{.NodeName}}TreeNode{key: key, value: value, parent: node, color: {{.NodeName}}TreeNodeRed}
				m.fixAfterInsert(node.left)
				m.size++
				return {{.ValZero}}, false
			}
			node = node.left
		} else if {{if .KeyIsFloat}}{{.CmpFn}}(key, node.key) > 0{{else}}key > node.key{{end}} {
			if node.right == nil {
				node.right = &{{.NodeName}}TreeNode{key: key, value: value, parent: node, color: {{.NodeName}}TreeNodeRed}
				m.fixAfterInsert(node.right)
				m.size++
				return {{.ValZero}}, false
			}
			node = node.right
		} else {
			old := node.value
			node.value = value
			return old, true
		}
	}
}

// Get returns the value for the key, or the zero value and false if not found.
func (m *{{.MapName}}) Get(key {{.KeyType}}) ({{.ValType}}, bool) {
	node := m.findNode(key)
	if node == nil {
		return {{.ValZero}}, false
	}
	return node.value, true
}

// GetOrDefault returns the value for the key if present, or the default value otherwise.
func (m *{{.MapName}}) GetOrDefault(key {{.KeyType}}, defaultValue {{.ValType}}) {{.ValType}} {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// ContainsKey returns true if the map contains the given key.
func (m *{{.MapName}}) ContainsKey(key {{.KeyType}}) bool {
	return m.findNode(key) != nil
}

// Remove removes the entry for the given key. Returns the previous value and true if found.
func (m *{{.MapName}}) Remove(key {{.KeyType}}) ({{.ValType}}, bool) {
	node := m.findNode(key)
	if node == nil {
		return {{.ValZero}}, false
	}
	old := node.value
	m.deleteNode(node)
	m.size--
	return old, true
}

// Len returns the number of elements. Use m.Len() == 0 to test for emptiness.
func (m *{{.MapName}}) Len() int {
	return m.size
}

// Clear removes all entries.
func (m *{{.MapName}}) Clear() {
	m.root = nil
	m.size = 0
}

// Min returns the smallest key and its value, or zero values and false if empty.
func (m *{{.MapName}}) Min() ({{.KeyType}}, {{.ValType}}, bool) {
	if m.root == nil {
		return 0, {{.ValZero}}, false
	}
	node := m.minNode(m.root)
	return node.key, node.value, true
}

// Max returns the largest key and its value, or zero values and false if empty.
func (m *{{.MapName}}) Max() ({{.KeyType}}, {{.ValType}}, bool) {
	if m.root == nil {
		return 0, {{.ValZero}}, false
	}
	node := m.maxNode(m.root)
	return node.key, node.value, true
}

// Floor returns the largest key <= the given key, or zero values and false.
func (m *{{.MapName}}) Floor(key {{.KeyType}}) ({{.KeyType}}, {{.ValType}}, bool) {
	var result *{{.NodeName}}TreeNode
	node := m.root
	for node != nil {
		if {{if .KeyIsFloat}}{{.CmpFn}}(key, node.key) == 0{{else}}key == node.key{{end}} {
			return node.key, node.value, true
		}
		if {{if .KeyIsFloat}}{{.CmpFn}}(key, node.key) > 0{{else}}key > node.key{{end}} {
			result = node
			node = node.right
		} else {
			node = node.left
		}
	}
	if result == nil {
		return 0, {{.ValZero}}, false
	}
	return result.key, result.value, true
}

// Ceiling returns the smallest key >= the given key, or zero values and false.
func (m *{{.MapName}}) Ceiling(key {{.KeyType}}) ({{.KeyType}}, {{.ValType}}, bool) {
	var result *{{.NodeName}}TreeNode
	node := m.root
	for node != nil {
		if {{if .KeyIsFloat}}{{.CmpFn}}(key, node.key) == 0{{else}}key == node.key{{end}} {
			return node.key, node.value, true
		}
		if {{if .KeyIsFloat}}{{.CmpFn}}(key, node.key) < 0{{else}}key < node.key{{end}} {
			result = node
			node = node.left
		} else {
			node = node.right
		}
	}
	if result == nil {
		return 0, {{.ValZero}}, false
	}
	return result.key, result.value, true
}

// All returns an iter.Seq2 that yields all key-value pairs in ascending key order.
func (m *{{.MapName}}) All() iter.Seq2[{{.KeyType}}, {{.ValType}}] {
	return func(yield func({{.KeyType}}, {{.ValType}}) bool) {
		var inorder func(node *{{.NodeName}}TreeNode) bool
		inorder = func(node *{{.NodeName}}TreeNode) bool {
			if node == nil {
				return true
			}
			if !inorder(node.left) {
				return false
			}
			if !yield(node.key, node.value) {
				return false
			}
			return inorder(node.right)
		}
		inorder(m.root)
	}
}

// Keys returns an iter.Seq that yields all keys in ascending order.
func (m *{{.MapName}}) Keys() iter.Seq[{{.KeyType}}] {
	return func(yield func({{.KeyType}}) bool) {
		for k, _ := range m.All() {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iter.Seq that yields all values in key order.
func (m *{{.MapName}}) Values() iter.Seq[{{.ValType}}] {
	return func(yield func({{.ValType}}) bool) {
		for _, v := range m.All() {
			if !yield(v) {
				return
			}
		}
	}
}

// RangeKeys returns an iter.Seq2 that yields entries with keys in [fromKey, toKey).
func (m *{{.MapName}}) RangeKeys(fromKey, toKey {{.KeyType}}) iter.Seq2[{{.KeyType}}, {{.ValType}}] {
	return func(yield func({{.KeyType}}, {{.ValType}}) bool) {
		for k, v := range m.All() {
			if {{if .KeyIsFloat}}{{.CmpFn}}(k, fromKey) < 0{{else}}k < fromKey{{end}} {
				continue
			}
			if {{if .KeyIsFloat}}{{.CmpFn}}(k, toKey) >= 0{{else}}k >= toKey{{end}} {
				return
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

// Higher returns the smallest key strictly greater than ` + "`key`" + ` (and its value),
// or zero values and false. Unlike Ceiling, never returns ` + "`key`" + ` itself.
func (m *{{.MapName}}) Higher(key {{.KeyType}}) ({{.KeyType}}, {{.ValType}}, bool) {
	var result *{{.NodeName}}TreeNode
	node := m.root
	for node != nil {
		if {{if .KeyIsFloat}}{{.CmpFn}}(key, node.key) < 0{{else}}key < node.key{{end}} {
			result = node
			node = node.left
		} else {
			node = node.right
		}
	}
	if result == nil {
		return 0, {{.ValZero}}, false
	}
	return result.key, result.value, true
}

// Lower returns the largest key strictly less than ` + "`key`" + ` (and its value),
// or zero values and false. Unlike Floor, never returns ` + "`key`" + ` itself.
func (m *{{.MapName}}) Lower(key {{.KeyType}}) ({{.KeyType}}, {{.ValType}}, bool) {
	var result *{{.NodeName}}TreeNode
	node := m.root
	for node != nil {
		if {{if .KeyIsFloat}}{{.CmpFn}}(key, node.key) > 0{{else}}key > node.key{{end}} {
			result = node
			node = node.right
		} else {
			node = node.left
		}
	}
	if result == nil {
		return 0, {{.ValZero}}, false
	}
	return result.key, result.value, true
}

// HeadMap returns an iter.Seq2 over entries with keys strictly less than toKey.
// Matches Java NavigableMap.headMap(toKey) (exclusive by default).
func (m *{{.MapName}}) HeadMap(toKey {{.KeyType}}) iter.Seq2[{{.KeyType}}, {{.ValType}}] {
	return func(yield func({{.KeyType}}, {{.ValType}}) bool) {
		for k, v := range m.All() {
			if {{if .KeyIsFloat}}{{.CmpFn}}(k, toKey) >= 0{{else}}k >= toKey{{end}} {
				return
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

// TailMap returns an iter.Seq2 over entries with keys >= fromKey.
// Matches Java NavigableMap.tailMap(fromKey) (inclusive by default).
func (m *{{.MapName}}) TailMap(fromKey {{.KeyType}}) iter.Seq2[{{.KeyType}}, {{.ValType}}] {
	return func(yield func({{.KeyType}}, {{.ValType}}) bool) {
		for k, v := range m.All() {
			if {{if .KeyIsFloat}}{{.CmpFn}}(k, fromKey) < 0{{else}}k < fromKey{{end}} {
				continue
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

// SubMap returns an iter.Seq2 over entries with keys in [fromKey, toKey).
// Alias for RangeKeys; exists for Java-NavigableMap API parity.
func (m *{{.MapName}}) SubMap(fromKey, toKey {{.KeyType}}) iter.Seq2[{{.KeyType}}, {{.ValType}}] {
	return m.RangeKeys(fromKey, toKey)
}

// FirstEntry is an alias of Min — the smallest key and its value, or zero/false.
func (m *{{.MapName}}) FirstEntry() ({{.KeyType}}, {{.ValType}}, bool) { return m.Min() }

// LastEntry is an alias of Max — the largest key and its value, or zero/false.
func (m *{{.MapName}}) LastEntry() ({{.KeyType}}, {{.ValType}}, bool) { return m.Max() }

// FloorKey returns the largest key <= the given key, or zero and false. Key form
// of Floor (drops the value).
func (m *{{.MapName}}) FloorKey(key {{.KeyType}}) ({{.KeyType}}, bool) {
	k, _, ok := m.Floor(key)
	return k, ok
}

// CeilingKey returns the smallest key >= the given key, or zero and false.
func (m *{{.MapName}}) CeilingKey(key {{.KeyType}}) ({{.KeyType}}, bool) {
	k, _, ok := m.Ceiling(key)
	return k, ok
}

// LowerKey returns the largest key strictly < the given key, or zero and false.
func (m *{{.MapName}}) LowerKey(key {{.KeyType}}) ({{.KeyType}}, bool) {
	k, _, ok := m.Lower(key)
	return k, ok
}

// HigherKey returns the smallest key strictly > the given key, or zero and false.
func (m *{{.MapName}}) HigherKey(key {{.KeyType}}) ({{.KeyType}}, bool) {
	k, _, ok := m.Higher(key)
	return k, ok
}

// FirstKey returns the smallest key, or zero and false if empty.
func (m *{{.MapName}}) FirstKey() ({{.KeyType}}, bool) {
	k, _, ok := m.Min()
	return k, ok
}

// LastKey returns the largest key, or zero and false if empty.
func (m *{{.MapName}}) LastKey() ({{.KeyType}}, bool) {
	k, _, ok := m.Max()
	return k, ok
}

// FloorEntry is an alias of Floor — the largest entry with key <= the given key.
func (m *{{.MapName}}) FloorEntry(key {{.KeyType}}) ({{.KeyType}}, {{.ValType}}, bool) {
	return m.Floor(key)
}

// CeilingEntry is an alias of Ceiling — the smallest entry with key >= the given key.
func (m *{{.MapName}}) CeilingEntry(key {{.KeyType}}) ({{.KeyType}}, {{.ValType}}, bool) {
	return m.Ceiling(key)
}

// LowerEntry is an alias of Lower — the largest entry with key strictly < the given key.
func (m *{{.MapName}}) LowerEntry(key {{.KeyType}}) ({{.KeyType}}, {{.ValType}}, bool) {
	return m.Lower(key)
}

// HigherEntry is an alias of Higher — the smallest entry with key strictly > the given key.
func (m *{{.MapName}}) HigherEntry(key {{.KeyType}}) ({{.KeyType}}, {{.ValType}}, bool) {
	return m.Higher(key)
}

// PollFirstEntry removes and returns the smallest entry, or zero/false if empty.
func (m *{{.MapName}}) PollFirstEntry() ({{.KeyType}}, {{.ValType}}, bool) {
	k, v, ok := m.Min()
	if !ok {
		return 0, {{.ValZero}}, false
	}
	m.Remove(k)
	return k, v, true
}

// PollLastEntry removes and returns the largest entry, or zero/false if empty.
func (m *{{.MapName}}) PollLastEntry() ({{.KeyType}}, {{.ValType}}, bool) {
	k, v, ok := m.Max()
	if !ok {
		return 0, {{.ValZero}}, false
	}
	m.Remove(k)
	return k, v, true
}

// DescendingMap returns an iter.Seq2 over entries in descending key order.
func (m *{{.MapName}}) DescendingMap() iter.Seq2[{{.KeyType}}, {{.ValType}}] {
	return func(yield func({{.KeyType}}, {{.ValType}}) bool) {
		var reverse func(node *{{.NodeName}}TreeNode) bool
		reverse = func(node *{{.NodeName}}TreeNode) bool {
			if node == nil {
				return true
			}
			if !reverse(node.right) {
				return false
			}
			if !yield(node.key, node.value) {
				return false
			}
			return reverse(node.left)
		}
		reverse(m.root)
	}
}

// DescendingKeys returns an iter.Seq over keys in descending order.
func (m *{{.MapName}}) DescendingKeys() iter.Seq[{{.KeyType}}] {
	return func(yield func({{.KeyType}}) bool) {
		for k := range m.DescendingMap() {
			if !yield(k) {
				return
			}
		}
	}
}

// ForEach calls the function for each key-value pair in ascending order.
func (m *{{.MapName}}) ForEach(f func({{.KeyType}}, {{.ValType}})) {
	for k, v := range m.All() {
		f(k, v)
	}
}

// Select returns a new map with entries satisfying the predicate.
func (m *{{.MapName}}) Select(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}} {
	result := New{{.MapName}}()
	for k, v := range m.All() {
		if predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// Reject returns a new map with entries NOT satisfying the predicate.
func (m *{{.MapName}}) Reject(predicate func({{.KeyType}}, {{.ValType}}) bool) *{{.MapName}} {
	result := New{{.MapName}}()
	for k, v := range m.All() {
		if !predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// Detect returns the first entry satisfying the predicate (in key order), or (zero, zero, false).
func (m *{{.MapName}}) Detect(predicate func({{.KeyType}}, {{.ValType}}) bool) ({{.KeyType}}, {{.ValType}}, bool) {
	for k, v := range m.All() {
		if predicate(k, v) {
			return k, v, true
		}
	}
	var zk {{.KeyType}}
	var zv {{.ValType}}
	return zk, zv, false
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *{{.MapName}}) AnySatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	for k, v := range m.All() {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *{{.MapName}}) AllSatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	for k, v := range m.All() {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *{{.MapName}}) NoneSatisfy(predicate func({{.KeyType}}, {{.ValType}}) bool) bool {
	for k, v := range m.All() {
		if predicate(k, v) {
			return false
		}
	}
	return true
}

// Count returns the number of entries satisfying the predicate.
func (m *{{.MapName}}) Count(predicate func({{.KeyType}}, {{.ValType}}) bool) int {
	c := 0
	for k, v := range m.All() {
		if predicate(k, v) {
			c++
		}
	}
	return c
}

// String returns a string representation with entries in sorted key order.
func (m *{{.MapName}}) String() string {
	if m.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for k, v := range m.All() {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v: %v", k, v)
		first = false
	}
	sb.WriteString("}")
	return sb.String()
}

// --- Red-black tree internals ---

func (m *{{.MapName}}) findNode(key {{.KeyType}}) *{{.NodeName}}TreeNode {
	node := m.root
	for node != nil {
		if {{if .KeyIsFloat}}{{.CmpFn}}(key, node.key) < 0{{else}}key < node.key{{end}} {
			node = node.left
		} else if {{if .KeyIsFloat}}{{.CmpFn}}(key, node.key) > 0{{else}}key > node.key{{end}} {
			node = node.right
		} else {
			return node
		}
	}
	return nil
}

func (m *{{.MapName}}) minNode(node *{{.NodeName}}TreeNode) *{{.NodeName}}TreeNode {
	for node.left != nil {
		node = node.left
	}
	return node
}

func (m *{{.MapName}}) maxNode(node *{{.NodeName}}TreeNode) *{{.NodeName}}TreeNode {
	for node.right != nil {
		node = node.right
	}
	return node
}

func (m *{{.MapName}}) rotateLeft(x *{{.NodeName}}TreeNode) {
	y := x.right
	x.right = y.left
	if y.left != nil {
		y.left.parent = x
	}
	y.parent = x.parent
	if x.parent == nil {
		m.root = y
	} else if x == x.parent.left {
		x.parent.left = y
	} else {
		x.parent.right = y
	}
	y.left = x
	x.parent = y
}

func (m *{{.MapName}}) rotateRight(x *{{.NodeName}}TreeNode) {
	y := x.left
	x.left = y.right
	if y.right != nil {
		y.right.parent = x
	}
	y.parent = x.parent
	if x.parent == nil {
		m.root = y
	} else if x == x.parent.right {
		x.parent.right = y
	} else {
		x.parent.left = y
	}
	y.right = x
	x.parent = y
}

func (m *{{.MapName}}) fixAfterInsert(z *{{.NodeName}}TreeNode) {
	for z.parent != nil && z.parent.color == {{.NodeName}}TreeNodeRed {
		if z.parent == z.parent.parent.left {
			y := z.parent.parent.right
			if y != nil && y.color == {{.NodeName}}TreeNodeRed {
				z.parent.color = {{.NodeName}}TreeNodeBlack
				y.color = {{.NodeName}}TreeNodeBlack
				z.parent.parent.color = {{.NodeName}}TreeNodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.right {
					z = z.parent
					m.rotateLeft(z)
				}
				z.parent.color = {{.NodeName}}TreeNodeBlack
				z.parent.parent.color = {{.NodeName}}TreeNodeRed
				m.rotateRight(z.parent.parent)
			}
		} else {
			y := z.parent.parent.left
			if y != nil && y.color == {{.NodeName}}TreeNodeRed {
				z.parent.color = {{.NodeName}}TreeNodeBlack
				y.color = {{.NodeName}}TreeNodeBlack
				z.parent.parent.color = {{.NodeName}}TreeNodeRed
				z = z.parent.parent
			} else {
				if z == z.parent.left {
					z = z.parent
					m.rotateRight(z)
				}
				z.parent.color = {{.NodeName}}TreeNodeBlack
				z.parent.parent.color = {{.NodeName}}TreeNodeRed
				m.rotateLeft(z.parent.parent)
			}
		}
	}
	m.root.color = {{.NodeName}}TreeNodeBlack
}

func (m *{{.MapName}}) deleteNode(z *{{.NodeName}}TreeNode) {
	if z.left != nil && z.right != nil {
		succ := m.minNode(z.right)
		z.key = succ.key
		z.value = succ.value
		z = succ
	}
	var child *{{.NodeName}}TreeNode
	if z.left != nil {
		child = z.left
	} else {
		child = z.right
	}
	if child != nil {
		child.parent = z.parent
		if z.parent == nil {
			m.root = child
		} else if z == z.parent.left {
			z.parent.left = child
		} else {
			z.parent.right = child
		}
		if z.color == {{.NodeName}}TreeNodeBlack {
			m.fixAfterDelete(child)
		}
	} else if z.parent == nil {
		m.root = nil
	} else {
		if z.color == {{.NodeName}}TreeNodeBlack {
			m.fixAfterDelete(z)
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

func (m *{{.MapName}}) fixAfterDelete(x *{{.NodeName}}TreeNode) {
	for x != m.root && x.color == {{.NodeName}}TreeNodeBlack {
		if x == x.parent.left {
			w := x.parent.right
			if w == nil {
				x = x.parent
				continue
			}
			if w.color == {{.NodeName}}TreeNodeRed {
				w.color = {{.NodeName}}TreeNodeBlack
				x.parent.color = {{.NodeName}}TreeNodeRed
				m.rotateLeft(x.parent)
				w = x.parent.right
			}
			if w == nil {
				x = x.parent
				continue
			}
			leftBlack := w.left == nil || w.left.color == {{.NodeName}}TreeNodeBlack
			rightBlack := w.right == nil || w.right.color == {{.NodeName}}TreeNodeBlack
			if leftBlack && rightBlack {
				w.color = {{.NodeName}}TreeNodeRed
				x = x.parent
			} else {
				if rightBlack {
					if w.left != nil {
						w.left.color = {{.NodeName}}TreeNodeBlack
					}
					w.color = {{.NodeName}}TreeNodeRed
					m.rotateRight(w)
					w = x.parent.right
				}
				w.color = x.parent.color
				x.parent.color = {{.NodeName}}TreeNodeBlack
				if w.right != nil {
					w.right.color = {{.NodeName}}TreeNodeBlack
				}
				m.rotateLeft(x.parent)
				x = m.root
			}
		} else {
			w := x.parent.left
			if w == nil {
				x = x.parent
				continue
			}
			if w.color == {{.NodeName}}TreeNodeRed {
				w.color = {{.NodeName}}TreeNodeBlack
				x.parent.color = {{.NodeName}}TreeNodeRed
				m.rotateRight(x.parent)
				w = x.parent.left
			}
			if w == nil {
				x = x.parent
				continue
			}
			leftBlack := w.left == nil || w.left.color == {{.NodeName}}TreeNodeBlack
			rightBlack := w.right == nil || w.right.color == {{.NodeName}}TreeNodeBlack
			if leftBlack && rightBlack {
				w.color = {{.NodeName}}TreeNodeRed
				x = x.parent
			} else {
				if leftBlack {
					if w.right != nil {
						w.right.color = {{.NodeName}}TreeNodeBlack
					}
					w.color = {{.NodeName}}TreeNodeRed
					m.rotateLeft(w)
					w = x.parent.left
				}
				w.color = x.parent.color
				x.parent.color = {{.NodeName}}TreeNodeBlack
				if w.left != nil {
					w.left.color = {{.NodeName}}TreeNodeBlack
				}
				m.rotateRight(x.parent)
				x = m.root
			}
		}
	}
	x.color = {{.NodeName}}TreeNodeBlack
}
`
