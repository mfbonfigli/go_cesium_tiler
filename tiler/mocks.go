package tiler

import (
	"context"

	"github.com/mfbonfigli/gocesiumtiler/v2/tiler/mutator"
	"github.com/mfbonfigli/gocesiumtiler/v2/version"
)

type MockTiler struct {
	InputFiles          []string
	InputFolder         string
	OutputFolder        string
	SourceCRS           string
	Mutators            []mutator.Mutator
	Opts                *TilerOptions
	Ctx                 context.Context
	ProcessFilesCalled  bool
	ProcessFolderCalled bool
	// opts settings
	EightBit   bool
	GridSize   float64
	PtsPerTile int
	Depth      int
	Version    version.TilesetVersion
	err        error
}

func (m *MockTiler) ProcessFiles(inputLasFiles []string, outputFolder string, sourceCRS string, opts *TilerOptions, ctx context.Context) error {
	m.InputFiles = inputLasFiles
	m.OutputFolder = outputFolder
	m.SourceCRS = sourceCRS
	m.Opts = opts
	m.Ctx = ctx
	m.ProcessFilesCalled = true
	m.EightBit = opts.eightBitColors
	m.GridSize = opts.gridSize
	m.PtsPerTile = opts.minPointsPerTile
	m.Depth = opts.maxDepth
	m.Version = opts.version
	m.Mutators = opts.mutators
	return m.err
}

func (m *MockTiler) ProcessFolder(inputFolder, outputFolder string, sourceCRS string, opts *TilerOptions, ctx context.Context) error {
	m.InputFolder = inputFolder
	m.OutputFolder = outputFolder
	m.SourceCRS = sourceCRS
	m.Opts = opts
	m.Ctx = ctx
	m.ProcessFolderCalled = true
	m.EightBit = opts.eightBitColors
	m.GridSize = opts.gridSize
	m.PtsPerTile = opts.minPointsPerTile
	m.Depth = opts.maxDepth
	m.Version = opts.version
	m.Mutators = opts.mutators
	return m.err
}
