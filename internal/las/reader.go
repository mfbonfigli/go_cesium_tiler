package las

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
)

var recLengths = [11][4]int{
	{20, 18, 19, 17}, // Point format 0
	{28, 26, 27, 25}, // Point format 1
	{26, 24, 25, 23}, // Point format 2
	{34, 32, 33, 31}, // Point format 3
	{57, 55, 56, 54}, // Point format 4
	{63, 61, 62, 60}, // Point format 5
	{30, 28, 29, 27}, // Point format 6
	{36, 34, 35, 33}, // Point format 7
	{38, 36, 37, 35}, // Point format 8
	{59, 57, 58, 56}, // Point format 9
	{67, 65, 66, 64}, // Point format 10
}

var xyzOffets = [11][3]int{
	{0, 4, 8}, // Point format 0
	{0, 4, 8}, // Point format 1
	{0, 4, 8}, // Point format 2
	{0, 4, 8}, // Point format 3
	{0, 4, 8}, // Point format 4
	{0, 4, 8}, // Point format 5
	{0, 4, 8}, // Point format 6
	{0, 4, 8}, // Point format 7
	{0, 4, 8}, // Point format 8
	{0, 4, 8}, // Point format 9
	{0, 4, 8}, // Point format 10
}

var rgbOffets = [11][]int{
	nil,          // Point format 0
	nil,          // Point format 1
	{20, 22, 24}, // Point format 2
	{28, 30, 32}, // Point format 3
	nil,          // Point format 4
	{28, 30, 32}, // Point format 5
	nil,          // Point format 6
	{30, 32, 34}, // Point format 7
	{30, 32, 34}, // Point format 8
	nil,          // Point format 9
	{30, 32, 34}, // Point format 10
}

// intensity offset is always 12

var classificationOffets = [11]int{
	15, // Point format 0
	15, // Point format 1
	15, // Point format 2
	15, // Point format 3
	15, // Point format 4
	15, // Point format 5
	16, // Point format 6
	16, // Point format 7
	16, // Point format 8
	16, // Point format 9
	16, // Point format 10
}

type LasReader interface {
	// NumberOfPoints returns the number of points stored in the LAS file
	NumberOfPoints() int
	// GetNext returns the next point in the las file
	GetNext() (geom.Point64, error)
	// GetCRS returns a string defining the CRS. This is typically a string of the form EPSG:XYZ where XYZ is the EPSG code of the CRS.
	GetCRS() string
}

// CombinedFileLasReader enables reading a a list of LAS files as if they were a single one
// the files MUST have the same properties (SRID, etc)
type CombinedFileLasReader struct {
	currentReader int
	currentCount  int
	readers       []*FileLasReader
	numPts        int
	crs           string
}

func NewCombinedFileLasReader(files []string, crs string, eightBitColor bool) (*CombinedFileLasReader, error) {
	r := &CombinedFileLasReader{
		crs: crs,
	}
	for _, f := range files {
		fr, err := NewFileLasReader(f, crs, eightBitColor)
		if err != nil {
			return nil, err
		}
		r.numPts += fr.NumberOfPoints()
		r.readers = append(r.readers, fr)
	}
	return r, nil
}

func (m *CombinedFileLasReader) NumberOfPoints() int {
	return m.numPts
}

func (m *CombinedFileLasReader) GetCRS() string {
	return m.crs
}

func (m *CombinedFileLasReader) GetNext() (geom.Point64, error) {
	if m.currentReader >= len(m.readers) {
		return geom.Point64{}, fmt.Errorf("no points to read")
	}
	r := m.readers[m.currentReader]
	if m.currentCount == r.NumberOfPoints() {
		m.currentReader++
		m.currentCount = 0
		if m.currentReader >= len(m.readers) {
			return geom.Point64{}, fmt.Errorf("no points to read")
		}
		r = m.readers[m.currentReader]
	}
	m.currentCount++
	return r.GetNext()
}

// FileLasReader enables reading a single LAS file
type FileLasReader struct {
	f             *lasFile
	eightBitColor bool
	crs           string
	r             io.Reader
	current       int
	sync.Mutex
}

func NewFileLasReader(fileName string, crs string, eightBitColor bool) (*FileLasReader, error) {
	vlrs := []VLR{}
	las := lasFile{fileName: fileName, Header: lasHeader{}, VlrData: vlrs}
	var err error
	if las.f, err = os.Open(las.fileName); err != nil {
		return nil, err
	}
	if err = las.readHeader(); err != nil {
		return nil, err
	}
	if err := las.readVLRs(); err != nil {
		return nil, err
	}
	return &FileLasReader{
		f:             &las,
		eightBitColor: eightBitColor,
		crs:           crs,
	}, nil
}

func (f *FileLasReader) NumberOfPoints() int {
	return f.f.Header.NumberPoints
}

func (f *FileLasReader) GetNext() (geom.Point64, error) {
	data := make([]byte, f.f.Header.PointRecordLength)
	out := geom.Point64{}
	f.Lock()
	if f.current == 0 {
		f.f.f.Seek(int64(f.f.Header.OffsetToPoints), 0)
		f.r = bufio.NewReaderSize(f.f.f, 64*1024)
	}
	f.current = f.current + 1
	if _, err := io.ReadFull(f.r, data); err != nil {
		f.Unlock()
		return geom.Point64{}, err
	}
	f.Unlock()
	header := f.f.Header
	xyzOffsetValues := xyzOffets[header.PointFormatID]
	xOffset := xyzOffsetValues[0]
	yOffset := xyzOffsetValues[1]
	zOffset := xyzOffsetValues[2]
	out.X = float64(int32(binary.LittleEndian.Uint32(data[xOffset:xOffset+4])))*header.XScaleFactor + header.XOffset
	out.Y = float64(int32(binary.LittleEndian.Uint32(data[yOffset:yOffset+4])))*header.YScaleFactor + header.YOffset
	out.Z = float64(int32(binary.LittleEndian.Uint32(data[zOffset:zOffset+4])))*header.ZScaleFactor + header.ZOffset

	rgbOffsetValues := rgbOffets[header.PointFormatID]
	if rgbOffsetValues != nil {
		rOffset := rgbOffsetValues[0]
		gOffset := rgbOffsetValues[1]
		bOffset := rgbOffsetValues[2]
		var conversionFactor = uint16(256)
		if f.eightBitColor {
			conversionFactor = uint16(1)
		}

		out.R = uint8(binary.LittleEndian.Uint16(data[rOffset:rOffset+2]) / conversionFactor)
		out.G = uint8(binary.LittleEndian.Uint16(data[gOffset:gOffset+2]) / conversionFactor)
		out.B = uint8(binary.LittleEndian.Uint16(data[bOffset:bOffset+2]) / conversionFactor)
	}
	intensityOffset := 12
	out.Intensity = uint8(binary.LittleEndian.Uint16(data[intensityOffset : intensityOffset+2]))
	classificationOffset := classificationOffets[header.PointFormatID]
	classification := data[classificationOffset]
	// the upper 3 high bits are used for metadata and not for the actual classification
	// so wipe them out
	out.Classification = uint8(classification & 0b00011111)

	return out, nil
}

func (f *FileLasReader) GetCRS() string {
	return f.crs
}
