package coor

import (
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
)

type Converter interface {
	Transform(sourceCRS string, targetCRS string, coord geom.Vector3) (geom.Vector3, error)
	ToWGS84Cartesian(sourceCRS string, coord geom.Vector3) (geom.Vector3, error)
	Cleanup()
}

// ConverterFactory returns a new CoordinateConverter that should only be used in the same goroutine
// to avoid race conditions
type ConverterFactory func() (Converter, error)
