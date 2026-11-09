// Gallery shows every icon in a native MyGo window: search by name, pick an accent, the soft look and a size, and hover an icon to play it. A click copies its Go call.
//
//	go run ./examples/gallery
//	go run ./examples/gallery -capture gallery.png -query folder
package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/yldm-tech/mygo-kubai-icons/all"
	"github.com/yldm-tech/mygo-kubai-icons/kubai"
)

var accents = []struct {
	name  string
	color ui.Color
}{
	{"Blue", kubai.DefaultColor},
	{"Violet", ui.Hex("#8B5CF6")},
	{"Pink", ui.Hex("#EC4899")},
	{"Orange", ui.Hex("#F97316")},
	{"Green", ui.Hex("#22C55E")},
	{"Slate", ui.Hex("#64748B")},
}

type gallery struct {
	query    string
	accent   int
	variant  int
	size     float64
	grid     ui.GridState
	copied   string
	shown    []all.Entry
	shownQ   string
	hoverAll bool
}

func (g *gallery) filtered() []all.Entry {
	q := strings.ToLower(strings.TrimSpace(g.query))
	if g.shown != nil && q == g.shownQ {
		return g.shown
	}
	g.shown = g.shown[:0]
	for _, e := range all.Entries {
		if q == "" || strings.Contains(e.Name, q) || strings.Contains(strings.ToLower(e.GoName), q) || strings.Contains(e.Category, q) {
			g.shown = append(g.shown, e)
		}
	}
	g.shownQ = q
	return g.shown
}

func (g *gallery) view(c *ui.Context) {
	t := c.Theme()
	entries := g.filtered()
	opts := kubai.Options{Size: float32(g.size), Color: accents[g.accent].color, Variant: kubai.Variant(g.variant), SemanticBadge: true}
	ui.Column(c).Fill().Children(func() {
		ui.Row(c).Gap(12).Padding(12, 16).Children(func() {
			ui.SearchField(c, &g.query).Label("Search icons").Grow(1)
			ui.Segmented(c, &g.variant, "Color", "Soft").Label("Look")
			ui.Row(c).Gap(6).Children(func() {
				for i, a := range accents {
					sw := ui.Box(c.Key(a.name)).Size(20, 20).Radius(10).Background(a.color).Label(a.name).Cursor(ui.CursorPointer)
					if i == g.accent {
						sw.Border(2, t.Text)
					}
					if sw.Clicked() {
						g.accent = i
					}
				}
			})
			ui.Slider(c, &g.size, 16, 64).Label("Size").Width(120)
		})
		cell := float32(g.size) + 56
		ui.GridView(c, &g.grid, len(entries), cell+24, cell, func(i int) {
			e := entries[i]
			card := ui.Column(c.Key(e.Name)).Center().Gap(10).Padding(8).Radius(10).Cursor(ui.CursorPointer)
			if card.Hovered() {
				card.Background(t.SurfaceHover)
			}
			if card.Clicked() {
				c.WriteClipboard("kubai." + e.GoName + "()")
				g.copied = "Copied kubai." + e.GoName + "()"
			}
			card.Children(func() {
				o := opts
				o.Hovered = g.hoverAll || card.Hovered()
				e.Icon().View(c, o)
				ui.Text(c, e.Name).FontSize(11).TextColor(t.TextMuted).MaxLines(1)
			})
		}).Grow(1).Padding(0, 12)
		ui.Row(c).Padding(8, 16).Children(func() {
			status := g.copied
			if status == "" {
				status = "Hover an icon to play it; click to copy its Go call."
			}
			ui.Textf(c, "%d of %d icons · %s", len(entries), kubai.Count, status).FontSize(12).TextColor(t.TextMuted)
		})
	})
}

func main() {
	capture := flag.String("capture", "", "write a PNG of the window to this path once it shows, then quit")
	query := flag.String("query", "", "start with this search")
	hover := flag.Bool("hover", false, "hold every icon at the end of its hover")
	soft := flag.Bool("soft", false, "start with the soft look")
	size := flag.Float64("size", 32, "start with this icon size")
	flag.Parse()
	g := &gallery{query: *query, size: *size, hoverAll: *hover}
	if *soft {
		g.variant = 1
	}
	mygo.App.WhenReady(func() {
		w := mygo.NewWindow(mygo.WindowOptions{
			Title:   "Kubai Icons for MyGo",
			Width:   960,
			Height:  720,
			Content: ui.View(g.view),
		})
		if *capture != "" {
			go func() {
				time.Sleep(1500 * time.Millisecond)
				png, err := w.CapturePage()
				if err == nil {
					err = os.WriteFile(*capture, png, 0o644)
				}
				if err != nil {
					log.Print(err)
				}
				mygo.App.Quit()
			}()
		}
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
