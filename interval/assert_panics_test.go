package interval

import "testing"

// assertPanics fails the test unless fn panics. Used by the generated tests to
// verify that out-of-range index access panics like a native Go slice.
func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("expected panic, got none")
		}
	}()
	fn()
}
