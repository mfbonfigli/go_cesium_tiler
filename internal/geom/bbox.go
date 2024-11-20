package geom

type BoundingBox struct {
	Xmin, Xmax, Ymin, Ymax, Zmin, Zmax, Xmid, Ymid, Zmid float64
}

// Constructor to properly initialize a boundingBox struct computing the mids
func NewBoundingBox(Xmin, Xmax, Ymin, Ymax, Zmin, Zmax float64) BoundingBox {
	bbox := BoundingBox{
		Xmin: Xmin,
		Xmax: Xmax,
		Ymin: Ymin,
		Ymax: Ymax,
		Zmin: Zmin,
		Zmax: Zmax,
		Xmid: (Xmin + Xmax) / 2,
		Ymid: (Ymin + Ymax) / 2,
		Zmid: (Zmin + Zmax) / 2,
	}
	return bbox
}

// Computes a bounding box from the given box and the given octant index
func NewBoundingBoxFromParent(parent BoundingBox, octant int) BoundingBox {
	var xMin, xMax, yMin, yMax, zMin, zMax float64
	switch octant {
	case 0, 2, 4, 6:
		xMin = parent.Xmin
		xMax = parent.Xmid
	case 1, 3, 5, 7:
		xMin = parent.Xmid
		xMax = parent.Xmax
	}
	switch octant {
	case 0, 1, 4, 5:
		yMin = parent.Ymin
		yMax = parent.Ymid
	case 2, 3, 6, 7:
		yMin = parent.Ymid
		yMax = parent.Ymax
	}
	switch octant {
	case 0, 1, 2, 3:
		zMin = parent.Zmin
		zMax = parent.Zmid
	case 4, 5, 6, 7:
		zMin = parent.Zmid
		zMax = parent.Zmax
	}
	return NewBoundingBox(xMin, xMax, yMin, yMax, zMin, zMax)
}

// AsCesiumBox returns the bounding box expressed according to the cesium "box" format
func (b BoundingBox) AsCesiumBox() [12]float64 {
	return [12]float64{
		b.Xmid, b.Ymid, b.Zmid,
		(b.Xmax - b.Xmid), 0, 0,
		0, (b.Ymax - b.Ymid), 0,
		0, 0, (b.Zmax - b.Zmid),
	}
}
