package geom

import (
	"math"

	"github.com/mfbonfigli/gocesiumtiler/v2/tiler/model"
)

// LocalToGlobalTransformFromPoint takes in input a set of x,y,z coordinates
// assumed to be in EPSG 4978 CRS, ie based on a earth-centered cartesian system
// wrt the WGS84 ellipsoid and returns a Transform from the global CRS to a local CRS
// that has the following properties:
// - Has origin located on the x,y,z point
// - Has a Z-up axis normal to the WGS84 ellipsoid
func LocalToGlobalTransformFromPoint(x, y, z float64) model.Transform {
	zAxis := normalToWGS84FromPoint(x, y, z)
	xAxis, yAxis := normals(zAxis)

	toGlobal := [4][4]float64{
		{xAxis.X, yAxis.X, zAxis.X, x},
		{xAxis.Y, yAxis.Y, zAxis.Y, y},
		{xAxis.Z, yAxis.Z, zAxis.Z, z},
		{0, 0, 0, 1},
	}

	return model.NewTransform(toGlobal)
}

// normals returns a set of two arbitrary unit vectors guaranteed to be
// normal to the input one and between each other
func normals(v model.Vector) (model.Vector, model.Vector) {
	arbitraryVector := model.Vector{X: 0, Y: 1, Z: 0}
	if v.Cross(arbitraryVector).Norm() < 0.05 {
		arbitraryVector = model.Vector{X: 1, Y: 0, Z: 0}
	}
	xAxis := arbitraryVector.Cross(v).Unit()
	yAxis := v.Cross(xAxis).Unit()
	return xAxis, yAxis
}

// normalToWGS84FromPoint returns a Unit vector that is normal to the WGS84
// ellipsoid surface from the given point
func normalToWGS84FromPoint(x, y, z float64) model.Vector {
	a := 6378137.0        // Semi-major axis in meters (equatorial radius)
	b := 6356752.31424518 // Semi-minor axis in meters (polar radius)
	if x == 0 && y == 0 && z == 0 {
		// origin, choose the global z axis arbitrarily
		return model.Vector{X: 0, Y: 0, Z: 1}
	}

	return model.Vector{
		X: 2 * x / math.Pow(a, 2),
		Y: 2 * y / math.Pow(a, 2),
		Z: 2 * z / math.Pow(b, 2),
	}.Unit()
}
