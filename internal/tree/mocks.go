package tree

import (
	"context"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/elev"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/las"
)

type MockNode struct {
	Region                    geom.BoundingBox
	Children                  [8]Node
	Pts                       geom.Point32List
	TotalNumPts               int
	Root                      bool
	Leaf                      bool
	GeomError                 float64
	CenterX, CenterY, CenterZ float64
	// invocation params
	Las         las.LasReader
	Conv        coor.CoordinateConverter
	Elev        elev.ElevationConverter
	Ctx         context.Context
	LoadCalled  bool
	BuildCalled bool
}

func (n *MockNode) GetBoundingBoxRegion(converter coor.CoordinateConverter) (geom.BoundingBox, error) {
	n.Conv = converter
	return n.Region, nil
}
func (n *MockNode) GetChildren() [8]Node {
	return n.Children
}
func (n *MockNode) GetPoints(converter coor.CoordinateConverter) geom.Point32List {
	n.Conv = converter
	return n.Pts
}
func (n *MockNode) TotalNumberOfPoints() int {
	return n.TotalNumPts
}
func (n *MockNode) NumberOfPoints() int {
	return n.Pts.Len()
}
func (n *MockNode) IsRoot() bool {
	return n.Root
}
func (n *MockNode) IsLeaf() bool {
	return n.Leaf
}
func (n *MockNode) ComputeGeometricError() float64 {
	return n.GeomError
}
func (n *MockNode) GetCenter(converter coor.CoordinateConverter) (float64, float64, float64, error) {
	n.Conv = converter
	return n.CenterX, n.CenterY, n.CenterZ, nil
}
func (n *MockNode) Build() error {
	n.BuildCalled = true
	return nil
}
func (n *MockNode) GetRootNode() Node {
	return n
}
func (n *MockNode) IsBuilt() bool {
	return true
}
func (n *MockNode) Load(l las.LasReader, c coor.CoordinateConverter, e elev.ElevationConverter, ctx context.Context) error {
	n.LoadCalled = true
	n.Ctx = ctx
	n.Las = l
	n.Conv = c
	n.Elev = e
	return nil
}
