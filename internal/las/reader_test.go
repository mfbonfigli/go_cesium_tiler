package las

import (
	"fmt"
	"os"
	"testing"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
)

func TestCombinedReader(t *testing.T) {
	entries, err := os.ReadDir("./testdata")
	if err != nil {
		t.Fatal(err)
	}

	files := []string{}
	for _, e := range entries {
		filename := e.Name()
		files = append(files, fmt.Sprintf("./testdata/%s", filename))
	}

	r, err := NewCombinedFileLasReader(files, "EPSG:32633", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if actual := r.NumberOfPoints(); actual != 10*len(files) {
		t.Errorf("expected %d points got %d", 10*len(files), actual)
	}

	if actual := r.GetCRS(); actual != "EPSG:32633" {
		t.Errorf("expected epsg %d got epsg %s", 32633, actual)
	}

	for i := 0; i < r.NumberOfPoints(); i++ {
		_, err := r.GetNext()
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
	}
	_, err = r.GetNext()
	if err == nil {
		t.Errorf("expected error, got none")
	}
}

func TestReader(t *testing.T) {
	entries, err := os.ReadDir("./testdata")
	if err != nil {
		t.Fatal(err)
	}

	expectedNumPoints := 10

	expectedWOcolor := []geom.Point64{
		{X: 432488.4714159001, Y: 4.705678720925195e+06, Z: 2.550538045727727, R: 0, G: 0, B: 0, Intensity: 7, Classification: 3},
		{X: 432466.58372129063, Y: 4.705686414284739e+06, Z: 4.457767479175496, R: 0, G: 0, B: 0, Intensity: 7, Classification: 3},
		{X: 432469.98167019564, Y: 4.705681849831438e+06, Z: 4.655027394763156, R: 0, G: 0, B: 0, Intensity: 7, Classification: 3},
		{X: 432456.1696575739, Y: 4.705683863015449e+06, Z: 1.7922372985151949, R: 0, G: 0, B: 0, Intensity: 7, Classification: 3},
		{X: 432465.34641605255, Y: 4.705682466042181e+06, Z: 1.9543476118949155, R: 0, G: 0, B: 0, Intensity: 7, Classification: 3},
		{X: 432471.7795208326, Y: 4.705664509499522e+06, Z: 1.9162578617315837, R: 0, G: 0, B: 0, Intensity: 7, Classification: 3},
		{X: 432449.48655283387, Y: 4.705672817604579e+06, Z: 2.7424278273838762, R: 0, G: 0, B: 0, Intensity: 7, Classification: 3},
		{X: 432457.86106384156, Y: 4.705682649635591e+06, Z: 2.081788046497537, R: 0, G: 0, B: 0, Intensity: 7, Classification: 3},
		{X: 432466.3000286405, Y: 4.705680013894157e+06, Z: 1.9357872046388636, R: 0, G: 0, B: 0, Intensity: 7, Classification: 3},
		{X: 432455.7336230836, Y: 4.705657784890012e+06, Z: 6.120847715554042, R: 0, G: 0, B: 0, Intensity: 7, Classification: 3},
	}
	expectedWcolor := []geom.Point64{
		{X: 432488.4714159001, Y: 4.705678720925195e+06, Z: 2.550538045727727, R: 160, G: 166, B: 203, Intensity: 7, Classification: 3},
		{X: 432466.58372129063, Y: 4.705686414284739e+06, Z: 4.457767479175496, R: 186, G: 200, B: 237, Intensity: 7, Classification: 3},
		{X: 432469.98167019564, Y: 4.705681849831438e+06, Z: 4.655027394763156, R: 156, G: 167, B: 204, Intensity: 7, Classification: 3},
		{X: 432456.1696575739, Y: 4.705683863015449e+06, Z: 1.7922372985151949, R: 107, G: 114, B: 156, Intensity: 7, Classification: 3},
		{X: 432465.34641605255, Y: 4.705682466042181e+06, Z: 1.9543476118949155, R: 165, G: 176, B: 206, Intensity: 7, Classification: 3},
		{X: 432471.7795208326, Y: 4.705664509499522e+06, Z: 1.9162578617315837, R: 90, G: 136, B: 213, Intensity: 7, Classification: 3},
		{X: 432449.48655283387, Y: 4.705672817604579e+06, Z: 2.7424278273838762, R: 51, G: 66, B: 87, Intensity: 7, Classification: 3},
		{X: 432457.86106384156, Y: 4.705682649635591e+06, Z: 2.081788046497537, R: 80, G: 97, B: 123, Intensity: 7, Classification: 3},
		{X: 432466.3000286405, Y: 4.705680013894157e+06, Z: 1.9357872046388636, R: 160, G: 169, B: 202, Intensity: 7, Classification: 3},
		{X: 432455.7336230836, Y: 4.705657784890012e+06, Z: 6.120847715554042, R: 73, G: 78, B: 110, Intensity: 7, Classification: 3},
	}

	for _, e := range entries {
		filename := e.Name()
		r, err := NewFileLasReader(fmt.Sprintf("./testdata/%s", filename), "EPSG:32633", false)
		if err != nil {
			t.Fatalf("unexpected error opening %v", err)
		}
		if actual := r.NumberOfPoints(); actual != expectedNumPoints {
			t.Errorf("unexpected number of points for %s: expected %d got %d", filename, expectedNumPoints, actual)
		}
		for i := 0; i < 10; i++ {
			actual, err := r.GetNext()
			if err != nil {
				t.Errorf("unexpected error returned: %v", err)
			}
			if r.f.Header.PointFormatID == 1 || r.f.Header.PointFormatID == 4 {
				if actual != expectedWOcolor[i] {
					t.Errorf("for file %s, expected point %v got %v", filename, expectedWOcolor[i], actual)
				}
			} else {
				if actual != expectedWcolor[i] {
					t.Errorf("for file %s, expected point %v got %v", filename, expectedWcolor[i], actual)
				}
			}
		}
	}

}

func TestString(t *testing.T) {
	r, _ := NewFileLasReader("./testdata/las-12-pf1.las", "EPSG:123", false)
	expected := `File Signature: LASF
File Source ID: 0
Global Encoding: 
GpsTime: GpsWeekTime
WaveformDataInternal: false
WaveformDataExternal: false
ReturnDataSynthetic: false
CoordinateReferenceSystemMethod: GeoTiff
Project ID (GUID): 0-0-0-00-000000
System ID: 
Generating Software: LASzip DLL 3.4 r3 (191111)
Las Version: 1.2
File Creation Day/Year: 119/2024
Header Size: 227
Offset to points: 227
Number of VLRs: 0
Point Format: 1
Point Record Length: 28
Number of points: 10
Number of points by Return: [0, 0, 0, 0, 0]
X Scale Factor: 0.000001
Y Scale Factor: 0.000002
Z Scale Factor: 0.000000
X Offset: 431746.265869
Y Offset: 4704754.958740
Z Offset: -5.751223
Max X: 432488.471416
Min X: 432449.486553
Max Y: 4705686.414285
Min Y: 4705657.784890
Max Z: 6.120848
Min Z: 1.792237
Waveform Data Start: 0
`
	if actual := r.f.Header.String(); actual != expected {
		t.Errorf("unexpected header string, expected:\n%s\n\ngot:\n%s\n", expected, actual)
	}

}
