package mutator

import "github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"

// ZOffset is a mutator that shifts the points vertically for the given offset
type ZOffset struct {
	Offset float32
}

func NewZOffset(offset float32) *ZOffset {
	return &ZOffset{
		Offset: offset,
	}
}

func (z *ZOffset) Mutate(pt geom.Point32, t geom.Transform) (geom.Point32, bool) {
	pt.Z += z.Offset
	return pt, true
}
