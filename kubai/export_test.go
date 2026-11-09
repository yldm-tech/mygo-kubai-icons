package kubai

import "github.com/egoist/mygo/ui"

// Frame draws the icon at hover progress v (0 to 1) on its way in, for tests that compare poses.
func (ic *Icon) Frame(c *ui.Context, o Options, size float32, v float64) {
	ic.layers(c, o, size, v, 0, false)
}

// Layers renders the icon at hover progress v and toggle progress a, on its way out (back to rest) or in.
func (ic *Icon) Layers(v, a float64, out bool) (main, shadow *ui.SVG) {
	l := ic.frame(ic.paint(Options{}), v, a, out)
	return l.main, l.shadow
}
