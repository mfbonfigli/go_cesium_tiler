package mutator

import "github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"

// Mutator defines a generic interface to manipulate coordinates or attributes of points.
type Mutator interface {
	// Mutate transforms or discards the points in input.
	//
	// The function receives in input the point, with coordinates expressed in
	// the local CRS with Z-up, and a transform object that can be used to
	// transform to and from the local CRS and EPSG 4978 CRS.
	//
	// The function returns the manipulated point and true if the point is to be used
	// or false if the point should be discarded from the final point cloud
	Mutate(geom.Point32, geom.Transform) (geom.Point32, bool)
}
