package arraylist

import "testing"

func TestNegativeCapacityHasNamedPanic(t *testing.T) {
	for _, tc := range []struct {
		name string
		make func()
		want string
	}{
		{"Int32", func() { NewInt32WithCapacity(-1) }, "mapdb: NewInt32WithCapacity: negative capacity"},
		{"Float64", func() { NewFloat64WithCapacity(-1) }, "mapdb: NewFloat64WithCapacity: negative capacity"},
		{"SynchronizedInt32", func() { NewSynchronizedInt32WithCapacity(-1) }, "mapdb: NewInt32WithCapacity: negative capacity"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != tc.want {
					t.Errorf("panic = %v, want %q", got, tc.want)
				}
			}()
			tc.make()
		})
	}
}
