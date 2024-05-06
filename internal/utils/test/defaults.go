package test

import (
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor/proj4"
)

// GetTestCoordinateConverter returns the function to use to convert coordinates in tests
func GetTestCoordinateConverter() (coor.CoordinateConverter, error) {
	return proj4.NewProj4CoordinateConverter()
}
