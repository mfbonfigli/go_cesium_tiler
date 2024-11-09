package geom

import (
	"math"
)

var IdentityQuaternion Quaternion = [4][4]float64{
	{1, 0, 0, 0},
	{0, 1, 0, 0},
	{0, 0, 1, 0},
	{0, 0, 0, 1},
}

var IdentityTransform Transform = Transform{
	LocalToGlobal: IdentityQuaternion,
	GlobalToLocal: IdentityQuaternion,
}

// Vector3 represents a Vector in a 3D space
type Vector3 struct {
	X float64
	Y float64
	Z float64
}

// Quaternion represents a 4x4 transformation matrix representing a rigid roto-translation
type Quaternion [4][4]float64

// Transform wraps two quaternions, one the inverse of the other, and is used as a convenience
// to perform direct and inverse transform between cartesian systems
type Transform struct {
	LocalToGlobal Quaternion
	GlobalToLocal Quaternion
}

// Transform applies the quaternion to the input Vector3 returning
// the resulting transformed Vector3
func (q Quaternion) Transform(v Vector3) Vector3 {
	return Vector3{
		X: q[0][0]*v.X + q[0][1]*v.Y + q[0][2]*v.Z + q[0][3],
		Y: q[1][0]*v.X + q[1][1]*v.Y + q[1][2]*v.Z + q[1][3],
		Z: q[2][0]*v.X + q[2][1]*v.Y + q[2][2]*v.Z + q[2][3],
	}
}

// ColumnMajor returns the quaternion represented as a linear array
// with column-major ordering of values
func (q Quaternion) ColumnMajor() [16]float64 {
	return [16]float64{
		q[0][0], q[1][0], q[2][0], q[3][0],
		q[0][1], q[1][1], q[2][1], q[3][1],
		q[0][2], q[1][2], q[2][2], q[3][2],
		q[0][3], q[1][3], q[2][3], q[3][3],
	}
}

// Unit returns the unit vector with same direction as the vector
func (v Vector3) Unit() Vector3 {
	n := v.Norm()
	return Vector3{
		X: v.X / n,
		Y: v.Y / n,
		Z: v.Z / n,
	}
}

// Norm return the euclidean norm of the vector
func (v Vector3) Norm() float64 {
	return math.Sqrt(math.Pow(v.X, 2) + math.Pow(v.Y, 2) + math.Pow(v.Z, 2))
}

// Cross returns the result of the cross product with the vector passed as input
func (v Vector3) Cross(w Vector3) Vector3 {
	return Vector3{
		X: v.Y*w.Z - v.Z*w.Y,
		Y: v.Z*w.X - v.X*w.Z,
		Z: v.X*w.Y - v.Y*w.X,
	}
}

// LocalCRSFromPoint takes in input a set of x,y,z coordinates
// assumed to be in EPSG 4978 CRS, ie based on a earth-centered cartesian system
// wrt the WGS84 ellipsoid and returns a Transform to a local CRS with the following properties:
// - Has origin located on the x,y,z point
// - Has a Z-up axis normal to the WGS84 ellipsoid
func LocalCRSFromPoint(x, y, z float64) Transform {
	zAxis := normalToWGS84FromPoint(x, y, z)
	xAxis, yAxis := zAxis.normals()

	toGlobal := Quaternion{
		{xAxis.X, yAxis.X, zAxis.X, x},
		{xAxis.Y, yAxis.Y, zAxis.Y, y},
		{xAxis.Z, yAxis.Z, zAxis.Z, z},
		{0, 0, 0, 1},
	}

	inverseTranslationVector := [3]float64{
		-xAxis.X*x - xAxis.Y*y - xAxis.Z*z,
		-yAxis.X*x - yAxis.Y*y - yAxis.Z*z,
		-zAxis.X*x - zAxis.Y*y - zAxis.Z*z,
	}

	toLocal := Quaternion{
		{xAxis.X, xAxis.Y, xAxis.Z, inverseTranslationVector[0]},
		{yAxis.X, yAxis.Y, yAxis.Z, inverseTranslationVector[1]},
		{zAxis.X, zAxis.Y, zAxis.Z, inverseTranslationVector[2]},
		{0, 0, 0, 1},
	}

	return Transform{
		LocalToGlobal: toGlobal,
		GlobalToLocal: toLocal,
	}
}

// normals returns a set of two arbitrary unit vectors guaranteed to be
// normal to the input one and between each other
func (v Vector3) normals() (Vector3, Vector3) {
	arbitraryVector := Vector3{X: 0, Y: 1, Z: 0}
	if v.Cross(arbitraryVector).Norm() < 0.05 {
		arbitraryVector = Vector3{X: 1, Y: 0, Z: 0}
	}
	xAxis := arbitraryVector.Cross(v).Unit()
	yAxis := v.Cross(xAxis).Unit()
	return xAxis, yAxis
}

// normalToWGS84FromPoint returns a Unit vector that is normal to the WGS84
// ellipsoid surface from the given point
func normalToWGS84FromPoint(x, y, z float64) Vector3 {
	a := 6378137.0        // Semi-major axis in meters (equatorial radius)
	b := 6356752.31424518 // Semi-minor axis in meters (polar radius)
	if x == 0 && y == 0 && z == 0 {
		// origin, choose the global z axis arbitrarily
		return Vector3{X: 0, Y: 0, Z: 1}
	}

	return Vector3{
		X: 2 * x / math.Pow(a, 2),
		Y: 2 * y / math.Pow(a, 2),
		Z: 2 * z / math.Pow(b, 2),
	}.Unit()
}
