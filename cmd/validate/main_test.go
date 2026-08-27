// Copyright (c) 2026 Jan Kotek.
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.

package main

import (
	"encoding/json"
	"testing"
)

func TestRenderExpectedFenwickTreeUnquotes(t *testing.T) {
	got := renderExpected(json.RawMessage(`["5","5","0","7","0","0","0","16"]`), "tree", modeNone)
	if got != "[5,5,0,7,0,0,0,16]" {
		t.Fatalf("tree: got %q", got)
	}
}

func TestRenderExpectedBoolArray(t *testing.T) {
	got := renderExpected(json.RawMessage(`[true]`), "contains_results", modeNone)
	if got != "[true]" {
		t.Fatalf("contains_results: got %q", got)
	}
}

func TestRenderExpectedNestedSnapshotEntries(t *testing.T) {
	got := renderExpected(json.RawMessage(`[[[2,20],[3,30],[1,10]]]`), "snapshot_entries_log", modeNone)
	if got != "[[[2,20],[3,30],[1,10]]]" {
		t.Fatalf("snapshot_entries_log: got %q", got)
	}
}

func TestRenderExpectedEvictionLog(t *testing.T) {
	got := renderExpected(json.RawMessage(`[[1,10,"size"]]`), "eviction_log", modeNone)
	if got != `[[1,10,"size"]]` {
		t.Fatalf("eviction_log: got %q", got)
	}
}
