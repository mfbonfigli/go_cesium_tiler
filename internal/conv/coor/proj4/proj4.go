package proj4

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/assets"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
	proj "github.com/xeonx/proj4"
)

const toRadians = math.Pi / 180
const toDeg = 180 / math.Pi

type proj4CoordinateConverter struct {
	epsgDatabase   map[int]*epsgProjection
	assetTmpFolder string
}

func NewProj4CoordinateConverter() (*proj4CoordinateConverter, error) {
	// first step is to unpack the proj4 definitions to a temp location from the assets
	tempDir, err := unpackAssets()
	if err != nil {
		return nil, fmt.Errorf("unable to unpack proj4 assets: %v", err)
	}

	// Set path for retrieving projection assets data
	proj.SetFinder([]string{tempDir})

	// Initialization of EPSG Proj4 database
	db, err := loadEPSGProjectionDatabase()
	if err != nil {
		return nil, err
	}
	return &proj4CoordinateConverter{
		epsgDatabase:   db,
		assetTmpFolder: tempDir,
	}, nil
}

// unpackAssets unpacks the embedded assets into a temporary directory and returns its location
func unpackAssets() (string, error) {
	entries, err := assets.GetAssets().ReadDir("share")
	if err != nil {
		return "", fmt.Errorf("unable to unpack assets: %v", err)
	}
	temp, err := os.MkdirTemp(os.TempDir(), "gocesiumtiler-proj4-assets-*")
	if err != nil {
		return "", fmt.Errorf("unable to create temp folder for proj4 assets: %v", err)
	}
	for _, entry := range entries {
		data, err := assets.GetAssets().ReadFile(path.Join("share", entry.Name()))
		if err != nil {
			return "", fmt.Errorf("unable to read asset: %v", err)
		}
		f, err := os.Create(path.Join(temp, entry.Name()))
		if err != nil {
			return "", fmt.Errorf("unable to create asset file: %v", err)
		}
		defer f.Close()
		_, err = f.Write(data)
		if err != nil {
			return "", fmt.Errorf("unable to write asset: %v", err)
		}
	}
	return temp, nil
}

func loadEPSGProjectionDatabase() (map[int]*epsgProjection, error) {
	file, err := assets.GetAssets().Open("epsg_projections.txt")
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var epsgDatabase = make(map[int]*epsgProjection)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		record := scanner.Text()
		code, projection, err := parseEPSGProjectionDatabaseRecord(record)
		if err != nil {
			return nil, err
		}
		epsgDatabase[code] = projection
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return epsgDatabase, nil
}

func parseEPSGProjectionDatabaseRecord(databaseRecord string) (int, *epsgProjection, error) {
	tokens := strings.Split(databaseRecord, "\t")
	code, err := strconv.Atoi(strings.Replace(tokens[0], "EPSG:", "", -1))
	if err != nil {
		return 0, nil, fmt.Errorf("error while parsing the epsg projection file: %v", err)
	}
	desc := tokens[1]
	proj4 := tokens[2]

	return code, &epsgProjection{
		EpsgCode:    code,
		Description: desc,
		Proj4:       proj4,
	}, nil
}

// Converts the given coordinate from the given source Srid to the given target srid.
func (cc *proj4CoordinateConverter) ToSrid(sourceSrid int, targetSrid int, coord geom.Coord) (geom.Coord, error) {
	if sourceSrid == targetSrid {
		return coord, nil
	}

	src, err := cc.initProjection(sourceSrid)
	if err != nil {
		return coord, err
	}

	dst, err := cc.initProjection(targetSrid)
	if err != nil {
		return coord, err
	}

	var converted, result = executeConversion(&coord, src, dst)

	return *converted, result
}

// Converts the input coordinate from the given srid to EPSG:4326 srid
func (cc *proj4CoordinateConverter) ToWGS84Cartesian(coord geom.Coord, sourceSrid int) (geom.Coord, error) {
	if sourceSrid == 4978 {
		return coord, nil
	}

	res, err := cc.ToSrid(sourceSrid, 4326, coord)
	if err != nil {
		return coord, err
	}
	res2, err := cc.ToSrid(4329, 4978, res)
	return res2, err
}

// Releases all projection objects from memory
func (cc *proj4CoordinateConverter) Cleanup() {
	for _, val := range cc.epsgDatabase {
		if val.Projection != nil {
			val.Projection.Close()
		}
	}
	os.Remove(cc.assetTmpFolder)
}

func executeConversion(coord *geom.Coord, sourceProj *proj.Proj, destinationProj *proj.Proj) (*geom.Coord, error) {
	var x, y, z = getCoordinateArraysForConversion(coord, sourceProj)

	var err = proj.TransformRaw(sourceProj, destinationProj, x, y, z)

	var converted = geom.Coord{
		X: getCoordinateFromRadiansToSridFormat(x[0], destinationProj),
		Y: getCoordinateFromRadiansToSridFormat(y[0], destinationProj),
		Z: extractZPointerIfPresent(z),
	}

	return &converted, err
}

// From a input Coordinate object and associated Proj object, return a set of arrays to be used for coordinate coversion
func getCoordinateArraysForConversion(coord *geom.Coord, srid *proj.Proj) ([]float64, []float64, []float64) {
	var x, y, z []float64

	x = []float64{*getCoordinateInRadiansFromSridFormat(coord.X, srid)}
	y = []float64{*getCoordinateInRadiansFromSridFormat(coord.Y, srid)}

	if !math.IsNaN(coord.Z) {
		z = []float64{coord.Z}
	}

	return x, y, z
}

// Returns the input coordinate expressed in the given srid converting it into radians if necessary
func getCoordinateInRadiansFromSridFormat(coord float64, srid *proj.Proj) *float64 {
	var radians = coord

	if srid.IsLatLong() {
		radians = coord * toRadians
	}

	return &radians
}

func extractZPointerIfPresent(zContainer []float64) float64 {
	if zContainer != nil {
		return zContainer[0]
	}

	return math.NaN()
}

// Returns the input coordinate expressed in the given srid converting it into radians if necessary
func getCoordinateFromRadiansToSridFormat(coord float64, srid *proj.Proj) float64 {
	var angle = coord

	if srid.IsLatLong() {
		angle = coord * toDeg
	}

	return angle
}

// Returns the projection corresponding to the given EPSG code, storing it in the relevant EpsgDatabase entry for caching
func (cc *proj4CoordinateConverter) initProjection(code int) (*proj.Proj, error) {
	val, ok := cc.epsgDatabase[code]
	if !ok {
		return &proj.Proj{}, errors.New("epsg code not found")
	} else if val.Projection == nil {
		projection, err := proj.InitPlus(val.Proj4)
		if err != nil {
			return &proj.Proj{}, errors.New("unable to init projection")
		}
		val.Projection = projection
	}
	return val.Projection, nil
}
