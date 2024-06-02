package writer

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/tree"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/utils"
	"github.com/mfbonfigli/gocesiumtiler/v2/version"
)

// GeometryEncoder encodes a tree.Node into a binary file, like a .pnts or .glb/.gltf files.
type GeometryEncoder interface {
	Write(n tree.Node, c coor.CoordinateConverter, folderPath string) error
	TilesetVersion() version.TilesetVersion
	Filename() string
}

type Consumer interface {
	Consume(workchan chan *WorkUnit, errchan chan error, waitGroup *sync.WaitGroup)
}

type StandardConsumer struct {
	encoder GeometryEncoder
	conv    coor.CoordinateConverter
}

func NewStandardConsumer(coordinateConverter coor.CoordinateConverter, optFn ...func(*StandardConsumer)) Consumer {
	c := &StandardConsumer{
		conv:    coordinateConverter,
		encoder: NewPntsEncoder(),
	}
	for _, fn := range optFn {
		fn(c)
	}
	return c
}

// WithGeometryEncoder sets the consumer geometry encoder to the given one
func WithGeometryEncoder(e GeometryEncoder) func(*StandardConsumer) {
	return func(c *StandardConsumer) {
		c.encoder = e
	}
}

// Continually consumes WorkUnits submitted to a work channel producing corresponding gometry .pnts/.glb files and tileset.json files
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

// Takes a workunit and writes the corresponding content.glb/.pnts and tileset.json files
func (c *StandardConsumer) doWork(workUnit *WorkUnit) error {
	parentFolder := workUnit.BasePath
	node := workUnit.Node

	// Create base folder if it does not exist
	err := utils.CreateDirectoryIfDoesNotExist(parentFolder)
	if err != nil {
		return err
	}
	// encodes and writes the geometries to the disk as a .pnts/.glb file
	err = c.encoder.Write(node, c.conv, parentFolder)
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
		Content:        Content{c.encoder.Filename()},
		BoundingVolume: BoundingVolume{reg.GetAsArray()},
		GeometricError: node.ComputeGeometricError(),
		Refine:         "ADD",
		Children:       children,
	}, nil
}

func (c *StandardConsumer) generateTileset(node tree.Node, root Root) Tileset {
	tileset := Tileset{}
	tileset.Asset = Asset{Version: c.encoder.TilesetVersion()}
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
		filename = c.encoder.Filename()
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
