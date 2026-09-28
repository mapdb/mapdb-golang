// Copyright (c) 2026 Jan Kotek.
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.

package main

import (
	"encoding/json"
	"testing"
)

func TestPanicPassedJudge(t *testing.T) {
	m1 := reachMarkerLine(1, 1) + "\n"
	m2 := reachMarkerLine(2, 2) + "\n"
	cases := []struct {
		name     string
		exit     int
		stdout   string
		timedOut bool
		ops      int
		want     bool
	}{
		{"exit 0 empty", 0, m1, false, 1, false},
		{"exit 0 banner", 0, m1 + "=== scenario: x ===\n", false, 1, false},
		{"exit 1 size", 1, m1 + "size: 1\n", false, 1, false},
		{"exit 1 marker only", 1, m1, false, 1, true},
		{"exit 101 boom", 101, m1 + "boom\n", false, 1, true},
		{"timed out", 1, m1, true, 1, false},
		{"fail line", 1, m1 + "FAIL name expect_panic\n", false, 1, true},
		{"expect_panic line", 1, m1 + "expect_panic: true\n", false, 1, false},
		{"summary line", 1, m1 + "SUMMARY: 1\n", false, 1, true},
		{"colon without space", 1, m1 + "boom:detail\n", false, 1, true},
		{"fail-count key", 1, m1 + "FAIL-count: 1\n", false, 1, false},
		// astra25/25 F4: a crash before the product leaves no reach marker.
		{"crash before product", 1, "", false, 1, false},
		{"crash before product with noise", 1, "boom\n", false, 1, false},
		{"trapped on op 1 of 2", 1, reachMarkerLine(1, 2) + "\n", false, 2, false},
		{"reached op 2 of 2", 1, reachMarkerLine(1, 2) + "\n" + m2, false, 2, true},
		{"no ops", 1, m1, false, 0, false},
		{"marker with trailing space", 1, "[panic-child] reached op 1/1 \n", false, 1, false},
		{"marker with CR", 1, "[panic-child] reached op 1/1\r\n", false, 1, true},
		{"last call returned", 1, m1 + returnMarkerLine(1, 1) + "\n", false, 1, false},
		{"op 1 returned, op 2 trapped", 1, reachMarkerLine(1, 2) + "\n" + returnMarkerLine(1, 2) + "\n" + m2, false, 2, true},
	}
	for _, c := range cases {
		if got := panicPassed(c.exit, c.stdout, c.timedOut, c.ops); got != c.want {
			t.Fatalf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

func TestReachMarkerIsNotASentinel(t *testing.T) {
	if stdoutHasSentinel(reachMarkerLine(1, 1) + "\n" + returnMarkerLine(1, 1) + "\n") {
		t.Fatal("the reach/return markers must not be assertion sentinels")
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
