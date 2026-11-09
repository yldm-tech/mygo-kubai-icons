// Gallery shows every icon in a MyGo window: browse by category, search by name (Chinese names too), pick an accent, the soft look and a size, hover an icon to play it (the icons in view also play one after another on their own, and choosing one plays it once, for a touch screen), and click it for a large preview that plays on its own, its Go code and its other versions. It speaks Chinese when the system or browser does, and English otherwise. The same program is the site at https://kubai.yldm.ai/, compiled to WebAssembly.
//
//	go run ./examples/gallery
//	go run ./examples/gallery -capture gallery.png -query folder
package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/egoist/mygo"
	_ "github.com/egoist/mygo/locales/zh-Hans" // MyGo's own strings in Chinese
	"github.com/egoist/mygo/ui"
)

func main() {
	capture := flag.String("capture", "", "write a PNG of the window to this path once it shows, then quit")
	query := flag.String("query", "", "start with this search")
	hover := flag.Bool("hover", false, "hold every icon at the end of its hover")
	soft := flag.Bool("soft", false, "start with the soft look")
	size := flag.Float64("size", 32, "start with this icon size")
	pick := flag.String("pick", "", "start with this icon chosen, by its upstream name")
	flag.Parse()
	g := newGallery()
	g.query, g.size, g.hoverAll, g.selName = *query, *size, *hover, *pick
	if *soft {
		g.variant = 1
	}
	// The gallery is written in English and Chinese; MyGo shows the one the user reads, in its own controls too. MYGO_LOCALE=en forces one.
	mygo.App.SetLanguages("en", "zh-Hans")
	mygo.App.WhenReady(func() {
		w := mygo.NewWindow(mygo.WindowOptions{
			Title:   "Kubai Icons for MyGo",
			Width:   1280,
			Height:  820,
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
