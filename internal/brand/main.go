// Command brand draws this repository's logo and its GitHub social preview with MyGo and this package, into .github/assets.
//
//	go run ./internal/brand
//
// The logo is the set's own shapes icon (MIT, like every icon here), not a Kubai mark: the Kubai name and logos are not licensed.
package main

import (
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/egoist/mygo/ui"
	"github.com/yldm-tech/mygo-kubai-icons/kubai"
)

const out = ".github/assets"

func main() {
	write("logo.png", logo())
	write("social-preview.png", preview())
}

// logo is the shapes icon at rest, 512 px square with room for its shadow.
func logo() *image.RGBA {
	const side, size = 256, 200
	return ui.Render(func(c *ui.Context) {
		ui.Box(c).Size(side, side).Children(func() {
			ui.Box(c).Absolute().Left((side-size)/2).Top((side-size)/2-6).Size(size, size).Children(func() {
				kubai.Shapes().View(c, kubai.Options{Size: size, Still: true})
			})
		})
	}, side, side, 2)
}

// preview is GitHub's 1280×640 social preview: the logo, the name, one line on what it is and a row of icons.
func preview() *image.RGBA {
	const w, h = 640, 320
	row := []*kubai.Icon{kubai.FolderOpen(), kubai.MailOpen(), kubai.Calendar(), kubai.BellCheck(), kubai.LockOpen(), kubai.Rocket(), kubai.Camera(), kubai.Send()}
	return ui.Render(func(c *ui.Context) {
		ui.Row(c).Size(w, h).Background(ui.Hex("#F4F7FF")).Padding(0, 44).Gap(36).Center().Children(func() {
			kubai.Shapes().View(c, kubai.Options{Size: 132, Still: true})
			ui.Column(c).Gap(10).Children(func() {
				ui.Text(c, "Kubai Icons for MyGo").FontSize(34).Bold().TextColor(ui.Hex("#1B2440"))
				ui.Text(c, "2,901 icons that move on hover, for MyGo native UI.").FontSize(15).TextColor(ui.Hex("#56607A"))
				ui.Row(c).Gap(14).Padding(14, 0, 0, 0).Children(func() {
					for i, ic := range row {
						ic.View(c.Key(i), kubai.Options{Size: 30, Still: true, SemanticBadge: true})
					}
				})
			})
		})
	}, w, h, 2)
}

func write(name string, img *image.RGBA) {
	f, err := os.Create(filepath.Join(out, name))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	log.Printf("%s: %dx%d", name, img.Bounds().Dx(), img.Bounds().Dy())
}
