package mutator

import "github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"

// Pipeline is a mutator that applies all registered mutators sequentially
// and returns the result as output
type Pipeline struct {
	mutators []Mutator
}

func NewPipeline(m ...Mutator) *Pipeline {
	return &Pipeline{
		mutators: m,
	}
}

func (p *Pipeline) Mutate(pt geom.Point32, t geom.Transform) (geom.Point32, bool) {
	for _, m := range p.mutators {
		keep := true
		pt, keep = m.Mutate(pt, t)
		if !keep {
			return pt, false
		}
	}
	return pt, true
}
