package geoid2ellipsoid

import (
	"math"
	"sync"
)

// Represent minimal data necessary to provide an efficient, cache-based solution for the massive geodetic to ellipsoidic height conversion
type BufferedCalculator struct {
	cellSize       float64
	geoidHeightMap sync.Map
	calc           Calculator
}

// Inits a new instance of EllipsoidToGeoidBufferedCalculator for the given Srid and with given caching cell size. To all points with X,Y coordinates
// falling inside a square cell with side equal to cellSize will be applied the same, eventually cached, elevation transformation.
// Choosing a small value for cell size improves the accuracy but increases computation times. Cell sizes approximately
// equivalent to 1m are acceptable approximations
func NewBufferedCalculator(cellSize float64, calc Calculator) Calculator {
	return &BufferedCalculator{
		cellSize: cellSize,
		calc:     calc,
	}
}

func (bc *BufferedCalculator) GetEllipsoidToGeoidOffset(lon, lat float64, srid int) (float64, error) {
	x := bc.getCellIndex(lon)
	y := bc.getCellIndex(lat)

	yMap, yMapPresent := bc.geoidHeightMap.Load(x)

	if !yMapPresent {
		var temp sync.Map
		bc.geoidHeightMap.Store(x, &temp)
		yMap = &temp
	}

	yVal, yValPresent := yMap.(*sync.Map).Load(y)
	if yValPresent {
		// return cached result
		return yVal.(float64), nil
	}

	// else compute offset and store in cache
	off, err := computeAndStoreInMap(x, y, srid, bc, yMap.(*sync.Map))

	if err != nil {
		return 0, err
	}

	// return offset + input elevation
	return off, nil
}

func (bc *BufferedCalculator) getCellIndex(dimensionValue float64) int {
	return int(math.Floor(dimensionValue / bc.cellSize))
}

func (bc *BufferedCalculator) getCellCenter(x, y int) (float64, float64) {
	return float64(x)*bc.cellSize + bc.cellSize/2, float64(y)*bc.cellSize + bc.cellSize/2
}

func computeAndStoreInMap(x, y, srid int, calc *BufferedCalculator, yMap *sync.Map) (float64, error) {
	off, err := calc.getCellEllipsoidToGeoidOffset(x, y, srid)

	if err != nil {
		return 0, err
	}

	yMap.Store(y, off)
	return off, nil
}

func (bc *BufferedCalculator) getCellEllipsoidToGeoidOffset(x, y, srid int) (float64, error) {
	cX, cY := bc.getCellCenter(x, y)

	return bc.calc.GetEllipsoidToGeoidOffset(cX, cY, srid)
}
