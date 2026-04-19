// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Real-world examples demonstrating HashingStrategy, Comparator, TreeMap, TreeSet.
// These run as regular tests but double as usage documentation.

package object

import (
	"slices"
	"strings"
	"testing"
)

// ── Example 1: Case-insensitive HTTP headers ──────────────────────────
//
// HTTP header names are case-insensitive per RFC 7230. A plain map would
// treat "Content-Type" and "content-type" as different keys. Using a
// CaseInsensitiveHashingStrategy fixes this.

func TestExample_HTTPHeaders(t *testing.T) {
	headers := NewHashMapWithStrategy[string, string](CaseInsensitiveHashingStrategy())

	headers.Put("Content-Type", "application/json")
	headers.Put("Content-Length", "42")
	headers.Put("Authorization", "Bearer xyz")

	// Case-insensitive lookup
	if ct, _ := headers.Get("content-type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
	if auth, _ := headers.Get("AUTHORIZATION"); auth != "Bearer xyz" {
		t.Fatalf("expected Bearer xyz, got %s", auth)
	}

	// Overwriting with different case
	headers.Put("content-TYPE", "text/html")
	if headers.Size() != 3 {
		t.Fatalf("expected 3 entries, got %d", headers.Size())
	}
	if ct, _ := headers.Get("Content-Type"); ct != "text/html" {
		t.Fatalf("expected overwrite; got %s", ct)
	}
}

// ── Example 2: Deduplicating users by email ───────────────────────────
//
// Real-world scenario: you're processing a stream of user records from
// multiple sources. The same user may appear with different casing in
// email or different metadata. You want unique users by email only.

type User struct {
	Email    string
	Name     string
	Source   string
	LoginCnt int
}

func TestExample_DeduplicateUsers(t *testing.T) {
	// Strategy: two users are "the same" if their emails match case-insensitively.
	ciStr := CaseInsensitiveHashingStrategy()
	emailStrategy := HashingStrategy[User]{
		HashCode: func(u User) uint64 { return ciStr.HashCode(u.Email) },
		Equals:   func(a, b User) bool { return ciStr.Equals(a.Email, b.Email) },
	}

	unique := NewHashSetWithStrategy(emailStrategy)

	// Duplicate-ish records from multiple sources
	unique.Add(User{"alice@example.com", "Alice", "source-a", 5})
	unique.Add(User{"ALICE@example.com", "Alice A.", "source-b", 10})
	unique.Add(User{"bob@example.com", "Bob", "source-a", 3})
	unique.Add(User{"Alice@Example.Com", "Alice", "source-c", 0})

	if unique.Size() != 2 {
		t.Fatalf("expected 2 unique users (alice, bob), got %d", unique.Size())
	}
}

// ── Example 3: Log lines sorted by timestamp, then severity ───────────

type LogLine struct {
	Timestamp int64
	Severity  int // 0=debug, 1=info, 2=warn, 3=error
	Message   string
}

func TestExample_LogSorting(t *testing.T) {
	// Sort: timestamp ascending, then severity descending (errors first
	// within the same timestamp).
	cmp := ThenComparing(
		ComparatorByField(func(l LogLine) int64 { return l.Timestamp }),
		ReverseComparatorByField(func(l LogLine) int { return l.Severity }),
	)

	logs := NewTreeSet[LogLine](cmp)
	logs.Add(LogLine{100, 1, "info first"})
	logs.Add(LogLine{100, 3, "error same time"})
	logs.Add(LogLine{50, 0, "debug earliest"})
	logs.Add(LogLine{200, 2, "warn latest"})

	ordered := logs.ToSlice()

	// Expected order:
	//   t=50  severity=0 "debug earliest"
	//   t=100 severity=3 "error same time"  (higher severity first)
	//   t=100 severity=1 "info first"
	//   t=200 severity=2 "warn latest"
	expectedMsgs := []string{
		"debug earliest",
		"error same time",
		"info first",
		"warn latest",
	}
	got := make([]string, len(ordered))
	for i, l := range ordered {
		got[i] = l.Message
	}
	if !slices.Equal(got, expectedMsgs) {
		t.Fatalf("expected %v, got %v", expectedMsgs, got)
	}
}

// ── Example 4: Leaderboard with TreeMap ───────────────────────────────
//
// Use a TreeMap keyed by score (descending) to get sorted leaderboards.

func TestExample_Leaderboard(t *testing.T) {
	// Key: score, Value: player name.
	// Higher scores first → reverse comparator.
	board := NewTreeMap[int, string](ReverseComparator[int]())

	board.Put(100, "Alice")
	board.Put(250, "Bob")
	board.Put(175, "Charlie")
	board.Put(50, "Dave")

	// Top player
	topScore, topPlayer, _ := board.Min() // Min under reverse = highest score
	if topScore != 250 || topPlayer != "Bob" {
		t.Fatalf("expected Bob@250, got %s@%d", topPlayer, topScore)
	}

	// Iterate in rank order
	var ranked []string
	board.ForEach(func(_ int, name string) {
		ranked = append(ranked, name)
	})
	if !slices.Equal(ranked, []string{"Bob", "Charlie", "Alice", "Dave"}) {
		t.Fatalf("unexpected ranking: %v", ranked)
	}
}

// ── Example 5: Grouping by normalized name ────────────────────────────
//
// Merging data from external systems where names have inconsistent
// whitespace or casing ("New York", "new york", "NEW  YORK" are the same).

func normalizeName(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}

func TestExample_NormalizedGrouping(t *testing.T) {
	// Reuse one strategy so hash seed is stable across calls.
	base := StringHashingStrategy()
	normStrategy := HashingStrategy[string]{
		HashCode: func(s string) uint64 { return base.HashCode(normalizeName(s)) },
		Equals:   func(a, b string) bool { return normalizeName(a) == normalizeName(b) },
	}

	m := NewHashMapWithStrategy[string, int](normStrategy)
	m.Put("New York", 1)
	m.Put("new york", 2)  // merges with above
	m.Put("NEW  YORK", 3) // merges with above
	m.Put("Boston", 10)

	if m.Size() != 2 {
		t.Fatalf("expected 2 distinct cities, got %d", m.Size())
	}
}
