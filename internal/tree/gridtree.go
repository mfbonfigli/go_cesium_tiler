package tree

import (
	"context"
	"math"
	"sync"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/elev"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/las"
)

// GridTreeNode implements both the Tree and Node interfaces. The points of the point cloud
// are internally stored in EPSG 4978, which is a metric, cartesian CRS and the same internal
// reference system of Cesium. The sampling is performed by determining a virtual "grid" at each level
// of detail. The grid has a spacing in meters.
//
// The build operation will:
//   - partition the space according to the grid
//   - given a space partition, retain the point that belongs to the partition and that is closest to its center
//     unless the maximum depth of the tree is reached, in which case all points are retained.
//   - store all other points no retained to be used to build the children
//
// The tree is "lazy". It never builds the children until they are queried.
type GridTreeNode struct {
	cX, cY, cZ           float64
	pts                  *geom.LinkedPoint
	childrenPts          [8]*geom.LinkedPoint
	children             [8]Node
	childrenBuilt        bool
	bounds               geom.BoundingBox
	gridSize             float64
	built                bool
	maxDepth             int
	depth                int
	numPoints            int
	totalNumPoints       int
	loadWorkersNumber    int
	minPointsPerChildren int
	sync.Mutex
}

func NewGridTree(opts ...func(*GridTreeNode)) *GridTreeNode {
	t := &GridTreeNode{
		built:                false,
		maxDepth:             10,
		depth:                0,
		childrenBuilt:        false,
		gridSize:             1,
		loadWorkersNumber:    1,
		minPointsPerChildren: 10000,
	}
	for _, optFn := range opts {
		optFn(t)
	}
	return t
}

func WithGridSize(size float64) func(t *GridTreeNode) {
	return func(t *GridTreeNode) {
		t.gridSize = size
	}
}

func WithMaxDepth(depth int) func(t *GridTreeNode) {
	return func(t *GridTreeNode) {
		t.maxDepth = depth
	}
}

func WithLoadWorkersNumber(num int) func(t *GridTreeNode) {
	return func(t *GridTreeNode) {
		t.loadWorkersNumber = num
	}
}

func WithMinPointsPerChildren(num int) func(t *GridTreeNode) {
	return func(t *GridTreeNode) {
		t.minPointsPerChildren = num
	}
}

func (t *GridTreeNode) Load(reader las.LasReader, coorConv coor.ConverterFactory, elevConv elev.ElevationConverter, ctx context.Context) error {
	return t.loadPoints(reader, coorConv, elevConv, ctx)
}

func (t *GridTreeNode) IsBuilt() bool {
	return t.built
}

func (t *GridTreeNode) GetRootNode() Node {
	if t.depth == 0 {
		return Node(t)
	}
	return nil
}

func (t *GridTreeNode) Build() error {
	if t.depth >= t.maxDepth {
		// reached maxDepth, swallow in all points
		current := t.pts
		// traverse just to update the internal counters
		for current != nil {
			t.totalNumPoints++
			t.numPoints++
			current = current.Next
		}
		// max depth, no further subdivision possible, mark as built and return
		t.built = true
		return nil
	}

	// nX, nY, nZ represent the number of grid cells in each direction, should always be >= 1
	nX := math.Ceil((t.bounds.Xmax - t.bounds.Xmin) / t.gridSize)
	nY := math.Ceil((t.bounds.Ymax - t.bounds.Ymin) / t.gridSize)
	nZ := math.Ceil((t.bounds.Zmax - t.bounds.Zmin) / t.gridSize)

	// these are the actual gridSizes after the rounding
	gridSizeX := (t.bounds.Xmax - t.bounds.Xmin) / nX
	gridSizeY := (t.bounds.Ymax - t.bounds.Ymin) / nY
	gridSizeZ := (t.bounds.Zmax - t.bounds.Zmin) / nZ

	// we need to keep track of the closest point to each grid cell center
	// define an inner type so that it's not leaked outside the scope of the build method
	type cell struct {
		pt   *geom.LinkedPoint
		dist float64
	}

	childrenCount := [8]int{}

	// start from the first point
	cur := t.pts

	// the winners (i.e. closest points to each cell center are stored in a map)
	// the key to the map is a [3]float array of the grid cell center.
	grid := map[[3]int32]cell{}

	for cur != nil {
		// keep track of the number of points seen overall
		t.totalNumPoints++
		// store the next point for the next iteration in the loop,
		// then detach the current point from the linked list by wiping the 'next' pointer
		next := cur.Next
		cur.Next = nil

		// compute 3D integer coordinates of the cell the point falls into
		iX := int32(math.Min(math.Max(1, math.Ceil((float64(cur.Pt.X)-t.bounds.Xmin)/gridSizeX)), float64(nX)))
		iY := int32(math.Min(math.Max(1, math.Ceil((float64(cur.Pt.Y)-t.bounds.Ymin)/gridSizeY)), float64(nY)))
		iZ := int32(math.Min(math.Max(1, math.Ceil((float64(cur.Pt.Z)-t.bounds.Zmin)/gridSizeZ)), float64(nZ)))
		// this is the unique id of the cell the point belongs to
		cellIndex := [3]int32{iX, iY, iZ}

		// compute the cell center coordinates
		cX := t.bounds.Xmin + float64(iX-1)*gridSizeX + gridSizeX/2
		cY := t.bounds.Ymin + float64(iY-1)*gridSizeY + gridSizeY/2
		cZ := t.bounds.Zmin + float64(iZ-1)*gridSizeZ + gridSizeZ/2

		// get the (squared, to save some CPU) distance of the point to the cell center
		curDist := (cX-float64(cur.Pt.X))*(cX-float64(cur.Pt.X)) + (cY-float64(cur.Pt.Y))*(cY-float64(cur.Pt.Y)) + (cZ-float64(cur.Pt.Z))*(cZ-float64(cur.Pt.Z))

		// find if we already have a "winner" (closest point) for the identified grid cell
		oldWinner, ok := grid[cellIndex]
		if !ok {
			// no winner? then the current point is the new cell winner
			grid[cellIndex] = cell{pt: cur, dist: curDist}
		} else {
			// we have a winner, check if it loses against the current point
			if curDist < oldWinner.dist {
				// current point wins, old winner needs to go
				grid[cellIndex] = cell{pt: cur, dist: curDist}
				// oldWinner needs to be moved to the linked list of the child octant it belongs to
				idx := t.getChildrenIndex(oldWinner.pt.Pt)
				childrenCount[idx]++

				if t.childrenPts[idx] == nil {
					t.childrenPts[idx] = oldWinner.pt
				} else {
					oldWinner.pt.Next = t.childrenPts[idx]
					t.childrenPts[idx] = oldWinner.pt
				}
			} else {
				// oldWinner wins against current point, so just push current point to the
				// relevant octant list
				idx := t.getChildrenIndex(cur.Pt)
				childrenCount[idx]++
				if t.childrenPts[idx] == nil {
					t.childrenPts[idx] = cur
				} else {
					cur.Next = t.childrenPts[idx]
					t.childrenPts[idx] = cur
				}
			}
		}
		// update cur with the next one
		cur = next
	}

	// now we need to extract all points in the map as they are
	// the ones left belonging to this node
	t.pts = nil
	for _, pt := range grid {
		point := pt.pt
		point.Next = t.pts
		t.pts = point
		t.numPoints++
	}

	// are we done? Not really. If there are children with a number of points < minPointsPerChildren
	// then merge them with the current node
	for i, count := range childrenCount {
		if count < t.minPointsPerChildren {
			current := t.childrenPts[i]
			for current != nil {
				next := current.Next
				current.Next = t.pts
				t.pts = current
				current = next
				t.numPoints++
			}
			t.childrenPts[i] = nil
		}
	}
	t.built = true
	return nil
}

func (t *GridTreeNode) GetInternalCRS() string {
	return "EPSG:4978"
}

func (t *GridTreeNode) IsRoot() bool {
	return t.depth == 0
}

func (t *GridTreeNode) GetBoundingBoxRegion(conv coor.CoordinateConverter) (geom.BoundingBox, error) {
	// the bounds are in a 3D earth centric coordinate system
	// to be translated in Lat Lon EPSG 4979 we convert each corner of the box and
	// take the min/max latitude, longitude and elevation of each point
	bbox := t.bounds
	p1 := geom.Coord{
		X: bbox.Xmin + t.cX,
		Y: bbox.Ymin + t.cY,
		Z: bbox.Zmin + t.cZ,
	}
	p2 := geom.Coord{
		X: bbox.Xmax + t.cX,
		Y: bbox.Ymin + t.cY,
		Z: bbox.Zmin + t.cZ,
	}
	p3 := geom.Coord{
		X: bbox.Xmin + t.cX,
		Y: bbox.Ymax + t.cY,
		Z: bbox.Zmin + t.cZ,
	}
	p4 := geom.Coord{
		X: bbox.Xmax + t.cX,
		Y: bbox.Ymax + t.cY,
		Z: bbox.Zmin + t.cZ,
	}
	p5 := geom.Coord{
		X: bbox.Xmin + t.cX,
		Y: bbox.Ymin + t.cY,
		Z: bbox.Zmax + t.cZ,
	}
	p6 := geom.Coord{
		X: bbox.Xmax + t.cX,
		Y: bbox.Ymin + t.cY,
		Z: bbox.Zmax + t.cZ,
	}
	p7 := geom.Coord{
		X: bbox.Xmin + t.cX,
		Y: bbox.Ymax + t.cY,
		Z: bbox.Zmax + t.cZ,
	}
	p8 := geom.Coord{
		X: bbox.Xmax + t.cX,
		Y: bbox.Ymax + t.cY,
		Z: bbox.Zmax + t.cZ,
	}
	target := "EPSG:4979"
	p1c, err := conv.Transform(t.GetInternalCRS(), target, p1)
	if err != nil {
		return geom.BoundingBox{}, err
	}
	p2c, err := conv.Transform(t.GetInternalCRS(), target, p2)
	if err != nil {
		return geom.BoundingBox{}, err
	}
	p3c, err := conv.Transform(t.GetInternalCRS(), target, p3)
	if err != nil {
		return geom.BoundingBox{}, err
	}
	p4c, err := conv.Transform(t.GetInternalCRS(), target, p4)
	if err != nil {
		return geom.BoundingBox{}, err
	}
	p5c, err := conv.Transform(t.GetInternalCRS(), target, p5)
	if err != nil {
		return geom.BoundingBox{}, err
	}
	p6c, err := conv.Transform(t.GetInternalCRS(), target, p6)
	if err != nil {
		return geom.BoundingBox{}, err
	}
	p7c, err := conv.Transform(t.GetInternalCRS(), target, p7)
	if err != nil {
		return geom.BoundingBox{}, err
	}
	p8c, err := conv.Transform(t.GetInternalCRS(), target, p8)
	if err != nil {
		return geom.BoundingBox{}, err
	}
	minFunc := func(nums ...float64) float64 {
		min := math.MaxFloat64
		for _, num := range nums {
			if num < min {
				min = num
			}
		}
		return min
	}
	maxFunc := func(nums ...float64) float64 {
		min := -1 * math.MaxFloat64
		for _, num := range nums {
			if num > min {
				min = num
			}
		}
		return min
	}

	return geom.NewBoundingBox(
		minFunc(p1c.X, p2c.X, p3c.X, p4c.X, p5c.X, p6c.X, p7c.X, p8c.X)*math.Pi/180,
		maxFunc(p1c.X, p2c.X, p3c.X, p4c.X, p5c.X, p6c.X, p7c.X, p8c.X)*math.Pi/180,
		minFunc(p1c.Y, p2c.Y, p3c.Y, p4c.Y, p5c.Y, p6c.Y, p7c.Y, p8c.Y)*math.Pi/180,
		maxFunc(p1c.Y, p2c.Y, p3c.Y, p4c.Y, p5c.Y, p6c.Y, p7c.Y, p8c.Y)*math.Pi/180,
		minFunc(p1c.Z, p2c.Z, p3c.Z, p4c.Z, p5c.Z, p6c.Z, p7c.Z, p8c.Z),
		maxFunc(p1c.Z, p2c.Z, p3c.Z, p4c.Z, p5c.Z, p6c.Z, p7c.Z, p8c.Z),
	), nil
}

func (t *GridTreeNode) GetChildren() [8]Node {
	t.Lock()
	defer t.Unlock()
	if t.childrenBuilt {
		return t.children
	}
	t.children = [8]Node{}
	if !t.built {
		// not built? return nothing
		return t.children
	}
	for i, c := range t.childrenPts {
		if c == nil {
			continue
		}
		v := &GridTreeNode{
			pts:                  c,
			childrenPts:          [8]*geom.LinkedPoint{},
			bounds:               geom.NewBoundingBoxFromParent(t.bounds, i),
			depth:                t.depth + 1,
			maxDepth:             t.maxDepth,
			gridSize:             t.gridSize / 2,
			childrenBuilt:        false,
			minPointsPerChildren: t.minPointsPerChildren,
			cX:                   t.cX,
			cY:                   t.cY,
			cZ:                   t.cZ,
		}
		// Children MUST be built before returned
		v.Build()
		t.children[i] = Node(v)
	}
	t.childrenBuilt = true
	return t.children
}

func (t *GridTreeNode) GetPoints(c coor.CoordinateConverter) geom.Point32List {
	return geom.NewLinkedPointStream(t.pts, t.numPoints)
}

func (t *GridTreeNode) TotalNumberOfPoints() int {
	return t.totalNumPoints
}

func (t *GridTreeNode) NumberOfPoints() int {
	return t.numPoints
}

func (t *GridTreeNode) IsLeaf() bool {
	for _, v := range t.GetChildren() {
		if v != nil {
			return false
		}
	}
	return true
}

func (t *GridTreeNode) ComputeGeometricError() float64 {
	return math.Sqrt(t.gridSize * t.gridSize * 3)
}

func (t *GridTreeNode) getChildrenIndex(p geom.Point32) int {
	if float64(p.X) < t.bounds.Xmid && float64(p.Y) < t.bounds.Ymid && float64(p.Z) < t.bounds.Zmid {
		return 0
	} else if float64(p.X) >= t.bounds.Xmid && float64(p.Y) < t.bounds.Ymid && float64(p.Z) < t.bounds.Zmid {
		return 1
	} else if float64(p.X) < t.bounds.Xmid && float64(p.Y) >= t.bounds.Ymid && float64(p.Z) < t.bounds.Zmid {
		return 2
	} else if float64(p.X) >= t.bounds.Xmid && float64(p.Y) >= t.bounds.Ymid && float64(p.Z) < t.bounds.Zmid {
		return 3
	} else if float64(p.X) < t.bounds.Xmid && float64(p.Y) < t.bounds.Ymid && float64(p.Z) >= t.bounds.Zmid {
		return 4
	} else if float64(p.X) >= t.bounds.Xmid && float64(p.Y) < t.bounds.Ymid && float64(p.Z) >= t.bounds.Zmid {
		return 5
	} else if float64(p.X) < t.bounds.Xmid && float64(p.Y) >= t.bounds.Ymid && float64(p.Z) >= t.bounds.Zmid {
		return 6
	}
	return 7
}

func (t *GridTreeNode) loadPoints(reader las.LasReader, convFactory coor.ConverterFactory, eConv elev.ElevationConverter, ctx context.Context) error {
	numPts := reader.NumberOfPoints()

	// Store all the points in a continuous memory space
	// While not required, storing points in a contiguous array makes
	// the system more CPU cache friendly and thus measurably faster
	backingArray := make([]geom.LinkedPoint, numPts)

	// all coordinates are referred as relative to the coordinates of the first point
	baselinePt, err := reader.GetNext()
	if err != nil {
		return err
	}

	conv, err := convFactory()
	if err != nil {
		return err
	}
	defer conv.Cleanup()
	baselinePt, err = t.transformPoint(baselinePt, conv, eConv, reader.GetCRS())
	if err != nil {
		return err
	}
	backingArray[0] = geom.LinkedPoint{Pt: baselinePt.ToPointFromBaseline(baselinePt)}
	baselineGeomPt := &backingArray[0]

	minX := baselinePt.X
	minY := baselinePt.Y
	minZ := baselinePt.Z
	maxX := baselinePt.X
	maxY := baselinePt.Y
	maxZ := baselinePt.Z

	var wg sync.WaitGroup
	var mutex sync.Mutex
	var errchan chan error = make(chan error)
	var ptchan chan geom.Point64 = make(chan geom.Point64, t.loadWorkersNumber*10)

	// the consumers store their artifacts in these structures
	startPts := make([]*geom.LinkedPoint, t.loadWorkersNumber)
	endPts := make([]*geom.LinkedPoint, t.loadWorkersNumber)
	averages := make([][3]float64, t.loadWorkersNumber)
	ptCounts := make([]int, t.loadWorkersNumber)

	wg.Add(1)

	// PRODUCER: reads the points one after another and pushes them to a channel
	produce := func() {
		defer close(ptchan)
		defer wg.Done()
		for i := 1; i < numPts; i++ { // 1 point was already consumed (baseline pt)
			if err := ctx.Err(); err != nil {
				errchan <- err
				return
			}
			pt, err := reader.GetNext()
			if err != nil {
				errchan <- err
				return
			}
			ptchan <- pt
		}
	}

	// CONSUMER: reads from the channel, transforms and stores the points
	consume := func(workerId int, start int, count int) {
		conv, err := convFactory()
		if err != nil {
			errchan <- err
			return
		}
		defer conv.Cleanup()
		defer wg.Done()
		var curNode *geom.LinkedPoint
		c := 0
		for c < count {
			if err := ctx.Err(); err != nil {
				errchan <- err
				return
			}
			// get work from channel
			pt, ok := <-ptchan
			if !ok {
				// channel was closed by producer, quit infinite loop
				return
			}

			pt, err := t.transformPoint(pt, conv, eConv, reader.GetCRS())
			if err != nil {
				errchan <- err
				return
			}

			averages[workerId][0] = (averages[workerId][0]*float64(ptCounts[workerId]) + pt.X)
			ptCounts[workerId]++
			// update bounds estimation
			mutex.Lock()
			minX = math.Min(float64(pt.X), minX)
			minY = math.Min(float64(pt.Y), minY)
			minZ = math.Min(float64(pt.Z), minZ)
			maxX = math.Max(float64(pt.X), maxX)
			maxY = math.Max(float64(pt.Y), maxY)
			maxZ = math.Max(float64(pt.Z), maxZ)
			backingArray[start+c] = geom.LinkedPoint{Pt: pt.ToPointFromBaseline(baselinePt)}
			newNode := &backingArray[start+c]
			if curNode == nil {
				curNode = newNode
				startPts[workerId] = curNode
			} else {
				curNode.Next = newNode
				curNode = newNode
				endPts[workerId] = curNode
			}
			mutex.Unlock()
			c++
		}
	}

	go produce()

	// Launch the consumers, but first compute how many points each should read
	// one point was already read at the beginning, so remove it from the total point count
	basePtPerWorker := (numPts - 1) / t.loadWorkersNumber
	residual := (numPts - 1) - basePtPerWorker*t.loadWorkersNumber
	start := 1
	for i := 0; i < t.loadWorkersNumber; i++ {
		workerPtsNum := basePtPerWorker
		if i < residual {
			workerPtsNum += 1
		}
		wg.Add(1)
		go consume(i, start, workerPtsNum)
		start = start + workerPtsNum
	}

	errs := []error{}

	// ERROR LISTENER
	errWg := &sync.WaitGroup{}
	errWg.Add(1)
	go func() {
		defer errWg.Done()
		for {
			err, ok := <-errchan
			if !ok {
				// channel was closed by producer, quit infinite loop
				return
			}
			errs = append(errs, err)
		}
	}()

	wg.Wait()

	// retrieve errors
	close(errchan)
	errWg.Wait()
	if len(errs) != 0 {
		return errs[0]
	}

	for i, startPt := range startPts {
		if t.pts != nil {
			endPts[i].Next = t.pts
		}
		t.pts = startPt
	}
	baselineGeomPt.Next = t.pts
	t.pts = baselineGeomPt
	t.bounds = geom.NewBoundingBox(minX-baselinePt.X, maxX-baselinePt.X, minY-baselinePt.Y, maxY-baselinePt.Y, minZ-baselinePt.Z, maxZ-baselinePt.Z)
	t.cX = baselinePt.X
	t.cY = baselinePt.Y
	t.cZ = baselinePt.Z
	return nil
}

func (t *GridTreeNode) GetCenter(cConv coor.CoordinateConverter) (float64, float64, float64, error) {
	return t.cX, t.cY, t.cZ, nil
}

func (t *GridTreeNode) transformPoint(pt geom.Point64, conv coor.CoordinateConverter, eConv elev.ElevationConverter, sourceCRS string) (geom.Point64, error) {
	var err error
	z := pt.Z
	if eConv != nil {
		z, err = eConv.ConvertElevation(pt.X, pt.Y, pt.Z)
		if err != nil {
			return pt, err
		}
	}

	out, err := conv.ToWGS84Cartesian(sourceCRS, geom.Coord{X: pt.X, Y: pt.Y, Z: z})
	if err != nil {
		return pt, err
	}
	pt.X, pt.Y, pt.Z = out.X, out.Y, out.Z
	return pt, nil
}
