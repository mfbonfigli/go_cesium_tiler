package writer

import (
	"context"
	"math"
	"sync"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/tree"
	"github.com/mfbonfigli/gocesiumtiler/v2/version"
)

// Writer writes a tree as a 3D Cesium Point cloud to the given output folder
type Writer interface {
	Write(t tree.Tree, folderName string, ctx context.Context) error
}

type StandardWriter struct {
	numWorkers   int
	bufferRatio  int
	basePath     string
	version      version.TilesetVersion
	producerFunc func(basepath, folder string) Producer
	consumerFunc func(version.TilesetVersion) Consumer
}

func NewWriter(basePath string, options ...func(*StandardWriter)) (*StandardWriter, error) {
	w := &StandardWriter{
		basePath:     basePath,
		numWorkers:   1,
		bufferRatio:  5,
		version:      version.TilesetVersion_1_0,
		producerFunc: NewStandardProducer,
		consumerFunc: func(v version.TilesetVersion) Consumer {
			if v == version.TilesetVersion_1_0 {
				return NewStandardConsumer(WithGeometryEncoder(NewPntsEncoder()))
			}
			return NewStandardConsumer(WithGeometryEncoder(NewGltfEncoder()))
		},
	}
	for _, optFn := range options {
		optFn(w)
	}
	return w, nil
}

// WithNumWorkers defines how many writer goroutines to launch when writing the tiles.
func WithNumWorkers(n int) func(*StandardWriter) {
	return func(w *StandardWriter) {
		w.numWorkers = n
	}
}

// WithBufferRation defines how many jobs per writer worker to allow enqueuing.
func WithBufferRatio(n int) func(*StandardWriter) {
	return func(w *StandardWriter) {
		w.bufferRatio = int(math.Max(1, float64(n)))
	}
}

// WithTilesetVersion sets the version of the generated tilesets. version 1.0 generates .pnts gemetries
// while version 1.1 generates .glb (gltf) geometries.
func WithTilesetVersion(v version.TilesetVersion) func(*StandardWriter) {
	return func(w *StandardWriter) {
		w.version = v
	}
}

func (w *StandardWriter) Write(t tree.Tree, folderName string, ctx context.Context) error {
	// init channel where consumers can eventually submit errors that prevented them to finish the job
	errorChannel := make(chan error)

	// init channel where to submit work with a buffer N times greater than the number of consumer
	workChannel := make(chan *WorkUnit, w.numWorkers*w.bufferRatio)

	var waitGroup sync.WaitGroup
	var errorWaitGroup sync.WaitGroup

	// producing is easy, only 1 producer
	producer := w.producerFunc(w.basePath, folderName)
	waitGroup.Add(1)
	go producer.Produce(workChannel, errorChannel, &waitGroup, t.RootNode(), ctx)

	// add consumers to waitgroup and launch them
	for i := 0; i < w.numWorkers; i++ {
		waitGroup.Add(1)
		// instantiate a new converter per each goroutine for thread safety
		consumer := w.consumerFunc(w.version)
		go consumer.Consume(workChannel, errorChannel, &waitGroup)
	}

	// launch error listener
	errs := []error{}
	errorWaitGroup.Add(1)
	go func() {
		defer errorWaitGroup.Done()
		for {
			err, ok := <-errorChannel
			if !ok {
				return
			}
			errs = append(errs, err)
		}
	}()

	// wait for producers and consumers to finish
	waitGroup.Wait()

	// close error chan
	close(errorChannel)
	errorWaitGroup.Wait()

	if len(errs) != 0 {
		return errs[0]
	}
	return nil
}
