package tiler

import (
	"testing"

	"github.com/mfbonfigli/gocesiumtiler/v2/tiler/mutator"
)

func TestOptions(t *testing.T) {
	m := mutator.NewSubsampler(0.5)
	opts := NewTilerOptions(
		WithCallback(func(event TilerEvent, filename string, elapsed int64, msg string) {}),
		WithEightBitColors(true),
		WithGridSize(11.1),
		WithMaxDepth(12),
		WithMinPointsPerTile(10),
		WithWorkerNumber(3),
		WithMutators([]mutator.Mutator{m}),
	)

	if opts.callback == nil {
		t.Errorf("unexpected nil callback")
	}
	if opts.eightBitColors != true {
		t.Errorf("expected eightbitcolor to be %v got %v", true, opts.eightBitColors)
	}
	if opts.gridSize != 11.1 {
		t.Errorf("expected gridSize to be %v got %v", 11.1, opts.gridSize)
	}
	if opts.maxDepth != 12 {
		t.Errorf("expected maxDepth to be %v got %v", 12, opts.maxDepth)
	}
	if opts.minPointsPerTile != 10 {
		t.Errorf("expected minPointsPerTile to be %v got %v", 10, opts.minPointsPerTile)
	}
	if opts.numWorkers != 3 {
		t.Errorf("expected numWorkers to be %v got %v", 3, opts.numWorkers)
	}
	if opts.mutators[0] != m && len(opts.mutators) != 1 {
		t.Error("expected 1 mutator to be registered")
	}
}
