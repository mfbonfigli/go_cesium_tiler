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
	ChildNodes                [8]Node
	Pts                       geom.Point32List
	TotalNumPts               int
	Root                      bool
	Leaf                      bool
	GeomError                 float64
	CenterX, CenterY, CenterZ float64
	// invocation params
	Las         las.LasReader
	ConvFactory coor.ConverterFactory
	Elev        elev.Converter
	Ctx         context.Context
	LoadCalled  bool
	BuildCalled bool
	Transform   *geom.Transform
}

func (n *MockNode) TransformMatrix() *geom.Transform {
	return n.Transform
}
func (n *MockNode) BoundingBox() geom.BoundingBox {
	return n.Region
}
func (n *MockNode) Children() [8]Node {
	return n.ChildNodes
}
func (n *MockNode) Points() geom.Point32List {
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
func (n *MockNode) Build() error {
	n.BuildCalled = true
	return nil
}
func (n *MockNode) RootNode() Node {
	return n
}
func (n *MockNode) Load(l las.LasReader, c coor.ConverterFactory, e elev.Converter, ctx context.Context) error {
	n.LoadCalled = true
	n.Ctx = ctx
	n.Las = l
	n.ConvFactory = c
	n.Elev = e
	return nil
}
