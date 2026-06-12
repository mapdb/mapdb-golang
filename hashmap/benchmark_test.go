package hashmap

import (
	"testing"
)

func BenchmarkInt32Int64_Put(b *testing.B) {
	m := NewInt32Int64WithCapacity(b.N * 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Put(int32(i), int64(i*10))
	}
}

func BenchmarkInt32Int64_Get(b *testing.B) {
	m := NewInt32Int64()
	for i := int32(0); i < 10000; i++ {
		m.Put(i, int64(i*10))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(int32(i % 10000))
	}
}

func BenchmarkGoBuiltinMap_Put(b *testing.B) {
	m := make(map[int32]int64, b.N*2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m[int32(i)] = int64(i * 10)
	}
}

func BenchmarkGoBuiltinMap_Get(b *testing.B) {
	m := make(map[int32]int64)
	for i := int32(0); i < 10000; i++ {
		m[i] = int64(i * 10)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m[int32(i%10000)]
	}
}

func BenchmarkInt32Int64_Iterate(b *testing.B) {
	m := NewInt32Int64()
	for i := int32(0); i < 10000; i++ {
		m.Put(i, int64(i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for range m.All() {
		}
	}
}

func BenchmarkGoBuiltinMap_Iterate(b *testing.B) {
	m := make(map[int32]int64)
	for i := int32(0); i < 10000; i++ {
		m[i] = int64(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for range m {
		}
	}
}
