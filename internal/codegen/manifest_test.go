package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
)

// directiveRe matches the codegen go:generate directive and captures the
// subcommand argument, e.g. "arraylist" in
//
//	//go:generate go run ../internal/codegen arraylist
var directiveRe = regexp.MustCompile(`go:generate go run \.\./internal/codegen (\w+)`)

// TestManifestMatchesGenerators makes the manifest load-bearing: it asserts that
// Families lists exactly the collection packages wired to the generator via a
// go:generate directive. Adding a generated family without a manifest row (so
// the family matrix silently omits it), or leaving a manifest row for a family
// whose generator was removed, fails here. The "matrix" subcommand is the
// manifest's own renderer, not a family, so it is excluded.
func TestManifestMatchesGenerators(t *testing.T) {
	// The test runs with cwd = internal/codegen; the collection packages are
	// siblings under the repo root.
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(repoRoot)
	if err != nil {
		t.Fatal(err)
	}

	wired := map[string]bool{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		doc := filepath.Join(repoRoot, e.Name(), "doc.go")
		src, err := os.ReadFile(doc)
		if err != nil {
			continue // package has no doc.go / no directive
		}
		for _, m := range directiveRe.FindAllSubmatch(src, -1) {
			if name := string(m[1]); name != "matrix" {
				wired[name] = true
			}
		}
	}

	declared := map[string]bool{}
	for _, f := range Families {
		if declared[f.Package] {
			t.Errorf("duplicate manifest package %q", f.Package)
		}
		declared[f.Package] = true
	}

	for pkg := range wired {
		if !declared[pkg] {
			t.Errorf("family %q has a codegen directive but no manifest row (add it to Families)", pkg)
		}
	}
	for pkg := range declared {
		if !wired[pkg] {
			t.Errorf("manifest lists %q but no package has its codegen go:generate directive", pkg)
		}
	}

	if t.Failed() {
		t.Logf("wired=%v declared=%v", sortedKeys(wired), sortedKeys(declared))
	}
}

func sortedKeys(m map[string]bool) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}
