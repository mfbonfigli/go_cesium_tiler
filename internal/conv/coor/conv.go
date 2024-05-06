package coor

import (
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
)

type CoordinateConverter interface {
	ToSrid(sourceSrid int, targetSrid int, coord geom.Coord) (geom.Coord, error)
	ToWGS84Cartesian(coord geom.Coord, sourceSrid int) (geom.Coord, error)
	Cleanup()
}
