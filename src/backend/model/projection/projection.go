// Package projection maps vectors/images into a model's eigenspace. Read-only: never mutates the model.
package projection

import (
	"fmt"
	"image"

	"eigentsar/core/pca"
	"eigentsar/model/builder"
)

func Project(m *builder.Model, x []float64) ([]float64, error) {
	if len(x) != m.Width*m.Height {
		return nil, fmt.Errorf("panjang vektor %d tidak sesuai dimensi model %d", len(x), m.Width*m.Height)
	}
	return pca.Project(m.Mean, m.Basis, x)
}

func ProjectImage(m *builder.Model, img image.Image) ([]float64, error) {
	return Project(m, builder.ImageVector(img))
}
