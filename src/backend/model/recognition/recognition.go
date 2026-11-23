
package recognition

import (
	"errors"
	"fmt"
	"math"

	"eigentsar/model/builder"
)

type Result struct {
	Index     int
	Subject   string
	Image     string 
	Distance  float64
	Known     bool
	Threshold float64
}

func Nearest(m *builder.Model, q []float64, threshold float64) (Result, error) {
	if len(m.Projections) == 0 {
		return Result{}, errors.New("model tidak memiliki image untuk training")
	}
	if len(q) != m.K {
		return Result{}, fmt.Errorf("query memiliki %d components, model memerlukan %d", len(q), m.K)
	}
	if threshold < 0 || math.IsNaN(threshold) || math.IsInf(threshold, 0) {
		return Result{}, fmt.Errorf("invalid threshold %v", threshold)
	}
	best, bestD := -1, math.Inf(1)
	for i, p := range m.Projections {
		d := 0.0
		for j, v := range p {
			t := v - q[j]
			d += t * t
		}
		if d < bestD {
			best, bestD = i, d
		}
	}
	if best < 0 {
		return Result{}, errors.New("komputasi distance gagal(NaN)")
	}
	dist := math.Sqrt(bestD)
	return Result{best, m.Labels[best], m.Images[best], dist, threshold == 0 || dist <= threshold, threshold}, nil
}
