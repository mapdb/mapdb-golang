// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package countmin

import "sort"

// SSEntry is a Space-Saving monitored (item, count, error) triple, the port's
// idiom for the projection serialized as [item, count, error].
type SSEntry struct {
	Item  int32
	Count uint64
	Error uint64
}

// ssState is a monitored entry's (count, error) pair (the item is the map key).
type ssState struct {
	count uint64
	err   uint64
}

// SpaceSaving is a bounded heavy-hitters / top-k summary tracking at most m
// monitored (item, count, error) triples with a deterministic eviction rule.
//
// Unlike the Count-Min Sketch, Space-Saving is order-DEPENDENT (eviction
// depends on which item is the current min when the set is full, which depends
// on add order). For an identical capacity m and an identical add-sequence in
// the same order, the monitored set in canonical order is bit-identical across
// all five ports. No floating point appears in any asserted value.
type SpaceSaving struct {
	capacity  uint32
	monitored map[int32]ssState
}

// NewSpaceSaving constructs an empty summary monitoring at most m items.
//
// Panics if m == 0 (a zero-capacity summary can monitor nothing; every add
// would have to evict from an empty set).
func NewSpaceSaving(m uint32) *SpaceSaving {
	if m == 0 {
		panic("SpaceSaving capacity m must be non-zero")
	}
	return &SpaceSaving{
		capacity:  m,
		monitored: make(map[int32]ssState),
	}
}

// Add adds item with weight count.
//
//   - count == 0 is a no-op (no admit, increment, or eviction).
//   - If item is already monitored: its count grows (saturating); its error is
//     unchanged.
//   - If there is room (size < m): admit with error = 0.
//   - If full: evict the (count, signed item)-min victim; the new item takes
//     count = evictedCount + count (saturating) and error = evictedCount.
func (s *SpaceSaving) Add(item int32, count uint64) {
	if count == 0 {
		return // zero-weight add changes nothing.
	}
	if e, ok := s.monitored[item]; ok {
		e.count = saturatingAddU64(e.count, count)
		// error unchanged for an already-monitored item.
		s.monitored[item] = e
		return
	}
	if uint32(len(s.monitored)) < s.capacity {
		s.monitored[item] = ssState{count: count, err: 0}
		return
	}
	// Full + unmonitored item: evict the (count, signed item)-min victim.
	victim := s.argminVictim()
	evictedCount := s.monitored[victim].count
	delete(s.monitored, victim)
	s.monitored[item] = ssState{
		count: saturatingAddU64(evictedCount, count),
		err:   evictedCount,
	}
}

// AddOne is a convenience for Add(item, 1).
func (s *SpaceSaving) AddOne(item int32) {
	s.Add(item, 1)
}

// argminVictim returns the monitored item minimizing (count, signed item):
// smallest count, then smallest signed i32 item on a count tie. Items are
// distinct, so the victim is unique. Caller guarantees the set is non-empty.
func (s *SpaceSaving) argminVictim() int32 {
	var victim int32
	var best ssState
	first := true
	for item, e := range s.monitored {
		if first || e.count < best.count || (e.count == best.count && item < victim) {
			victim = item
			best = e
			first = false
		}
	}
	return victim
}

// Count returns the monitored count for item, or 0 if not monitored.
func (s *SpaceSaving) Count(item int32) uint64 {
	if e, ok := s.monitored[item]; ok {
		return e.count
	}
	return 0
}

// Error returns the monitored error for item, or 0 if not monitored.
func (s *SpaceSaving) Error(item int32) uint64 {
	if e, ok := s.monitored[item]; ok {
		return e.err
	}
	return 0
}

// IsMonitored reports whether item is currently monitored.
func (s *SpaceSaving) IsMonitored(item int32) bool {
	_, ok := s.monitored[item]
	return ok
}

// Size returns the number of currently monitored items (<= m).
func (s *SpaceSaving) Size() uint32 {
	return uint32(len(s.monitored))
}

// Capacity returns the capacity m.
func (s *SpaceSaving) Capacity() uint32 {
	return s.capacity
}

// MonitoredSet returns the entire monitored set as (item, count, error) triples
// in canonical order: count DESCENDING, then signed item ASCENDING.
func (s *SpaceSaving) MonitoredSet() []SSEntry {
	out := make([]SSEntry, 0, len(s.monitored))
	for item, e := range s.monitored {
		out = append(out, SSEntry{Item: item, Count: e.count, Error: e.err})
	}
	// count DESC, then signed item ASC.
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Item < out[j].Item
	})
	return out
}

// TopK returns the k highest-count monitored items in canonical order (the
// first k of MonitoredSet). k > size returns all monitored items (no padding);
// k == 0 returns the empty list.
func (s *SpaceSaving) TopK(k uint32) []SSEntry {
	all := s.MonitoredSet()
	if uint64(k) < uint64(len(all)) {
		all = all[:k]
	}
	return all
}
