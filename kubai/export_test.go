package kubai

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Frame draws the icon at hover progress v (0 to 1) on its way in, size DIPs square, for tests that compare poses.
func (ic *Icon) Frame(c *ui.Context, o Options, size float32, v float64) {
	at := &motion{from: v, to: v}
	ui.Box(c).Size(size, size).Draw(ic.painter(ic.paint(o), at, nil, false, !o.NoShadow))
}

// Pose renders the icon at hover progress v and toggle progress a, on its way out (back to rest) or in.
func (ic *Icon) Pose(v, a float64, out bool) *ui.SVG {
	return ic.frame(ic.paint(Options{}), v, a, out)
}

// Bitmap draws the icon at rest with its shadow, px device pixels square.
func (ic *Icon) Bitmap(px int) *ui.Bitmap {
	return ic.bitmap(ic.paint(Options{}), 0, 0, false, px, true)
}

// FreshBudget starts a new frame's budget for drawing bitmaps. The budget runs on the wall clock and is shared by every icon, so a test that draws many frames can leave it spent for the next one, which would then see a bitmap held back.
func FreshBudget() {
	budget.mu.Lock()
	budget.window, budget.spent = time.Time{}, 0
	budget.mu.Unlock()
}
