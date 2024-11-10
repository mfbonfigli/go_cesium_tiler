package geom

import (
	"math"
	"testing"
)

const tolerance = 1e-7

func compareWithTolerance(u Vector3, v Vector3, t *testing.T) {
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

func TestVectorUnit(t *testing.T) {
	v := Vector3{
		X: 2.0,
		Y: 2.0,
		Z: 0.0,
	}
	expected := Vector3{X: math.Sqrt(2) / 2, Y: math.Sqrt(2) / 2, Z: 0}
	compareWithTolerance(v.Unit(), expected, t)

	v = Vector3{
		X: 2.0,
		Y: 0.0,
		Z: 2.0,
	}
	expected = Vector3{X: math.Sqrt(2) / 2, Y: 0, Z: math.Sqrt(2) / 2}
	compareWithTolerance(v.Unit(), expected, t)

	v = Vector3{
		X: 0.0,
		Y: 2.0,
		Z: 2.0,
	}
	expected = Vector3{X: 0, Y: math.Sqrt(2) / 2, Z: math.Sqrt(2) / 2}
	compareWithTolerance(v.Unit(), expected, t)

	v = Vector3{
		X: 2.0,
		Y: 2.0,
		Z: 2.0,
	}
	expected = Vector3{X: math.Sqrt(3) / 3, Y: math.Sqrt(3) / 3, Z: math.Sqrt(3) / 3}
	compareWithTolerance(v.Unit(), expected, t)
}

func TestVectorNorm(t *testing.T) {
	v := Vector3{
		X: 6.0,
		Y: 8.0,
		Z: 0.0,
	}
	expected := 10.0
	if actual := v.Norm(); actual != expected {
		t.Errorf("expected norm %f, got %f", expected, actual)
	}

	v = Vector3{
		X: 0.0,
		Y: 8.0,
		Z: 6.0,
	}
	expected = 10.0
	if actual := v.Norm(); actual != expected {
		t.Errorf("expected norm %f, got %f", expected, actual)
	}

	v = Vector3{
		X: -6.0,
		Y: -7.0,
		Z: 6.0,
	}
	expected = 11.0
	if actual := v.Norm(); actual != expected {
		t.Errorf("expected norm %f, got %f", expected, actual)
	}
}

func TestVectorCross(t *testing.T) {
	u := Vector3{X: 2, Y: 0, Z: 0}
	v := Vector3{X: 0, Y: 2, Z: 0}
	expected := Vector3{X: 0, Y: 0, Z: 4}
	compareWithTolerance(expected, u.Cross(v), t)

	u = Vector3{X: 1, Y: 1, Z: 0}
	v = Vector3{X: 0, Y: 1, Z: 0}
	expected = Vector3{X: 0, Y: 0, Z: 1}
	compareWithTolerance(expected, u.Cross(v), t)

	u = Vector3{X: 0, Y: 1, Z: 0}
	v = Vector3{X: 0, Y: 0, Z: 1}
	expected = Vector3{X: 1, Y: 0, Z: 0}
	compareWithTolerance(expected, u.Cross(v), t)
}

func TestQuaternionTransform(t *testing.T) {
	// pure translation
	q := Quaternion{
		{1, 0, 0, 10},
		{0, 1, 0, 20},
		{0, 0, 1, 30},
		{0, 0, 0, 1},
	}
	actual := q.transformVector(Vector3{X: 5, Y: -4, Z: 7})
	expected := Vector3{X: 15, Y: 16, Z: 37}
	compareWithTolerance(expected, actual, t)

	// pure rotation around z
	q = Quaternion{
		{0, -1, 0, 0},
		{1, 0, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
	actual = q.transformVector(Vector3{X: 5, Y: -4, Z: 7})
	expected = Vector3{X: 4, Y: 5, Z: 7}
	compareWithTolerance(expected, actual, t)

	// pure rotation around x
	q = Quaternion{
		{1, 0, 0, 0},
		{0, 0, -1, 0},
		{0, 1, 0, 0},
		{0, 0, 0, 1},
	}
	actual = q.transformVector(Vector3{X: 5, Y: -4, Z: 7})
	expected = Vector3{X: 5, Y: -7, Z: -4}
	compareWithTolerance(expected, actual, t)

	// pure rotation around y
	q = Quaternion{
		{0, 0, 1, 0},
		{0, 1, 0, 0},
		{-1, 0, 0, 0},
		{0, 0, 0, 1},
	}
	actual = q.transformVector(Vector3{X: 5, Y: -4, Z: 7})
	expected = Vector3{X: 7, Y: -4, Z: -5}
	compareWithTolerance(expected, actual, t)

	// translation and rotation
	q = Quaternion{
		{0, -1, 0, 10},
		{1, 0, 0, 20},
		{0, 0, 1, 30},
		{0, 0, 0, 1},
	}
	actual = q.transformVector(Vector3{X: 5, Y: -4, Z: 7})
	expected = Vector3{X: 14, Y: 25, Z: 37}
	compareWithTolerance(expected, actual, t)
}

func TestLocalCRSFromPoint(t *testing.T) {
	origin := Vector3{X: 100, Y: 0, Z: 0}
	trans := LocalCRSFromPoint(origin.X, origin.Y, origin.Z)
	// assert correctness indirectly
	// should be centered in the input point
	compareWithTolerance(origin, trans.LocalToGlobal.transformVector(Vector3{}), t)
	// Z axis should be oriented correctly
	compareWithTolerance(Vector3{X: 100 + 1, Y: 0, Z: 0}, trans.LocalToGlobal.transformVector(Vector3{X: 0, Y: 0, Z: 1}), t)
	compareWithTolerance(Vector3{X: 0, Y: 0, Z: 0}, trans.GlobalToLocal.transformVector(Vector3{X: 100, Y: 0, Z: 0}), t)
	// X axis should be oriented correctly
	compareWithTolerance(Vector3{X: 100, Y: 0, Z: -1}, trans.LocalToGlobal.transformVector(Vector3{X: 1, Y: 0, Z: 0}), t)
	compareWithTolerance(Vector3{X: 1, Y: 0, Z: 0}, trans.GlobalToLocal.transformVector(Vector3{X: 100, Y: 0, Z: -1}), t)
	// Y axis should be oriented correctly
	compareWithTolerance(Vector3{X: 100, Y: 1, Z: 0}, trans.LocalToGlobal.transformVector(Vector3{X: 0, Y: 1, Z: 0}), t)
	compareWithTolerance(Vector3{X: 0, Y: 1, Z: 0}, trans.GlobalToLocal.transformVector(Vector3{X: 100, Y: 1, Z: 0}), t)

	origin = Vector3{X: 0, Y: 100, Z: 0}
	trans = LocalCRSFromPoint(origin.X, origin.Y, origin.Z)
	// assert correctness indirectly
	// should be centered in the input point
	compareWithTolerance(origin, trans.LocalToGlobal.transformVector(Vector3{}), t)
	// Z axis should be oriented correctly
	compareWithTolerance(Vector3{X: 0, Y: 100 + 1, Z: 0}, trans.LocalToGlobal.transformVector(Vector3{X: 0, Y: 0, Z: 1}), t)
	// X axis should be oriented correctly
	compareWithTolerance(Vector3{X: 0, Y: 100, Z: 1}, trans.LocalToGlobal.transformVector(Vector3{X: 1, Y: 0, Z: 0}), t)
	// Y axis should be oriented correctly
	compareWithTolerance(Vector3{X: 1, Y: 100, Z: 0}, trans.LocalToGlobal.transformVector(Vector3{X: 0, Y: 1, Z: 0}), t)

	origin = Vector3{X: 0, Y: 100, Z: 0}
	trans = LocalCRSFromPoint(origin.X, origin.Y, origin.Z)
	// assert correctness indirectly
	// should be centered in the input point
	compareWithTolerance(origin, trans.LocalToGlobal.transformVector(Vector3{}), t)
	// Z axis should be oriented correctly
	compareWithTolerance(Vector3{X: 0, Y: 100 + 1, Z: 0}, trans.LocalToGlobal.transformVector(Vector3{X: 0, Y: 0, Z: 1}), t)

	origin = Vector3{X: -100, Y: 0, Z: 0}
	trans = LocalCRSFromPoint(origin.X, origin.Y, origin.Z)
	// assert correctness indirectly
	// should be centered in the input point
	compareWithTolerance(origin, trans.LocalToGlobal.transformVector(Vector3{}), t)
	// Z axis should be oriented correctly
	compareWithTolerance(Vector3{X: -100 - 1, Y: 0, Z: 0}, trans.LocalToGlobal.transformVector(Vector3{X: 0, Y: 0, Z: 1}), t)

	origin = Vector3{X: 0, Y: -100, Z: 0}
	trans = LocalCRSFromPoint(origin.X, origin.Y, origin.Z)
	// assert correctness indirectly
	// should be centered in the input point
	compareWithTolerance(origin, trans.LocalToGlobal.transformVector(Vector3{}), t)
	// Z axis should be oriented correctly
	compareWithTolerance(Vector3{X: 0, Y: -100 - 1, Z: 0}, trans.LocalToGlobal.transformVector(Vector3{X: 0, Y: 0, Z: 1}), t)

	// test transform
	// Z axis should be oriented correctly
	compareWithTolerance(Vector3{X: 0, Y: -100 - 1, Z: 0}, trans.LocalToGlobal.Transform(0, 0, 1), t)
}

func TestQuaternionColumnMajor(t *testing.T) {
	q := Quaternion{
		{0, -1, 0, 10},
		{1, 0, 0, 20},
		{0, 0, 1, 30},
		{0, 0, 0, 1},
	}

	expected := [16]float64{
		0, 1, 0, 0,
		-1, 0, 0, 0,
		0, 0, 1, 0,
		10, 20, 30, 1,
	}

	if actual := q.ColumnMajor(); actual != expected {
		t.Errorf("expected %v got %v", expected, actual)
	}
}
