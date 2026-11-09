package kubai

import (
	"image"
	"math"
)

// castShadow returns img with a drop shadow under it: img's silhouette moved down by dy pixels, blurred by a Gaussian of sigma pixels and painted in c at opacity op, as upstream's feOffset, feGaussianBlur and feFlood do. img is premultiplied, as image.RGBA is.
func castShadow(img *image.RGBA, c rgb, dy, sigma, op float64) *image.RGBA {
	w, h := img.Rect.Dx(), img.Rect.Dy()
	alpha := make([]float32, w*h)
	off := int(math.Round(dy))
	for y := off; y < h; y++ {
		src := img.Pix[(y-off)*img.Stride:]
		row := alpha[y*w : (y+1)*w]
		for x := range row {
			row[x] = float32(src[x*4+3])
		}
	}
	gaussian(alpha, w, h, sigma)

	out := image.NewRGBA(img.Rect)
	k := float32(op) / 255
	cr, cg, cb := float32(c[0]), float32(c[1]), float32(c[2])
	for y := range h {
		src := img.Pix[y*img.Stride:]
		dst := out.Pix[y*out.Stride:]
		for x := range w {
			i := x * 4
			ia := float32(src[i+3]) / 255
			sa := alpha[y*w+x] * k * (1 - ia) // the shadow's coverage left showing under the icon
			dst[i+0] = clamp8(float32(src[i+0]) + cr*sa)
			dst[i+1] = clamp8(float32(src[i+1]) + cg*sa)
			dst[i+2] = clamp8(float32(src[i+2]) + cb*sa)
			dst[i+3] = clamp8(float32(src[i+3]) + 255*sa)
		}
	}
	return out
}

func clamp8(v float32) uint8 {
	switch {
	case v <= 0:
		return 0
	case v >= 255:
		return 255
	}
	return uint8(v + 0.5)
}

// gaussian blurs a w×h plane in place by three box blurs of the widths that come closest to a Gaussian of sigma, horizontally and then vertically. What lies outside the plane counts as empty, as it does for a shadow.
func gaussian(a []float32, w, h int, sigma float64) {
	if sigma <= 0.3 {
		return
	}
	tmp := make([]float32, len(a))
	for _, r := range boxRadii(sigma) {
		boxRows(a, tmp, w, h, r)
		boxCols(tmp, a, w, h, r)
	}
}

// boxRadii returns the radii of three box blurs whose sum approximates a Gaussian of sigma (Kovesi, "Fast almost-Gaussian filtering").
func boxRadii(sigma float64) [3]int {
	const n = 3
	ideal := math.Sqrt(12*sigma*sigma/n + 1)
	lo := int(math.Floor(ideal))
	if lo%2 == 0 {
		lo--
	}
	hi := lo + 2
	m := int(math.Round((12*sigma*sigma - n*float64(lo*lo) - 4*n*float64(lo) - 3*n) / (-4*float64(lo) - 4)))
	var r [3]int
	for i := range r {
		size := hi
		if i < m {
			size = lo
		}
		r[i] = (size - 1) / 2
	}
	return r
}

// boxRows averages each pixel of src with r on either side along its row into dst.
func boxRows(src, dst []float32, w, h, r int) {
	if r <= 0 {
		copy(dst, src)
		return
	}
	inv := 1 / float32(2*r+1)
	for y := range h {
		row, out := src[y*w:(y+1)*w], dst[y*w:(y+1)*w]
		var sum float32
		for x := 0; x <= r && x < w; x++ {
			sum += row[x]
		}
		for x := range w {
			out[x] = sum * inv
			if x+r+1 < w {
				sum += row[x+r+1]
			}
			if x-r >= 0 {
				sum -= row[x-r]
			}
		}
	}
}

// boxCols is boxRows down the columns.
func boxCols(src, dst []float32, w, h, r int) {
	if r <= 0 {
		copy(dst, src)
		return
	}
	inv := 1 / float32(2*r+1)
	for x := range w {
		var sum float32
		for y := 0; y <= r && y < h; y++ {
			sum += src[y*w+x]
		}
		for y := range h {
			dst[y*w+x] = sum * inv
			if y+r+1 < h {
				sum += src[(y+r+1)*w+x]
			}
			if y-r >= 0 {
				sum -= src[(y-r)*w+x]
			}
		}
	}
}
