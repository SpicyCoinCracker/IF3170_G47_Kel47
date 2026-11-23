
package pca

import (
	"errors"
	"fmt"

	"eigentsar/core/matrix"
	"eigentsar/core/svd"
)

type Model struct {
	Mean     []float64      
	Basis    *matrix.Matrix 
	Singular []float64      
	Rank     int
}

func Fit(x *matrix.Matrix, k int) (*Model, error) {
	if k < 1 {
		return nil, fmt.Errorf("pca: k harus >= 1, k = %d", k)
	}
	if x.Cols < 2 {
		return nil, errors.New("pca: sampel minimal 2")
	}
	if err := x.Validate(); err != nil {
		return nil, err
	}
	mean := make([]float64, x.Rows)
	for i := 0; i < x.Rows; i++ {
		sum := 0.0
		for _, v := range x.Row(i) {
			sum += v
		}
		mean[i] = sum / float64(x.Cols)
	}
	a, _ := matrix.New(x.Rows, x.Cols)
	for i := 0; i < x.Rows; i++ {
		src, dst := x.Row(i), a.Row(i)
		for j, v := range src {
			dst[j] = v - mean[i]
		}
	}
	res, err := svd.Decompose(a, k)
	if errors.Is(err, svd.ErrNoVariance) {
		return nil, errors.New("pca: semua gambar identis, tidak ada variansi positif pca")
	}
	if err != nil {
		return nil, err
	}
	if k > res.Rank {
		return nil, fmt.Errorf("pca: k=%d melebihi jumlah komponen seharusnya (%d)", k, res.Rank)
	}
	if err := res.U.Validate(); err != nil {
		return nil, err
	}
	return &Model{Mean: mean, Basis: res.U, Singular: res.S, Rank: res.Rank}, nil
}

func Project(mean []float64, basis *matrix.Matrix, x []float64) ([]float64, error) {
	if len(x) != basis.Rows || len(mean) != basis.Rows {
		return nil, fmt.Errorf("pca: panjang vektor %d tidak sama dengan basis dimensi %d", len(x), basis.Rows)
	}
	c := make([]float64, basis.Cols)
	for i := 0; i < basis.Rows; i++ {
		y := x[i] - mean[i]
		for j, u := range basis.Row(i) {
			c[j] += u * y
		}
	}
	return c, nil
}

func VarianceRetained(sing []float64, k int) float64 {
	var part, total float64
	for i, s := range sing {
		total += s * s
		if i < k {
			part += s * s
		}
	}
	if total == 0 {
		return 0
	}
	return part / total
}
