package matrix

import (
	"errors"
	"fmt"
	"math"
)

var ErrDim = errors.New("matrix: dimensi tidak kompatibel")

type Matrix struct {
	Rows, Cols int
	Data       []float64
}

func New(rows, cols int) (*Matrix, error) {
	if rows <= 0 || cols <= 0 {
		return nil, fmt.Errorf("%w: invalid size %dx%d", ErrDim, rows, cols)
	}
	return &Matrix{rows, cols, make([]float64, rows*cols)}, nil
}

func FromData(rows, cols int, data []float64) (*Matrix, error) {
	if rows <= 0 || cols <= 0 || len(data) != rows*cols {
		return nil, fmt.Errorf("%w: %d values for %dx%d", ErrDim, len(data), rows, cols)
	}
	return &Matrix{rows, cols, data}, nil
}

func (m *Matrix) At(i, j int) float64     { return m.Data[i*m.Cols+j] }
func (m *Matrix) Set(i, j int, v float64) { m.Data[i*m.Cols+j] = v }

func (m *Matrix) Row(i int) []float64 { return m.Data[i*m.Cols : (i+1)*m.Cols] }


func (m *Matrix) Col(j int) []float64 {
	c := make([]float64, m.Rows)
	for i := range c {
		c[i] = m.Data[i*m.Cols+j]
	}
	return c
}

func (m *Matrix) Transpose() *Matrix {
	t := &Matrix{m.Cols, m.Rows, make([]float64, len(m.Data))}
	for i := 0; i < m.Rows; i++ {
		for j := 0; j < m.Cols; j++ {
			t.Data[j*m.Rows+i] = m.Data[i*m.Cols+j]
		}
	}
	return t
}

func Mul(a, b *Matrix) (*Matrix, error) {
	if a.Cols != b.Rows {
		return nil, fmt.Errorf("%w: %dx%d * %dx%d", ErrDim, a.Rows, a.Cols, b.Rows, b.Cols)
	}
	c := &Matrix{a.Rows, b.Cols, make([]float64, a.Rows*b.Cols)}
	for i := 0; i < a.Rows; i++ {
		ci := c.Data[i*c.Cols : (i+1)*c.Cols]
		for k := 0; k < a.Cols; k++ {
			aik := a.Data[i*a.Cols+k]
			if aik == 0 {
				continue
			}
			bk := b.Data[k*b.Cols : (k+1)*b.Cols]
			for j, v := range bk {
				ci[j] += aik * v
			}
		}
	}
	return c, nil
}

func (m *Matrix) Validate() error {
	if m == nil || m.Rows <= 0 || m.Cols <= 0 || len(m.Data) != m.Rows*m.Cols {
		return fmt.Errorf("%w: malformed matrix", ErrDim)
	}
	for _, v := range m.Data {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return errors.New("matrix: nilai NaN atau Inf tidak valid")
		}
	}
	return nil
}
