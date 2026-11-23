package builder

import (
	"fmt"
	"image"
	"time"

	"eigentsar/core/matrix"
	"eigentsar/core/pca"
	"eigentsar/image/loader"
	"eigentsar/image/preprocessing"
	"eigentsar/image/vectorization"
	"eigentsar/storage/dataset"
)

const Version = 1

type Model struct {
	Version       int
	Width, Height int
	K, Rank       int
	Mean          []float64      
	Basis         *matrix.Matrix 
	Singular      []float64      
	Labels        []string       
	Images        []string       
	Projections   [][]float64    
	Fingerprint   string         
	BuiltAt       time.Time
	BuildMillis   int64
}

func ImageVector(img image.Image) []float64 {
	return vectorization.Vectorize(preprocessing.Process(img))
}

func (m *Model) Validate() error {
	d := m.Width * m.Height
	if m.Version != Version || m.Width != preprocessing.Width || m.Height != preprocessing.Height {
		return fmt.Errorf("model versi/preprocessing (v%d, %dx%d) berbeda dengan build ini (v%d, %dx%d)",
			m.Version, m.Width, m.Height, Version, preprocessing.Width, preprocessing.Height)
	}
	if m.Basis == nil || m.Basis.Validate() != nil || m.Basis.Rows != d || m.Basis.Cols != m.K || m.K < 1 ||
		len(m.Mean) != d || len(m.Labels) == 0 || len(m.Labels) != len(m.Images) || len(m.Labels) != len(m.Projections) {
		return fmt.Errorf("model artifact tidak konsisten")
	}
	for _, p := range m.Projections {
		if len(p) != m.K {
			return fmt.Errorf("model artifact tidak konsisten")
		}
	}
	return nil
}

func (m *Model) VarianceRetained() float64 { return pca.VarianceRetained(m.Singular, m.K) }

func Build(store dataset.Store, k int) (*Model, error) {
	start := time.Now()
	if k < 1 {
		return nil, fmt.Errorf("k harus >= 1, k = %d", k)
	}
	entries, err := store.List(dataset.Train)
	if err != nil {
		return nil, fmt.Errorf("membaca dataset: %w", err)
	}
	if len(entries) < 2 {
		return nil, fmt.Errorf("dataset training memiliki %d image; minimal perlu dua(2)", len(entries))
	}
	fp, err := store.Fingerprint(entries)
	if err != nil {
		return nil, err
	}
	d := preprocessing.Width * preprocessing.Height
	x, _ := matrix.New(d, len(entries))
	for j, e := range entries {
		img, err := loader.Open(store.Path(e))
		if err != nil {
			return nil, fmt.Errorf("training image %s: %w", e.ID(), err)
		}
		for i, v := range ImageVector(img) {
			x.Set(i, j, v)
		}
	}
	p, err := pca.Fit(x, k)
	if err != nil {
		return nil, err
	}
	m := &Model{
		Version: Version, Width: preprocessing.Width, Height: preprocessing.Height,
		K: k, Rank: p.Rank, Mean: p.Mean, Basis: p.Basis, Singular: p.Singular,
		Fingerprint: fp, BuiltAt: time.Now(),
	}
	for j, e := range entries {
		c, err := pca.Project(p.Mean, p.Basis, x.Col(j))
		if err != nil {
			return nil, err
		}
		m.Labels = append(m.Labels, e.Subject)
		m.Images = append(m.Images, e.ID())
		m.Projections = append(m.Projections, c)
	}
	m.BuildMillis = time.Since(start).Milliseconds()
	return m, nil
}
