
package svd

import (
	"errors"
	"fmt"
	"math"
	"sort"

	"eigentsar/core/matrix"
)

const (
	maxSweeps = 100
	offTol    = 1e-13 
	rankTol   = 1e-10 
	negTol    = 1e-8  
)

var ErrNoVariance = errors.New("svd: matriks nol(0), tidak ada singular value")

type Result struct {
	U    *matrix.Matrix 
	S    []float64      
	V    *matrix.Matrix 
	Rank int
}


func Decompose(a *matrix.Matrix, k int) (*Result, error) {
	if err := a.Validate(); err != nil {
		return nil, err
	}
	g, err := matrix.Mul(a.Transpose(), a)
	if err != nil {
		return nil, err
	}
	vals, vt, err := jacobiEigen(g)
	if err != nil {
		return nil, err
	}
	n := len(vals)
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(x, y int) bool { return vals[order[x]] > vals[order[y]] })

	gmax := vals[order[0]]
	if gmax <= 0 || gmax < 1e-300 {
		return nil, ErrNoVariance
	}
	if vals[order[n-1]] < -negTol*gmax {
		return nil, fmt.Errorf("svd: eigenvalue negatif terlalu besar, sekitar %g dari Gram matrix (gagal dihitung)", vals[order[n-1]])
	}
	rank := 0
	for _, idx := range order {
		if vals[idx] > rankTol*gmax {
			rank++
		}
	}
	m := rank
	if k > 0 && k < rank {
		m = k
	}

	s := make([]float64, rank)
	for i := range s {
		s[i] = math.Sqrt(vals[order[i]])
	}
	v, _ := matrix.New(n, m)
	for c := 0; c < m; c++ {
		row := vt.Row(order[c]) 
		for j := 0; j < n; j++ {
			v.Set(j, c, row[j])
		}
	}
	u, err := matrix.Mul(a, v)
	if err != nil {
		return nil, err
	}
	for c := 0; c < m; c++ {
		inv := 1 / s[c]
		for i := 0; i < u.Rows; i++ {
			u.Data[i*u.Cols+c] *= inv
		}
	}
	return &Result{U: u, S: s, V: v, Rank: rank}, nil
}



func jacobiEigen(b *matrix.Matrix) ([]float64, *matrix.Matrix, error) {
	n := b.Rows
	if b.Cols != n {
		return nil, nil, fmt.Errorf("%w: eigen needs a square matrix", matrix.ErrDim)
	}
	d := make([]float64, len(b.Data))
	copy(d, b.Data)
	q, _ := matrix.New(n, n) 
	for i := 0; i < n; i++ {
		q.Set(i, i, 1)
	}
	frob := 0.0
	for _, x := range d {
		frob += x * x
	}
	frob = math.Sqrt(frob)

	converged := false
	for sweep := 0; sweep < maxSweeps; sweep++ {
		off := 0.0
		for i := 0; i < n; i++ {
			for j := i + 1; j < n; j++ {
				off += d[i*n+j] * d[i*n+j]
			}
		}
		if math.Sqrt(2*off) <= offTol*frob {
			converged = true
			break
		}
		for p := 0; p < n-1; p++ {
			for r := p + 1; r < n; r++ {
				apq := d[p*n+r]
				if math.Abs(apq) <= 1e-18*frob {
					continue
				}
				theta := (d[r*n+r] - d[p*n+p]) / (2 * apq)
				t := 1 / (math.Abs(theta) + math.Sqrt(theta*theta+1))
				if theta < 0 {
					t = -t
				}
				c := 1 / math.Sqrt(t*t+1)
				s := t * c
				for k := 0; k < n; k++ {
					if k == p || k == r {
						continue
					}
					dkp, dkq := d[k*n+p], d[k*n+r]
					np, nq := c*dkp-s*dkq, s*dkp+c*dkq
					d[k*n+p], d[p*n+k] = np, np
					d[k*n+r], d[r*n+k] = nq, nq
				}
				app, aqq := d[p*n+p], d[r*n+r]
				d[p*n+p] = app - t*apq
				d[r*n+r] = aqq + t*apq
				d[p*n+r], d[r*n+p] = 0, 0
				qp, qr := q.Row(p), q.Row(r)
				for k := 0; k < n; k++ {
					a, b2 := qp[k], qr[k]
					qp[k], qr[k] = c*a-s*b2, s*a+c*b2
				}
			}
		}
	}
	if !converged {
		return nil, nil, errors.New("svd: Jacobi eigen-solver gagal konvergen")
	}
	vals := make([]float64, n)
	for i := range vals {
		vals[i] = d[i*n+i]
	}
	return vals, q, nil
}
