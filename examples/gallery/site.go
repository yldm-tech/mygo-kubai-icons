package main

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/yldm-tech/mygo-kubai-icons/all"
	"github.com/yldm-tech/mygo-kubai-icons/kubai"
)

const (
	repoURL     = "https://github.com/yldm-tech/mygo-kubai-icons"
	docsURL     = "https://pkg.go.dev/github.com/yldm-tech/mygo-kubai-icons/kubai"
	upstreamURL = "https://icons.kubai.design/"
	mygoURL     = "https://mygo.run/"
	awesomeURL  = "https://awesomemygo.dev/"
	aiLogoURL   = "https://ailogo.yldm.ai/"
	// compactWidth is the window width below which the sidebar folds into a drop-down and the details open in a dialog: three columns need about this much to fit their text.
	compactWidth = 1100
	// roomyWidth is the width from which the header has room for the install command beside the title, and the footer for the related sites.
	roomyWidth = 1280
)

var accents = []struct {
	key   string
	color ui.Color
}{
	{"blue", kubai.DefaultColor},
	{"violet", ui.Hex("#8B5CF6")},
	{"pink", ui.Hex("#EC4899")},
	{"orange", ui.Hex("#F97316")},
	{"green", ui.Hex("#22C55E")},
	{"slate", ui.Hex("#64748B")},
}

// categoryIcons is the icon each category shows in the sidebar.
var categoryIcons = map[string]func() *kubai.Icon{
	"files": kubai.Folder, "text": kubai.Bold, "interface": kubai.Bookmark, "layout": kubai.LayoutDashboard,
	"arrows": kubai.ArrowRight, "communication": kubai.Mail, "people": kubai.Hand, "security": kubai.Lock,
	"devices": kubai.Smartphone, "media": kubai.Camera, "charts": kubai.ChartBar, "commerce": kubai.CreditCard,
	"time": kubai.Clock, "weather": kubai.CloudSun, "nature": kubai.Flower, "home": kubai.Armchair,
	"food": kubai.Burger, "travel": kubai.Compass, "transport": kubai.Car, "tools": kubai.Hammer,
	"development": kubai.Code, "science": kubai.Atom, "education": kubai.GraduationCap, "health": kubai.HeartPulse,
	"sports": kubai.Basketball, "shapes": kubai.Shapes,
}

type gallery struct {
	query     string
	category  string // "" for every category
	accent    int
	variant   int
	size      float64
	basesOnly bool
	hoverAll  bool

	grid    ui.GridState
	selIdx  int    // the chosen icon's place in shown, -1 for none, which the grid's arrows move
	selName string // the chosen icon, which outlives a new search
	scroll  bool   // bring the chosen icon into view once the grid is laid out
	laidOut bool   // the grid was built once, so it knows its columns
	dialog  bool   // the details' dialog, open in a compact window
	active  bool   // the toggled state the preview shows, for icons a click toggles

	demoName string    // the icon the preview plays on its own, "" for the placeholder
	demoFrom time.Time // when its loop started, again each time the pointer leaves it
	pulse    string    // the icon the grid plays once because it was just chosen
	pulseAt  time.Time

	tour     map[string]time.Time // the icons the grid plays on its own, by when each started
	tourAt   int                  // the item the tour played last
	tourNext time.Time            // when it plays the next, zero while it waits
	inView   []int                // the items that showed in the last frame
	building []int                // those of this one

	copied string // what was copied last, the install command or a Go call

	shown  []all.Entry
	shownK string
	counts map[string]int
	countK bool
}

// newGallery starts on base icons only: about two icons in three are badged or slashed versions of a base, which would fill the first screen with one object five times over. Each base lists its versions in the details, and a search looks through all of them.
func newGallery() *gallery {
	return &gallery{selIdx: -1, basesOnly: true, tourAt: -1, tour: map[string]time.Time{}}
}

// basesShown reports whether the grid leaves out the versions: with the switch on and no search, as a search for a version must find it.
func (g *gallery) basesShown() bool {
	return g.basesOnly && strings.TrimSpace(g.query) == ""
}

// filtered returns the icons the search, the category and the base-only switch leave.
func (g *gallery) filtered() []all.Entry {
	q := strings.ToLower(strings.TrimSpace(g.query))
	bases := g.basesShown()
	key := fmt.Sprintf("%s\x00%s\x00%v", q, g.category, bases)
	if g.shown != nil && key == g.shownK {
		return g.shown
	}
	g.shown = g.shown[:0]
	for _, e := range all.Entries {
		if g.category != "" && e.Category != g.category {
			continue
		}
		if bases && e.Name != e.Base {
			continue
		}
		if q == "" || strings.Contains(e.Name, q) || strings.Contains(strings.ToLower(e.GoName), q) || strings.Contains(e.Zh, q) {
			g.shown = append(g.shown, e)
		}
	}
	g.shownK = key
	g.scroll = true
	return g.shown
}

// categoryCounts counts the icons of each category, all of them under "".
func (g *gallery) categoryCounts() map[string]int {
	bases := g.basesShown()
	if g.counts != nil && g.countK == bases {
		return g.counts
	}
	g.counts = map[string]int{}
	for _, e := range all.Entries {
		if bases && e.Name != e.Base {
			continue
		}
		g.counts[e.Category]++
		g.counts[""]++
	}
	g.countK = bases
	return g.counts
}

func (g *gallery) options() kubai.Options {
	return kubai.Options{Size: float32(g.size), Color: accents[g.accent].color, Variant: kubai.Variant(g.variant), SemanticBadge: true}
}

func (g *gallery) chosen() (all.Entry, bool) {
	for _, e := range all.Entries {
		if e.Name == g.selName {
			return e, true
		}
	}
	return all.Entry{}, false
}

func (g *gallery) choose(name string) {
	if name != g.selName {
		g.selName, g.active, g.scroll = name, false, true
	}
}

// The preview's loop: at rest, then the hover pose from demoIn to demoOut, then back, every demoPeriod. kubai.Motion takes 460 ms each way, so each pose holds a moment.
const (
	demoIn     = 700 * time.Millisecond
	demoOut    = 1900 * time.Millisecond
	demoPeriod = 2800 * time.Millisecond
	pulseFor   = 1100 * time.Millisecond // how long a just-chosen icon holds its hover pose in the grid
)

// demo reports whether the preview of name shows its hover pose: under the pointer it does, as everywhere, and otherwise it plays on its own in a loop, so that a visitor who never hovers, or cannot, still sees the icon move. The loop starts over at rest for each icon and each time the pointer leaves; it does not run while the system asks for less motion.
func (g *gallery) demo(c *ui.Context, name string, hovered bool) bool {
	now := c.Now()
	if hovered || name != g.demoName || c.Preferences().ReduceMotion {
		g.demoName, g.demoFrom = name, now
		return hovered
	}
	at := now.Sub(g.demoFrom) % demoPeriod
	switch {
	case at < demoIn:
		c.After(demoIn - at)
		return false
	case at < demoOut:
		c.After(demoOut - at)
		return true
	default:
		c.After(demoPeriod - at)
		return false
	}
}

// pulsing reports whether name, just chosen in the grid, plays its hover once: a tap shows the motion where there is no pointer to hover with, and the arrows show it too.
func (g *gallery) pulsing(c *ui.Context, name string) bool {
	if name != g.pulse || c.Preferences().ReduceMotion {
		return false
	}
	left := pulseFor - c.Now().Sub(g.pulseAt)
	if left <= 0 {
		g.pulse = ""
		return false
	}
	c.After(left)
	return true
}

// The grid's tour: every tourStep the next icon in view plays its hover for tourHold, left to right and row by row, so that the page moves before anyone hovers it, two or three icons at a time rather than all of them.
const (
	tourStep = 500 * time.Millisecond
	tourHold = 1200 * time.Millisecond
)

// advanceTour starts the next icon of the tour when it is due. It goes on under the pointer, which rests on the grid most of the time on a desktop, and the icon hovered plays as any hovered icon does; it waits while every icon is held hovered and while the system asks for less motion, and starts again a step after.
func (g *gallery) advanceTour(c *ui.Context, entries []all.Entry) {
	now := c.Now()
	for name, at := range g.tour {
		if now.Sub(at) >= tourHold {
			delete(g.tour, name)
		}
	}
	if g.hoverAll || c.Preferences().ReduceMotion {
		g.tourNext = time.Time{}
		return
	}
	if len(g.inView) == 0 {
		// What shows is known a frame after it is laid out.
		g.tourNext = time.Time{}
		if len(entries) > 0 {
			c.After(tourStep)
		}
		return
	}
	if g.tourNext.IsZero() {
		g.tourNext = now.Add(tourStep)
	}
	if now.Before(g.tourNext) {
		c.After(g.tourNext.Sub(now))
		return
	}
	next := g.inView[0]
	for _, i := range g.inView {
		if i > g.tourAt {
			next = i
			break
		}
	}
	if next < len(entries) {
		g.tour[entries[next].Name] = now
	}
	g.tourAt, g.tourNext = next, now.Add(tourStep)
	c.After(tourStep)
}

// touring reports whether name plays as part of the tour.
func (g *gallery) touring(c *ui.Context, name string) bool {
	at, ok := g.tour[name]
	if !ok {
		return false
	}
	left := tourHold - c.Now().Sub(at)
	if left <= 0 {
		return false
	}
	c.After(left)
	return true
}

func (g *gallery) copy(c *ui.Context, s string) {
	c.WriteClipboard(s)
	g.copied = s
}

func (g *gallery) view(c *ui.Context) {
	t := c.Theme()
	w := say(c)
	width, _ := c.Size()
	compact := width < compactWidth
	entries := g.filtered()

	ui.Column(c).Fill().Background(t.Background).Children(func() {
		g.header(c, w, compact, width >= roomyWidth)
		ui.Row(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
			if !compact {
				g.sidebar(c, w)
			}
			ui.Column(c).Grow(1).MinWidth(0).Children(func() {
				g.toolbar(c, w, compact)
				g.gridView(c, w, entries, compact)
			})
			if !compact {
				ui.Column(c).Width(340).Shrink(0).BorderWidth(0, 0, 0, 1).BorderColor(t.Border).Background(t.Surface).Children(func() {
					g.detail(c, w)
				})
			}
		})
		g.footer(c, w, len(entries), width >= roomyWidth)
	})
	if compact {
		if _, ok := g.chosen(); ok && g.dialog {
			ui.Modal(c, &g.dialog, func() {
				ui.Column(c).Width(min(width-32, 420)).Children(func() { g.detail(c, w) })
			})
		}
	}
}

func (g *gallery) header(c *ui.Context, w words, compact, roomy bool) {
	t := c.Theme()
	ui.Row(c).Padding(14, 20).Gap(14).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
		kubai.Shapes().View(c, kubai.Options{Size: 34, Label: w.title})
		ui.Column(c).Gap(2).Shrink(1).Children(func() {
			ui.Text(c, w.title).FontSize(18).Bold()
			if !compact {
				ui.Text(c, w.tagline).FontSize(12).TextColor(t.TextMuted).MaxLines(1)
			}
		})
		ui.Spacer(c)
		if roomy {
			chip := ui.Row(c).Gap(10).AlignItems(ui.Center).Padding(6, 6, 6, 12).Radius(8).Background(t.Surface).Border(1, t.Border)
			chip.Children(func() {
				ui.Text(c, w.install).Font("monospace").FontSize(12).Selectable()
				label := w.copy
				if g.copied == w.install {
					label = w.copied
				}
				if ui.Button(c, label).Clicked() {
					g.copy(c, w.install)
				}
			})
		}
		ui.Link(c, w.github, repoURL)
		ui.Link(c, w.docs, docsURL)
		if ui.Button(c, w.otherLang).Label(w.otherLang).Clicked() {
			if w.zh {
				mygo.App.SetLocale("en")
			} else {
				mygo.App.SetLocale("zh-Hans")
			}
		}
	})
}

func (g *gallery) sidebar(c *ui.Context, w words) {
	t := c.Theme()
	counts := g.categoryCounts()
	ui.Scroll(c).Width(232).Shrink(0).BorderWidth(0, 1, 0, 0).BorderColor(t.Border).Children(func() {
		ui.Column(c).Padding(12, 10).Gap(2).Children(func() {
			item := func(id, label string, icon *kubai.Icon) {
				row := ui.Row(c.Key("cat-"+id)).Gap(10).AlignItems(ui.Center).Padding(6, 10).Radius(8).Cursor(ui.CursorPointer).Label(label)
				switch {
				case g.category == id:
					row.Background(t.SurfacePressed)
				case row.Hovered():
					row.Background(t.SurfaceHover)
				}
				if row.Clicked() {
					g.category = id
				}
				row.Children(func() {
					// The categories keep the default look whatever the toolbar sets, so that a switch redraws only the icons it shows.
					o := kubai.Options{Size: 20, NoShadow: true, Hovered: row.Hovered()}
					icon.View(c, o)
					txt := ui.Text(c, label).FontSize(13).Grow(1).MaxLines(1)
					if g.category == id {
						txt.Bold()
					}
					ui.Textf(c, "%d", counts[id]).FontSize(11).TextColor(t.TextMuted)
				})
			}
			item("", w.all, kubai.Shapes())
			ui.Box(c).Height(8)
			for _, cat := range all.Categories {
				item(cat.ID, w.categoryName(cat.ID), categoryIcons[cat.ID]())
			}
		})
	})
}

func (g *gallery) toolbar(c *ui.Context, w words, compact bool) {
	t := c.Theme()
	ui.Column(c).Padding(12, 16).Gap(10).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
		ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
			ui.SearchField(c, &g.query).Label(w.search).Placeholder(w.search).Grow(1)
			if compact {
				names := []string{w.all}
				for _, cat := range all.Categories {
					names = append(names, w.categoryName(cat.ID))
				}
				pick := w.all
				if g.category != "" {
					pick = w.categoryName(g.category)
				}
				if ui.Select(c, &pick, names).Label(w.category).Changed() {
					g.category = ""
					for _, cat := range all.Categories {
						if w.categoryName(cat.ID) == pick {
							g.category = cat.ID
						}
					}
				}
			}
		})
		ui.Row(c).Gap(14).AlignItems(ui.Center).Wrap().Children(func() {
			ui.Segmented(c, &g.variant, w.color, w.soft).Label(w.look)
			ui.Row(c).Gap(6).Children(func() {
				for i, a := range accents {
					sw := ui.Box(c.Key(a.key)).Size(20, 20).Radius(10).Background(a.color).Label(w.accents[i]).Tooltip(w.accents[i]).Cursor(ui.CursorPointer)
					if i == g.accent {
						sw.Border(2, t.Text)
					}
					if sw.Clicked() {
						g.accent = i
					}
				}
			})
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Slider(c, &g.size, 16, 64).Label(w.size).Width(110)
				ui.Textf(c, "%.0f", g.size).FontSize(12).TextColor(t.TextMuted).Width(20)
			})
			ui.Spacer(c)
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				// A search looks through every version, so the switch has nothing to do meanwhile.
				searching := strings.TrimSpace(g.query) != ""
				sw := ui.Switch(c, &g.basesOnly).Label(w.basesOnly).Disabled(searching)
				label := ui.Text(c, w.basesOnly).FontSize(13).MaxLines(1)
				if searching {
					sw.Tooltip(w.searchAll)
					label.TextColor(t.TextMuted).Tooltip(w.searchAll)
				}
			})
		})
	})
}

func (g *gallery) gridView(c *ui.Context, w words, entries []all.Entry, compact bool) {
	t := c.Theme()
	if len(entries) == 0 {
		ui.Column(c).Grow(1).Center().Children(func() {
			ui.Text(c, w.empty).TextColor(t.TextMuted)
		})
		return
	}
	g.selIdx = -1
	for i, e := range entries {
		if e.Name == g.selName {
			g.selIdx = i
			break
		}
	}
	// A version chosen while the grid shows bases only, from the details or a link, marks its base, lightly: the grid's own mark stays off, so that a click on the base chooses it.
	group := -1
	if e, ok := g.chosen(); ok && g.selIdx < 0 && e.Base != e.Name {
		for i, x := range entries {
			if x.Name == e.Base {
				group = i
				break
			}
		}
	}
	g.grid.Selected = &g.selIdx
	g.advanceTour(c, entries)
	g.building = g.building[:0]
	opts := g.options()
	cell := float32(g.size) + 60
	grid := ui.GridView(c, &g.grid, len(entries), cell+28, cell, func(i int) {
		e := entries[i]
		// The card fills the grid's cell, which marks the chosen item itself with its border and background; the card only lights up under the pointer.
		card := ui.Column(c.Key(e.Name)).Grow(1).Center().Gap(10).Padding(8).Radius(t.Radius).Cursor(ui.CursorPointer).Tooltip(e.Name)
		// The grid builds a row or two beyond what shows; the tour plays only what does.
		if card.InViewAmount(0.8) {
			g.building = append(g.building, i)
		}
		switch {
		case i == group:
			card.Background(t.SurfacePressed)
		case i != g.selIdx && card.Hovered():
			card.Background(t.SurfaceHover)
		}
		card.Children(func() {
			o := opts
			// pulsing and touring first, so that they end under the pointer too
			pulse, tour := g.pulsing(c, e.Name), g.touring(c, e.Name)
			o.Hovered = pulse || tour || g.hoverAll || card.Hovered()
			e.Icon().View(c, o)
			ui.Text(c, w.label(e)).FontSize(11).TextColor(t.TextMuted).MaxLines(1)
		})
	}).Grow(1).Padding(4, 12)
	slices.Sort(g.building)
	g.inView, g.building = g.building, g.inView
	// Bring the chosen icon into view. The grid counts its columns from the window's width until it was laid out once, and when the count changes it scrolls to keep its rows; so the first frame leaves this for the next, and the request comes after the grid's own, which it overrides.
	if g.scroll && g.laidOut {
		switch {
		case g.selIdx >= 0:
			g.grid.ScrollTo(g.selIdx, ui.Center)
		case group >= 0:
			g.grid.ScrollTo(group, ui.Center)
		}
		g.scroll = false
	} else if g.scroll {
		c.After(16 * time.Millisecond)
	}
	g.laidOut = true
	if grid.Changed() && g.selIdx >= 0 && g.selIdx < len(entries) {
		g.selName, g.active = entries[g.selIdx].Name, false
		g.dialog = compact
		// The cards were built before the grid took the choice, so the one chosen starts playing in the next frame.
		g.pulse, g.pulseAt = g.selName, c.Now()
		c.After(time.Millisecond)
	}
}

func (g *gallery) footer(c *ui.Context, w words, n int, roomy bool) {
	t := c.Theme()
	ui.Row(c).Padding(8, 20).Gap(6).AlignItems(ui.Center).BorderWidth(1, 0, 0, 0).BorderColor(t.Border).Children(func() {
		status := w.hint
		if !c.Preferences().Hover() {
			status = w.hintTouch
		}
		if g.copied != "" {
			status = fmt.Sprintf(w.copiedCall, g.copied)
		}
		count := w.count
		if g.basesShown() {
			count = w.countBases
		}
		ui.Textf(c, count+" · %s", n, kubai.Count, status).FontSize(12).TextColor(t.TextMuted).MaxLines(1).Shrink(1)
		ui.Spacer(c)
		muted := func(s string) { ui.Text(c, s).FontSize(12).TextColor(t.TextMuted) }
		muted(w.credit)
		ui.Link(c, w.upstream, upstreamURL).FontSize(12)
		muted("(" + w.mit + ") · " + w.unofficial + " · " + w.drawnBy)
		// The related sites show where they leave the hint its room: a narrower window keeps the footer to the count and the credit.
		if roomy {
			muted("· " + w.related)
			ui.Link(c, w.mygo, mygoURL).FontSize(12)
			muted("·")
			ui.Link(c, w.awesome, awesomeURL).FontSize(12)
			muted("·")
			ui.Link(c, w.aiLogo, aiLogoURL).FontSize(12)
		}
	})
}
