package writer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/tree"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/utils/test"
)

func TestConsume(t *testing.T) {
	cv, err := test.GetTestCoordinateConverter()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	c := NewStandardConsumer(cv)
	wc := make(chan *WorkUnit)
	ec := make(chan error)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go c.Consume(wc, ec, wg)

	baseline := geom.Point64{X: 4.566154228810897e+06, Y: 1.1535788743923444e+06, Z: 4.286759667199761e+06, R: 160, G: 166, B: 203, Intensity: 7, Classification: 3}
	absPts := []geom.Point64{
		{X: 4.566154228810897e+06, Y: 1.1535788743923444e+06, Z: 4.286759667199761e+06, R: 160, G: 166, B: 203, Intensity: 7, Classification: 3},
		{X: 4.566156071317037e+06, Y: 1.1535566808619653e+06, Z: 4.286766473069598e+06, R: 186, G: 200, B: 237, Intensity: 7, Classification: 3},
		{X: 4.566158338063568e+06, Y: 1.1535608049065806e+06, Z: 4.286763264386577e+06, R: 156, G: 167, B: 204, Intensity: 7, Classification: 3},
	}

	pts := make([]geom.Point32, len(absPts))
	for i := range absPts {
		pts[i] = absPts[i].ToPointFromBaseline(baseline)
	}

	pt1 := &geom.LinkedPoint{
		Pt: pts[0],
	}
	pt2 := &geom.LinkedPoint{
		Pt: pts[1],
	}
	pt3 := &geom.LinkedPoint{
		Pt: pts[2],
	}
	pt1.Next = pt2
	pt2.Next = pt3

	stream := geom.NewLinkedPointStream(pt1, 3)
	n := &tree.MockNode{
		TotalNumPts: 3,
		Pts:         stream,
		Region: geom.NewBoundingBox(
			14.17808166914851,
			14.178348934756807,
			42.500527651607015,
			42.50059501992474,
			2.5505380453541875,
			4.655027396045625,
		),
		Root:      true,
		Leaf:      true,
		GeomError: 20,
		CenterX:   baseline.X,
		CenterY:   baseline.Y,
		CenterZ:   baseline.Z,
	}

	tmp, err := os.MkdirTemp(os.TempDir(), "tst")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmp)
	})

	tmpPath := filepath.Join(tmp, "tst")

	wc <- &WorkUnit{
		Node:     n,
		BasePath: tmpPath,
	}
	close(wc)
	wg.Wait()
	sb, err := os.ReadFile(filepath.Join(tmpPath, "tileset.json"))
	if err != nil {
		t.Fatalf("unable to read tileset.json: %v", err)
	}
	expected := Tileset{
		Asset: Asset{
			Version: "1.0",
		},
		GeometricError: 20,
		Root: Root{
			Children: nil,
			Content: Content{
				Url: "content.pnts",
			},
			BoundingVolume: BoundingVolume{
				Region: []float64{
					14.17808166914851,
					42.500527651607015,
					14.178348934756807,
					42.50059501992474,
					2.5505380453541875,
					4.655027396045625,
				},
			},
			GeometricError: 20,
			Refine:         "ADD",
		},
	}

	actual := Tileset{}
	err = json.Unmarshal(sb, &actual)
	if err != nil {
		t.Fatalf("unable to decode tileset.json: %v", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected tileset.json, expected:\n*%v*\n\ngot:\n\n*%v*\n", expected, actual)
	}

	actualPnts, err := os.ReadFile(filepath.Join(tmpPath, "content.pnts"))
	if err != nil {
		t.Fatalf("unable to read content.pnts: %v", err)
	}
	expectedPnts, err := os.ReadFile("./testdata/content.pnts")
	if err != nil {
		t.Fatalf("unable to read tileset.json: %v", err)
	}
	if !reflect.DeepEqual(actualPnts, expectedPnts) {
		t.Errorf("expected pnts:\n%v\n\ngot:\n\n%v\n", expectedPnts, actualPnts)
	}
}

func TestConsumeGltf(t *testing.T) {
	cv, err := test.GetTestCoordinateConverter()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	c := NewStandardConsumer(cv, WithGeometryEncoder(NewGltfEncoder()))
	wc := make(chan *WorkUnit)
	ec := make(chan error)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go c.Consume(wc, ec, wg)

	baseline := geom.Point64{X: 4.566154228810897e+06, Y: 1.1535788743923444e+06, Z: 4.286759667199761e+06, R: 160, G: 166, B: 203, Intensity: 7, Classification: 3}
	absPts := []geom.Point64{
		{X: 4.566154228810897e+06, Y: 1.1535788743923444e+06, Z: 4.286759667199761e+06, R: 160, G: 166, B: 203, Intensity: 7, Classification: 3},
		{X: 4.566156071317037e+06, Y: 1.1535566808619653e+06, Z: 4.286766473069598e+06, R: 186, G: 200, B: 237, Intensity: 7, Classification: 3},
		{X: 4.566158338063568e+06, Y: 1.1535608049065806e+06, Z: 4.286763264386577e+06, R: 156, G: 167, B: 204, Intensity: 7, Classification: 3},
	}

	pts := make([]geom.Point32, len(absPts))
	for i := range absPts {
		pts[i] = absPts[i].ToPointFromBaseline(baseline)
	}

	pt1 := &geom.LinkedPoint{
		Pt: pts[0],
	}
	pt2 := &geom.LinkedPoint{
		Pt: pts[1],
	}
	pt3 := &geom.LinkedPoint{
		Pt: pts[2],
	}
	pt1.Next = pt2
	pt2.Next = pt3

	stream := geom.NewLinkedPointStream(pt1, 3)
	n := &tree.MockNode{
		TotalNumPts: 3,
		Pts:         stream,
		Region: geom.NewBoundingBox(
			14.17808166914851,
			14.178348934756807,
			42.500527651607015,
			42.50059501992474,
			2.5505380453541875,
			4.655027396045625,
		),
		Root:      true,
		Leaf:      true,
		GeomError: 20,
		CenterX:   baseline.X,
		CenterY:   baseline.Y,
		CenterZ:   baseline.Z,
	}

	tmp, err := os.MkdirTemp(os.TempDir(), "tst")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmp)
	})

	tmpPath := filepath.Join(tmp, "tst")

	wc <- &WorkUnit{
		Node:     n,
		BasePath: tmpPath,
	}
	close(wc)
	wg.Wait()
	sb, err := os.ReadFile(filepath.Join(tmpPath, "tileset.json"))
	if err != nil {
		t.Fatalf("unable to read tileset.json: %v", err)
	}
	expected := Tileset{
		Asset: Asset{
			Version: "1.1",
		},
		GeometricError: 20,
		Root: Root{
			Children: nil,
			Content: Content{
				Url: "content.glb",
			},
			BoundingVolume: BoundingVolume{
				Region: []float64{
					14.17808166914851,
					42.500527651607015,
					14.178348934756807,
					42.50059501992474,
					2.5505380453541875,
					4.655027396045625,
				},
			},
			GeometricError: 20,
			Refine:         "ADD",
		},
	}

	actual := Tileset{}
	err = json.Unmarshal(sb, &actual)
	if err != nil {
		t.Fatalf("unable to decode tileset.json: %v", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected tileset.json, expected:\n*%v*\n\ngot:\n\n*%v*\n", expected, actual)
	}

	actualPnts, err := os.ReadFile(filepath.Join(tmpPath, "content.glb"))
	if err != nil {
		t.Fatalf("unable to read content.pnts: %v", err)
	}
	expectedPnts, err := os.ReadFile("./testdata/content.glb")
	if err != nil {
		t.Fatalf("unable to read tileset.json: %v", err)
	}
	if !reflect.DeepEqual(actualPnts, expectedPnts) {
		t.Errorf("expected pnts:\n%v\n\ngot:\n\n%v\n", expectedPnts, actualPnts)
	}
}
