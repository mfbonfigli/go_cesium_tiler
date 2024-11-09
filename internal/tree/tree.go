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
	RootNode() Node
	// Load loads the points into the tree. Must be called before any other operation on the tree.
	// requires providing a coordinate and an elevation converter that will be used by the tree
	// to internally perform coordinate conversions, as appropriate. The elevation converter can be nil.
	Load(las.LasReader, coor.ConverterFactory, elev.Converter, context.Context) error
}

// Node models a generic node of a Tree. A node contains the points to show on its corresponding LoD.
// It must also be able to compute and return its children.
type Node interface {
	// BoundingBox returns the bounding box of the node, expressed in local coordinates
	BoundingBox() geom.BoundingBox
	// Children returns the 8 children of the current tree node. Some or
	// all of these could be nil if not present.
	Children() [8]Node
	// Points returns  the points stored in the current node, not including those in the children.
	// Points will have coordinates expressed relative to the local reference system
	Points() geom.Point32List
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
	// TransformMatrix returns the Transform object to use to transform the coordinates from the
	// node local CRS to the parent CRS. For a root node this can be used to transform the
	// coordinates back to the EPSG 4978 (ECEF) coordinate system.
	TransformMatrix() *geom.Transform
}
