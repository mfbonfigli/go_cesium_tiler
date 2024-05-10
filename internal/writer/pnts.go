package writer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/tree"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/utils"
	"github.com/mfbonfigli/gocesiumtiler/v2/version"
)

// PntsEncoder writes a node data as Pnts file (3D Tiles 1.0 specs)
type PntsEncoder struct{}

func (e *PntsEncoder) TilesetVersion() version.TilesetVersion {
	return version.TilesetVersion_1_0
}

func (e *PntsEncoder) Filename() string {
	return "content.pnts"
}

func NewPntsEncoder() GeometryEncoder {
	return &PntsEncoder{}
}

func (e *PntsEncoder) Write(node tree.Node, conv coor.CoordinateConverter, folderPath string) error {
	pts := node.GetPoints(conv)
	cX, cY, cZ, err := node.GetCenter(conv)
	if err != nil {
		return err
	}
	// Evaluating average X, Y, Z to express coords relative to tile center
	averageXYZ, err := e.computeAverageXYZFromPointStream(pts, cX, cY, cZ)
	if err != nil {
		return err
	}

	// Feature table
	featureTableBytes, featureTableLen := e.generateFeatureTable(averageXYZ[0], averageXYZ[1], averageXYZ[2], pts.Len())

	// Batch table
	batchTableBytes, batchTableLen := e.generateBatchTable(pts.Len())

	// Write binary content to file
	pntsFilePath := path.Join(folderPath, e.Filename())
	f, err := os.Create(pntsFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	wr := bufio.NewWriter(f)

	err = e.writePntsHeader(pts.Len(), featureTableLen, batchTableLen, wr)
	if err != nil {
		return err
	}
	err = e.writeTable(featureTableBytes, wr)
	if err != nil {
		return err
	}

	err = e.writePointCoords(pts, averageXYZ, cX, cY, cZ, wr)
	if err != nil {
		return err
	}

	err = e.writePointColors(pts, wr)
	if err != nil {
		return err
	}

	err = e.writeTable(batchTableBytes, wr)
	if err != nil {
		return err
	}

	err = e.writePointIntensities(pts, wr)
	if err != nil {
		return err
	}

	err = e.writePointClassifications(pts, wr)
	if err != nil {
		return err
	}

	err = wr.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (e *PntsEncoder) generateFeatureTable(avgX float64, avgY float64, avgZ float64, numPoints int) ([]byte, int) {
	featureTableStr := e.generateFeatureTableJsonContent(avgX, avgY, avgZ, numPoints, 0)
	featureTableLen := len(featureTableStr)
	return []byte(featureTableStr), featureTableLen
}

func (e *PntsEncoder) generateBatchTable(numPoints int) ([]byte, int) {
	batchTableStr := e.generateBatchTableJsonContent(numPoints, 0)
	batchTableLen := len(batchTableStr)
	return []byte(batchTableStr), batchTableLen
}

func (e *PntsEncoder) writePntsHeader(numPoints int, featureTableLen int, batchTableLen int, wr io.Writer) error {
	_, err := wr.Write([]byte("pnts")) // magic
	if err != nil {
		return err
	}
	err = utils.WriteIntAs4ByteNumber(1, wr) // version number
	if err != nil {
		return err
	}
	positionBytesLen := 4 * 3 * numPoints                                                  // 4 bytes per coordinate component (x,y,z) -> 12 bytes per point
	err = utils.WriteIntAs4ByteNumber(28+featureTableLen+positionBytesLen+numPoints*3, wr) // numpoints*3 is colorbytes (1 byte per color component)
	if err != nil {
		return err
	}
	err = utils.WriteIntAs4ByteNumber(featureTableLen, wr) // feature table length
	if err != nil {
		return err
	}
	err = utils.WriteIntAs4ByteNumber(positionBytesLen+numPoints*3, wr) // feature table binary length (position len + colors len)
	if err != nil {
		return err
	}
	err = utils.WriteIntAs4ByteNumber(batchTableLen, wr) // batch table length
	if err != nil {
		return err
	}
	err = utils.WriteIntAs4ByteNumber(2*numPoints, wr) // intensity + classification
	if err != nil {
		return err
	}
	return nil
}

func (e *PntsEncoder) writeTable(tableBytes []byte, wr io.Writer) error {
	_, err := wr.Write(tableBytes)
	if err != nil {
		return err
	}
	return nil
}

func (e *PntsEncoder) writePointCoords(pts geom.Point32List, avgCoords []float64, cX, cY, cZ float64, wr io.Writer) error {
	n := pts.Len()
	// write coords
	for i := 0; i < n; i++ {
		pt, err := pts.Next()
		if err != nil {
			return err
		}
		err = utils.WriteTruncateFloat64ToFloat32(float64(pt.X)-avgCoords[0]+cX, wr)
		if err != nil {
			return err
		}
		err = utils.WriteTruncateFloat64ToFloat32(float64(pt.Y)-avgCoords[1]+cY, wr)
		if err != nil {
			return err
		}
		err = utils.WriteTruncateFloat64ToFloat32(float64(pt.Z)-avgCoords[2]+cZ, wr)
		if err != nil {
			return err
		}
	}
	pts.Reset()
	return nil
}

func (e *PntsEncoder) writePointColors(pts geom.Point32List, wr io.Writer) error {
	n := pts.Len()
	// write colors
	for i := 0; i < n; i++ {
		pt, err := pts.Next()
		if err != nil {
			return err
		}
		_, err = wr.Write([]byte{pt.R, pt.G, pt.B})
		if err != nil {
			return err
		}
	}
	pts.Reset()
	return nil
}

func (e *PntsEncoder) writePointIntensities(pts geom.Point32List, wr io.Writer) error {
	n := pts.Len()
	// write colors
	for i := 0; i < n; i++ {
		pt, err := pts.Next()
		if err != nil {
			return err
		}
		_, err = wr.Write([]byte{pt.Intensity})
		if err != nil {
			return err
		}
	}
	pts.Reset()
	return nil
}

func (e *PntsEncoder) writePointClassifications(pts geom.Point32List, wr io.Writer) error {
	n := pts.Len()
	// write colors
	for i := 0; i < n; i++ {
		pt, err := pts.Next()
		if err != nil {
			return err
		}
		_, err = wr.Write([]byte{pt.Classification})
		if err != nil {
			return err
		}
	}
	pts.Reset()
	return nil
}

func (e *PntsEncoder) computeAverageXYZFromPointStream(pts geom.Point32List, cX, cY, cZ float64) ([]float64, error) {
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
func (e *PntsEncoder) generateFeatureTableJsonContent(x, y, z float64, pointNo int, spaceNo int) string {
	s := fmt.Sprintf(`{"POINTS_LENGTH":%d,"RTC_CENTER":[%f%s,%f%s,%f%s],"POSITION":{"byteOffset":0},"RGB":{"byteOffset":%d}}`,
		pointNo,
		x, strings.Repeat("0", spaceNo), y, strings.Repeat("0", spaceNo), z, strings.Repeat("0", spaceNo),
		pointNo*12,
	)
	headerByteLength := len([]byte(s))
	paddingSize := headerByteLength % 4
	if paddingSize != 0 {
		return e.generateFeatureTableJsonContent(x, y, z, pointNo, 4-paddingSize)
	}
	return s
}

// Generates the json representation of the batch table
func (e *PntsEncoder) generateBatchTableJsonContent(pointNumber, spaceNumber int) string {
	s := fmt.Sprintf(`{"INTENSITY":{"byteOffset":0,"componentType":"UNSIGNED_BYTE","type":"SCALAR"},
	"CLASSIFICATION":{"byteOffset":%d,"componentType":"UNSIGNED_BYTE","type":"SCALAR"}}%s`, 1*pointNumber, strings.Repeat(" ", spaceNumber))
	headerByteLength := len([]byte(s))
	paddingSize := headerByteLength % 4
	if paddingSize != 0 {
		return e.generateBatchTableJsonContent(pointNumber, 4-paddingSize)
	}
	return s
}
