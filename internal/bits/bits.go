// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package bits holds small, non-generic integer helpers shared by the generated
// collection packages, so a single definition is imported rather than stamped
// into every generated file.
package bits

// NextPowerOfTwo returns the smallest power of two >= n, or 16 for n <= 0 (the
// default hash-table capacity floor). It is the sizing helper used by the
// open-addressed and map-backed hash collections.
func NextPowerOfTwo(n int) int {
	if n <= 0 {
		return 16
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32 // no-op on 32-bit platforms (Go shifts are width-defined), required on 64-bit
	n++
	return n
}
