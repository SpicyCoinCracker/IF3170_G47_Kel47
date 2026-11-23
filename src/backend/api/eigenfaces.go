package api

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"strconv"
	"strings"

	"eigentsar/image/vectorization"
)

func (a *API) eigenfaces(w http.ResponseWriter, r *http.Request) {
	m, err := a.Mgr.Model()
	if err != nil {
		writeErr(w, 503, err)
		return
	}
	v := m.BuiltAt.UnixNano()
	type comp struct {
		Index    int     `json:"index"`
		Singular float64 `json:"singular"`
		URL      string  `json:"url"`
	}
	comps := make([]comp, m.K)
	for i := range comps {
		comps[i] = comp{i + 1, m.Singular[i], fmt.Sprintf("/api/model/eigenfaces/%d.png?v=%d", i+1, v)}
	}
	writeJSON(w, 200, map[string]any{
		"k": m.K, "meanUrl": fmt.Sprintf("/api/model/eigenfaces/0.png?v=%d", v), "components": comps,
	})
}

func (a *API) eigenfaceImage(w http.ResponseWriter, r *http.Request) {
	m, err := a.Mgr.Model()
	if err != nil {
		writeErr(w, 503, err)
		return
	}
	n, err := strconv.Atoi(strings.TrimSuffix(r.PathValue("n"), ".png"))
	if err != nil || n < 0 || n > m.K {
		writeErr(w, 404, errors.New("no such component"))
		return
	}
	var vec []float64
	if n == 0 {
		vec = append(vec, m.Mean...)
		for i := range vec {
			vec[i] *= 255
		}
	} else {
		vec = m.Basis.Col(n - 1)
		lo, hi := vec[0], vec[0]
		for _, x := range vec {
			lo, hi = min(lo, x), max(hi, x)
		}
		for i, x := range vec {
			if hi > lo {
				vec[i] = 255 * (x - lo) / (hi - lo)
			} else {
				vec[i] = 128
			}
		}
	}
	g, err := vectorization.Unvectorize(vec, m.Width, m.Height)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	img := image.NewGray(image.Rect(0, 0, g.W, g.H))
	for i, x := range g.Pix {
		img.Pix[i] = uint8(max(0, min(255, x+0.5)))
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	png.Encode(w, img)
}
