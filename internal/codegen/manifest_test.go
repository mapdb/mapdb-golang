package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// directiveRe matches an ACTIVE codegen go:generate directive and captures the
// subcommand argument, e.g. "arraylist" in
//
//	//go:generate go run ../internal/codegen arraylist
//
// It is anchored to the start of a line (Go only honours a go:generate directive
// whose "//go:generate" begins the line) so a commented-out or disabled variant
// like "// disabled: //go:generate …" or "////go:generate …" is NOT counted as
// wired — otherwise the consistency guard could pass while a package's real
// generation was switched off.
var directiveRe = regexp.MustCompile(`(?m)^//go:generate go run \.\./internal/codegen (\w+)\s*$`)

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

	// nonFamily are subcommands that are not per-family generators: the "matrix"
	// renderer and the "interfaces" vocabulary generator. They are excluded from
	// every family set below.
	nonFamily := map[string]bool{"matrix": true, "interfaces": true}

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
			if name := string(m[1]); !nonFamily[name] {
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

	registered := map[string]bool{}
	for name := range generators {
		if !nonFamily[name] {
			registered[name] = true
		}
	}

	for pkg := range wired {
		if !declared[pkg] {
			t.Errorf("family %q has a codegen directive but no manifest row (add it to Families)", pkg)
		}
		if !registered[pkg] {
			t.Errorf("family %q has a codegen directive but no generators[] entry", pkg)
		}
	}
	for pkg := range declared {
		if !wired[pkg] {
			t.Errorf("manifest lists %q but no package has its codegen go:generate directive", pkg)
		}
		if !registered[pkg] {
			t.Errorf("manifest lists %q but it has no generators[] entry", pkg)
		}
	}
	for pkg := range registered {
		if !declared[pkg] {
			t.Errorf("generators[] has %q but the manifest does not (add it to Families)", pkg)
		}
	}

	if t.Failed() {
		t.Logf("wired=%v declared=%v registered=%v",
			sortedKeys(wired), sortedKeys(declared), sortedKeys(registered))
	}
}

// TestManifestVariantsMatchFiles guards the manifest's structured Immutable and
// Synchronized facts against reality: they must agree with the presence of
// immutable_*.go / synchronized_*.go generated files in the family's package.
// This is the automated check for the class of error a manual audit caught in
// the descriptive columns — a stale boolean here now fails the build instead of
// silently shipping a wrong cell in the family matrix.
func TestManifestVariantsMatchFiles(t *testing.T) {
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range Families {
		dir := filepath.Join(repoRoot, f.Package)
		if got := hasGenerated(t, dir, "immutable_*.go"); got != f.Immutable {
			t.Errorf("%s: manifest Immutable=%v but immutable_*.go present=%v", f.Package, f.Immutable, got)
		}
		if got := hasGenerated(t, dir, "synchronized_*.go"); got != f.Synchronized {
			t.Errorf("%s: manifest Synchronized=%v but synchronized_*.go present=%v", f.Package, f.Synchronized, got)
		}
	}
}

// hasGenerated reports whether dir holds at least one non-test source file
// matching pattern.
func hasGenerated(t *testing.T, dir, pattern string) bool {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		t.Fatalf("glob %s: %v", pattern, err)
	}
	for _, m := range matches {
		if !strings.HasSuffix(m, "_test.go") {
			return true
		}
	}
	return false
}

func sortedKeys(m map[string]bool) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}
