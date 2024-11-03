package test

import (
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor/proj"
)

// GetTestCoordinateConverter returns the function to use to convert coordinates in tests
func GetTestCoordinateConverterFactory() coor.ConverterFactory {
	return func() (coor.CoordinateConverter, error) {
		return proj.NewProjCoordinateConverter()
	}
}
