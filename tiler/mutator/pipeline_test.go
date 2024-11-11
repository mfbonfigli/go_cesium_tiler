package mutator

import (
	"testing"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
	"github.com/mfbonfigli/gocesiumtiler/v2/tiler/model"
)

type discardMutator struct{}

func (p *discardMutator) Mutate(pt model.Point, t model.Transform) (model.Point, bool) {
	return pt, false
}

func TestPipeline(t *testing.T) {
	p := NewPipeline(
		NewZOffset(1.5),
		NewZOffset(2.5),
	)
	actual, keep := p.Mutate(geom.NewPoint(1, 2, 3, 1, 2, 3, 4, 5), model.Transform{})
	expected := geom.NewPoint(1, 2, 7, 1, 2, 3, 4, 5)
	if actual != expected {
		t.Errorf("expected %v, got %v", expected, actual)
	}
	if !keep {
		t.Errorf("expected keep to be true but is false")
	}
}

func TestPipelineDiscard(t *testing.T) {
	p := NewPipeline(
		NewZOffset(1.5),
		&discardMutator{},
		NewZOffset(2.5),
	)
	actual, keep := p.Mutate(geom.NewPoint(1, 2, 3, 1, 2, 3, 4, 5), model.Transform{})
	expected := geom.NewPoint(1, 2, 4.5, 1, 2, 3, 4, 5)
	if actual != expected {
		t.Errorf("expected %v, got %v", expected, actual)
	}
	if keep {
		t.Errorf("expected point to be discarded but was not")
	}
}
