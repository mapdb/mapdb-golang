// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package boundedlru implements the bounded LRU map (max-size v1) described in
// spec/features/bounded-lru.md.
//
// BoundedLruInt32Int32Map is a fixed-capacity map from int32 to int32 that
// evicts its least-recently-used entry when a new-key insert would exceed the
// capacity. Recency order is kept by an intrusive doubly-linked LRU list over
// an arena + slot-index (mirroring the Rust reference src/bounded_lru.rs):
// each live entry owns a slot in a contiguous node arena; nodes carry
// prev/next slot INDICES (never pointers) plus a back-reference key, and freed
// slots recycle through a free-list. The list head is the LRU end (the
// eviction victim); the tail is the MRU end. A recency refresh is an O(1)
// unlink + push-to-tail; eviction is an O(1) pop-from-head.
//
// Recency is position-implicit (head = least-recently-used): there is no
// stored last_use stamp. v1 has NO wall clock — all time is the
// caller-supplied logical tick. TTL is an after-write
// expire_at = saturating(now + ttl); ExpireEntries(now) removes every entry
// with expire_at <= now (inclusive), firing the callback with cause EXPIRED in
// ascending-expire_at then ascending-last_use (LRU) order. Plain Put(k, v) is
// defined as PutAt(k, v, 0).
package boundedlru

import "sort"

// EvictionCause is why an entry left the map (the eviction-callback cause).
// Only CauseSize and CauseExpired exist in v1 — put-update, Remove, and Clear
// are NOT evictions and never invoke the callback.
type EvictionCause int

const (
	// CauseSize: evicted because an insert exceeded maxSize (the LRU victim).
	CauseSize EvictionCause = iota
	// CauseExpired: removed by ExpireEntries(now) because its logical expiry
	// tick passed.
	CauseExpired
)

// String returns the lower-case serialized name used by the cross-language
// suite ("size" / "expired").
func (c EvictionCause) String() string {
	switch c {
	case CauseSize:
		return "size"
	case CauseExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// nilIdx is the sentinel slot index meaning "no node" (list end / free-list
// end).
const nilIdx = -1

// neverExpire is the saturated "+inf / never" expiry sentinel: an entry whose
// expire_at equals this never expires, even at the maximum now.
const neverExpire = ^uint64(0) // math.MaxUint64

// node is one arena slot: an intrusive doubly-linked-list node. When the slot
// is live it links into the LRU list (prev/next are slot indices, key is the
// back-reference into the index). When the slot is free it sits on the
// free-list (next chains the free-list; prev/key/value/expireAt are dead).
type node struct {
	prev     int
	next     int
	key      int32
	value    int32
	expireAt uint64
}

// BoundedLruInt32Int32Map is a fixed-capacity LRU map from int32 to int32.
//
// Construct it with NewBoundedLruInt32Int32Map or the builder
// (BuilderBoundedLruInt32Int32Map). The map holds at most maxSize entries; a
// new-key insert that would exceed it evicts the least-recently-used entry
// first (evict-before-insert), so the inserted key is never its own victim.
type BoundedLruInt32Int32Map struct {
	// index maps key -> arena slot index. The arena slot holds the value + LRU
	// links.
	index map[int32]int
	// arena is the slot-index-addressed pool of LRU-list nodes.
	arena []node
	// freeHead is the free-list head (a slot index), or nilIdx when no free
	// slot is available.
	freeHead int
	// head is the LRU-list head = least-recently-used (the eviction victim),
	// or nilIdx.
	head int
	// tail is the LRU-list tail = most-recently-used, or nilIdx.
	tail int
	// maxSize is the capacity n (0 => the map is permanently empty; every
	// insert drops).
	maxSize int
	// ttl is the after-write TTL in logical ticks; hasTTL distinguishes "TTL 0"
	// from "no TTL configured".
	ttl    uint64
	hasTTL bool
	// onEvict is the optional eviction callback (key, value-at-eviction, cause).
	onEvict func(key int32, value int32, cause EvictionCause)
	// inCallback is true while onEvict is running. Re-entering any mutating
	// method from the callback corrupts the arena (e.g. ExpireEntries holds
	// collected victim indices that a reentrant mutation invalidates), so the
	// guarded methods panic instead of silently corrupting.
	inCallback bool
}

// assertNotInCallback panics if called while the eviction callback is running.
// Mutating (or recency-refreshing) the map from inside onEvict is forbidden.
func (m *BoundedLruInt32Int32Map) assertNotInCallback() {
	if m.inCallback {
		panic("boundedlru: reentrant call from eviction callback")
	}
}

// BuilderBoundedLruInt32Int32Map builds a BoundedLruInt32Int32Map. maxSize is
// required; TTL and on-evict are optional.
type BuilderBoundedLruInt32Int32Map struct {
	maxSize int
	ttl     uint64
	hasTTL  bool
	onEvict func(key int32, value int32, cause EvictionCause)
}

// NewBoundedLruInt32Int32Map returns a pure max-size LRU map of capacity n (no
// TTL, no callback).
func NewBoundedLruInt32Int32Map(maxSize int) *BoundedLruInt32Int32Map {
	return NewBuilderBoundedLruInt32Int32Map().MaxSize(maxSize).Build()
}

// NewBuilderBoundedLruInt32Int32Map starts building a map. maxSize defaults to
// 0 (drop-everything) until set; call MaxSize(n) to fix the capacity.
func NewBuilderBoundedLruInt32Int32Map() *BuilderBoundedLruInt32Int32Map {
	return &BuilderBoundedLruInt32Int32Map{}
}

// MaxSize sets the capacity n (the maximum number of resident entries).
func (b *BuilderBoundedLruInt32Int32Map) MaxSize(n int) *BuilderBoundedLruInt32Int32Map {
	b.maxSize = n
	return b
}

// TTL sets the after-write TTL (logical ticks). expire_at = saturating(now+ttl).
func (b *BuilderBoundedLruInt32Int32Map) TTL(ttl uint64) *BuilderBoundedLruInt32Int32Map {
	b.ttl = ttl
	b.hasTTL = true
	return b
}

// OnEvict installs the eviction callback. It is invoked once per evicted entry
// with (key, value-at-eviction, cause), synchronously, for causes CauseSize
// and CauseExpired only. It MUST NOT mutate the map.
func (b *BuilderBoundedLruInt32Int32Map) OnEvict(cb func(key int32, value int32, cause EvictionCause)) *BuilderBoundedLruInt32Int32Map {
	b.onEvict = cb
	return b
}

// Build constructs the map.
func (b *BuilderBoundedLruInt32Int32Map) Build() *BoundedLruInt32Int32Map {
	// A negative capacity is meaningless; clamp it to 0 (drop-everything), the
	// natural floor. The Rust reference uses an unsigned `usize`, so 0 is the
	// smallest representable capacity there; this keeps the Go API total
	// (never panics on a bad capacity) and behaviorally faithful.
	maxSize := b.maxSize
	if maxSize < 0 {
		maxSize = 0
	}
	return &BoundedLruInt32Int32Map{
		index:    make(map[int32]int),
		arena:    nil,
		freeHead: nilIdx,
		head:     nilIdx,
		tail:     nilIdx,
		maxSize:  maxSize,
		ttl:      b.ttl,
		hasTTL:   b.hasTTL,
		onEvict:  b.onEvict,
	}
}

// Size returns the current entry count (0 ..= maxSize).
func (m *BoundedLruInt32Int32Map) Size() int { return len(m.index) }

// Len is an alias for Size (idiomatic-Go length).
func (m *BoundedLruInt32Int32Map) Len() int { return len(m.index) }

// IsEmpty reports whether the map holds no entries.
func (m *BoundedLruInt32Int32Map) IsEmpty() bool { return len(m.index) == 0 }

// Capacity returns the configured capacity n.
func (m *BoundedLruInt32Int32Map) Capacity() int { return m.maxSize }

// --- arena / intrusive-list primitives (non-observable) -------------------

// allocNode allocates a slot for a fresh entry (reusing a free slot if
// available), returning its index. The node is NOT yet linked into the LRU
// list.
func (m *BoundedLruInt32Int32Map) allocNode(key int32, value int32, expireAt uint64) int {
	if m.freeHead != nilIdx {
		idx := m.freeHead
		m.freeHead = m.arena[idx].next
		n := &m.arena[idx]
		n.prev = nilIdx
		n.next = nilIdx
		n.key = key
		n.value = value
		n.expireAt = expireAt
		return idx
	}
	idx := len(m.arena)
	m.arena = append(m.arena, node{
		prev:     nilIdx,
		next:     nilIdx,
		key:      key,
		value:    value,
		expireAt: expireAt,
	})
	return idx
}

// freeNode returns a slot to the free-list (the node must already be unlinked
// from the LRU list and removed from the index).
func (m *BoundedLruInt32Int32Map) freeNode(idx int) {
	m.arena[idx].next = m.freeHead
	m.arena[idx].prev = nilIdx
	m.freeHead = idx
}

// unlink removes a node from the LRU list (O(1)); leaves the slot allocated.
func (m *BoundedLruInt32Int32Map) unlink(idx int) {
	prev := m.arena[idx].prev
	next := m.arena[idx].next
	if prev != nilIdx {
		m.arena[prev].next = next
	} else {
		m.head = next
	}
	if next != nilIdx {
		m.arena[next].prev = prev
	} else {
		m.tail = prev
	}
	m.arena[idx].prev = nilIdx
	m.arena[idx].next = nilIdx
}

// pushTail pushes a (currently unlinked) node onto the MRU end (tail).
func (m *BoundedLruInt32Int32Map) pushTail(idx int) {
	oldTail := m.tail
	m.arena[idx].prev = oldTail
	m.arena[idx].next = nilIdx
	if oldTail != nilIdx {
		m.arena[oldTail].next = idx
	} else {
		m.head = idx
	}
	m.tail = idx
}

// touch moves an existing live node to the MRU end (a recency refresh).
func (m *BoundedLruInt32Int32Map) touch(idx int) {
	if m.tail == idx {
		return // already MRU
	}
	m.unlink(idx)
	m.pushTail(idx)
}

// --- map surface ----------------------------------------------------------

// Put is PutAt(k, v, 0) (no hidden clock). On a no-TTL map the now is
// irrelevant; on a TTL map this writes with now = 0. Returns the previous
// value and whether one was present.
func (m *BoundedLruInt32Int32Map) Put(key int32, value int32) (int32, bool) {
	return m.PutAt(key, value, 0)
}

// PutAt inserts-or-updates with a logical write tick. It refreshes the recency
// of key; a new-key insert at capacity evicts the LRU entry first
// (evict-before-insert). Returns the previous value and whether one existed.
func (m *BoundedLruInt32Int32Map) PutAt(key int32, value int32, now uint64) (int32, bool) {
	m.assertNotInCallback()
	expireAt := neverExpire
	if m.hasTTL {
		expireAt = saturatingAdd(now, m.ttl)
	}

	if idx, ok := m.index[key]; ok {
		// Update: value replaced, expiry reset, recency refreshed; NO evict.
		old := m.arena[idx].value
		m.arena[idx].value = value
		m.arena[idx].expireAt = expireAt
		m.touch(idx)
		return old, true
	}

	// Genuine insertion of a new key.
	if m.maxSize == 0 {
		// Capacity 0: the entry is dropped, never resident, no callback.
		return 0, false
	}

	// Evict-before-insert: the invariant len <= maxSize holds on every op, so a
	// new-key insert raises size by one and needs AT MOST ONE size eviction
	// when maxSize >= 1. The `if` (not a loop) makes that one-eviction contract
	// explicit.
	if len(m.index) >= m.maxSize {
		victim := m.head // LRU end; always valid since len >= 1.
		m.evictNode(victim, CauseSize)
	}

	idx := m.allocNode(key, value, expireAt)
	m.pushTail(idx)
	m.index[key] = idx
	return 0, false
}

// Get looks up key. On a hit it refreshes recency; on a miss it does nothing.
func (m *BoundedLruInt32Int32Map) Get(key int32) (int32, bool) {
	m.assertNotInCallback() // a hit refreshes recency, mutating LRU order
	if idx, ok := m.index[key]; ok {
		v := m.arena[idx].value
		m.touch(idx)
		return v, true
	}
	return 0, false
}

// GetOrDefault returns the value for key, or defaultValue on a miss. A hit
// refreshes recency exactly like Get; a miss does NOT refresh recency and does
// NOT insert defaultValue.
func (m *BoundedLruInt32Int32Map) GetOrDefault(key int32, defaultValue int32) int32 {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// ContainsKey is a membership test. It does NOT refresh recency and never
// evicts.
func (m *BoundedLruInt32Int32Map) ContainsKey(key int32) bool {
	_, ok := m.index[key]
	return ok
}

// Remove deletes key. It does not evict and does NOT invoke the eviction
// callback (manual removal is not an eviction). Returns the removed value and
// whether it was present.
func (m *BoundedLruInt32Int32Map) Remove(key int32) (int32, bool) {
	m.assertNotInCallback()
	idx, ok := m.index[key]
	if !ok {
		return 0, false
	}
	v := m.arena[idx].value
	delete(m.index, key)
	m.unlink(idx)
	m.freeNode(idx)
	return v, true
}

// Clear removes all entries. It does NOT invoke the eviction callback for the
// cleared entries (bulk manual removal is not eviction).
func (m *BoundedLruInt32Int32Map) Clear() {
	m.assertNotInCallback()
	m.index = make(map[int32]int)
	m.arena = m.arena[:0]
	m.freeHead = nilIdx
	m.head = nilIdx
	m.tail = nilIdx
}

// evictNode removes a victim node entirely (unlink + free + index-remove) and
// fires the eviction callback with the given cause and the value-at-eviction.
func (m *BoundedLruInt32Int32Map) evictNode(idx int, cause EvictionCause) {
	key := m.arena[idx].key
	value := m.arena[idx].value
	delete(m.index, key)
	m.unlink(idx)
	m.freeNode(idx)
	if m.onEvict != nil {
		m.inCallback = true
		// Clear the flag even if the callback panics, so a recovered panic does
		// not leave the map permanently wedged.
		func() {
			defer func() { m.inCallback = false }()
			m.onEvict(key, value, cause)
		}()
	}
}

// ExpireEntries is the logical-time expiry pass: it removes every entry with
// expire_at <= now (inclusive), firing the callback with cause CauseExpired in
// ascending expire_at, then ascending last_use (LRU) order. It returns the
// count removed. This is the only time-driven eviction; surviving entries'
// recency is unchanged. A no-TTL map never expires anything.
func (m *BoundedLruInt32Int32Map) ExpireEntries(now uint64) int {
	m.assertNotInCallback()
	if !m.hasTTL {
		return 0
	}
	// Collect victims: every live node with expire_at <= now. Walk the LRU list
	// head->tail (ascending last_use) and STABLE-sort by expire_at so equal
	// expire_at entries keep that ascending-LRU order.
	type victim struct {
		expireAt uint64
		idx      int
	}
	var victims []victim
	for cur := m.head; cur != nilIdx; {
		n := &m.arena[cur]
		next := n.next
		// neverExpire is the saturated "+inf" sentinel: such an entry never
		// expires, even at now == math.MaxUint64.
		if n.expireAt != neverExpire && n.expireAt <= now {
			victims = append(victims, victim{n.expireAt, cur})
		}
		cur = next
	}
	sort.SliceStable(victims, func(i, j int) bool {
		return victims[i].expireAt < victims[j].expireAt
	})
	for _, v := range victims {
		m.evictNode(v.idx, CauseExpired)
	}
	return len(victims)
}

// --- iteration (LRU order, read-only snapshots) ---------------------------

// Keys returns all keys in LRU order (least-recently-used first). It is a
// read-only snapshot: it does NOT refresh recency and never evicts.
func (m *BoundedLruInt32Int32Map) Keys() []int32 {
	out := make([]int32, 0, len(m.index))
	for cur := m.head; cur != nilIdx; cur = m.arena[cur].next {
		out = append(out, m.arena[cur].key)
	}
	return out
}

// Values returns all values in LRU order, parallel to Keys. Read-only snapshot.
func (m *BoundedLruInt32Int32Map) Values() []int32 {
	out := make([]int32, 0, len(m.index))
	for cur := m.head; cur != nilIdx; cur = m.arena[cur].next {
		out = append(out, m.arena[cur].value)
	}
	return out
}

// Entry is a (key, value) pair returned by Entries in LRU order.
type Entry struct {
	Key   int32
	Value int32
}

// Entries returns all (key, value) entries in LRU order. Read-only snapshot.
func (m *BoundedLruInt32Int32Map) Entries() []Entry {
	out := make([]Entry, 0, len(m.index))
	for cur := m.head; cur != nilIdx; cur = m.arena[cur].next {
		n := &m.arena[cur]
		out = append(out, Entry{Key: n.key, Value: n.value})
	}
	return out
}

// saturatingAdd returns a+b clamped to math.MaxUint64 on overflow.
func saturatingAdd(a, b uint64) uint64 {
	s := a + b
	if s < a {
		return neverExpire
	}
	return s
}
