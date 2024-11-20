package geom

import (
	"fmt"

	"github.com/mfbonfigli/gocesiumtiler/v2/tiler/model"
)

// Point64 contains data of a Point Cloud Point, namely X,Y,Z coords,
// R,G,B color components, Intensity and Classification. Coordinates are expressed
// as double precision float64 numbers.
type Point64 struct {
	model.Vector
	R              uint8
	G              uint8
	B              uint8
	Intensity      uint8
	Classification uint8
}

// Builds a new model.Point from the given coordinates, colors, intensity and classification values
func NewPoint(X, Y, Z float32, R, G, B, Intensity, Classification uint8) model.Point {
	return model.Point{
		X:              X,
		Y:              Y,
		Z:              Z,
		R:              R,
		G:              G,
		B:              B,
		Intensity:      Intensity,
		Classification: Classification,
	}
}

// PointList models a list of model.Point. Points are immutable and returned by value.
type PointList interface {
	Len() int
	Next() (model.Point, error)
	Reset()
}

// LinkedPoint wraps a model.Point to create a Linked List
type LinkedPoint struct {
	Next *LinkedPoint
	Pt   model.Point
}

// LinkedPointStream is a wrapper helper that allows a LinkedPoint to implement the PointList interface
type LinkedPointStream struct {
	len     int
	current *LinkedPoint
	start   *LinkedPoint
}

// NewLinkedPointStream initializes a linked stream from the given root.
// the length is not cross-verified, it must be coherent with the actual point count in the linked list.
func NewLinkedPointStream(root *LinkedPoint, len int) *LinkedPointStream {
	return &LinkedPointStream{
		len:     len,
		current: root,
		start:   root,
	}
}

func (l *LinkedPointStream) Next() (model.Point, error) {
	if l.current == nil {
		return model.Point{}, fmt.Errorf("no more points")
	}
	pt := l.current.Pt
	l.current = l.current.Next
	return pt, nil
}

func (l *LinkedPointStream) Len() int {
	return l.len
}

func (l *LinkedPointStream) Reset() {
	l.current = l.start
}
