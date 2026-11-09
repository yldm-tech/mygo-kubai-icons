// Command glyphs writes gen/glyphs.json, the outlines gen/generate.mjs turns the text of a few icons into, since MyGo does not draw SVG text. Latin letters, digits and punctuation come from Go Bold (golang.org/x/image/font/gofont, BSD licensed); 文, which it lacks, is drawn here in strokes of the same weight.
//
//	go run ./gen/glyphs > gen/glyphs.json
//
// Each glyph is its advance and its outline as SVG path data, in thousandths of an em, the baseline at y=0 and y growing down.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

const em = 1000

type glyph struct {
	Advance float64 `json:"advance"`
	D       string  `json:"d"`
}

func main() {
	f, err := sfnt.Parse(gobold.TTF)
	if err != nil {
		log.Fatal(err)
	}
	var b sfnt.Buffer
	ppem := fixed.I(em)
	out := map[string]glyph{}
	for r := rune(0x21); r < 0x7f; r++ {
		i, err := f.GlyphIndex(&b, r)
		if err != nil || i == 0 {
			continue
		}
		adv, err := f.GlyphAdvance(&b, i, ppem, font.HintingNone)
		if err != nil {
			log.Fatal(err)
		}
		segs, err := f.LoadGlyph(&b, i, ppem, nil)
		if err != nil {
			log.Fatal(err)
		}
		var d strings.Builder
		for _, s := range segs {
			switch s.Op {
			case sfnt.SegmentOpMoveTo:
				fmt.Fprintf(&d, "M%s", pts(s.Args[:1]))
			case sfnt.SegmentOpLineTo:
				fmt.Fprintf(&d, "L%s", pts(s.Args[:1]))
			case sfnt.SegmentOpQuadTo:
				fmt.Fprintf(&d, "Q%s", pts(s.Args[:2]))
			case sfnt.SegmentOpCubeTo:
				fmt.Fprintf(&d, "C%s", pts(s.Args[:3]))
			}
		}
		if d.Len() > 0 {
			d.WriteString("Z")
		}
		out[string(r)] = glyph{Advance: num(adv), D: d.String()}
	}
	out[" "] = glyph{Advance: 280}
	out["文"] = wen()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	if err := enc.Encode(out); err != nil {
		log.Fatal(err)
	}
}

// wen draws 文 in strokes about as heavy as Go Bold's stems: a dot, a bar, and two falling strokes that bend as they leave the bar and cross below it.
func wen() glyph {
	const w = 120
	strokes := [][][2]float64{
		{{500, -840}, {530, -730}},            // dot
		{{70, -610}, {930, -610}},             // bar
		{{650, -610}, {560, -330}, {110, 30}}, // falling left
		{{330, -610}, {440, -330}, {910, 30}}, // falling right
	}
	var d strings.Builder
	for _, s := range strokes {
		for i := 1; i < len(s); i++ {
			a, z := s[i-1], s[i]
			dx, dy := z[0]-a[0], z[1]-a[1]
			l := math.Hypot(dx, dy)
			nx, ny := -dy/l*w/2, dx/l*w/2
			fmt.Fprintf(&d, "M%s %sL%s %sL%s %sL%s %sZ", g(a[0]+nx), g(a[1]+ny), g(z[0]+nx), g(z[1]+ny), g(z[0]-nx), g(z[1]-ny), g(a[0]-nx), g(a[1]-ny))
		}
	}
	return glyph{Advance: 1000, D: d.String()}
}

func g(v float64) string { return fmt.Sprint(math.Round(v*10) / 10) }

func pts(ps []fixed.Point26_6) string {
	s := make([]string, len(ps))
	for i, p := range ps {
		s[i] = fmt.Sprintf("%g %g", num(p.X), num(p.Y))
	}
	return strings.Join(s, " ")
}

func num(v fixed.Int26_6) float64 { return math.Round(float64(v)/64*10) / 10 }
