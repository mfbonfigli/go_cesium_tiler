package coor

import (
	"github.com/mfbonfigli/gocesiumtiler/v2/tiler/model"
)

type Converter interface {
	Transform(sourceCRS string, targetCRS string, coord model.Vector) (model.Vector, error)
	ToWGS84Cartesian(sourceCRS string, coord model.Vector) (model.Vector, error)
	Cleanup()
}

// ConverterFactory returns a new CoordinateConverter that should only be used in the same goroutine
// to avoid race conditions
type ConverterFactory func() (Converter, error)
