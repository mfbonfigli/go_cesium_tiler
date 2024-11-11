package mutator

import (
	"github.com/mfbonfigli/gocesiumtiler/v2/tiler/model"
)

// Mutator defines a generic interface to manipulate coordinates or attributes of points.
type Mutator interface {
	// Mutate transforms or discards the points in input.
	//
	// The function receives in input the point, with coordinates expressed in
	// the local CRS with Z-up, and a transform object that can be used to
	// forward transform from the local CRS to the global EPSG 4978 CRS and inverse transform
	// from the global CRS to the local CRS.
	//
	// The function returns the manipulated point and true if the point is to be used
	// or false if the point should be discarded from the final point cloud
	Mutate(pt model.Point, localToGlobal model.Transform) (model.Point, bool)
}
