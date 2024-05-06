package tree

import (
	"context"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/elev"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/las"
)

// Tree represents the interface that an Octree representation of the point cloud should implement.
// A tree has a root node with up to 8 children and each can have up to 8 recursively.
// A tree must be "loaded" with points, and then "built" before being used.
type Tree interface {
	// Initializes the tree. Must be called before calling GetRootNode but after having called Load.
	Build() error
	// RootNode returns the root node of the tree
	GetRootNode() Node
	// IsBuilt returns if the tree has been previously initialized (built).
	IsBuilt() bool
	// Load loads the points into the tree. Must be called before any other operation on the tree.
	// requires providing a coordinate and an elevation converter that will be used by the tree
	// to internally perform coordinate conversions, as appropriate. The elevation converter can be nil.
	Load(las.LasReader, coor.CoordinateConverter, elev.ElevationConverter, context.Context) error
}

// Node models a generic node of a Tree. A node contains the points to show on its corresponding LoD.
// It must also be able to compute and return its children.
type Node interface {
	// GetBoundingBoxRegion returns the bounding box of the node, expressed
	// in EPSG:4979 (WGS 84) coordinates. A coordinate converter must be passed as input.
	GetBoundingBoxRegion(converter coor.CoordinateConverter) (geom.BoundingBox, error)
	// GetChildren returns the 8 children of the current tree node. Some or
	// all of these could be nil if not present.
	GetChildren() [8]Node
	// GetPoints returns  the points stored in the current node, not including those in the children.
	// Points MUST be returned in EPSG 4978 coordinate system expressed as offsets from the Node Center (see GetCenter)
	// a converter is provided if needed to perform the conversion
	GetPoints(converter coor.CoordinateConverter) geom.Point32List
	// TotalNumberOfPoints returns the number of points contained in this node AND all its children
	TotalNumberOfPoints() int
	// NumberOfPoints returns the number of points contained in this node, EXCLUDING its children
	NumberOfPoints() int
	// IsRoot returns true if the node is the root node of the tree representation
	IsRoot() bool
	// IsLeaf returns true if the current node does not have any children
	IsLeaf() bool
	// ComputeGeometricError returns an estimation, in meters, of the geometric error modeled
	// by the current tree node.
	ComputeGeometricError() float64
	// GetCenter return the EPSG 4978 x,y,z coordinates relative to which the points for the node are referred to
	GetCenter(converter coor.CoordinateConverter) (float64, float64, float64, error)
}
