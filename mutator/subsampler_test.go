package mutator

import (
	"testing"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
)

func TestSubsample(t *testing.T) {
	s := NewSubsampler(0.1)
	pt := geom.Point32{X: 1, Y: 2, Z: 3}
	out, keep := s.Mutate(pt, geom.Transform{})
	if !keep {
		// first point should always be kept
		t.Error("expected first Mutated point to be kept but was not")
	}
	if out != pt {
		t.Errorf("expected point %v, got %v", pt, out)
	}

	samples := 100000
	kept := 0
	for i := 0; i < samples; i++ {
		out, keep := s.Mutate(pt, geom.Transform{})
		if keep {
			kept++
			if out != pt {
				t.Errorf("expected point %v, got %v", pt, out)
			}
		}
	}
	// approximately 10000 pts should have been kept (0.1 or 10%)
	if kept < 9000 || kept > 11000 {
		t.Errorf("expected approx. %d of samples to be kept but %d were kept", samples/10, kept)
	}
}
