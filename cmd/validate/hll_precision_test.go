package main

import "testing"

func TestBuildHLLChecksPrecisionBeforeNarrowing(t *testing.T) {
	for _, p := range []int{4, 18} {
		h, ok := buildHLL([]map[string]any{{"op": "with_precision", "p": float64(p)}}, nil)
		if !ok || int(h.Precision()) != p || h.RegisterCount() != 1<<p {
			t.Fatalf("valid precision %d: accepted=%v", p, ok)
		}
	}
	for _, p := range []int{3, 19, 260, -252, 65540} {
		if _, ok := buildHLL([]map[string]any{{"op": "with_precision", "p": float64(p)}}, nil); ok {
			t.Errorf("out-of-range precision %d accepted after narrowing", p)
		}
	}
}
