package geom

import (
	"testing"
)

func TestNewBBox(t *testing.T) {
	actual := NewBoundingBox(-10, 0, 10, 20, -100, 20)
	expected := BoundingBox{
		Xmin: -10,
		Xmax: 0,
		Xmid: -5,
		Ymin: 10,
		Ymax: 20,
		Ymid: 15,
		Zmin: -100,
		Zmax: 20,
		Zmid: -40,
	}
	if actual != expected {
		t.Errorf("expected boundingbox %v got %v", expected, actual)
	}
}

func TestNewBBoxFromParent(t *testing.T) {
	parent := NewBoundingBox(-10, 0, 10, 20, -100, 20)
	actual := NewBoundingBoxFromParent(parent, 0)
	expected := BoundingBox{
		Xmin: -10,
		Xmax: -5,
		Xmid: -7.5,
		Ymin: 10,
		Ymax: 15,
		Ymid: 12.5,
		Zmin: -100,
		Zmax: -40,
		Zmid: -70,
	}
	if actual != expected {
		t.Errorf("expected boundingbox %v got %v", expected, actual)
	}

	actual = NewBoundingBoxFromParent(parent, 1)
	expected = BoundingBox{
		Xmin: -5,
		Xmax: -0,
		Xmid: -2.5,
		Ymin: 10,
		Ymax: 15,
		Ymid: 12.5,
		Zmin: -100,
		Zmax: -40,
		Zmid: -70,
	}
	if actual != expected {
		t.Errorf("expected boundingbox %v got %v", expected, actual)
	}

	actual = NewBoundingBoxFromParent(parent, 2)
	expected = BoundingBox{
		Xmin: -10,
		Xmax: -5,
		Xmid: -7.5,
		Ymin: 15,
		Ymax: 20,
		Ymid: 17.5,
		Zmin: -100,
		Zmax: -40,
		Zmid: -70,
	}
	if actual != expected {
		t.Errorf("expected boundingbox %v got %v", expected, actual)
	}

	actual = NewBoundingBoxFromParent(parent, 3)
	expected = BoundingBox{
		Xmin: -5,
		Xmax: -0,
		Xmid: -2.5,
		Ymin: 15,
		Ymax: 20,
		Ymid: 17.5,
		Zmin: -100,
		Zmax: -40,
		Zmid: -70,
	}
	if actual != expected {
		t.Errorf("expected boundingbox %v got %v", expected, actual)
	}

	actual = NewBoundingBoxFromParent(parent, 4)
	expected = BoundingBox{
		Xmin: -10,
		Xmax: -5,
		Xmid: -7.5,
		Ymin: 10,
		Ymax: 15,
		Ymid: 12.5,
		Zmin: -40,
		Zmax: 20,
		Zmid: -10,
	}
	if actual != expected {
		t.Errorf("expected boundingbox %v got %v", expected, actual)
	}

	actual = NewBoundingBoxFromParent(parent, 5)
	expected = BoundingBox{
		Xmin: -5,
		Xmax: -0,
		Xmid: -2.5,
		Ymin: 10,
		Ymax: 15,
		Ymid: 12.5,
		Zmin: -40,
		Zmax: 20,
		Zmid: -10,
	}
	if actual != expected {
		t.Errorf("expected boundingbox %v got %v", expected, actual)
	}

	actual = NewBoundingBoxFromParent(parent, 6)
	expected = BoundingBox{
		Xmin: -10,
		Xmax: -5,
		Xmid: -7.5,
		Ymin: 15,
		Ymax: 20,
		Ymid: 17.5,
		Zmin: -40,
		Zmax: 20,
		Zmid: -10,
	}
	if actual != expected {
		t.Errorf("expected boundingbox %v got %v", expected, actual)
	}

	actual = NewBoundingBoxFromParent(parent, 7)
	expected = BoundingBox{
		Xmin: -5,
		Xmax: -0,
		Xmid: -2.5,
		Ymin: 15,
		Ymax: 20,
		Ymid: 17.5,
		Zmin: -40,
		Zmax: 20,
		Zmid: -10,
	}
	if actual != expected {
		t.Errorf("expected boundingbox %v got %v", expected, actual)
	}
}

func TestGetAsArray(t *testing.T) {
	parent := NewBoundingBox(-10, 0, 10, 20, -100, 20)
	expected := [6]float64{-10, 10, 0, 20, -100, 20}
	if actual := parent.GetAsArray(); *(*[6]float64)(actual) != expected {
		t.Errorf("expected %v got %v", expected, actual)
	}
}
