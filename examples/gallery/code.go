package main

import (
	"regexp"
	"strings"

	"github.com/egoist/mygo/ui"
)

// codePalette colours Go as GitHub's light and dark themes do.
type codePalette struct {
	plain, pkg, fn, typ, field, str, num, keyword ui.Color
}

var (
	lightCode = codePalette{
		plain: ui.Hex("#1F2328"), pkg: ui.Hex("#57606A"), fn: ui.Hex("#8250DF"), typ: ui.Hex("#953800"),
		field: ui.Hex("#0550AE"), str: ui.Hex("#0A3069"), num: ui.Hex("#0550AE"), keyword: ui.Hex("#CF222E"),
	}
	darkCode = codePalette{
		plain: ui.Hex("#E6EDF3"), pkg: ui.Hex("#8B949E"), fn: ui.Hex("#D2A8FF"), typ: ui.Hex("#FFA657"),
		field: ui.Hex("#79C0FF"), str: ui.Hex("#A5D6FF"), num: ui.Hex("#79C0FF"), keyword: ui.Hex("#FF7B72"),
	}
)

var goToken = regexp.MustCompile(`"[^"]*"|[0-9]+(?:\.[0-9]+)?|[A-Za-z_][A-Za-z0-9_]*|\s+|.`)

// packages are the package names the gallery's code uses.
var packages = map[string]bool{"kubai": true, "ui": true}

// highlight splits the Go code the gallery writes into spans coloured by what each token is: a package, a function or method called, a type or constant, a field, a string, a number or true and false.
func highlight(code string, pal codePalette) []ui.Span {
	toks := goToken.FindAllString(code, -1)
	next := func(i int) string {
		for j := i + 1; j < len(toks); j++ {
			if strings.TrimSpace(toks[j]) != "" {
				return toks[j]
			}
		}
		return ""
	}
	spans := make([]ui.Span, 0, len(toks))
	for i, tok := range toks {
		col := pal.plain
		switch {
		case strings.HasPrefix(tok, `"`):
			col = pal.str
		case tok[0] >= '0' && tok[0] <= '9':
			col = pal.num
		case tok == "true" || tok == "false":
			col = pal.keyword
		case isIdent(tok):
			switch n := next(i); {
			case packages[tok] && n == ".":
				col = pal.pkg
			case n == "(":
				col = pal.fn
			case n == ":":
				col = pal.field
			case tok[0] >= 'A' && tok[0] <= 'Z':
				col = pal.typ
			}
		}
		// Neighbours of one colour join, so that the text has few spans.
		if k := len(spans) - 1; k >= 0 && spans[k].Color == col {
			spans[k].Text += tok
			continue
		}
		spans = append(spans, ui.Span{Text: tok, Color: col})
	}
	return spans
}

func isIdent(s string) bool {
	c := s[0]
	return c == '_' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z'
}
