package preprocessing

import (
	"image"
	"math"
)

const (
	Width  = 64
	Height = 80
)

type Gray struct {
	W, H int
	Pix  []float64
}

func Process(img image.Image) *Gray {
	b := img.Bounds()
	sw, sh := b.Dx(), b.Dy()
	src := make([]float64, sw*sh)
	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA() // 16-bit channels
			src[y*sw+x] = (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(bl)) / 65535
		}
	}
	out := &Gray{W: Width, H: Height, Pix: make([]float64, Width*Height)}
	for dy := 0; dy < Height; dy++ {
		y0 := dy * sh / Height
		y1 := max(y0+1, int(math.Ceil(float64((dy+1)*sh)/Height)))
		for dx := 0; dx < Width; dx++ {
			x0 := dx * sw / Width
			x1 := max(x0+1, int(math.Ceil(float64((dx+1)*sw)/Width)))
			sum := 0.0
			for y := y0; y < y1; y++ {
				for x := x0; x < x1; x++ {
					sum += src[y*sw+x]
				}
			}
			out.Pix[dy*Width+dx] = sum / float64((y1-y0)*(x1-x0))
		}
	}
	return out
}
