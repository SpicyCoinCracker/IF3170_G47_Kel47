package vectorization

import (
	"fmt"

	"eigentsar/image/preprocessing"
)

func Vectorize(g *preprocessing.Gray) []float64 {
	v := make([]float64, g.W*g.H)
	for x := 0; x < g.W; x++ {
		for y := 0; y < g.H; y++ {
			v[x*g.H+y] = g.Pix[y*g.W+x]
		}
	}
	return v
}

func Unvectorize(v []float64, w, h int) (*preprocessing.Gray, error) {
	if w <= 0 || h <= 0 || len(v) != w*h {
		return nil, fmt.Errorf("panjang vektor %d tidak sesuai %dx%d", len(v), w, h)
	}
	g := &preprocessing.Gray{W: w, H: h, Pix: make([]float64, w*h)}
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			g.Pix[y*w+x] = v[x*h+y]
		}
	}
	return g, nil
}
