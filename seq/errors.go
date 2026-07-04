// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package seq

import (
	"fmt"
	"iter"
)

// The core seq algebra is infallible: in-memory collections cannot fail. Fallible
// sources (deserializers, file/network-backed pumps) live at the edges and reach
// the algebra through the bridges here. Two shapes are supported because the Go
// ecosystem is split and neither is canonical:
//
//   - per-element iter.Seq2[T, error]  — each element individually carries an
//     error; meaningful when skip-and-continue makes sense (SkipErr).
//   - the error-func pair (Seq[T], func() error) — the bufio.Scanner shape: one
//     check after the loop, setup/teardown errors representable. This is the
//     recommended shape for the framework's own IO-backed sources (StopOnErr,
//     Checked).
//
// Rule of thumb: error-func for single-use IO sources (an error terminates the
// stream); Seq2[T, error] only when elements individually carry errors and
// continuing past one is meaningful.

// StopOnErr adapts a per-element fallible sequence into the error-func shape: the
// returned Seq yields values until the first error, then stops; the returned
// func reports that error (nil if none) and is valid AFTER the Seq has been
// ranged. Single-shot in its error reporting: the error reflects the most recent
// iteration, so range once, then check — do not read the func concurrently with
// ranging.
func StopOnErr[T any](s iter.Seq2[T, error]) (Seq[T], func() error) {
	var err error
	out := func(yield func(T) bool) {
		err = nil // reset so a re-range starts clean
		for v, e := range s {
			if e != nil {
				err = e
				return
			}
			if !yield(v) {
				return
			}
		}
	}
	return out, func() error { return err }
}

// SkipErr adapts a per-element fallible sequence into an infallible Seq by
// dropping the errored elements, calling onErr (if non-nil) on each error so the
// caller can log or count it. Lazy, O(1) memory, and re-runnable when s is (it
// holds no cross-call state).
func SkipErr[T any](s iter.Seq2[T, error], onErr func(error)) Seq[T] {
	return func(yield func(T) bool) {
		for v, e := range s {
			if e != nil {
				if onErr != nil {
					onErr(e)
				}
				continue
			}
			if !yield(v) {
				return
			}
		}
	}
}

// MustAll adapts a per-element fallible sequence into an infallible Seq that
// panics on the first error. For tests and known-good data. Lazy, O(1) memory,
// re-runnable when s is.
func MustAll[T any](s iter.Seq2[T, error]) Seq[T] {
	return func(yield func(T) bool) {
		for v, e := range s {
			if e != nil {
				panic(fmt.Sprintf("seq: MustAll: %v", e))
			}
			if !yield(v) {
				return
			}
		}
	}
}

// WithErr lifts an infallible Seq into the per-element Seq2[T, error] shape,
// pairing every element with a nil error. For APIs that demand that shape. Lazy,
// O(1) memory, re-runnable when s is.
func WithErr[T any](s Seq[T]) Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for v := range s {
			if !yield(v, nil) {
				return
			}
		}
	}
}

// Checked is the explicit owner returned by resource-owning source constructors
// (files, decoders): the element stream plus an error accessor valid once
// iteration completes and an optional source-owned cleanup. A Checked Seq is
// single-shot unless materialized or passed through Cache. Constructors, not the
// seq algebra, produce these — the algebra stays infallible.
type Checked[T any] struct {
	// Seq is the element stream. Range it to completion, then consult Err.
	Seq iter.Seq[T]
	// Err reports the terminating error (nil on clean completion). Valid only
	// after Seq has been fully ranged.
	Err func() error
	// Close releases source-owned resources; may be nil. Safe to call after
	// iteration; idempotency is the constructor's contract.
	Close func() error
}
