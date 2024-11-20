package mutator

import (
	"testing"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
	"github.com/mfbonfigli/gocesiumtiler/v2/tiler/model"
)

func TestZOffset(t *testing.T) {
	actual, keep := NewZOffset(2).Mutate(geom.NewPoint(1, 2, 3, 1, 2, 3, 4, 5), model.Transform{})
	expected := geom.NewPoint(1, 2, 5, 1, 2, 3, 4, 5)
	if actual != expected {
		t.Errorf("expected %v, got %v", expected, actual)
	}
	if !keep {
		t.Errorf("expected keep to be true but is false")
	}
}
