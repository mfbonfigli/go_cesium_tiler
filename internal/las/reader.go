package las

import (
	"fmt"
	"os"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/las/golas"
	"github.com/mfbonfigli/gocesiumtiler/v2/tiler/model"
)

// LasReader wraps
type LasReader interface {
	// NumberOfPoints returns the number of points stored in the LAS file
	NumberOfPoints() int
	// GetNext returns the next point in the las file
	GetNext() (geom.Point64, error)
	// GetCRS returns a string defining the CRS. This is typically a string of the form EPSG:XYZ where XYZ is the EPSG code of the CRS.
	GetCRS() string
	// Close closes the reader
	Close()
}

// CombinedFileLasReader enables reading a a list of LAS files as if they were a single one
// the files MUST have the same properties (SRID, etc)
type CombinedFileLasReader struct {
	currentReader int
	currentCount  int
	readers       []LasReader
	numPts        int
	crs           string
}

// NewCombinedFileReader creates a new file reader for the files passed as input. If crs is the empty string, the
// reader will autodetect the CRS from the input files, however an error is returned if the CRS is not consistent across
// all of them or if it's not found in the files.
func NewCombinedFileLasReader(files []string, crs string, eightBitColor bool) (*CombinedFileLasReader, error) {
	r := &CombinedFileLasReader{}
	crsProvided := crs != ""
	for _, f := range files {
		fr, err := NewGoLasReader(f, crs, eightBitColor)
		if err != nil {
			return nil, err
		}
		r.numPts += fr.NumberOfPoints()
		r.readers = append(r.readers, fr)
		if !crsProvided {
			if crs != "" && crs != fr.GetCRS() {
				return nil, fmt.Errorf("no CRS was provided and inconsistent CRS were detected:\n%s\n\n and\n\n%s", crs, fr.GetCRS())
			}
			crs = fr.GetCRS()
		}
	}
	r.crs = crs
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

func (m *CombinedFileLasReader) Close() {
	for _, r := range m.readers {
		r.Close()
	}
}

// GoLasReader wraps a golas.Las object implementing the specific interface LasReader required by gocesiumtiler
type GoLasReader struct {
	file          *os.File
	f             *golas.Las
	eightBitColor bool
	crs           string
}

// NewGoLasReader returns a GoLasReader instance. If crs is empty the system will attempt to autodetect
// the CRS from the LAS metadata and return an error in case of issues.
func NewGoLasReader(fileName string, crs string, eightBitColor bool) (*GoLasReader, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	g, err := golas.NewLas(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	if crs == "" {
		crs = g.CRS()
		if crs == "" {
			f.Close()
			return nil, fmt.Errorf("no CRS provided and was not possible to determine CRS from LAS file %s", fileName)
		}
	}
	return &GoLasReader{
		file:          f,
		f:             g,
		eightBitColor: eightBitColor,
		crs:           crs,
	}, nil
}

func (f *GoLasReader) NumberOfPoints() int {
	return int(f.f.NumberOfPoints())
}

func (f *GoLasReader) GetCRS() string {
	return f.f.CRS()
}

func (f *GoLasReader) Close() {
	f.file.Close()
}

func (f *GoLasReader) GetNext() (geom.Point64, error) {
	pt, err := f.f.Next()
	if err != nil {
		return geom.Point64{}, err
	}
	var corr uint16 = 256
	if f.eightBitColor {
		corr = 1
	}
	return geom.Point64{
		Vector: model.Vector{
			X: pt.X,
			Y: pt.Y,
			Z: pt.Z,
		},
		R:              uint8(pt.Red / corr),
		G:              uint8(pt.Green / corr),
		B:              uint8(pt.Blue / corr),
		Intensity:      uint8(pt.Intensity),
		Classification: pt.Classification,
	}, nil
}
