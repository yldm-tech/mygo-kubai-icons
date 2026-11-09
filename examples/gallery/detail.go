package main

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/yldm-tech/mygo-kubai-icons/all"
	"github.com/yldm-tech/mygo-kubai-icons/kubai"
)

// detail shows the chosen icon: large, its names, the Go code that shows it as the toolbar sets it, and the other versions of its base icon.
func (g *gallery) detail(c *ui.Context, w words) {
	t := c.Theme()
	e, ok := g.chosen()
	ui.Scroll(c).Grow(1).Children(func() {
		ui.Column(c).Padding(20).Gap(16).Children(func() {
			if !ok {
				ui.Column(c).Gap(14).Padding(40, 8).Center().Children(func() {
					kubai.Shapes().View(c, kubai.Options{Size: 72, Color: accents[g.accent].color, Variant: kubai.Variant(g.variant)})
					ui.Text(c, w.pick).FontSize(13).TextColor(t.TextMuted).Wrap()
				})
				return
			}
			ic := e.Icon()

			ui.Column(c).Gap(4).Children(func() {
				ui.Text(c, w.label(e)).FontSize(18).Bold().Wrap()
				ui.Text(c, w.categoryName(e.Category)).FontSize(12).TextColor(t.TextMuted)
			})

			stage := ui.Column(c).Height(220).Center().Gap(14).Radius(14).Background(t.Background).Border(1, t.Border)
			stage.Children(func() {
				o := g.options()
				o.Size = 112
				o.Hovered = stage.Hovered()
				o.Active = g.active
				o.Label = w.label(e)
				ic.View(c, o)
				ui.Text(c, w.previewHint).FontSize(11).TextColor(t.TextMuted)
			})
			if ic.Tap() {
				ui.Toggle(c, &g.active, w.active)
			}

			ui.Column(c).Gap(8).Children(func() {
				row := func(k, v string, mono bool) {
					ui.Row(c).Gap(10).AlignItems(ui.Baseline).Children(func() {
						ui.Text(c, k).FontSize(12).TextColor(t.TextMuted).Width(52)
						txt := ui.Text(c, v).FontSize(12).Selectable().Wrap().Shrink(1)
						if mono {
							txt.Font("monospace")
						}
					})
				}
				row(w.name, e.Name, true)
				row(w.function, "kubai."+e.GoName+"()", true)
				row(w.inCategory, w.categoryName(e.Category), false)
			})

			code := g.code(e, ic)
			ui.Column(c).Gap(8).Children(func() {
				ui.Row(c).AlignItems(ui.Center).Children(func() {
					ui.Text(c, w.code).FontSize(12).Bold()
					ui.Spacer(c)
					label := w.copyCode
					if g.copied == code {
						label = w.copied
					}
					if ui.PrimaryButton(c, label).Clicked() {
						g.copy(c, code)
					}
				})
				ui.Box(c).Padding(12).Radius(10).Background(t.Background).Border(1, t.Border).Children(func() {
					ui.Text(c, code).Font("monospace").FontSize(12).Selectable().Wrap()
				})
			})

			versions := g.versions(e)
			if len(versions) > 1 {
				ui.Column(c).Gap(8).Children(func() {
					ui.Textf(c, "%s · %d", w.variants, len(versions)).FontSize(12).Bold()
					ui.Grid(c).Columns(5).Gap(6).Children(func() {
						for _, v := range versions {
							cell := ui.Column(c.Key("v-" + v.Name)).Center().Padding(8).Radius(8).Cursor(ui.CursorPointer).Tooltip(w.label(v))
							switch {
							case v.Name == e.Name:
								cell.Background(t.SurfacePressed).Border(1, t.Accent)
							case cell.Hovered():
								cell.Background(t.SurfaceHover)
							}
							if cell.Clicked() {
								g.choose(v.Name)
							}
							cell.Children(func() {
								o := g.options()
								o.Size = 28
								o.Hovered = cell.Hovered()
								v.Icon().View(c, o)
							})
						}
					})
				})
			}
		})
	})
}

// code is the Go call that shows the icon with the toolbar's settings, leaving out the defaults.
func (g *gallery) code(e all.Entry, ic *kubai.Icon) string {
	var opts []string
	if g.size != 24 {
		opts = append(opts, fmt.Sprintf("Size: %.0f", g.size))
	}
	if g.accent != 0 {
		c := accents[g.accent].color
		opts = append(opts, fmt.Sprintf(`Color: ui.Hex("#%02X%02X%02X")`, c.R, c.G, c.B))
	}
	if g.variant == 1 {
		opts = append(opts, "Variant: kubai.Soft")
	}
	if ic.Tone() != "" {
		opts = append(opts, "SemanticBadge: true")
	}
	if ic.Tap() && g.active {
		opts = append(opts, "Active: true")
	}
	return fmt.Sprintf("kubai.%s().View(c, kubai.Options{%s})", e.GoName, strings.Join(opts, ", "))
}

// versions lists the base icon of e and every composite of it.
func (g *gallery) versions(e all.Entry) []all.Entry {
	var out []all.Entry
	for _, v := range all.Entries {
		if v.Base == e.Base {
			out = append(out, v)
		}
	}
	return out
}
