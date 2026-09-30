package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout runs f and returns what it printed to os.Stdout.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	return captureFile(t, &os.Stdout, f)
}

// captureFile runs f with *target redirected and returns what was written.
func captureFile(t *testing.T, target **os.File, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	saved := *target
	*target = w
	defer func() { *target = saved }()
	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	f()
	w.Close()
	return <-done
}

// runSketch decodes a scenario the way loadScenario does (UseNumber) and runs
// it through run, restoring anyFail afterwards.
func runSketch(t *testing.T, raw string, run func(scenario)) string {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	var s scenario
	if err := dec.Decode(&s); err != nil {
		t.Fatal(err)
	}
	savedFail := anyFail
	defer func() { anyFail = savedFail }()
	return captureStdout(t, func() { run(s) })
}

// Wide u32 sketch parameters must SKIP, never wrap (4294967312 -> 16).
func TestSketchParamsRejectWideValuesBeforeNarrowing(t *testing.T) {
	for _, c := range []struct {
		name, raw, key string
		run            func(scenario)
		emits          bool
	}{
		{"cms control", `{"name":"c","operations":[{"op":"with_params","d":4,"w":16}],"assertions":{"width":16}}`, "width:", runCountMin, true},
		{"cms wide w", `{"name":"c","operations":[{"op":"with_params","d":4,"w":4294967312}],"assertions":{"width":16}}`, "width:", runCountMin, false},
		{"cms wide d", `{"name":"c","operations":[{"op":"with_params","d":4294967300,"w":16}],"assertions":{"depth":4}}`, "depth:", runCountMin, false},
		{"cms beyond int64", `{"name":"c","operations":[{"op":"with_params","d":4,"w":18446744073709551632}],"assertions":{"width":16}}`, "width:", runCountMin, false},
		{"ss control", `{"name":"s","operations":[{"op":"with_capacity","m":10}],"assertions":{"capacity":10}}`, "capacity:", runSpaceSaving, true},
		{"ss wide m", `{"name":"s","operations":[{"op":"with_capacity","m":4294967306}],"assertions":{"capacity":10}}`, "capacity:", runSpaceSaving, false},
		{"positions control", `{"name":"p","operations":[{"op":"positions","value":7,"m":16,"k":4}],"assertions":{"positions":[0]}}`, "positions:", runHashPipeline, true},
		{"positions wide m", `{"name":"p","operations":[{"op":"positions","value":7,"m":4294967312,"k":4}],"assertions":{"positions":[0]}}`, "positions:", runHashPipeline, false},
		{"positions wide k", `{"name":"p","operations":[{"op":"positions","value":7,"m":16,"k":4294967297}],"assertions":{"positions":[0]}}`, "positions:", runHashPipeline, false},
	} {
		out := runSketch(t, c.raw, c.run)
		if got := strings.Contains(out, c.key); got != c.emits {
			t.Errorf("%s: emitted %s = %v, want %v; output:\n%s", c.name, c.key, got, c.emits, out)
		}
	}
}

func TestExactU32(t *testing.T) {
	for _, v := range []any{json.Number("0"), json.Number("4294967295"), float64(16)} {
		if _, ok := exactU32(v); !ok {
			t.Errorf("in-domain %v rejected", v)
		}
	}
	for _, v := range []any{json.Number("-1"), json.Number("4294967296"), json.Number("18446744073709551620"), float64(-1), float64(4294967312)} {
		if _, ok := exactU32(v); ok {
			t.Errorf("out-of-domain %v accepted", v)
		}
	}
}

// The HLL [4, 18] range is production's check: the runner only guards the
// uint8 parameter domain, so p = 3 / 19 must reach NewHyperLogLogWithPrecision
// and SKIP on its error (not on a copied harness range).
func TestBuildHLLDelegatesPrecisionRangeToProduction(t *testing.T) {
	for _, p := range []string{"3", "19", "255"} {
		var ok bool
		errOut := captureFile(t, &os.Stderr, func() {
			_, ok = buildHLL([]map[string]any{{"op": "with_precision", "p": json.Number(p)}}, nil)
		})
		if ok || !strings.Contains(errOut, "with_precision error") {
			t.Errorf("p=%s: accepted=%v, want production rejection; stderr: %q", p, ok, errOut)
		}
	}
	for _, p := range []string{"260", "-252", "4294967300", "18446744073709551620"} {
		var ok bool
		errOut := captureFile(t, &os.Stderr, func() {
			_, ok = buildHLL([]map[string]any{{"op": "with_precision", "p": json.Number(p)}}, nil)
		})
		if ok || !strings.Contains(errOut, "uint8 domain") {
			t.Errorf("p=%s: accepted=%v, want uint8-domain SKIP; stderr: %q", p, ok, errOut)
		}
	}
}
