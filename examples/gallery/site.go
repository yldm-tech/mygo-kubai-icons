package main

import (
	"fmt"
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
	upstreamURL = "https://github.com/kubai087/kubai-icons"
	// compactWidth is the window width below which the sidebar folds into a drop-down and the details open in a dialog.
	compactWidth = 1000
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

	copied string // what was copied last, the install command or a Go call

	shown  []all.Entry
	shownK string
	counts map[string]int
	countK bool
}

func newGallery() *gallery { return &gallery{selIdx: -1} }

// filtered returns the icons the search, the category and the base-only switch leave.
func (g *gallery) filtered() []all.Entry {
	q := strings.ToLower(strings.TrimSpace(g.query))
	key := fmt.Sprintf("%s\x00%s\x00%v", q, g.category, g.basesOnly)
	if g.shown != nil && key == g.shownK {
		return g.shown
	}
	g.shown = g.shown[:0]
	for _, e := range all.Entries {
		if g.category != "" && e.Category != g.category {
			continue
		}
		if g.basesOnly && e.Name != e.Base {
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
	if g.counts != nil && g.countK == g.basesOnly {
		return g.counts
	}
	g.counts = map[string]int{}
	for _, e := range all.Entries {
		if g.basesOnly && e.Name != e.Base {
			continue
		}
		g.counts[e.Category]++
		g.counts[""]++
	}
	g.countK = g.basesOnly
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
		g.header(c, w, compact)
		ui.Row(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
			if !compact {
				g.sidebar(c, w)
			}
			ui.Column(c).Grow(1).Children(func() {
				g.toolbar(c, w, compact)
				g.gridView(c, w, entries, compact)
			})
			if !compact {
				ui.Column(c).Width(340).BorderWidth(0, 0, 0, 1).BorderColor(t.Border).Background(t.Surface).Children(func() {
					g.detail(c, w)
				})
			}
		})
		g.footer(c, w, len(entries))
	})
	if compact {
		if _, ok := g.chosen(); ok && g.dialog {
			ui.Modal(c, &g.dialog, func() {
				ui.Column(c).Width(min(width-32, 420)).Children(func() { g.detail(c, w) })
			})
		}
	}
}

func (g *gallery) header(c *ui.Context, w words, compact bool) {
	t := c.Theme()
	ui.Row(c).Padding(14, 20).Gap(14).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
		kubai.Shapes().View(c, kubai.Options{Size: 34, Label: w.title})
		ui.Column(c).Gap(2).Shrink(1).Children(func() {
			ui.Text(c, w.title).FontSize(18).Bold()
			if !compact {
				ui.Text(c, w.tagline).FontSize(12).TextColor(t.TextMuted)
			}
		})
		ui.Spacer(c)
		if !compact {
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
	ui.Scroll(c).Width(232).BorderWidth(0, 1, 0, 0).BorderColor(t.Border).Children(func() {
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
					o := kubai.Options{Size: 20, Color: accents[g.accent].color, Variant: kubai.Variant(g.variant), NoShadow: true, Hovered: row.Hovered()}
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
		ui.Row(c).Gap(14).AlignItems(ui.Center).Children(func() {
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
				ui.Switch(c, &g.basesOnly).Label(w.basesOnly)
				ui.Text(c, w.basesOnly).FontSize(13)
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
	g.grid.Selected = &g.selIdx
	opts := g.options()
	cell := float32(g.size) + 60
	grid := ui.GridView(c, &g.grid, len(entries), cell+28, cell, func(i int) {
		e := entries[i]
		card := ui.Column(c.Key(e.Name)).Center().Gap(10).Padding(8).Radius(10).Cursor(ui.CursorPointer).Tooltip(e.Name)
		switch {
		case i == g.selIdx:
			card.Border(1.5, t.Accent) // the grid paints the chosen item's background itself
		case card.Hovered():
			card.Background(t.SurfaceHover)
		}
		card.Children(func() {
			o := opts
			o.Hovered = g.hoverAll || card.Hovered()
			e.Icon().View(c, o)
			ui.Text(c, w.label(e)).FontSize(11).TextColor(t.TextMuted).MaxLines(1)
		})
	}).Grow(1).Padding(4, 12)
	// Bring the chosen icon into view. The grid counts its columns from the window's width until it was laid out once, and when the count changes it scrolls to keep its rows; so the first frame leaves this for the next, and the request comes after the grid's own, which it overrides.
	if g.scroll && g.laidOut {
		if g.selIdx >= 0 {
			g.grid.ScrollTo(g.selIdx, ui.Center)
		}
		g.scroll = false
	} else if g.scroll {
		c.After(16 * time.Millisecond)
	}
	g.laidOut = true
	if grid.Changed() && g.selIdx >= 0 && g.selIdx < len(entries) {
		g.selName, g.active = entries[g.selIdx].Name, false
		g.dialog = compact
	}
}

func (g *gallery) footer(c *ui.Context, w words, n int) {
	t := c.Theme()
	ui.Row(c).Padding(8, 20).Gap(6).AlignItems(ui.Center).BorderWidth(1, 0, 0, 0).BorderColor(t.Border).Children(func() {
		status := w.hint
		if g.copied != "" {
			status = fmt.Sprintf(w.copiedCall, g.copied)
		}
		ui.Textf(c, w.count+" · %s", n, kubai.Count, status).FontSize(12).TextColor(t.TextMuted).MaxLines(1).Shrink(1)
		ui.Spacer(c)
		muted := func(s string) { ui.Text(c, s).FontSize(12).TextColor(t.TextMuted) }
		muted(w.credit)
		ui.Link(c, w.upstream, upstreamURL).FontSize(12)
		muted("(" + w.mit + ") · " + w.unofficial + " · " + w.drawnBy)
	})
}
