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

// BenchmarkDraw measures rasterising a frame at 48 px, which MyGo does for each new SVG.
func BenchmarkDraw(b *testing.B) {
	l := Folder().frame(Folder().paint(Options{}), 0.5, 0, false)
	for range b.N {
		l.main.Image(96, 96, nil)
	}
}
