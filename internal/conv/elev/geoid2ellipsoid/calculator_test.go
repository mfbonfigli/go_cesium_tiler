package geoid2ellipsoid

import (
	"testing"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/utils"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/utils/test"
)

var tolerance = 0.1

func TestGetEllipsoidToGeoid(t *testing.T) {
	conv, err := test.GetTestCoordinateConverter()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	c, err := NewEGMCalculator(conv)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	actual, err := c.GetEllipsoidToGeoidOffset(4707614.798256041, 431097.9816898434, 32633)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected := 45.16844841225744
	if diff, err := utils.CompareWithTolerance(actual, expected, tolerance); err != nil {
		t.Errorf("expected elevation diff to be within %f from %f but got %f (diff=%f)", tolerance, expected, actual, diff)
	}

	actual, err = c.GetEllipsoidToGeoidOffset(5209982.276799, 1576416.971624, 3395)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if diff, err := utils.CompareWithTolerance(actual, expected, tolerance); err != nil {
		t.Errorf("expected elevation diff to be within %f from %f but got %f (diff=%f)", tolerance, expected, actual, diff)
	}

	actual, err = c.GetEllipsoidToGeoidOffset(42.511443111445345, 14.167646172063971, 4326)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if diff, err := utils.CompareWithTolerance(actual, expected, tolerance); err != nil {
		t.Errorf("expected elevation diff to be within %f from %f but got %f (diff=%f)", tolerance, expected, actual, diff)
	}
}

func TestCachedCalc(t *testing.T) {
	conv, err := test.GetTestCoordinateConverter()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	underlyingCalc, err := NewEGMCalculator(conv)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	c := NewCachedCalculator(underlyingCalc)

	actual, err := c.GetEllipsoidToGeoidOffset(4707614.798256041, 431097.9816898434, 32633)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected := 45.16844841225744
	if diff, err := utils.CompareWithTolerance(actual, expected, tolerance); err != nil {
		t.Errorf("expected elevation diff to be within %f from %f but got %f (diff=%f)", tolerance, expected, actual, diff)
	}

	actual, err = c.GetEllipsoidToGeoidOffset(5209982.276799, 1576416.971624, 3395)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if diff, err := utils.CompareWithTolerance(actual, expected, tolerance); err != nil {
		t.Errorf("expected elevation diff to be within %f from %f but got %f (diff=%f)", tolerance, expected, actual, diff)
	}

	actual, err = c.GetEllipsoidToGeoidOffset(42.511443111445345, 14.167646172063971, 4326)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if diff, err := utils.CompareWithTolerance(actual, expected, tolerance); err != nil {
		t.Errorf("expected elevation diff to be within %f from %f but got %f (diff=%f)", tolerance, expected, actual, diff)
	}
}

func TestBufferedCalc(t *testing.T) {
	conv, err := test.GetTestCoordinateConverter()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	underlyingCalc, err := NewEGMCalculator(conv)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	c := NewBufferedCalculator(20, underlyingCalc)

	actual, err := c.GetEllipsoidToGeoidOffset(4707614.798256041, 431097.9816898434, 32633)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected := 45.16844841225744
	if diff, err := utils.CompareWithTolerance(actual, expected, tolerance); err != nil {
		t.Errorf("expected elevation diff to be within %f from %f but got %f (diff=%f)", tolerance, expected, actual, diff)
	}

	actual, err = c.GetEllipsoidToGeoidOffset(5209982.276799, 1576416.971624, 3395)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if diff, err := utils.CompareWithTolerance(actual, expected, tolerance); err != nil {
		t.Errorf("expected elevation diff to be within %f from %f but got %f (diff=%f)", tolerance, expected, actual, diff)
	}

	c = NewBufferedCalculator(0.01, underlyingCalc)
	actual, err = c.GetEllipsoidToGeoidOffset(42.511443111445345, 14.167646172063971, 4326)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if diff, err := utils.CompareWithTolerance(actual, expected, tolerance); err != nil {
		t.Errorf("expected elevation diff to be within %f from %f but got %f (diff=%f)", tolerance, expected, actual, diff)
	}
}
