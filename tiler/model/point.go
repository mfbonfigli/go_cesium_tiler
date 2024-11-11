package model

// Point models a point cloud point expressed in local, single precision, coordinates
type Point struct {
	X              float32
	Y              float32
	Z              float32
	R              uint8
	G              uint8
	B              uint8
	Intensity      uint8
	Classification uint8
}

// Vector returns a Vector representation of the position of the point in the local coordinate space
func (p Point) Vector() Vector {
	return Vector{
		X: float64(p.X),
		Y: float64(p.Y),
		Z: float64(p.Z),
	}
}
