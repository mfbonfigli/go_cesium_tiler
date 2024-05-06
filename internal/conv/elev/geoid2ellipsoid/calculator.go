package geoid2ellipsoid

import (
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
)

type Calculator interface {
	GetEllipsoidToGeoidOffset(lat, lon float64, sourceSrid int) (float64, error)
}

type EGMCalculator struct {
	egm  *egm
	conv coor.CoordinateConverter
}

func NewEGMCalculator(coordinateConverter coor.CoordinateConverter) (*EGMCalculator, error) {
	gravitationalModel, err := newDefaultEGM()
	if err != nil {
		return nil, err
	}

	return &EGMCalculator{
		egm:  gravitationalModel,
		conv: coordinateConverter,
	}, nil
}

func (e *EGMCalculator) GetEllipsoidToGeoidOffset(y, x float64, sourceSrid int) (float64, error) {
	coordinateInEPSG4326, err := e.conv.ToSrid(sourceSrid, 4326, geom.Coord{X: x, Y: y, Z: 0})
	if err != nil {
		return 0, err
	}

	return e.egm.heightOffset(coordinateInEPSG4326.X, coordinateInEPSG4326.Y, 0), err
}
