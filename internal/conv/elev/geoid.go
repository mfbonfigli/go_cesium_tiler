package elev

import (
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/elev/geoid2ellipsoid"
)

// GeoidElevationConverter transforms a geoid height to that referred to the WGS84 ellipsoid
type GeoidElevationConverter struct {
	srid             int
	offsetCalculator geoid2ellipsoid.Calculator
}

// NewGeoidElevationConverter instantiates a new converter instance using the provided geoid to ellipsoid offset calculator
func NewGeoidElevationConverter(srid int, offsetCalculator geoid2ellipsoid.Calculator) *GeoidElevationConverter {
	// TODO: by default we are using the ellipsoidToGeoidSinglePointConverter as the old buffered converter
	//  suffers of coupling problems with the srid of data. It needs a cell size but this depends on the SRID and
	//  thus either a way to dynamically estabilish according to the srid is found or we can only use it if data is in 4326 srid
	//  which introduces an undocumented requirement. We need to fix the EllipsoidToGeoidBufferedCalculator to allow it
	//  to be used here
	return &GeoidElevationConverter{
		srid:             srid,
		offsetCalculator: geoid2ellipsoid.NewCachedCalculator(offsetCalculator),
	}
}

func (c *GeoidElevationConverter) ConvertElevation(x, y, z float64) (float64, error) {
	zfix, err := c.offsetCalculator.GetEllipsoidToGeoidOffset(x, y, c.srid)
	if err != nil {
		return 0, err
	}
	return zfix + z, nil
}
