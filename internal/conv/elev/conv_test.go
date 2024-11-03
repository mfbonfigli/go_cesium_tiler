package elev

import (
	"testing"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/utils"
)

var tolerance = 0.1

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
	c1 := NewOffsetElevationConverter(5)
	c2 := NewOffsetElevationConverter(10)
	c := NewPipelineElevationCorrector(c1, c2)
	actual, err := c.ConvertElevation(4707614.798256041, 431097.9816898434, 20)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected := 35.0
	if diff, err := utils.CompareWithTolerance(actual, expected, tolerance); err != nil {
		t.Errorf("expected elevation %f within %f precision, got %f (diff=%f)", expected, tolerance, actual, diff)
	}
}
