package goproj

import (
	"fmt"
	"math"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
	"github.com/wroge/wgs84/v2"
)

type goprojCoordinateConverter struct {
	transformCache map[int]map[int]wgs84.Func
}

func NewGoProjCoordinateConverter() *goprojCoordinateConverter {
	return &goprojCoordinateConverter{transformCache: make(map[int]map[int]wgs84.Func)}
}

// Converts the given coordinate from the given source Srid to the given target srid.
func (cc *goprojCoordinateConverter) ToSrid(sourceSrid int, targetSrid int, coord geom.Coord) (geom.Coord, error) {
	if sourceSrid == targetSrid {
		return coord, nil
	}

	var transformFunc wgs84.Func
	if sourceMap, ok := cc.transformCache[sourceSrid]; ok {
		transformFunc = sourceMap[targetSrid]
	}
	if transformFunc == nil {
		transformFunc = wgs84.Transform(wgs84.EPSG(sourceSrid), wgs84.EPSG(targetSrid))
		sourceMap, ok := cc.transformCache[sourceSrid]
		if !ok {
			sourceMap = make(map[int]wgs84.Func)
			cc.transformCache[sourceSrid] = sourceMap
		}
		sourceMap[targetSrid] = transformFunc
	}
	x, y, z := transformFunc(coord.X, coord.Y, coord.Z)

	if math.IsNaN(x) || math.IsNaN(y) || math.IsNaN(z) {
		// unfortunately the wgs84 library does not return an error when, e.g., the EPSG code is unknown
		// but rather returns NaNs
		return coord, fmt.Errorf("coordinate conversion failed")
	}

	return geom.Coord{X: x, Y: y, Z: z}, nil

}

func (g *goprojCoordinateConverter) ToWGS84Cartesian(coord geom.Coord, sourceSrid int) (geom.Coord, error) {
	if sourceSrid == 4978 {
		return coord, nil
	}

	return g.ToSrid(sourceSrid, 4978, coord)
}

func (g *goprojCoordinateConverter) Cleanup() {}
