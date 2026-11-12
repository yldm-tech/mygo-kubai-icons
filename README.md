<p align="center"><img src=".github/assets/logo.png" width="128" alt="" /></p>

# Kubai Icons for MyGo

[![CI](https://github.com/yldm-tech/mygo-kubai-icons/actions/workflows/ci.yml/badge.svg)](https://github.com/yldm-tech/mygo-kubai-icons/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/yldm-tech/mygo-kubai-icons/kubai.svg)](https://pkg.go.dev/github.com/yldm-tech/mygo-kubai-icons/kubai)
[![Release](https://img.shields.io/github/v/release/yldm-tech/mygo-kubai-icons)](https://github.com/yldm-tech/mygo-kubai-icons/releases)
[![X](https://img.shields.io/badge/X-%40suzukaze__haduki-000000?logo=x&logoColor=white)](https://x.com/suzukaze_haduki)
[![GitHub Sponsors](https://img.shields.io/badge/GitHub%20Sponsors-ea4aaa?logo=githubsponsors&logoColor=white)](https://github.com/sponsors/houko)

English · [简体中文](README.zh-CN.md)

The 2,901 icons of [Kubai Icons](https://icons.kubai.design/) for [MyGo](https://github.com/egoist/mygo) native UI, moving on hover as they do on the web: a folder opens, an envelope unseals, a calendar turns to the next day. MyGo draws them itself, with no web view.

![Icons playing their hover motion, drawn by MyGo](.github/assets/motion.gif)

This is an unofficial port. Kubai Icons is designed by Kubai087; the icons are MIT licensed, and the Kubai name is theirs.

## Install

```sh
go get github.com/yldm-tech/mygo-kubai-icons@latest
```

Requires Go 1.27.3 and MyGo v0.8.3 or later, as MyGo does. Versions follow semantic versioning, and each [release](https://github.com/yldm-tech/mygo-kubai-icons/releases) lists what changed; `go get github.com/yldm-tech/mygo-kubai-icons@latest` again upgrades.

## Use

Each icon is a function returning a shared `*kubai.Icon`. `View` shows it and plays its motion while the pointer is over it:

```go
import (
	"github.com/egoist/mygo/ui"
	"github.com/yldm-tech/mygo-kubai-icons/kubai"
)

func sidebar(c *ui.Context) {
	ui.Row(c).Gap(8).Children(func() {
		kubai.Folder().View(c, kubai.Options{Size: 24})
		ui.Text(c, "Documents")
	})
}
```

Names follow upstream's ids in MixedCaps: `folder-plus` is `FolderPlus`, `wifi-off` is `WifiOff`. Every base icon is followed by its badged composites (`FolderCheck`, `BellX`) and, where upstream has one, a slashed `-off` version.

### Options

| Field | Effect |
|---|---|
| `Size` | Side of the icon's canvas in DIPs, 24 by default. The shadow reaches a little past it. |
| `Color` | The accent every colour of the icon derives from, `kubai.DefaultColor` (#5393FF) by default. |
| `Variant` | `kubai.Color`, or `kubai.Soft` for the clear look. |
| `BadgeColor`, `SemanticBadge` | Ink of a corner badge's disc, or each badge's own colour: green for add, red for remove. |
| `Hovered` | Plays the motion as if hovered, so a parent can drive it: `Options{Hovered: row.Hovered()}`. |
| `Still` | Ignores the pointer. |
| `Active` | The toggled state of an icon a click toggles (`Icon.Tap()`), such as a saved bookmark. |
| `NoShadow` | Leaves out the drop shadow. |
| `Label` | Names the icon for assistive technology; without one it is a decoration. |

Motion follows MyGo's reduced-motion preference.

For places that take an SVG rather than an element, such as a menu or tray icon, `Icon.SVG(options, hovered)` returns a `*ui.SVG` of the rest pose or the end of the hover, without the shadow.

### Binary size

Unused icons are left out of the executable. A base icon and its composites share one compressed blob, 1.2 KB at the median and under 16 KB at most, decoded the first time one of them is shown. Package `all` lists every icon by name for galleries and pickers; importing it keeps all of them, which made a test program 3.3 MB larger.

## Gallery

[Try it in your browser](https://kubai.yldm.ai/): the gallery example compiled to WebAssembly, drawn by MyGo as on the desktop, a download of about 5.5 MB. Or run it natively:

```sh
go run github.com/yldm-tech/mygo-kubai-icons/examples/gallery@latest
```

It starts on the base icons, with each one's badged and slashed versions in its details; a search looks through every version, and the switch shows them all in the grid. Browse by category, search by name (Chinese names too, `all.Entry.Zh`), pick an accent, the soft look and a size, and hover an icon to play it. The icons in view also play on their own, one after another, and choosing one, with a click, a tap or the arrows, plays it once. Click one for a large preview that plays its motion on its own, its Go code with those settings and its badged and slashed versions. It speaks Chinese or English after the system or browser, and holds still where the system asks for less motion. On the site, a link can open an icon or a search: [`?icon=folder-plus`](https://kubai.yldm.ai/?icon=folder-plus), `?q=文件夹`, `?soft`, `?size=48`, `?all` for every version.

![The gallery example](.github/assets/gallery.png)

## How it works

Upstream icons are JavaScript: each declares its layers, a rest and a hover target, and a `draw()` that recomputes its geometry from numbers interpolated between the two. Rather than translate 900 `draw()` functions, `gen/generate.mjs` runs upstream's own renderer at keyframes along each motion and stores the markup as a template with numeric slots. Keyframes are placed adaptively: an interval is split until interpolating its numbers in a straight line stays within 0.04 canvas units (about 0.06 px at 48 px) of what `draw()` gives, so a simple slide needs five keyframes and a damped shake about forty. Swings and pulses, which upstream plays only on the way in, get a separate track for the way back.

At run time, `View` moves progress along upstream's curve (460 ms, `cubic-bezier(.22, 1, .36, 1)`) as it paints, in 64 steps, and asks only to be painted again while it moves (`Painter.AnimationFrame`): an icon playing does not build the view again for each frame, however much else the window holds, and a motion playing again finds the bitmaps it drew before. It interpolates the slots, resolves the colours with a port of upstream's colour derivation and draws the markup with MyGo's SVG renderer. A frame the cache does not hold takes about 0.08 ms to build and parse.

## Differences from upstream

MyGo's SVG renderer draws no filters and no text, so:

- The drop shadow is cast from the whole icon's silhouette, where upstream casts it from the parts it marks: blurred, at upstream's offset and opacity, and drawn with the icon into one bitmap per pose, so that a screen of icons costs one texture each. It is close, not identical.
- The faint inner highlight along the top edge of bodies is left out.
- The text in seven icons (`calendar`, `circle-help`, `languages`, `list-ordered`, `parking`, `quiz`, `spell-check` and their composites) is drawn as glyph outlines from Go Bold, which differs a little from the system font upstream uses; the one CJK glyph, 文, is drawn by hand in strokes of the same weight.
- Keyboard focus does not play the motion by itself; pass `Hovered` from the focused control.

To compare the two, `internal/compare/compare.sh` renders icons with upstream's code through `rsvg-convert` and with this port through MyGo, side by side.

## Layout

| Path | What it holds |
|---|---|
| `kubai/` | The package: the runtime (`kubai.go`, `data.go`, `color.go`) and the generated icons (`icons_gen.go`). |
| `all/` | Every icon by name, for galleries and pickers. |
| `examples/gallery/` | The native gallery. |
| `site/` | The page that loads the gallery as WebAssembly on GitHub Pages; `.github/workflows/pages.yml` builds and publishes it. |
| `gen/` | The generator, the glyph table and the pinned upstream release. |
| `internal/compare/` | Scripts that render this port and upstream side by side, and the motion GIF. |
| `internal/brand/` | Draws the logo and the social preview with this package: `go run ./internal/brand`. |
| `licenses/` | Licences of the icons and the Go fonts. |

## Regenerate

```sh
git clone https://github.com/kubai087/kubai-icons ../kubai-icons
git -C ../kubai-icons checkout kubai-icons@0.4.1   # or a newer release
node gen/generate.mjs --upstream ../kubai-icons
go test ./...
```

Update `gen/upstream.json` to the release generated from; CI checks the generated files against it with `--check`. `go run ./gen/glyphs > gen/glyphs.json` rebuilds the glyph table.

## Release

Merging to `main` releases by itself once CI passes (`.github/workflows/release.yml`, configured in `.releaserc.json`): semantic-release reads the [conventional commits](https://www.conventionalcommits.org/) since the last tag, makes a patch release for `fix` and `perf`, a minor one for `feat` and a major one for a breaking change, tags `vX.Y.Z` and publishes the notes as a GitHub Release. Other types, such as `docs`, `ci` and `chore`, release nothing.

## License

MIT for this module's code. The icons are from Kubai Icons, MIT licensed by Kubai087 and the Kubai Icons contributors ([licenses/kubai-icons.txt](licenses/kubai-icons.txt)). The glyph outlines come from the Go fonts by Bigelow & Holmes ([licenses/go-fonts.txt](licenses/go-fonts.txt)).
