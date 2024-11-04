package tiler

import (
	"context"

	"github.com/mfbonfigli/gocesiumtiler/v2/version"
)

type MockTiler struct {
	InputFiles          []string
	InputFolder         string
	OutputFolder        string
	SourceCRS           string
	Opts                *TilerOptions
	Ctx                 context.Context
	ProcessFilesCalled  bool
	ProcessFolderCalled bool
	// opts settings
	EightBit   bool
	GridSize   float64
	PtsPerTile int
	Depth      int
	ElevOffset float64
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
	m.ElevOffset = opts.elevationOffset
	m.Version = opts.version
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
	m.ElevOffset = opts.elevationOffset
	m.Version = opts.version
	return m.err
}
