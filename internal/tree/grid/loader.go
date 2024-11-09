package grid

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/elev"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/las"
)

// loader is a utility class used to read points from a las reader and initialize
// a grid Node with the correct data
type loader struct {
	createCoorConverter coor.ConverterFactory
	elevationConverter  elev.Converter
	workers             int
}

// load the points from the LasReader r into the node n
func (l *loader) load(n *Node, r las.LasReader, ctx context.Context) (err error) {
	numPts := r.NumberOfPoints()
	if numPts == 0 {
		return fmt.Errorf("las with no points")
	}
	subCtx, cancelFunc := context.WithCancel(ctx)

	// Store all the points in a continuous memory space
	// While not required, storing points in a contiguous array makes
	// the system more CPU cache friendly and thus measurably faster
	backingArray := make([]geom.LinkedPoint, numPts)

	// read first point
	first, err := r.GetNext()
	if err != nil {
		return err
	}

	c, err := l.createCoorConverter()
	if err != nil {
		return err
	}
	pt, err := transformPoint(first, c, l.elevationConverter, r.GetCRS())
	if err != nil {
		return err
	}
	// use first point as baseline
	transform, base := l.baseline(pt)
	backingArray[0] = base

	// init concurrent vars
	var wg sync.WaitGroup
	var errchan chan error = make(chan error)

	// launch consumers
	basePtPerWorker := (numPts - 1) / l.workers
	residual := (numPts - 1) - basePtPerWorker*l.workers
	start := 1
	consumers := []*consumer{}
	for i := 0; i < l.workers; i++ {
		workerPtsNum := basePtPerWorker
		if i < residual {
			workerPtsNum += 1
		}
		conv, err := l.createCoorConverter()
		if err != nil {
			cancelFunc()
			return err
		}
		consumers = append(consumers, newConsumer(i, start, workerPtsNum, conv, l.elevationConverter, r.GetCRS(), &backingArray))
		wg.Add(1)
		go consumers[i].consume(r, transform, errchan, &wg, subCtx)
		start = start + workerPtsNum
	}

	// retrieve errors
	errs := []error{}
	errWg := &sync.WaitGroup{}
	errWg.Add(1)
	go func() {
		defer errWg.Done()
		for {
			err, ok := <-errchan
			if !ok {
				return
			}
			errs = append(errs, err)
		}
	}()
	wg.Wait()
	close(errchan)
	errWg.Wait()

	if len(errs) != 0 {
		return errs[0]
	}

	var pts *geom.LinkedPoint
	var bboxbuilder *boundingBoxBuilder

	// merge consumer points and bounding boxes
	for _, c := range consumers {
		c.endPt.Next = pts
		pts = c.startPt
		if bboxbuilder == nil {
			bboxbuilder = c.bboxBuilder
		} else {
			bboxbuilder.mergeWith(c.bboxBuilder)
		}
	}
	bboxbuilder.processPoint(base.Pt.X, base.Pt.Y, base.Pt.Z)
	bbox := bboxbuilder.build()
	base.Next = pts

	// set data into the gridnode
	n.bounds = bbox
	n.pts = &base
	n.transform = &transform
	return nil
}

// baseline returns a transform from EPSG 4978 to a local cartesian system centered in pt
// and with Z normal to the WGS84 ellipsoid
func (l *loader) baseline(pt geom.Point64) (geom.Transform, geom.LinkedPoint) {
	transform := geom.LocalCRSFromPoint(pt.X, pt.Y, pt.Z)
	baselinePtLocalCoords := pt.ToLocal(transform)
	return transform, geom.LinkedPoint{Pt: baselinePtLocalCoords}
}

// consumer consumes points from the las reader storing them internally and updating its internal bounds
type consumer struct {
	id           int
	start        int
	count        int
	conv         coor.Converter
	elev         elev.Converter
	crs          string
	backingArray *[]geom.LinkedPoint
	// output vars
	bboxBuilder *boundingBoxBuilder
	startPt     *geom.LinkedPoint
	endPt       *geom.LinkedPoint
}

func newConsumer(id int, start int, count int, conv coor.Converter, elev elev.Converter, crs string, backingArray *[]geom.LinkedPoint) *consumer {
	return &consumer{
		id:           id,
		start:        start,
		count:        count,
		conv:         conv,
		elev:         elev,
		crs:          crs,
		backingArray: backingArray,
		bboxBuilder:  newBoundingBoxBuilder(),
	}
}

func (c *consumer) consume(r las.LasReader, transform geom.Transform, errchan chan error, wg *sync.WaitGroup, ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			errchan <- fmt.Errorf("panic while reading from las: %v", r)
		}
	}()
	defer c.conv.Cleanup()
	defer wg.Done()
	var currentPt *geom.LinkedPoint
	i := 0
	for i < c.count {
		if err := ctx.Err(); err != nil {
			errchan <- err
			return
		}
		pt, err := r.GetNext()
		if err != nil {
			errchan <- err
			return
		}

		pt, err = transformPoint(pt, c.conv, c.elev, c.crs)
		if err != nil {
			errchan <- err
			return
		}

		// transform local and update bounds
		localPt := pt.ToLocal(transform)
		c.bboxBuilder.processPoint(localPt.X, localPt.Y, localPt.Z)

		// store point in backing array at the right offset
		(*c.backingArray)[c.start+i] = geom.LinkedPoint{Pt: localPt}
		newPt := &((*c.backingArray)[c.start+i])
		if currentPt == nil {
			currentPt = newPt
			c.startPt = currentPt
		} else {
			currentPt.Next = newPt
			currentPt = newPt
			c.endPt = currentPt
		}
		i++
	}
}

// transformPoint converts a point from the original CRS to EPSG:4978 eventually also manipulating the point elevation
func transformPoint(pt geom.Point64, conv coor.Converter, eConv elev.Converter, sourceCRS string) (geom.Point64, error) {
	var err error
	z := pt.Z
	if eConv != nil {
		z, err = eConv.ConvertElevation(pt.X, pt.Y, pt.Z)
		if err != nil {
			return pt, err
		}
	}

	out, err := conv.ToWGS84Cartesian(sourceCRS, geom.Vector3{X: pt.X, Y: pt.Y, Z: z})
	if err != nil {
		return pt, err
	}
	pt.X, pt.Y, pt.Z = out.X, out.Y, out.Z
	return pt, nil
}

// boundingBoxBuilder is a utility struct to compute bounds from input points
type boundingBoxBuilder struct {
	minX, minY, minZ, maxX, maxY, maxZ float64
}

func newBoundingBoxBuilder() *boundingBoxBuilder {
	return &boundingBoxBuilder{
		minX: math.Inf(1),
		minY: math.Inf(1),
		minZ: math.Inf(1),
		maxX: math.Inf(-1),
		maxY: math.Inf(-1),
		maxZ: math.Inf(-1),
	}
}

// processPoint examines the input coordinates and expands the bounds if necessary
func (b *boundingBoxBuilder) processPoint(x, y, z float32) {
	if float64(x) < b.minX {
		b.minX = float64(x)
	}
	if float64(y) < b.minY {
		b.minY = float64(y)
	}
	if float64(z) < b.minZ {
		b.minZ = float64(z)
	}
	if float64(x) > b.maxX {
		b.maxX = float64(x)
	}
	if float64(y) > b.maxY {
		b.maxY = float64(y)
	}
	if float64(z) > b.maxZ {
		b.maxZ = float64(z)
	}
}

// mergeWith merges the current boundingBoxBuilder bounds with the input
// boundingBoxBuilder
func (b *boundingBoxBuilder) mergeWith(o *boundingBoxBuilder) {
	if o.minX < b.minX {
		b.minX = o.minX
	}
	if o.minY < b.minY {
		b.minY = o.minY
	}
	if o.minZ < b.minZ {
		b.minZ = o.minZ
	}
	if o.maxX > b.maxX {
		b.maxX = o.maxX
	}
	if o.maxY > b.maxY {
		b.maxY = o.maxY
	}
	if o.maxZ > b.maxZ {
		b.maxZ = o.maxZ
	}
}

// build returns a BoundingBox instance from the current builder bounds
func (b *boundingBoxBuilder) build() geom.BoundingBox {
	return geom.NewBoundingBox(b.minX, b.maxX, b.minY, b.maxY, b.minZ, b.maxZ)
}