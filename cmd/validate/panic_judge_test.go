// Copyright (c) 2026 Jan Kotek.
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.

package main

import (
	"encoding/json"
	"testing"
)

func TestPanicPassedJudge(t *testing.T) {
	cases := []struct {
		name     string
		exit     int
		stdout   string
		timedOut bool
		want     bool
	}{
		{"exit 0 empty", 0, "", false, false},
		{"exit 0 banner", 0, "=== scenario: x ===\n", false, false},
		{"exit 1 size", 1, "size: 1\n", false, false},
		{"exit 1 empty", 1, "", false, true},
		{"exit 101 boom", 101, "boom\n", false, true},
		{"timed out", 1, "", true, false},
		{"fail line", 1, "FAIL name expect_panic\n", false, true},
		{"expect_panic line", 1, "expect_panic: true\n", false, false},
		{"summary line", 1, "SUMMARY: 1\n", false, true},
		{"colon without space", 1, "boom:detail\n", false, true},
		{"fail-count key", 1, "FAIL-count: 1\n", false, false},
	}
	for _, c := range cases {
		if got := panicPassed(c.exit, c.stdout, c.timedOut); got != c.want {
			t.Fatalf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

func TestStdoutHasSentinelStatusLines(t *testing.T) {
	if stdoutHasSentinel("ERROR: x\nSKIP: y\nSUMMARY: 1\nPASS\nFAIL name\n") {
		t.Fatal("status lines are not sentinels")
	}
	if !stdoutHasSentinel("size: 1\r\n") {
		t.Fatal("size line with CR")
	}
	if !stdoutHasSentinel("=== scenario:") {
		t.Fatal("banner prefix")
	}
	if stdoutHasSentinel("boom\n") {
		t.Fatal("boom is not a sentinel")
	}
}

func TestAssertionBoolTrue(t *testing.T) {
	if !assertionBoolTrue(json.RawMessage("true")) {
		t.Fatal("true")
	}
	for _, raw := range []string{"false", "1", "null", `"true"`} {
		if assertionBoolTrue(json.RawMessage(raw)) {
			t.Fatalf("accepted %s", raw)
		}
	}
}
