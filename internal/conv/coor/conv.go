package coor

import (
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
)

type CoordinateConverter interface {
	Transform(sourceCRS string, targetCRS string, coord geom.Coord) (geom.Coord, error)
	ToWGS84Cartesian(sourceCRS string, coord geom.Coord) (geom.Coord, error)
	Cleanup()
}

// ConverterFactory returns a new CoordinateConverter that should only be used in the same goroutine
// to avoid race conditions
type ConverterFactory func() (CoordinateConverter, error)
