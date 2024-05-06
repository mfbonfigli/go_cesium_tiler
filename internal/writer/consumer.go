package writer

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/tree"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/utils"
)

type Consumer interface {
	Consume(workchan chan *WorkUnit, errchan chan error, waitGroup *sync.WaitGroup)
}

type StandardConsumer struct {
	conv coor.CoordinateConverter
}

func NewStandardConsumer(coordinateConverter coor.CoordinateConverter) Consumer {
	return &StandardConsumer{
		conv: coordinateConverter,
	}
}

// Continually consumes WorkUnits submitted to a work channel producing corresponding content.pnts files and tileset.json files
// continues working until work channel is closed or if an error is raised. In this last case submits the error to an error
// channel before quitting
func (c *StandardConsumer) Consume(workchan chan *WorkUnit, errchan chan error, waitGroup *sync.WaitGroup) {
	// signal waitgroup finished work
	defer waitGroup.Done()
	for {
		// get work from channel
		work, ok := <-workchan
		if !ok {
			// channel was closed by producer, quit infinite loop
			break
		}

		// do work
		err := c.doWork(work)

		// if there were errors during work send in error channel and quit
		if err != nil {
			errchan <- err
			break
		}
	}

}

// Takes a workunit and writes the corresponding content.pnts and tileset.json files
func (c *StandardConsumer) doWork(workUnit *WorkUnit) error {
	// writes the content.pnts file
	err := c.writeBinaryPntsFile(*workUnit)
	if err != nil {
		return err
	}
	// as an edge case we could have a leaf root node. This needs a tileset.json even if it's leaf.
	if !workUnit.Node.IsLeaf() || workUnit.Node.IsRoot() {
		// if the node has children also writes the tileset.json file
		err := c.writeTilesetJsonFile(*workUnit)
		if err != nil {
			return err
		}
	}
	return nil
}

// Writes a content.pnts binary files from the given WorkUnit
func (c *StandardConsumer) writeBinaryPntsFile(workUnit WorkUnit) error {
	parentFolder := workUnit.BasePath
	node := workUnit.Node

	// Create base folder if it does not exist
	err := utils.CreateDirectoryIfDoesNotExist(parentFolder)
	if err != nil {
		return err
	}

	pts := node.GetPoints(c.conv)
	cX, cY, cZ, err := node.GetCenter(c.conv)
	if err != nil {
		return err
	}
	// Evaluating average X, Y, Z to express coords relative to tile center
	averageXYZ, err := c.computeAverageXYZFromPointStream(pts, cX, cY, cZ)
	if err != nil {
		return err
	}

	// Feature table
	featureTableBytes, featureTableLen := c.generateFeatureTable(averageXYZ[0], averageXYZ[1], averageXYZ[2], pts.Len())

	// Batch table
	batchTableBytes, batchTableLen := c.generateBatchTable(pts.Len())

	// Write binary content to file
	pntsFilePath := path.Join(parentFolder, "content.pnts")
	f, err := os.Create(pntsFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	err = c.writePntsHeader(pts.Len(), featureTableLen, batchTableLen, w)
	if err != nil {
		return err
	}
	err = c.writeTable(featureTableBytes, w)
	if err != nil {
		return err
	}

	err = c.writePointCoords(pts, averageXYZ, cX, cY, cZ, w)
	if err != nil {
		return err
	}

	err = c.writePointColors(pts, w)
	if err != nil {
		return err
	}

	err = c.writeTable(batchTableBytes, w)
	if err != nil {
		return err
	}

	err = c.writePointIntensities(pts, w)
	if err != nil {
		return err
	}

	err = c.writePointClassifications(pts, w)
	if err != nil {
		return err
	}

	err = w.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (c *StandardConsumer) generateFeatureTable(avgX float64, avgY float64, avgZ float64, numPoints int) ([]byte, int) {
	featureTableStr := c.generateFeatureTableJsonContent(avgX, avgY, avgZ, numPoints, 0)
	featureTableLen := len(featureTableStr)
	return []byte(featureTableStr), featureTableLen
}

func (c *StandardConsumer) generateBatchTable(numPoints int) ([]byte, int) {
	batchTableStr := c.generateBatchTableJsonContent(numPoints, 0)
	batchTableLen := len(batchTableStr)
	return []byte(batchTableStr), batchTableLen
}

func (c *StandardConsumer) writePntsHeader(numPoints int, featureTableLen int, batchTableLen int, w io.Writer) error {
	_, err := w.Write([]byte("pnts")) // magic
	if err != nil {
		return err
	}
	err = utils.WriteIntAs4ByteNumber(1, w) // version number
	if err != nil {
		return err
	}
	positionBytesLen := 4 * 3 * numPoints                                                 // 4 bytes per coordinate component (x,y,z) -> 12 bytes per point
	err = utils.WriteIntAs4ByteNumber(28+featureTableLen+positionBytesLen+numPoints*3, w) // numpoints*3 is colorbytes (1 byte per color component)
	if err != nil {
		return err
	}
	err = utils.WriteIntAs4ByteNumber(featureTableLen, w) // feature table length
	if err != nil {
		return err
	}
	err = utils.WriteIntAs4ByteNumber(positionBytesLen+numPoints*3, w) // feature table binary length (position len + colors len)
	if err != nil {
		return err
	}
	err = utils.WriteIntAs4ByteNumber(batchTableLen, w) // batch table length
	if err != nil {
		return err
	}
	err = utils.WriteIntAs4ByteNumber(2*numPoints, w) // intensity + classification
	if err != nil {
		return err
	}
	return nil
}

func (c *StandardConsumer) writeTable(tableBytes []byte, w io.Writer) error {
	_, err := w.Write(tableBytes)
	if err != nil {
		return err
	}
	return nil
}

func (c *StandardConsumer) writePointCoords(pts geom.Point32List, avgCoords []float64, cX, cY, cZ float64, w io.Writer) error {
	n := pts.Len()
	// write coords
	for i := 0; i < n; i++ {
		pt, err := pts.Next()
		if err != nil {
			return err
		}
		err = utils.WriteTruncateFloat64ToFloat32(float64(pt.X)-avgCoords[0]+cX, w)
		if err != nil {
			return err
		}
		err = utils.WriteTruncateFloat64ToFloat32(float64(pt.Y)-avgCoords[1]+cY, w)
		if err != nil {
			return err
		}
		err = utils.WriteTruncateFloat64ToFloat32(float64(pt.Z)-avgCoords[2]+cZ, w)
		if err != nil {
			return err
		}
	}
	pts.Reset()
	return nil
}

func (c *StandardConsumer) writePointColors(pts geom.Point32List, w io.Writer) error {
	n := pts.Len()
	// write colors
	for i := 0; i < n; i++ {
		pt, err := pts.Next()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte{pt.R, pt.G, pt.B})
		if err != nil {
			return err
		}
	}
	pts.Reset()
	return nil
}

func (c *StandardConsumer) writePointIntensities(pts geom.Point32List, w io.Writer) error {
	n := pts.Len()
	// write colors
	for i := 0; i < n; i++ {
		pt, err := pts.Next()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte{pt.Intensity})
		if err != nil {
			return err
		}
	}
	pts.Reset()
	return nil
}

func (c *StandardConsumer) writePointClassifications(pts geom.Point32List, w io.Writer) error {
	n := pts.Len()
	// write colors
	for i := 0; i < n; i++ {
		pt, err := pts.Next()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte{pt.Classification})
		if err != nil {
			return err
		}
	}
	pts.Reset()
	return nil
}

func (c *StandardConsumer) computeAverageXYZFromPointStream(pts geom.Point32List, cX, cY, cZ float64) ([]float64, error) {
	var avgX, avgY, avgZ float64
	n := pts.Len()
	for i := 0; i < n; i++ {
		pt, err := pts.Next()
		if err != nil {
			return nil, err
		}
		avgX = (avgX / float64(i+1) * float64(i)) + (float64(pt.X)+cX)/float64(i+1)
		avgY = (avgY / float64(i+1) * float64(i)) + (float64(pt.Y)+cY)/float64(i+1)
		avgZ = (avgZ / float64(i+1) * float64(i)) + (float64(pt.Z)+cZ)/float64(i+1)
	}
	pts.Reset()
	return []float64{avgX, avgY, avgZ}, nil
}

// Generates the json representation of the feature table
func (c *StandardConsumer) generateFeatureTableJsonContent(x, y, z float64, pointNo int, spaceNo int) string {
	s := fmt.Sprintf(`{"POINTS_LENGTH":%d,"RTC_CENTER":[%f%s,%f%s,%f%s],"POSITION":{"byteOffset":0},"RGB":{"byteOffset":%d}}`,
		pointNo,
		x, strings.Repeat("0", spaceNo), y, strings.Repeat("0", spaceNo), z, strings.Repeat("0", spaceNo),
		pointNo*12,
	)
	headerByteLength := len([]byte(s))
	paddingSize := headerByteLength % 4
	if paddingSize != 0 {
		return c.generateFeatureTableJsonContent(x, y, z, pointNo, 4-paddingSize)
	}
	return s
}

// Generates the json representation of the batch table
func (c *StandardConsumer) generateBatchTableJsonContent(pointNumber, spaceNumber int) string {
	s := fmt.Sprintf(`{"INTENSITY":{"byteOffset":0,"componentType":"UNSIGNED_BYTE","type":"SCALAR"},
	"CLASSIFICATION":{"byteOffset":%d,"componentType":"UNSIGNED_BYTE","type":"SCALAR"}}%s`, pointNumber, strings.Repeat(" ", spaceNumber))
	headerByteLength := len([]byte(s))
	paddingSize := headerByteLength % 4
	if paddingSize != 0 {
		return c.generateBatchTableJsonContent(pointNumber, 4-paddingSize)
	}
	return s
}

// Writes the tileset.json file for the given WorkUnit
func (c *StandardConsumer) writeTilesetJsonFile(workUnit WorkUnit) error {
	parentFolder := workUnit.BasePath
	node := workUnit.Node

	// Create base folder if it does not exist
	err := utils.CreateDirectoryIfDoesNotExist(parentFolder)
	if err != nil {
		return err
	}

	// tileset.json file
	file := path.Join(parentFolder, "tileset.json")
	jsonData, err := c.generateTilesetJson(node)
	if err != nil {
		return err
	}

	// Writes the tileset.json binary content to the given file
	err = os.WriteFile(file, jsonData, 0666)
	if err != nil {
		return err
	}

	return nil
}

// Generates the tileset.json content for the given tree node
func (c *StandardConsumer) generateTilesetJson(node tree.Node) ([]byte, error) {
	if !node.IsLeaf() || node.IsRoot() {
		root, err := c.generateTilesetRoot(node)
		if err != nil {
			return nil, err
		}

		tileset := c.generateTileset(node, root)

		// Outputting a formatted json file
		e, err := json.MarshalIndent(tileset, "", "\t")
		if err != nil {
			return nil, err
		}

		return e, nil
	}

	return nil, errors.New("this node is a non-root leaf, cannot create a tileset json for it")
}

func (c *StandardConsumer) generateTilesetRoot(node tree.Node) (Root, error) {
	reg, err := node.GetBoundingBoxRegion(c.conv)

	if err != nil {
		return Root{}, err
	}

	children, err := c.generateTilesetChildren(node)
	if err != nil {
		return Root{}, err
	}

	return Root{
		Content:        Content{"content.pnts"},
		BoundingVolume: BoundingVolume{reg.GetAsArray()},
		GeometricError: node.ComputeGeometricError(),
		Refine:         "ADD",
		Children:       children,
	}, nil
}

func (c *StandardConsumer) generateTileset(node tree.Node, root Root) Tileset {
	tileset := Tileset{}
	tileset.Asset = Asset{Version: "1.0"}
	tileset.GeometricError = node.ComputeGeometricError()
	tileset.Root = root

	return tileset
}

func (c *StandardConsumer) generateTilesetChildren(node tree.Node) ([]Child, error) {
	var children []Child
	for i, child := range node.GetChildren() {
		if c.nodeContainsPoints(child) {
			childJson, err := c.generateTilesetChild(child, i)
			if err != nil {
				return nil, err
			}
			children = append(children, childJson)
		}
	}
	return children, nil
}

func (c *StandardConsumer) nodeContainsPoints(node tree.Node) bool {
	return node != nil && node.TotalNumberOfPoints() > 0
}

func (c *StandardConsumer) generateTilesetChild(child tree.Node, childIndex int) (Child, error) {
	childJson := Child{}
	filename := "tileset.json"
	if child.IsLeaf() {
		filename = "content.pnts"
	}
	childJson.Content = Content{
		Url: strconv.Itoa(childIndex) + "/" + filename,
	}
	reg, err := child.GetBoundingBoxRegion(c.conv)
	if err != nil {
		return Child{}, err
	}
	childJson.BoundingVolume = BoundingVolume{
		Region: reg.GetAsArray(),
	}
	childJson.GeometricError = child.ComputeGeometricError()
	childJson.Refine = "ADD"
	return childJson, nil
}
