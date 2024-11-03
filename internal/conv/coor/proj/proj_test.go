package proj

import (
	"testing"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/utils"
	"github.com/twpayne/go-proj/v10"
)

var coordTolerance = 0.01

func TestToSrid(t *testing.T) {
	c, err := NewProjCoordinateConverter()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	// 4326 to 4978
	actual, err := c.Transform("EPSG:4326", "EPSG:4978", geom.Coord{X: 123.474003, Y: 8.099314, Z: 0})
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected := geom.Coord{X: -3483057.5277292132, Y: 5267517.241803079, Z: 892655.4197953615}
	if err := utils.CompareCoord(actual, expected, coordTolerance); err != nil {
		t.Errorf("expected coordinate %v, got %v. Err: %v", expected, actual, err)
	}

	// 4978 to 4326
	expected, err = c.Transform("EPSG:4978", "EPSG:4326", geom.Coord{X: -3483057.5277292132, Y: 5267517.241803079, Z: 892655.4197953615})
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	actual = geom.Coord{X: 123.474003, Y: 8.099314, Z: 0}
	if err := utils.CompareCoord(actual, expected, coordTolerance); err != nil {
		t.Errorf("expected coordinate %v, got %v. Err: %v", expected, actual, err)
	}

	// 4326 to 3124
	actual, err = c.Transform("EPSG:4326", "EPSG:3124", geom.Coord{X: 123.474003, Y: 8.099314, Z: 0})
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected = geom.Coord{X: 552074.5400524682, Y: 895674.6033419219, Z: 0}
	if err := utils.CompareCoord(actual, expected, coordTolerance); err != nil {
		t.Errorf("expected coordinate %v, got %v. Err: %v", expected, actual, err)
	}

	// 4978 to 3124
	actual, err = c.Transform("EPSG:4978", "EPSG:3124", geom.Coord{X: -3483057.5277292132, Y: 5267517.241803079, Z: 892655.4197953615})
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected = geom.Coord{X: 552074.5400524682, Y: 895674.6033419219, Z: 0}
	if err := utils.CompareCoord(actual, expected, coordTolerance); err != nil {
		t.Errorf("expected coordinate %v, got %v. Err: %v", expected, actual, err)
	}

	// 3124 to 4978
	actual, err = c.Transform("EPSG:3124", "EPSG:4978", geom.Coord{X: 552074.5400524682, Y: 895674.6033419219, Z: 0})
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected = geom.Coord{X: -3483057.5277292132, Y: 5267517.241803079, Z: 892655.4197953615}
	if err := utils.CompareCoord(actual, expected, coordTolerance); err != nil {
		t.Errorf("expected coordinate %v, got %v. Err: %v", expected, actual, err)
	}
	c.Cleanup()
}

func TestToWGS84Cartesian(t *testing.T) {
	c, err := NewProjCoordinateConverter()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	// 4326 to 4978
	actual, err := c.ToWGS84Cartesian("EPSG:4326", geom.Coord{X: 123.474003, Y: 8.099314, Z: 0})
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	expected := geom.Coord{X: -3483057.5277292132, Y: 5267517.241803079, Z: 892655.4197953615}
	if err := utils.CompareCoord(actual, expected, coordTolerance); err != nil {
		t.Errorf("expected coordinate %v, got %v. Err: %v", expected, actual, err)
	}
	c.Cleanup()
}
func TestTest(t *testing.T) {
	context := proj.NewContext()

	// The C function does not return any error hence we can only reasonably
	// validate that executing the SetSearchPaths function call does not panic
	// considering various boundary conditions
	context.SetSearchPaths(nil)
	context.SetSearchPaths([]string{})
	context.SetSearchPaths([]string{"/tmp/data"})
	context.SetSearchPaths([]string{"/tmp/data", "/tmp/data2"})
}
