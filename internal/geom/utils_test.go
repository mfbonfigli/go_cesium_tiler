package geom

import (
	"math"
	"testing"

	"github.com/mfbonfigli/gocesiumtiler/v2/tiler/model"
)

const tolerance = 1e-7

func compareWithTolerance(u model.Vector, v model.Vector, t *testing.T) {
	if math.Abs(u.X-v.X) > tolerance {
		t.Errorf("expected coordinate X %f, got %f", u.X, v.X)
	}
	if math.Abs(u.Y-v.Y) > tolerance {
		t.Errorf("expected coordinate Y %f, got %f", u.Y, v.Y)
	}
	if math.Abs(u.Z-v.Z) > tolerance {
		t.Errorf("expected coordinate Z %f, got %f", u.Z, v.Z)
	}
}

func TestLocalToGlobalTransformFromPoint(t *testing.T) {
	origin := model.Vector{X: 100, Y: 0, Z: 0}
	trans := LocalToGlobalTransformFromPoint(origin.X, origin.Y, origin.Z)
	// assert correctness indirectly
	// should be centered in the input point
	compareWithTolerance(origin, trans.Forward(model.Vector{}), t)
	// Z axis should be oriented correctly
	compareWithTolerance(model.Vector{X: 100 + 1, Y: 0, Z: 0}, trans.Forward(model.Vector{X: 0, Y: 0, Z: 1}), t)
	compareWithTolerance(model.Vector{X: 0, Y: 0, Z: 0}, trans.Inverse(model.Vector{X: 100, Y: 0, Z: 0}), t)
	// X axis should be oriented correctly
	compareWithTolerance(model.Vector{X: 100, Y: 0, Z: -1}, trans.Forward(model.Vector{X: 1, Y: 0, Z: 0}), t)
	compareWithTolerance(model.Vector{X: 1, Y: 0, Z: 0}, trans.Inverse(model.Vector{X: 100, Y: 0, Z: -1}), t)
	// Y axis should be oriented correctly
	compareWithTolerance(model.Vector{X: 100, Y: 1, Z: 0}, trans.Forward(model.Vector{X: 0, Y: 1, Z: 0}), t)
	compareWithTolerance(model.Vector{X: 0, Y: 1, Z: 0}, trans.Inverse(model.Vector{X: 100, Y: 1, Z: 0}), t)

	origin = model.Vector{X: 0, Y: 100, Z: 0}
	trans = LocalToGlobalTransformFromPoint(origin.X, origin.Y, origin.Z)
	// assert correctness indirectly
	// should be centered in the input point
	compareWithTolerance(origin, trans.Forward(model.Vector{}), t)
	// Z axis should be oriented correctly
	compareWithTolerance(model.Vector{X: 0, Y: 100 + 1, Z: 0}, trans.Forward(model.Vector{X: 0, Y: 0, Z: 1}), t)
	// X axis should be oriented correctly
	compareWithTolerance(model.Vector{X: 0, Y: 100, Z: 1}, trans.Forward(model.Vector{X: 1, Y: 0, Z: 0}), t)
	// Y axis should be oriented correctly
	compareWithTolerance(model.Vector{X: 1, Y: 100, Z: 0}, trans.Forward(model.Vector{X: 0, Y: 1, Z: 0}), t)

	origin = model.Vector{X: 0, Y: 100, Z: 0}
	trans = LocalToGlobalTransformFromPoint(origin.X, origin.Y, origin.Z)
	// assert correctness indirectly
	// should be centered in the input point
	compareWithTolerance(origin, trans.Forward(model.Vector{}), t)
	// Z axis should be oriented correctly
	compareWithTolerance(model.Vector{X: 0, Y: 100 + 1, Z: 0}, trans.Forward(model.Vector{X: 0, Y: 0, Z: 1}), t)

	origin = model.Vector{X: -100, Y: 0, Z: 0}
	trans = LocalToGlobalTransformFromPoint(origin.X, origin.Y, origin.Z)
	// assert correctness indirectly
	// should be centered in the input point
	compareWithTolerance(origin, trans.Forward(model.Vector{}), t)
	// Z axis should be oriented correctly
	compareWithTolerance(model.Vector{X: -100 - 1, Y: 0, Z: 0}, trans.Forward(model.Vector{X: 0, Y: 0, Z: 1}), t)

	origin = model.Vector{X: 0, Y: -100, Z: 0}
	trans = LocalToGlobalTransformFromPoint(origin.X, origin.Y, origin.Z)
	// assert correctness indirectly
	// should be centered in the input point
	compareWithTolerance(origin, trans.Forward(model.Vector{}), t)
	// Z axis should be oriented correctly
	compareWithTolerance(model.Vector{X: 0, Y: -100 - 1, Z: 0}, trans.Forward(model.Vector{X: 0, Y: 0, Z: 1}), t)
}
