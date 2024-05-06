package writer

import (
	"context"
	"math"
	"sync"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor/proj4"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/tree"
)

// Writer writes a tree as a 3D Cesium Point cloud to the given output folder
type Writer interface {
	Write(t tree.Tree, folderName string, ctx context.Context) error
}

type StandardWriter struct {
	numWorkers   int
	bufferRatio  int
	basePath     string
	conv         coor.CoordinateConverter
	producerFunc func(basepath, folder string) Producer
	consumerFunc func(coor.CoordinateConverter) Consumer
}

func NewWriter(basePath string, conv coor.CoordinateConverter, options ...func(*StandardWriter)) (*StandardWriter, error) {
	w := &StandardWriter{
		basePath:     basePath,
		numWorkers:   1,
		bufferRatio:  5,
		producerFunc: NewStandardProducer,
		consumerFunc: NewStandardConsumer,
	}
	for _, optFn := range options {
		optFn(w)
	}
	if w.conv == nil {
		conv, err := proj4.NewProj4CoordinateConverter()
		if err != nil {
			return nil, err
		}
		w.conv = conv
	}
	return w, nil
}

func WithNumWorkers(n int) func(*StandardWriter) {
	return func(w *StandardWriter) {
		w.numWorkers = n
	}
}

func WithBufferRatio(n int) func(*StandardWriter) {
	return func(w *StandardWriter) {
		w.bufferRatio = int(math.Max(1, float64(n)))
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
	go producer.Produce(workChannel, errorChannel, &waitGroup, t.GetRootNode(), ctx)

	// add consumers to waitgroup and launch them
	for i := 0; i < w.numWorkers; i++ {
		waitGroup.Add(1)
		consumer := w.consumerFunc(w.conv)
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
