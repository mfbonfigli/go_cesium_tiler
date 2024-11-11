package proj

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mfbonfigli/gocesiumtiler/v2/tiler/model"
	"github.com/twpayne/go-proj/v10"
)

const epsg4978crs = "EPSG:4978"

type projCoordinateConverter struct {
	projections map[string]*proj.PJ
	searchPath  string
}

// Returns a
func NewProjCoordinateConverter() (*projCoordinateConverter, error) {
	// Initialization of EPSG Proj4 database
	conv := &projCoordinateConverter{
		projections: make(map[string]*proj.PJ),
	}

	// set the search path to the share folder in the same folder as the executable path
	execPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	conv.searchPath = filepath.Join(filepath.Dir(execPath), "share")
	return conv, nil
}

// Converts the given coordinate from the given source crs to the given target crs.
func (cc *projCoordinateConverter) Transform(sourceCRS string, targetCRS string, coord model.Vector) (model.Vector, error) {
	if sourceCRS == targetCRS {
		return coord, nil
	}
	pj, err := cc.getProjection(sourceCRS, targetCRS)
	if err != nil {
		return model.Vector{}, err
	}
	c := proj.NewCoord(coord.X, coord.Y, coord.Z, 0)
	out, err := pj.Forward(c)
	if err != nil {
		return coord, fmt.Errorf("error while transforming coordinates: %w", err)
	}
	return model.Vector{X: out.X(), Y: out.Y(), Z: out.Z()}, nil
}

// Converts the input coordinate from the given CRS to EPSG:4978 srid
func (cc *projCoordinateConverter) ToWGS84Cartesian(sourceCRS string, coord model.Vector) (model.Vector, error) {
	if sourceCRS == epsg4978crs {
		return coord, nil
	}

	return cc.Transform(sourceCRS, epsg4978crs, coord)
}

// Releases all projection objects from memory
func (cc *projCoordinateConverter) Cleanup() {
	for _, pj := range cc.projections {
		if pj != nil {
			pj.Destroy()
		}
	}
	// reset the projection cache
	cc.projections = make(map[string]*proj.PJ)
}

// Returns the projection object corresponding to the given crs representations, caching it internally to be reused
// This object is not designed for concurrent usage by multiple goroutines
func (cc *projCoordinateConverter) getProjection(source string, target string) (*proj.PJ, error) {
	uniqueProjectionCode := source + "#" + target
	if val, ok := cc.projections[uniqueProjectionCode]; ok {
		return val, nil
	}
	ctx := proj.NewContext()
	// set the search path if it points to a valid folder
	if _, err := os.Stat(cc.searchPath); err == nil {
		ctx.SetSearchPaths([]string{cc.searchPath})
	}
	pj, err := ctx.NewCRSToCRS(source, target, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize projection between %s and %s: %w", source, target, err)
	}
	pj, err = pj.NormalizeForVisualization()
	if err != nil {
		return nil, fmt.Errorf("unable to normalize the projection between %s and %s: %w", source, target, err)
	}

	cc.projections[uniqueProjectionCode] = pj

	return pj, nil
}
