package mutator

import (
	"math/rand"
	"sync/atomic"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
)

type Subsampler struct {
	Percentage float64
	first      *atomic.Bool
}

func NewSubsampler(percentage float64) *Subsampler {
	first := atomic.Bool{}
	first.Store(true)
	return &Subsampler{
		Percentage: percentage,
		first:      &first,
	}
}

func (s *Subsampler) Mutate(pt geom.Point32, t geom.Transform) (geom.Point32, bool) {
	if s.first.Load() {
		// always take the first point to ensure the point cloud has at least one point
		s.first.Swap(false)
		return pt, true
	}
	if rand.Float64() < s.Percentage {
		return pt, true
	}
	return pt, false
}
