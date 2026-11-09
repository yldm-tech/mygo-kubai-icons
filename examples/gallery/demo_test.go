package main

import (
	"bytes"
	"image"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/yldm-tech/mygo-kubai-icons/all"
)

// previewPixels returns the pixels of the large preview, the stage above its hint.
func previewPixels(t *testing.T, tt *ui.Tester, hint string) []byte {
	t.Helper()
	r, ok := tt.Find(hint)
	if !ok {
		t.Fatalf("no %q on screen; texts: %q", hint, tt.Texts())
	}
	img := tt.Image()
	s := img.Bounds().Dx() / 1280
	box := image.Rect(int(r.X-40)*s, int(r.Y-150)*s, int(r.X+r.W+40)*s, int(r.Y-4)*s).Intersect(img.Bounds())
	var out []byte
	for y := box.Min.Y; y < box.Max.Y; y++ {
		out = append(out, img.Pix[img.PixOffset(box.Min.X, y):img.PixOffset(box.Max.X, y)]...)
	}
	return out
}

func demoGallery() *gallery {
	g := newGallery()
	g.size, g.selName = 32, "folder-plus"
	return g
}

// TestPreviewPlaysOnItsOwn checks that the preview moves to its hover pose and back with no pointer over it.
func TestPreviewPlaysOnItsOwn(t *testing.T) {
	tt := ui.NewTesterWith(demoGallery().view, 1280, 820, ui.TesterOptions{})
	tt.SetLocale("en")
	hint := english.previewDemo
	rest := previewPixels(t, tt, hint)
	tt.Advance(demoIn + 600*time.Millisecond)
	if bytes.Equal(rest, previewPixels(t, tt, hint)) {
		t.Fatal("the preview did not move on its own")
	}
	tt.Advance(demoPeriod - demoIn - 600*time.Millisecond + 300*time.Millisecond)
	if !bytes.Equal(rest, previewPixels(t, tt, hint)) {
		t.Fatal("the preview did not come back to rest at the end of its loop")
	}
}

// TestPreviewStillWithLessMotion checks that the loop does not run while the system asks for less motion.
func TestPreviewStillWithLessMotion(t *testing.T) {
	tt := ui.NewTesterWith(demoGallery().view, 1280, 820, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	tt.SetLocale("en")
	hint := english.previewHint
	rest := previewPixels(t, tt, hint)
	tt.Advance(demoIn + 600*time.Millisecond)
	if !bytes.Equal(rest, previewPixels(t, tt, hint)) {
		t.Fatal("the preview moved although the system asks for less motion")
	}
	if tt.FramePending() {
		t.Fatal("the preview keeps asking for frames although it does not move")
	}
}

// TestChosenIconPlaysOnce checks that the icon chosen in the grid plays for a moment and then stops.
func TestChosenIconPlaysOnce(t *testing.T) {
	g := demoGallery()
	g.selName = ""
	tt := ui.NewTesterWith(g.view, 1280, 820, ui.TesterOptions{})
	tt.SetLocale("en")
	first := all.Entries[0].Name
	if err := tt.Click(first); err != nil {
		t.Fatal(err)
	}
	tt.Advance(100 * time.Millisecond)
	if g.selName != first || g.pulse != first {
		t.Fatalf("chose %q and plays %q, want %q for both", g.selName, g.pulse, first)
	}
	tt.Advance(pulseFor + 100*time.Millisecond)
	if g.pulse != "" {
		t.Fatalf("still plays %q after %v", g.pulse, pulseFor)
	}
}
