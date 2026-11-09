package kubai

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
)

// Every colour in an icon is written against the reference blue #5393FF and moved to the chosen accent by its relative lightness, as upstream's color.js does.

var ref = [3]float64{83, 147, 255}

// DefaultColor is the accent icons take when Options.Color is unset.
var DefaultColor = ui.Hex("#5393FF")

type rgb [3]float64

func parseHex(h string) rgb {
	var c rgb
	for i := range 3 {
		v, _ := strconv.ParseUint(h[1+2*i:3+2*i], 16, 8)
		c[i] = float64(v)
	}
	return c
}

func (c rgb) hex() string {
	var b [3]uint8
	for i, v := range c {
		b[i] = uint8(math.Max(0, math.Min(255, math.Round(v))))
	}
	return fmt.Sprintf("#%02X%02X%02X", b[0], b[1], b[2])
}

func colorRGB(c ui.Color) rgb { return rgb{float64(c.R), float64(c.G), float64(c.B)} }

// tint moves the reference colour src to the accent sel.
func tint(src string, sel rgb) string {
	s := parseHex(src)
	var out rgb
	for i := range 3 {
		var amt float64
		if i == 2 {
			amt = (s[0] - ref[0]) / (255 - ref[0])
		} else {
			amt = (s[i] - ref[i]) / (255 - ref[i])
		}
		if amt >= 0 {
			out[i] = sel[i] + (255-sel[i])*amt
		} else {
			out[i] = sel[i] * (1 + amt)
		}
	}
	return out.hex()
}

// soften lifts a colour toward white by k.
func soften(c rgb, k float64) rgb {
	for i, v := range c {
		c[i] = v + (255-v)*k
	}
	return c
}

// paint is the resolved colours of one accent, variant and badge ink.
type paint struct {
	c, e  rgb
	deep  bool
	badge string // a hex ink for every badge, or "" for each badge's own default
}

func newPaint(accent ui.Color, v Variant, badge string) paint {
	a := colorRGB(accent)
	if v == Soft {
		var e rgb
		for i, x := range a {
			e[i] = x * 0.84
		}
		return paint{c: soften(a, 0.1), e: e, deep: true, badge: badge}
	}
	return paint{c: a, e: a, badge: badge}
}

var (
	tintToken = regexp.MustCompile(`\{t(#[0-9A-Fa-f]{6})(?:b(#[0-9A-Fa-f]{6})|([ce]))\}`)
	pureToken = regexp.MustCompile(`\{p(#[0-9A-Fa-f]{6})\}`)
)

// fill resolves the colour tokens of generated markup, as upstream's fillTemplate does: {t<src>c} and {t<src>e} tint a reference colour to the fill or the edge accent, {t<src>b<ink>} to a badge's ink, {p<base>} is the light top of a pure gradient, {w} the white of marks and {S} the shadow.
func (p paint) fill(s string) string {
	if !strings.Contains(s, "{") {
		return s
	}
	s = tintToken.ReplaceAllStringFunc(s, func(m string) string {
		g := tintToken.FindStringSubmatch(m)
		switch {
		case g[2] != "":
			ink := g[2]
			if p.badge != "" {
				ink = p.badge
			}
			return tint(g[1], parseHex(ink))
		case g[3] == "e":
			return tint(g[1], p.e)
		default:
			return tint(g[1], p.c)
		}
	})
	s = pureToken.ReplaceAllStringFunc(s, func(m string) string {
		base := m[2:9]
		if p.deep {
			return soften(parseHex(base), 0.72).hex()
		}
		return base
	})
	white := "#FFFFFF"
	if p.deep {
		white = tint("#2F6FE0", p.e)
	}
	s = strings.ReplaceAll(s, "{w}", white)
	return strings.ReplaceAll(s, "{S}", tint("#8AB6FF", p.c))
}
