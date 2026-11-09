// Package kubai is Kubai Icons for MyGo native UI: 2,800+ layered, colour-derived icons that move on hover, drawn by MyGo itself.
//
// Each icon is a function returning a shared *Icon; View shows it in a view, animating on hover:
//
//	ui.Row(c).Gap(8).Children(func() {
//		kubai.Folder().View(c, kubai.Options{Size: 24})
//		ui.Text(c, "Documents")
//	})
//
// The icons are generated from kubai087/kubai-icons (MIT) by gen/generate.mjs, which renders every icon with upstream's own code at keyframes along its motion. Unused icons are left out of the executable.
package kubai

import (
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/egoist/mygo/ui"
)

// Variant is the look of an icon.
type Variant uint8

const (
	// Color is the default look: the accent at full strength, with white marks.
	Color Variant = iota
	// Soft is the clear look: fills lifted toward white, outlines a step deeper, white marks turned deep.
	Soft
)

// Options says how View shows an icon. The zero value is a 24 DIP icon in the default blue that moves when the pointer is over it.
type Options struct {
	// Size is the side of the icon's canvas in DIPs; 24 when 0. The shadow reaches a little past it.
	Size float32
	// Color is the accent every colour of the icon derives from; DefaultColor when zero.
	Color ui.Color
	// Variant is Color or Soft.
	Variant Variant
	// BadgeColor is the ink of a corner badge's disc; each badge's dark grey when zero.
	BadgeColor ui.Color
	// SemanticBadge inks a corner badge in its own meaning's colour, green for add and red for remove, over BadgeColor.
	SemanticBadge bool
	// Hovered plays the hover motion as if the pointer were over the icon, so that a parent such as a button can drive it: kubai.Folder().View(c, kubai.Options{Hovered: button.Hovered()}).
	Hovered bool
	// Still keeps the icon at rest whatever the pointer does; Hovered still moves it.
	Still bool
	// Active is the toggled state of an icon a click toggles (Icon.Tap), such as a saved bookmark. Icons that do not toggle ignore it.
	Active bool
	// NoShadow leaves out the drop shadow.
	NoShadow bool
	// Label names the icon for assistive technology; without one it is a decoration it does not see.
	Label string
}

// Motion is the curve every icon moves along, upstream's 460 ms on cubic-bezier(.22, 1, .36, 1).
var Motion = ui.Animation{Duration: 460 * time.Millisecond, Ease: ui.CubicBezier(0.22, 1, 0.36, 1)}

// Icon is one icon. Icon functions return the same *Icon each time; it is safe to use from any goroutine.
type Icon struct {
	name      string
	tone      string
	tap       bool
	templates []template
	tracks    []track

	mu      sync.Mutex
	cache   map[frameKey]*ui.SVG
	fills   map[fillKey]*template
	bitmaps map[bitmapKey]*ui.Bitmap
	last    *ui.Bitmap // the bitmap drawn last, shown while a new one waits for the budget
}

// Name is the icon's upstream id, such as "folder-plus".
func (ic *Icon) Name() string { return ic.name }

// Tap reports whether a click toggles the icon (see Options.Active).
func (ic *Icon) Tap() bool { return ic.tap }

// Tone is the semantic colour of the icon's corner badge as a hex string, or "" when it has none.
func (ic *Icon) Tone() string { return ic.tone }

// Canvas geometry: icons are drawn on a 31-unit canvas whose viewBox is 0 -3 31 31, and rendered with room for the shadow round it, as upstream's standalone SVGs are.
const (
	canvas = 31
	pad    = 2.5
	side   = canvas + 2*pad
)

const svgHead = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="-2.5 -5.5 36 36" width="36" height="36" fill="none">`

type frameKey struct {
	p    paint
	v, a uint16
	out  bool
}

// The cache holds the poses icons rest in and those passed through recently; it is cleared when it grows past this many.
const cacheMax = 256

func newFrameKey(p paint, v, a float64, out bool) frameKey {
	return frameKey{p: p, v: uint16(v*1024 + 0.5), a: uint16(a*1024 + 0.5), out: out}
}

// frame renders the icon at hover progress v and toggle progress a, both 0 to 1, on its way out (back to rest) or in.
func (ic *Icon) frame(p paint, v, a float64, out bool) *ui.SVG {
	k := newFrameKey(p, v, a, out)
	ic.mu.Lock()
	s := ic.cache[k]
	ic.mu.Unlock()
	if s != nil {
		return s
	}
	v, a = float64(k.v)/1024, float64(k.a)/1024
	ps := ic.pose(v, 0, out)
	if ic.tap && a > 0 {
		pa := ic.pose(v, 1, out)
		switch {
		case a >= 1:
			ps = pa
		case pa.t == ps.t:
			ps = pose{t: ps.t, slots: lerp(ps.slots, pa.slots, a)}
		case a >= 0.5:
			ps = pa
		}
	}
	s = ui.MustParseSVG([]byte(svgHead + ic.filled(p, ps.t).render(ps.slots) + "</svg>"))
	ic.mu.Lock()
	if ic.cache == nil || len(ic.cache) >= cacheMax {
		ic.cache = make(map[frameKey]*ui.SVG)
	}
	ic.cache[k] = s
	ic.mu.Unlock()
	return s
}

type bitmapKey struct {
	frameKey
	px     int
	shadow bool
}

// bitmapMax bounds the bitmaps an icon keeps: those it rests in at the sizes and colours shown, and the frames of a motion, which come again when it plays again.
const bitmapMax = 96

// cached returns the bitmap of k if the icon keeps it.
func (ic *Icon) cached(k bitmapKey) *ui.Bitmap {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	return ic.bitmaps[k]
}

// Drawing new bitmaps is the work a switch of colour, look or size brings to every icon on screen at once. budget spreads it over frames: once a frame spent budgetPerFrame on it, the icons still waiting show the bitmap they showed before, or nothing yet, and ask for the next frame, so that the switch shows at once and the icons follow within a few frames.
var budget struct {
	mu     sync.Mutex
	window time.Time
	spent  time.Duration
}

const (
	budgetPerFrame = 8 * time.Millisecond
	// budgetWindow is how long after the first bitmap of a frame others count as the same frame's.
	budgetWindow = 12 * time.Millisecond
)

func overBudget() bool {
	budget.mu.Lock()
	defer budget.mu.Unlock()
	if now := time.Now(); now.Sub(budget.window) > budgetWindow {
		budget.window, budget.spent = now, 0
	}
	return budget.spent >= budgetPerFrame
}

func spend(d time.Duration) {
	budget.mu.Lock()
	budget.spent += d
	budget.mu.Unlock()
}

// bitmap returns the icon drawn at a moment of its motion, px device pixels square for the padded box, with its drop shadow under it unless shadow is false. Drawing an icon as one bitmap, rather than an SVG over a blurred one, keeps a screen of icons to one texture each, with no offscreen pass a frame per icon.
func (ic *Icon) bitmap(p paint, v, a float64, out bool, px int, shadow bool) *ui.Bitmap {
	k := bitmapKey{newFrameKey(p, v, a, out), px, shadow}
	ic.mu.Lock()
	b := ic.bitmaps[k]
	ic.mu.Unlock()
	if b != nil {
		return b
	}
	img := ic.frame(p, v, a, out).Image(px, px, nil)
	if shadow {
		// Upstream's shadow: the silhouette moved down 2 units and blurred by 2, in a light tint of the accent at 21% opacity.
		u := float64(px) / side
		img = castShadow(img, parseHex(tint("#8AB6FF", p.c)), 2*u, 2*u, 0.21)
	}
	b = ui.NewBitmap(img)
	ic.mu.Lock()
	if ic.bitmaps == nil || len(ic.bitmaps) >= bitmapMax {
		ic.bitmaps = make(map[bitmapKey]*ui.Bitmap)
	}
	ic.bitmaps[k] = b
	ic.mu.Unlock()
	return b
}

// filled returns template t with its colours resolved for p. Colour tokens sit wholly in the text between numbers, so each piece is filled on its own.
func (ic *Icon) filled(p paint, t int) *template {
	k := fillKey{p, t}
	ic.mu.Lock()
	defer ic.mu.Unlock()
	if f := ic.fills[k]; f != nil {
		return f
	}
	src := &ic.templates[t]
	f := &template{Text: make([]string, len(src.Text)), Refs: src.Refs}
	for i, s := range src.Text {
		f.Text[i] = p.fill(s)
	}
	if ic.fills == nil || len(ic.fills) >= cacheMax {
		ic.fills = make(map[fillKey]*template)
	}
	ic.fills[k] = f
	return f
}

type fillKey struct {
	p paint
	t int
}

// pose finds the track of the toggle state, preferring the one that plays back to rest when out.
func (ic *Icon) pose(v float64, active float64, out bool) pose {
	var pick *track
	for i := range ic.tracks {
		tr := &ic.tracks[i]
		if tr.Active != (active >= 1) {
			continue
		}
		if pick == nil || tr.Out == out {
			pick = tr
		}
	}
	return pick.at(v)
}

func (ic *Icon) paint(o Options) paint {
	accent := o.Color
	if accent == (ui.Color{}) {
		accent = DefaultColor
	}
	badge := ""
	switch {
	case o.SemanticBadge && ic.tone != "":
		badge = ic.tone
	case o.BadgeColor != (ui.Color{}):
		badge = hexOf(o.BadgeColor)
	}
	return newPaint(accent, o.Variant, badge)
}

func hexOf(c ui.Color) string { return colorRGB(c).hex() }

// SVG returns the icon as an SVG in its own colours, at rest or at the end of its hover, without the shadow (MyGo does not draw SVG filters): for ui.Image, a menu or a tray icon. Its viewBox has room for the shadow round the 31-unit canvas, so the artwork fills about 86% of it.
func (ic *Icon) SVG(o Options, hovered bool) *ui.SVG {
	v := 0.0
	if hovered {
		v = 1
	}
	a := 0.0
	if o.Active {
		a = 1
	}
	return ic.frame(ic.paint(o), v, a, false)
}

// View shows the icon, o.Size DIPs square, moving on hover. Give each icon of a view its own key (c.Key) when icons come and go, so that each keeps its motion.
func (ic *Icon) View(c *ui.Context, o Options) ui.Element {
	size := o.Size
	if size <= 0 {
		size = 24
	}
	box := ui.Box(c).Size(size, size)
	if o.Label != "" {
		box = box.Label(o.Label)
	}
	hovered := o.Hovered || (!o.Still && box.Hovered())
	v := box.AnimateTo("kubai.hover", b2f(hovered), Motion)
	a := 0.0
	if ic.tap {
		a = float64(box.AnimateTo("kubai.active", b2f(o.Active), Motion))
	}
	return box.Draw(ic.painter(ic.paint(o), float64(v), a, !hovered, !o.NoShadow))
}

// painter draws the icon's bitmap over the box it is given, the padded square round it where the shadow falls, at the display's resolution.
func (ic *Icon) painter(p paint, v, a float64, out, shadow bool) func(*ui.Painter, ui.Rect) {
	return func(pt *ui.Painter, r ui.Rect) {
		unit := r.W / canvas
		full := side * unit
		px := int(math.Ceil(float64(full * pt.Scale())))
		dst := ui.Rect{X: r.X - pad*unit, Y: r.Y - pad*unit, W: full, H: full}
		b := ic.cached(bitmapKey{newFrameKey(p, v, a, out), px, shadow})
		if b == nil {
			if overBudget() {
				ic.mu.Lock()
				last := ic.last
				ic.mu.Unlock()
				if last != nil {
					pt.Image(last, dst, ui.Contain)
				}
				pt.After(time.Millisecond)
				return
			}
			start := time.Now()
			b = ic.bitmap(p, v, a, out, px, shadow)
			spend(time.Since(start))
		}
		ic.mu.Lock()
		ic.last = b
		ic.mu.Unlock()
		pt.Image(b, dst, ui.Contain)
	}
}

func b2f(b bool) float32 {
	if b {
		return 1
	}
	return 0
}

// String returns the icon's name.
func (ic *Icon) String() string { return "kubai." + strconv.Quote(ic.name) }
