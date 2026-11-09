package main

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// render draws one frame of the gallery headlessly, as a window of the site's usual size would.
func render(g *gallery) time.Duration {
	start := time.Now()
	ui.Render(g.view, 1280, 820, 1)
	return time.Since(start)
}

// TestSwitchCost reports how long the frame after each kind of switch takes, the first time and again, to find what makes switching slow.
func TestSwitchCost(t *testing.T) {
	if !testing.Verbose() {
		t.Skip("reports timings only; run with -v")
	}
	g := newGallery()
	g.size = 32
	g.selName = "folder-plus"
	t.Logf("first frame        %v", render(g))
	t.Logf("same frame again   %v", render(g))
	g.accent = 2
	t.Logf("accent → pink      %v", render(g))
	t.Logf("  again            %v", render(g))
	g.variant = 1
	t.Logf("variant → soft     %v", render(g))
	g.size = 48
	t.Logf("size → 48          %v", render(g))
	g.category = "devices"
	t.Logf("category devices   %v", render(g))
	g.category = "food"
	t.Logf("category food      %v", render(g))
	g.category = "devices"
	t.Logf("back to devices    %v", render(g))
	g.accent = 0
	t.Logf("accent → blue      %v", render(g))
}

// BenchmarkSwitch renders the frames of switching back and forth between two accents, for a CPU profile of a switch.
func BenchmarkSwitch(b *testing.B) {
	g := newGallery()
	g.size = 32
	g.selName = "folder-plus"
	render(g)
	for i := range b.N {
		g.accent = i % 2 * 2
		g.category = []string{"", "devices", "food"}[i%3]
		render(g)
	}
}

// BenchmarkSteady renders the same frame again and again, the cost of every frame of a hover animation.
func BenchmarkSteady(b *testing.B) {
	g := newGallery()
	g.size = 32
	g.selName = "folder-plus"
	render(g)
	for range b.N {
		render(g)
	}
}
