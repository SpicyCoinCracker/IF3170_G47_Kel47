package unit

import (
	"errors"
	"math"
	"math/rand"
	"testing"

	"eigentsar/core/matrix"
	"eigentsar/core/pca"
	"eigentsar/core/svd"
)

func TestMulAndDims(t *testing.T) {
	a, _ := matrix.FromData(2, 3, []float64{1, 2, 3, 4, 5, 6})
	b, _ := matrix.FromData(3, 2, []float64{7, 8, 9, 10, 11, 12})
	c, err := matrix.Mul(a, b)
	if err != nil || c.Data[0] != 58 || c.Data[1] != 64 || c.Data[2] != 139 || c.Data[3] != 154 {
		t.Fatalf("bad product %v %v", c, err)
	}
	if _, err := matrix.Mul(a, a); !errors.Is(err, matrix.ErrDim) {
		t.Fatal("expected ErrDim")
	}
	if tr := a.Transpose(); tr.Rows != 3 || tr.At(2, 1) != 6 {
		t.Fatal("bad transpose")
	}
	if _, err := matrix.New(0, 3); err == nil {
		t.Fatal("expected error for empty matrix")
	}
}

func randMatrix(r, c int) *matrix.Matrix {
	rng := rand.New(rand.NewSource(1))
	m, _ := matrix.New(r, c)
	for i := range m.Data {
		m.Data[i] = rng.NormFloat64()
	}
	return m
}

func TestSVDReconstructionAndOrthogonality(t *testing.T) {
	a := randMatrix(40, 12)
	res, err := svd.Decompose(a, 0)
	if err != nil || res.Rank != 12 {
		t.Fatalf("svd: %v rank=%d", err, res.Rank)
	}
	for i := 1; i < len(res.S); i++ {
		if res.S[i] > res.S[i-1] {
			t.Fatal("singular values not descending")
		}
	}
	utu, _ := matrix.Mul(res.U.Transpose(), res.U)
	for i := 0; i < utu.Rows; i++ {
		for j := 0; j < utu.Cols; j++ {
			want := 0.0
			if i == j {
				want = 1
			}
			if math.Abs(utu.At(i, j)-want) > 1e-8 {
				t.Fatalf("UᵀU[%d,%d]=%g", i, j, utu.At(i, j))
			}
		}
	}
	// A ≈ U Σ Vᵀ
	us := *res.U
	us.Data = append([]float64(nil), res.U.Data...)
	for i := 0; i < us.Rows; i++ {
		for j := 0; j < us.Cols; j++ {
			us.Data[i*us.Cols+j] *= res.S[j]
		}
	}
	rec, _ := matrix.Mul(&us, res.V.Transpose())
	for i := range a.Data {
		if math.Abs(rec.Data[i]-a.Data[i]) > 1e-8 {
			t.Fatalf("reconstruction error %g", rec.Data[i]-a.Data[i])
		}
	}
}

func TestSVDZeroAndRankDeficient(t *testing.T) {
	z, _ := matrix.New(5, 3)
	if _, err := svd.Decompose(z, 0); !errors.Is(err, svd.ErrNoVariance) {
		t.Fatalf("want ErrNoVariance, got %v", err)
	}
	// duplicated column => rank 2 of 3
	a, _ := matrix.FromData(3, 3, []float64{1, 0, 1, 0, 1, 0, 1, 1, 1})
	res, err := svd.Decompose(a, 0)
	if err != nil || res.Rank != 2 {
		t.Fatalf("rank=%v err=%v", res, err)
	}
}

func TestPCAKValidationAndProjection(t *testing.T) {
	// spec example: x1=[10,20], x2=[30,40] -> mean [20,30]
	x, _ := matrix.FromData(2, 2, []float64{10, 30, 20, 40})
	m, err := pca.Fit(x, 1)
	if err != nil || m.Mean[0] != 20 || m.Mean[1] != 30 || m.Rank != 1 {
		t.Fatalf("fit: %+v %v", m, err)
	}
	if _, err := pca.Fit(x, 0); err == nil {
		t.Fatal("k=0 must fail")
	}
	if _, err := pca.Fit(x, 2); err == nil {
		t.Fatal("k>rank must fail")
	}
	same, _ := matrix.FromData(2, 2, []float64{1, 1, 2, 2})
	if _, err := pca.Fit(same, 1); err == nil {
		t.Fatal("identical images must fail")
	}
	c, _ := pca.Project(m.Mean, m.Basis, []float64{20, 30})
	if math.Abs(c[0]) > 1e-12 {
		t.Fatalf("mean must project to 0, got %v", c)
	}
	if _, err := pca.Project(m.Mean, m.Basis, []float64{1}); err == nil {
		t.Fatal("dimension mismatch must fail")
	}
}
