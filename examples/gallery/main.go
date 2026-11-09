// Gallery shows every icon in a native MyGo window: search by name, pick an accent, the soft look and a size, and hover an icon to play it. A click copies its Go call. It speaks Chinese when the system or browser does, and English otherwise.
//
//	go run ./examples/gallery
//	go run ./examples/gallery -capture gallery.png -query folder
package main

import (
	"flag"
	"fmt"
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

// words is what the gallery says, in one language.
type words struct {
	search, look, color, soft, size string
	accents                         []string
	hint, copied, count             string
}

var english = words{
	search: "Search icons", look: "Look", color: "Color", soft: "Soft", size: "Size",
	accents: []string{"Blue", "Violet", "Pink", "Orange", "Green", "Slate"},
	hint:    "Hover an icon to play it; click to copy its Go call.",
	copied:  "Copied %s",
	count:   "%d of %d icons · %s",
}

var chinese = words{
	search: "搜索图标", look: "风格", color: "彩色", soft: "清透", size: "尺寸",
	accents: []string{"蓝", "紫", "粉", "橙", "绿", "灰"},
	hint:    "悬停播放动画，点击复制 Go 调用代码。",
	copied:  "已复制 %s",
	count:   "%d / %d 个图标 · %s",
}

// say picks the language the app shows: Chinese for any zh locale, English otherwise.
func say(c *ui.Context) words {
	if strings.HasPrefix(c.Locale(), "zh") {
		return chinese
	}
	return english
}

type gallery struct {
	query    string
	accent   int
	variant  int
	size     float64
	grid     ui.GridState
	copied   string // the Go call last copied
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
		if q == "" || strings.Contains(e.Name, q) || strings.Contains(strings.ToLower(e.GoName), q) || strings.Contains(e.Zh, q) || strings.Contains(e.Category, q) {
			g.shown = append(g.shown, e)
		}
	}
	g.shownQ = q
	return g.shown
}

func (g *gallery) view(c *ui.Context) {
	t := c.Theme()
	w := say(c)
	entries := g.filtered()
	opts := kubai.Options{Size: float32(g.size), Color: accents[g.accent].color, Variant: kubai.Variant(g.variant), SemanticBadge: true}
	ui.Column(c).Fill().Children(func() {
		ui.Row(c).Gap(12).Padding(12, 16).Children(func() {
			ui.SearchField(c, &g.query).Label(w.search).Placeholder(w.search).Grow(1)
			ui.Segmented(c, &g.variant, w.color, w.soft).Label(w.look)
			ui.Row(c).Gap(6).Children(func() {
				for i, a := range accents {
					sw := ui.Box(c.Key(a.key)).Size(20, 20).Radius(10).Background(a.color).Label(w.accents[i]).Cursor(ui.CursorPointer)
					if i == g.accent {
						sw.Border(2, t.Text)
					}
					if sw.Clicked() {
						g.accent = i
					}
				}
			})
			ui.Slider(c, &g.size, 16, 64).Label(w.size).Width(120)
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
				g.copied = "kubai." + e.GoName + "()"
			}
			card.Children(func() {
				o := opts
				o.Hovered = g.hoverAll || card.Hovered()
				e.Icon().View(c, o)
				ui.Text(c, e.Name).FontSize(11).TextColor(t.TextMuted).MaxLines(1)
			})
		}).Grow(1).Padding(0, 12)
		ui.Row(c).Padding(8, 16).Children(func() {
			status := w.hint
			if g.copied != "" {
				status = fmt.Sprintf(w.copied, g.copied)
			}
			ui.Textf(c, w.count, len(entries), kubai.Count, status).FontSize(12).TextColor(t.TextMuted)
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
	// The gallery is written in English and Chinese; MyGo shows the one the user reads, in its own controls too. MYGO_LOCALE=en forces one.
	mygo.App.SetLanguages("en", "zh-Hans")
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
