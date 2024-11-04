package tiler

import (
	"testing"
)

func TestOptions(t *testing.T) {
	opts := NewTilerOptions(
		WithCallback(func(event TilerEvent, filename string, elapsed int64, msg string) {}),
		WithEightBitColors(true),
		WithElevationOffset(1),
		WithGridSize(11.1),
		WithMaxDepth(12),
		WithMinPointsPerTile(10),
		WithWorkerNumber(3),
	)

	if opts.callback == nil {
		t.Errorf("unexpected nil callback")
	}
	if opts.eightBitColors != true {
		t.Errorf("expected eightbitcolor to be %v got %v", true, opts.eightBitColors)
	}
	if opts.elevationOffset != 1 {
		t.Errorf("expected elevationOffset to be %v got %v", 1, opts.elevationOffset)
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
}
