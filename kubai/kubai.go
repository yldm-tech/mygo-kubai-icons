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

	mu    sync.Mutex
	cache map[frameKey]*layers
	fills map[fillKey]*template
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

// layers is one rendered moment: the icon, and the silhouette its shadow is blurred from (nil when nothing casts one).
type layers struct {
	main, shadow *ui.SVG
}

type frameKey struct {
	p    paint
	v, a uint16
	out  bool
}

// The cache holds the poses icons rest in and those passed through recently; it is cleared when it grows past this many.
const cacheMax = 256

// frame renders the icon at hover progress v and toggle progress a, both 0 to 1, on its way out (back to rest) or in.
func (ic *Icon) frame(p paint, v, a float64, out bool) *layers {
	k := frameKey{p: p, v: uint16(v*1024 + 0.5), a: uint16(a*1024 + 0.5), out: out}
	ic.mu.Lock()
	l := ic.cache[k]
	ic.mu.Unlock()
	if l != nil {
		return l
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
	main, shadow := ic.filled(p, ps.t).render(ps.slots)
	l = &layers{main: ui.MustParseSVG([]byte(svgHead + main + "</svg>"))}
	if shadow != "" {
		l.shadow = ui.MustParseSVG([]byte(svgHead + `<g transform="translate(0 2)">` + shadow + "</g></svg>"))
	}
	ic.mu.Lock()
	if ic.cache == nil || len(ic.cache) >= cacheMax {
		ic.cache = make(map[frameKey]*layers)
	}
	ic.cache[k] = l
	ic.mu.Unlock()
	return l
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
	return ic.frame(ic.paint(o), v, a, false).main
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
	return box.Children(func() { ic.layers(c, o, size, float64(v), a, !hovered) })
}

func (ic *Icon) layers(c *ui.Context, o Options, size float32, v, a float64, out bool) {
	l := ic.frame(ic.paint(o), v, a, out)
	unit := size / canvas
	full := side * unit
	if l.shadow != nil && !o.NoShadow {
		// Upstream's shadow: the silhouette blurred by 2 units, moved down 2 (in the markup) at 21% opacity.
		ui.Image(c, l.shadow).Size(full, full).Absolute().Left(-pad * unit).Top(-pad * unit).Blur(2 * unit).Opacity(0.21)
	}
	ui.Image(c, l.main).Size(full, full).Absolute().Left(-pad * unit).Top(-pad * unit)
}

func b2f(b bool) float32 {
	if b {
		return 1
	}
	return 0
}

// String returns the icon's name.
func (ic *Icon) String() string { return "kubai." + strconv.Quote(ic.name) }
