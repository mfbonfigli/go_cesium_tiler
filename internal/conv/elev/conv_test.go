package elev

import (
	"testing"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/elev/geoid2ellipsoid"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/utils"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/utils/test"
)

var tolerance = 0.1

func TestGeoidElevConverter(t *testing.T) {
	conv, err := test.GetTestCoordinateConverter()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	calc, err := geoid2ellipsoid.NewEGMCalculator(conv)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	c := NewGeoidElevationConverter(32633, calc)
	actual, err := c.ConvertElevation(4707614.798256041, 431097.9816898434, 20)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected := 65.16844841225745
	if diff, err := utils.CompareWithTolerance(actual, expected, tolerance); err != nil {
		t.Errorf("expected elevation %f within %f precision, got %f (diff=%f)", expected, tolerance, actual, diff)
	}
}
func TestOffsetConverter(t *testing.T) {
	c := NewOffsetElevationConverter(10)
	actual, err := c.ConvertElevation(4707614.798256041, 431097.9816898434, 20)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected := 30.0
	if diff, err := utils.CompareWithTolerance(actual, expected, tolerance); err != nil {
		t.Errorf("expected elevation %f within %f precision, got %f (diff=%f)", expected, tolerance, actual, diff)
	}
}
func TestPipelineConverter(t *testing.T) {
	conv, err := test.GetTestCoordinateConverter()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	calc, err := geoid2ellipsoid.NewEGMCalculator(conv)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	c1 := NewGeoidElevationConverter(32633, calc)
	c2 := NewOffsetElevationConverter(10)
	c := NewPipelineElevationCorrector(c1, c2)
	actual, err := c.ConvertElevation(4707614.798256041, 431097.9816898434, 20)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected := 75.16844841225745
	if diff, err := utils.CompareWithTolerance(actual, expected, tolerance); err != nil {
		t.Errorf("expected elevation %f within %f precision, got %f (diff=%f)", expected, tolerance, actual, diff)
	}
}
