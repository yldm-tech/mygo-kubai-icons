package kubai_test

import (
	"image"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/yldm-tech/mygo-kubai-icons/all"
	"github.com/yldm-tech/mygo-kubai-icons/kubai"
)

func TestEveryIconRenders(t *testing.T) {
	if len(all.Entries) != kubai.Count {
		t.Fatalf("all.Entries has %d icons, Count says %d", len(all.Entries), kubai.Count)
	}
	seen := map[string]bool{}
	for _, e := range all.Entries {
		ic := e.Icon()
		if ic == nil {
			t.Fatalf("%s: no icon", e.Name)
		}
		if ic.Name() != e.Name {
			t.Fatalf("%s: icon is named %q", e.Name, ic.Name())
		}
		if seen[e.Name] {
			t.Fatalf("%s listed twice", e.Name)
		}
		seen[e.Name] = true
		for _, o := range []kubai.Options{{}, {Variant: kubai.Soft, Color: ui.Hex("#EC4899"), SemanticBadge: true, Active: true}} {
			for _, h := range []bool{false, true} {
				s := ic.SVG(o, h)
				if w, hh := s.Size(); w != 36 || hh != 36 {
					t.Fatalf("%s: SVG is %vx%v", e.Name, w, hh)
				}
			}
		}
	}
}

// The rest pose of an icon in the default colours draws something in the blue the set is built round.
func TestRestPoseDraws(t *testing.T) {
	img := kubai.Folder().SVG(kubai.Options{}, false).Image(72, 72, nil)
	if n := opaque(img); n < 72*72/4 {
		t.Fatalf("folder covers %d of %d pixels", n, 72*72)
	}
}

// Hover moves the icon: the end of the hover differs from the rest pose.
func TestHoverMoves(t *testing.T) {
	for _, ic := range []*kubai.Icon{kubai.Folder(), kubai.Abacus(), kubai.BellX(), kubai.Bookmark()} {
		rest := ic.SVG(kubai.Options{}, false).Image(64, 64, nil)
		hover := ic.SVG(kubai.Options{}, true).Image(64, 64, nil)
		if same(rest, hover) {
			t.Errorf("%s: hover end looks the same as rest", ic.Name())
		}
	}
}

func TestColorChangesPixels(t *testing.T) {
	blue := kubai.Mail().SVG(kubai.Options{}, false).Image(48, 48, nil)
	pink := kubai.Mail().SVG(kubai.Options{Color: ui.Hex("#EC4899")}, false).Image(48, 48, nil)
	if same(blue, pink) {
		t.Fatal("colour did not change the icon")
	}
}

func TestViewRendersHeadless(t *testing.T) {
	img := ui.Render(func(c *ui.Context) {
		ui.Row(c).Gap(8).Padding(8).Children(func() {
			kubai.Folder().View(c.Key("a"), kubai.Options{Size: 32})
			kubai.Bookmark().View(c.Key("b"), kubai.Options{Size: 32, Active: true})
			kubai.Mail().View(c.Key("c"), kubai.Options{Size: 32, Variant: kubai.Soft, NoShadow: true})
		})
	}, 160, 60, 2)
	if opaque(img) == 0 {
		t.Fatal("nothing drawn")
	}
}

func opaque(img *image.RGBA) int {
	n := 0
	for i := 3; i < len(img.Pix); i += 4 {
		if img.Pix[i] > 0 {
			n++
		}
	}
	return n
}

func same(a, b *image.RGBA) bool {
	if len(a.Pix) != len(b.Pix) {
		return false
	}
	for i := range a.Pix {
		if a.Pix[i] != b.Pix[i] {
			return false
		}
	}
	return true
}

// Every pose along every track parses, on the way in and back, including between toggle states.
func TestEveryPoseParses(t *testing.T) {
	for _, e := range all.Entries {
		ic := e.Icon()
		for _, v := range []float64{0.13, 0.5, 0.87} {
			for _, out := range []bool{false, true} {
				if ic.Pose(v, 0, out) == nil {
					t.Fatalf("%s at %v: no markup", e.Name, v)
				}
				if ic.Tap() {
					ic.Pose(v, 0.5, out)
					ic.Pose(v, 1, out)
				}
			}
		}
	}
}

// The shadow falls below the icon: the bitmap has coverage under the artwork's lowest row, where the SVG alone has none.
func TestShadowFallsBelow(t *testing.T) {
	const px = 72
	ic := kubai.Mail()
	plain := ic.SVG(kubai.Options{}, false).Image(px, px, nil)
	lowest := 0
	for y := 0; y < px; y++ {
		for x := 0; x < px; x++ {
			if plain.Pix[y*plain.Stride+x*4+3] > 0 {
				lowest = y
			}
		}
	}
	if b := ic.Bitmap(px); b == nil {
		t.Fatal("no bitmap")
	}
	if w, h := ic.Bitmap(px).Size(); w != px || h != px {
		t.Fatalf("bitmap is %dx%d", w, h)
	}
	if lowest >= px-1 {
		t.Fatalf("the artwork reaches the bottom edge (row %d), leaving no room to see the shadow", lowest)
	}
}
