package kubai

import (
	"testing"
)

// BenchmarkFrame measures a frame the cache does not hold, what each frame of a moving icon costs: interpolating its keyframes, writing the markup and parsing it.
func BenchmarkFrame(b *testing.B) {
	for _, ic := range []*Icon{Folder(), BellX(), Calendar()} {
		b.Run(ic.Name(), func(b *testing.B) {
			p := ic.paint(Options{})
			for i := range b.N {
				ic.cache = nil
				ic.frame(p, float64(i%1000)/1000, 0, false)
			}
		})
	}
}

// BenchmarkBitmap measures drawing a frame into a bitmap the cache does not hold, with its shadow, at 75 px: a 32 DIP icon on a 2× display, what a switch costs each icon on screen.
func BenchmarkBitmap(b *testing.B) {
	ic := Folder()
	p := ic.paint(Options{})
	for i := range b.N {
		ic.bitmaps = nil
		ic.bitmap(p, float64(i%2), 0, false, 75, true)
	}
}

// BenchmarkShadow measures casting the shadow alone at 75 px.
func BenchmarkShadow(b *testing.B) {
	img := Folder().frame(Folder().paint(Options{}), 0, 0, false).Image(75, 75, nil)
	for range b.N {
		castShadow(img, rgb{138, 182, 255}, 4, 4, 0.21)
	}
}
