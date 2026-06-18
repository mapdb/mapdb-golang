package hashmap

import "testing"

func TestSwissControlMatchers(t *testing.T) {
	ctrl := []uint8{0x01, swissEmpty, 0x7F, swissDeleted, 0x01, 0x42, swissEmpty, 0x01}
	g := swissLoadGroup(ctrl, 0)

	var lanes []int
	for matches := swissMatchByte(g, 0x01); matches != 0; matches &= matches - 1 {
		lanes = append(lanes, swissLowestLane(matches))
	}
	want := []int{0, 4, 7}
	if len(lanes) != len(want) {
		t.Fatalf("match lanes length = %v, want %v", lanes, want)
	}
	for i := range want {
		if lanes[i] != want[i] {
			t.Fatalf("match lanes = %v, want %v", lanes, want)
		}
	}

	if got := swissLowestLane(swissMatchEmpty(g)); got != 1 {
		t.Fatalf("first empty lane = %d, want 1", got)
	}
	full := swissMatchFull(g)
	for _, lane := range []int{0, 2, 4, 5, 7} {
		if full&(uint64(0x80)<<(lane*8)) == 0 {
			t.Fatalf("full mask missing lane %d: %#x", lane, full)
		}
	}
	for _, lane := range []int{1, 3, 6} {
		if full&(uint64(0x80)<<(lane*8)) != 0 {
			t.Fatalf("full mask included non-full lane %d: %#x", lane, full)
		}
	}
}

func TestSwissInt64Int64TombstoneReuseAndChurn(t *testing.T) {
	m := NewInt64Int64WithCapacity(16)
	for i := int64(0); i < 14; i++ {
		m.Put(i*17, i)
	}
	capBefore := len(m.entries)
	for i := int64(0); i < 7; i++ {
		if _, ok := m.Remove(i * 17); !ok {
			t.Fatalf("remove %d failed", i)
		}
	}
	for i := int64(100); i < 107; i++ {
		m.Put(i*17, i)
	}
	if len(m.entries) != capBefore {
		t.Fatalf("tombstone reuse grew capacity: got %d want %d", len(m.entries), capBefore)
	}
	if m.Len() != 14 {
		t.Fatalf("len = %d, want 14", m.Len())
	}

	live := make([]int64, 0, 14)
	for k := range m.All() {
		live = append(live, k)
	}
	for round := 0; round < 2_000; round++ {
		slot := round % len(live)
		key := live[slot]
		m.Remove(key)
		nextKey := key + int64(round+1)*1_000_003
		m.Put(nextKey, int64(round))
		live[slot] = nextKey
	}
	if m.Len() != 14 {
		t.Fatalf("len after churn = %d, want 14", m.Len())
	}
	for k, v := range m.All() {
		got, ok := m.Get(k)
		if !ok || got != v {
			t.Fatalf("live key %d not reachable: got (%d,%v), want %d", k, got, ok, v)
		}
	}
}

func TestSwissInt64Int64Differential(t *testing.T) {
	m := NewInt64Int64WithCapacity(8)
	ref := make(map[int64]int64)
	var x uint64 = 0x123456789abcdef0
	next := func() uint64 {
		x ^= x << 7
		x ^= x >> 9
		x ^= x << 8
		return x
	}

	for i := 0; i < 50_000; i++ {
		k := int64(next() % 4_096)
		switch next() % 3 {
		case 0:
			v := int64(next())
			old, ok := m.Put(k, v)
			refOld, refOK := ref[k]
			if ok != refOK || (ok && old != refOld) {
				t.Fatalf("put old mismatch at %d: got (%d,%v), want (%d,%v)", i, old, ok, refOld, refOK)
			}
			ref[k] = v
		case 1:
			old, ok := m.Remove(k)
			refOld, refOK := ref[k]
			if ok != refOK || (ok && old != refOld) {
				t.Fatalf("remove mismatch at %d: got (%d,%v), want (%d,%v)", i, old, ok, refOld, refOK)
			}
			delete(ref, k)
		default:
			got, ok := m.Get(k)
			want, wantOK := ref[k]
			if ok != wantOK || (ok && got != want) {
				t.Fatalf("get mismatch at %d: got (%d,%v), want (%d,%v)", i, got, ok, want, wantOK)
			}
		}
		if m.Len() != len(ref) {
			t.Fatalf("len mismatch at %d: got %d, want %d", i, m.Len(), len(ref))
		}
	}
	for k, want := range ref {
		got, ok := m.Get(k)
		if !ok || got != want {
			t.Fatalf("final key %d mismatch: got (%d,%v), want %d", k, got, ok, want)
		}
	}
}

var swissProbeSink int64

func benchInt64Int64BuildN(b *testing.B, n int) {
	keys := make([]int64, n)
	for i := 0; i < n; i++ {
		keys[i] = int64(i*1140071481932319849) ^ int64(i>>3)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for iter := 0; iter < b.N; iter++ {
		m := NewInt64Int64WithCapacity(n)
		for _, k := range keys {
			m.Put(k, k+1)
		}
		swissProbeSink += int64(m.Len())
	}
}

func benchInt64Int64HitN(b *testing.B, n int) {
	m := NewInt64Int64WithCapacity(n)
	keys := make([]int64, n)
	for i := 0; i < n; i++ {
		k := int64(i*1140071481932319849) ^ int64(i>>3)
		keys[i] = k
		m.Put(k, k+1)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for iter := 0; iter < b.N; iter++ {
		var sum int64
		for _, k := range keys {
			v, _ := m.Get(k)
			sum += v
		}
		swissProbeSink += sum
	}
}

func benchInt64Int64MissN(b *testing.B, n int) {
	m := NewInt64Int64WithCapacity(n)
	keys := make([]int64, n)
	for i := 0; i < n; i++ {
		k := int64(i*1140071481932319849) ^ int64(i>>3)
		keys[i] = k + 0x4000000000000000
		m.Put(k, k+1)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for iter := 0; iter < b.N; iter++ {
		var sum int64
		for _, k := range keys {
			if v, ok := m.Get(k); ok {
				sum += v
			}
		}
		swissProbeSink += sum
	}
}

func benchInt64Int64RemoveN(b *testing.B, n int) {
	keys := make([]int64, n)
	for i := 0; i < n; i++ {
		keys[i] = int64(i*1140071481932319849) ^ int64(i>>3)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for iter := 0; iter < b.N; iter++ {
		m := NewInt64Int64WithCapacity(n)
		for _, k := range keys {
			m.Put(k, k+1)
		}
		for _, k := range keys {
			v, _ := m.Remove(k)
			swissProbeSink += v
		}
	}
}

func benchInt64Int64IterN(b *testing.B, n int) {
	m := NewInt64Int64WithCapacity(n)
	for i := 0; i < n; i++ {
		k := int64(i*1140071481932319849) ^ int64(i>>3)
		m.Put(k, k+1)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for iter := 0; iter < b.N; iter++ {
		var sum int64
		for k, v := range m.All() {
			sum += k ^ v
		}
		swissProbeSink += sum
	}
}

func BenchmarkSwissProbeBuild1M(b *testing.B)  { benchInt64Int64BuildN(b, 1_000_000) }
func BenchmarkSwissProbeHit1M(b *testing.B)    { benchInt64Int64HitN(b, 1_000_000) }
func BenchmarkSwissProbeMiss1M(b *testing.B)   { benchInt64Int64MissN(b, 1_000_000) }
func BenchmarkSwissProbeRemove1M(b *testing.B) { benchInt64Int64RemoveN(b, 1_000_000) }
func BenchmarkSwissProbeIter1M(b *testing.B)   { benchInt64Int64IterN(b, 1_000_000) }

func BenchmarkSwissProbeBuild8M(b *testing.B)  { benchInt64Int64BuildN(b, 8_000_000) }
func BenchmarkSwissProbeHit8M(b *testing.B)    { benchInt64Int64HitN(b, 8_000_000) }
func BenchmarkSwissProbeMiss8M(b *testing.B)   { benchInt64Int64MissN(b, 8_000_000) }
func BenchmarkSwissProbeRemove8M(b *testing.B) { benchInt64Int64RemoveN(b, 8_000_000) }
func BenchmarkSwissProbeIter8M(b *testing.B)   { benchInt64Int64IterN(b, 8_000_000) }
