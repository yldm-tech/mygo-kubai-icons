package main

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/yldm-tech/mygo-kubai-icons/all"
)

// TestStartsOnBases checks that the grid starts on every base icon and nothing else.
func TestStartsOnBases(t *testing.T) {
	bases := 0
	for _, e := range all.Entries {
		if e.Name == e.Base {
			bases++
		}
	}
	shown := newGallery().filtered()
	if len(shown) != bases {
		t.Fatalf("the grid starts with %d icons, want the %d bases", len(shown), bases)
	}
	for _, e := range shown {
		if e.Name != e.Base {
			t.Fatalf("%s, a version of %s, is in the grid of bases", e.Name, e.Base)
		}
	}
}

// TestSearchFindsVersions checks that a search finds versions while the switch shows bases only.
func TestSearchFindsVersions(t *testing.T) {
	g := newGallery()
	g.query = "folder-plus"
	if !slices.ContainsFunc(g.filtered(), func(e all.Entry) bool { return e.Name == "folder-plus" }) {
		t.Fatal("a search for folder-plus does not find it")
	}
	g.query = ""
	if slices.ContainsFunc(g.filtered(), func(e all.Entry) bool { return e.Name == "folder-plus" }) {
		t.Fatal("folder-plus stays in the grid of bases once the search is cleared")
	}
}

// TestFooterCountsBases checks that the footer says the grid holds base icons, out of how many.
func TestFooterCountsBases(t *testing.T) {
	tt := ui.NewTesterWith(newGallery().view, 1280, 820, ui.TesterOptions{})
	tt.SetLocale("en")
	want := fmt.Sprintf(english.countBases, len(newGallery().filtered()), len(all.Entries))
	if !slices.ContainsFunc(tt.Texts(), func(s string) bool { return strings.Contains(s, want) }) {
		t.Fatalf("no %q in the footer; texts: %q", want, tt.Texts())
	}
}

// TestVersionMarksItsBase checks that a version chosen while the grid shows bases, as a link to it does, brings its base into view, and that a click on the base chooses it.
func TestVersionMarksItsBase(t *testing.T) {
	g := newGallery()
	g.selName = "folder-plus"
	tt := ui.NewTesterWith(g.view, 1280, 820, ui.TesterOptions{})
	tt.Advance(100 * time.Millisecond)
	if err := tt.Click("folder"); err != nil {
		t.Fatalf("folder is not in view: %v", err)
	}
	tt.Advance(50 * time.Millisecond)
	if g.selName != "folder" {
		t.Fatalf("clicking folder chose %q", g.selName)
	}
}
