package kubai_test

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	kubai "github.com/yldm-tech/mygo-kubai-icons"
	"github.com/yldm-tech/mygo-kubai-icons/all"
)

var (
	compareOut    = flag.String("compare.out", "", "directory to write this port's frames to, for internal/compare/compare.sh")
	compareIcons  = flag.String("compare.icons", "", "comma-separated icon names")
	comparePoints = flag.String("compare.points", "0,0.5,1", "comma-separated hover progress values")
)

// TestWriteFrames writes this port's frames at 96 px for the 36-unit box upstream's standalone SVGs use, the same framing as internal/compare/reference.mjs, so the two can be compared pixel for pixel.
func TestWriteFrames(t *testing.T) {
	if *compareOut == "" {
		t.Skip("set -compare.out")
	}
	byName := map[string]*kubai.Icon{}
	for _, e := range all.Entries {
		byName[e.Name] = e.Icon()
	}
	const px = 96
	size := float32(px * 31.0 / 36.0)
	off := float32(px * 2.5 / 36.0)
	if err := os.MkdirAll(*compareOut, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range strings.Split(*compareIcons, ",") {
		ic := byName[name]
		if ic == nil {
			t.Fatalf("no icon %q", name)
		}
		for _, ps := range strings.Split(*comparePoints, ",") {
			v, err := strconv.ParseFloat(ps, 64)
			if err != nil {
				t.Fatal(err)
			}
			img := ui.Render(func(c *ui.Context) {
				ui.Box(c).Size(px, px).Children(func() {
					ui.Box(c).Absolute().Left(off).Top(off).Size(size, size).Children(func() {
						ic.Frame(c, kubai.Options{}, size, v)
					})
				})
			}, px, px, 1)
			f, err := os.Create(filepath.Join(*compareOut, fmt.Sprintf("%s@%s.png", name, ps)))
			if err != nil {
				t.Fatal(err)
			}
			if err := png.Encode(f, img); err != nil {
				t.Fatal(err)
			}
			f.Close()
		}
	}
}
