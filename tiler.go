package tiler

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor/proj4"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/elev"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/elev/geoid2ellipsoid"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/las"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/tree"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/utils"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/writer"
)

type Tiler interface {
	ProcessFiles(inputLasFiles []string, outputFolder string, epsgCode int, opts *TilerOptions, ctx context.Context) error
	ProcessFolder(inputFolder, outputFolder string, epsgCode int, opts *TilerOptions, ctx context.Context) error
}

// GoCesiumTiler wraps the logic required to convert
// LAS point clouds into Cesium 3D tiles
type GoCesiumTiler struct {
	cconv coor.CoordinateConverter
	treeProvider
	writerProvider
	lasReaderProvider
}

type treeProvider func(opts *TilerOptions) tree.Tree
type writerProvider func(folder string, c coor.CoordinateConverter, opts *TilerOptions) (writer.Writer, error)
type lasReaderProvider func(inputLasFiles []string, epsgCode int, eightbit bool) (las.LasReader, error)

// NewGoCesiumTiler returns a new tiler to be used to convert LAS files into Cesium 3D Tiles
func NewGoCesiumTiler() (*GoCesiumTiler, error) {
	cconv, err := proj4.NewProj4CoordinateConverter()
	if err != nil {
		return nil, err
	}
	return &GoCesiumTiler{
		cconv: cconv,
		treeProvider: func(opts *TilerOptions) tree.Tree {
			return tree.NewGridTree(
				tree.WithGridSize(opts.gridSize),
				tree.WithMaxDepth(opts.maxDepth),
				tree.WithLoadWorkersNumber(opts.numWorkers),
				tree.WithMinPointsPerChildren(opts.minPointsPerTile),
			)
		},
		writerProvider: func(folder string, c coor.CoordinateConverter, opts *TilerOptions) (writer.Writer, error) {
			return writer.NewWriter(folder, c, writer.WithNumWorkers(opts.numWorkers))
		},
		lasReaderProvider: func(inputLasFiles []string, epsgCode int, eightbit bool) (las.LasReader, error) {
			return las.NewCombinedFileLasReader(inputLasFiles, epsgCode, eightbit)
		},
	}, nil
}

// ProcessFolder converts all LAS files found in the provided input folder converting them into separate tilesets
// each tileset is stored in a subdirectory in the outputFolder named after the filename
func (t *GoCesiumTiler) ProcessFolder(inputFolder, outputFolder string, epsgCode int, opts *TilerOptions, ctx context.Context) error {
	files, err := utils.FindLasFilesInFolder(inputFolder)
	if err != nil {
		return err
	}
	for _, f := range files {
		subfolderName := strings.TrimSuffix(filepath.Base(f), filepath.Ext(f))
		err := t.ProcessFiles([]string{f}, filepath.Join(outputFolder, subfolderName), epsgCode, opts, ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// ProcessFiles converts the specified LAS files as a single cesium tileset and stores them in the
func (t *GoCesiumTiler) ProcessFiles(inputLasFiles []string, outputFolder string, epsgCode int, opts *TilerOptions, ctx context.Context) error {
	start := time.Now()
	tr := t.treeProvider(opts)

	inputDesc := fmt.Sprintf("%d files", len(inputLasFiles))
	if len(inputLasFiles) == 1 {
		inputDesc = inputLasFiles[0]
	}

	// PARSE LAS HEADER
	emitEvent(EventReadLasHeaderStarted, opts, start, inputDesc, "start reading las")
	lasFile, err := t.lasReaderProvider(inputLasFiles, epsgCode, opts.eightBitColors)
	if err != nil {
		emitEvent(EventReadLasHeaderError, opts, start, inputDesc, fmt.Sprintf("las read error: %v", err))
		return err
	}
	emitEvent(EventReadLasHeaderCompleted, opts, start, inputDesc, fmt.Sprintf("las header read completed: found %d points", lasFile.NumberOfPoints()))

	// LOAD POINTS
	emitEvent(EventPointLoadingStarted, opts, start, inputDesc, "point loading started")
	elevationConverters := []elev.ElevationConverter{
		elev.NewOffsetElevationConverter(opts.elevationOffset),
	}
	if opts.geoidElevation {
		egmCalc, err := geoid2ellipsoid.NewEGMCalculator(t.cconv)
		if err != nil {
			emitEvent(EventPointLoadingError, opts, start, inputDesc, fmt.Sprintf("converter init error: %v", err))
			return err
		}
		elevationConverters = append(elevationConverters, elev.NewGeoidElevationConverter(epsgCode, egmCalc))
	}
	eConv := elev.NewPipelineElevationCorrector(elevationConverters...)
	err = tr.Load(lasFile, t.cconv, eConv, ctx)
	if err != nil {
		emitEvent(EventPointLoadingError, opts, start, inputDesc, fmt.Sprintf("load error: %v", err))
		return err
	}
	emitEvent(EventPointLoadingCompleted, opts, start, inputDesc, "point loading completed")

	// BUILD TREE
	emitEvent(EventBuildStarted, opts, start, inputDesc, "build started")
	err = tr.Build()
	if err != nil {
		emitEvent(EventBuildError, opts, start, inputDesc, fmt.Sprintf("build error: %v", err))
		return err
	}
	emitEvent(EventBuildCompleted, opts, start, inputDesc, "build completed")

	// EXPORT
	emitEvent(EventExportStarted, opts, start, inputDesc, "export started")
	w, err := t.writerProvider(outputFolder, t.cconv, opts)
	if err != nil {
		emitEvent(EventBuildError, opts, start, inputDesc, fmt.Sprintf("export init error: %v", err))
		return err
	}
	err = w.Write(tr, "", ctx)
	if err != nil {
		emitEvent(EventBuildError, opts, start, inputDesc, fmt.Sprintf("export error: %v", err))
		return err
	}
	emitEvent(EventExportStarted, opts, start, inputDesc, fmt.Sprintf("export completed in %v seconds", time.Since(start).String()))
	return nil
}

func emitEvent(e TilerEvent, opts *TilerOptions, start time.Time, inputDesc string, msg string) {
	if opts.callback != nil {
		opts.callback(e, inputDesc, time.Since(start).Milliseconds(), msg)
	}
}
