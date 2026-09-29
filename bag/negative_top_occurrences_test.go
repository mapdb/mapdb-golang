package bag

import "testing"

func TestNegativeTopOccurrencesHasNamedPanic(t *testing.T) {
	for _, tc := range []struct {
		name string
		top  func()
	}{
		{"HashInt32", func() { NewHashInt32().TopOccurrences(-1) }},
		{"TreeInt32", func() { NewTreeInt32().TopOccurrences(-1) }},
		{"HashFloat64", func() { NewHashFloat64().TopOccurrences(-1) }},
		{"TreeFloat64", func() { NewTreeFloat64().TopOccurrences(-1) }},
		{"ImmutableHashInt32", func() { NewImmutableHashInt32().TopOccurrences(-1) }},
		{"SynchronizedHashInt32", func() { NewSynchronizedHashInt32().TopOccurrences(-1) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				const want = "mapdb: TopOccurrences: negative count"
				if got := recover(); got != want {
					t.Errorf("panic = %v, want %q", got, want)
				}
			}()
			tc.top()
		})
	}
}
