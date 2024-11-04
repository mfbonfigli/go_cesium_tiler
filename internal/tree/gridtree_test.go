package tree

import (
	"context"
	"testing"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/las"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/utils"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/utils/test"
)

func TestNewGridTreeWithMaxDepth(t *testing.T) {
	tree := NewGridTree(WithGridSize(11.5), WithMaxDepth(12), WithMinPointsPerChildren(1))
	if tree.IsBuilt() == true {
		t.Errorf("tree should not be built")
	}
	if tree.GetRootNode() != tree {
		t.Errorf("the root node of a tree is the tree itself but it was not")
	}
	if tree.IsRoot() != true {
		t.Errorf("the tree object should be a root node")
	}
	if tree.GetInternalCRS() != "EPSG:4978" {
		t.Errorf("A gridtree stores coordinates in EPSG:4978 SRID, but a different one was returned")
	}
	if tree.maxDepth != 12 {
		t.Errorf("expected maxDepth %d but got %d", 12, tree.maxDepth)
	}
	if tree.depth != 0 {
		t.Errorf("expected depth %d but got %d", 0, tree.depth)
	}
	if tree.gridSize != 11.5 {
		t.Errorf("expected gridSize %f but got %f", 11.5, tree.gridSize)
	}
}

func TestNewGridTreeWithDefaultDepth(t *testing.T) {
	tree := NewGridTree(WithGridSize(11.5), WithMinPointsPerChildren(1), WithMinPointsPerChildren(1))
	if tree.IsBuilt() == true {
		t.Errorf("tree should not be built")
	}
	if tree.GetRootNode() != tree {
		t.Errorf("the root node of a tree is the tree itself but it was not")
	}
	if tree.IsRoot() != true {
		t.Errorf("the tree object should be a root node")
	}
	if tree.GetInternalCRS() != "EPSG:4978" {
		t.Errorf("A gridtree stores coordinates in EPSG:4978 SRID, but a different one was returned")
	}
	if tree.maxDepth != 10 {
		t.Errorf("expected maxDepth %d but got %d", 12, tree.maxDepth)
	}
	if tree.depth != 0 {
		t.Errorf("expected depth %d but got %d", 0, tree.depth)
	}
	if tree.gridSize != 11.5 {
		t.Errorf("expected gridSize %f but got %f", 11.5, tree.gridSize)
	}
}

func TestGridTreeLoad(t *testing.T) {
	// the grid size is kept big intentionally so that we have at most 1 point per octant during the tests
	tree := NewGridTree(WithGridSize(1000000), WithMaxDepth(3), WithMinPointsPerChildren(1), WithLoadWorkersNumber(1))
	reader := &las.MockLasReader{
		CRS: "EPSG:32633",
		Pts: []geom.Point64{
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
		},
	}

	baseline := geom.Point64{X: 4.566154228810897e+06, Y: 1.1535788743923444e+06, Z: 4.286759667199761e+06, R: 160, G: 166, B: 203, Intensity: 7, Classification: 3}
	expectedAbsolute := []geom.Point64{
		{X: 4.566154228810897e+06, Y: 1.1535788743923444e+06, Z: 4.286759667199761e+06, R: 160, G: 166, B: 203, Intensity: 7, Classification: 3},
		{X: 4.566156071317037e+06, Y: 1.1535566808619653e+06, Z: 4.286766473069598e+06, R: 186, G: 200, B: 237, Intensity: 7, Classification: 3},
		{X: 4.566158338063568e+06, Y: 1.1535608049065806e+06, Z: 4.286763264386577e+06, R: 156, G: 167, B: 204, Intensity: 7, Classification: 3},
		{X: 4.566158449200715e+06, Y: 1.153546562658203e+06, Z: 4.286762716263729e+06, R: 107, G: 114, B: 156, Intensity: 7, Classification: 3},
		{X: 4.566157170414539e+06, Y: 1.15355572145217e+06, Z: 4.286761861131848e+06, R: 165, G: 176, B: 206, Intensity: 7, Classification: 3},
		{X: 4.566167248860708e+06, Y: 1.1535650843414892e+06, Z: 4.28674863859353e+06, R: 90, G: 136, B: 213, Intensity: 7, Classification: 3},
		{X: 4.566168019124784e+06, Y: 1.153542195651589e+06, Z: 4.286755164544903e+06, R: 51, G: 66, B: 87, Intensity: 7, Classification: 3},
		{X: 4.56615902316643e+06, Y: 1.1535484648586307e+06, Z: 4.286762029119904e+06, R: 80, G: 97, B: 123, Intensity: 7, Classification: 3},
		{X: 4.566158518303036e+06, Y: 1.1535570703581185e+06, Z: 4.286760046974066e+06, R: 160, G: 169, B: 202, Intensity: 7, Classification: 3},
		{X: 4.566178677722106e+06, Y: 1.1535514839264813e+06, Z: 4.28674640514511e+06, R: 73, G: 78, B: 110, Intensity: 7, Classification: 3},
	}

	expected := make([]geom.Point32, len(expectedAbsolute))
	for i := range expectedAbsolute {
		expected[i] = expectedAbsolute[i].ToPointFromBaseline(baseline)
	}
	conv := test.GetTestCoordinateConverterFactory()
	tree.Load(reader, conv, nil, context.TODO())
	// verify the points are stored
	cur := tree.pts
	for i := 0; i < len(reader.Pts); i++ {
		if cur.Pt != expected[i] {
			t.Errorf("expected pt %v got %v", expected[i], cur.Pt)
		}
		cur = cur.Next
	}

	// build
	err := tree.Build()
	if err != nil {
		t.Fatalf("unexpected error during tree build: %v", err)
	}

	root := tree.GetRootNode()
	if actual := root.NumberOfPoints(); actual != 1 {
		t.Errorf("expected 1 points, got %d", actual)
	}
	if actual := root.TotalNumberOfPoints(); actual != 10 {
		t.Errorf("expected 10 points, got %d", actual)
	}
}

func TestGridTreeBuild(t *testing.T) {
	// the grid size is kept big intentionally so that we have at most 1 point per octant during the tests
	tree := NewGridTree(WithGridSize(1000000), WithMaxDepth(3), WithMinPointsPerChildren(1))
	reader := &las.MockLasReader{
		CRS: "EPSG:4978",
		Pts: []geom.Point64{
			{X: 0, Y: 0, Z: 0, R: 160, G: 166, B: 203, Intensity: 7, Classification: 3},
			{X: -1, Y: -1, Z: -1, R: 186, G: 200, B: 237, Intensity: 7, Classification: 3},
			{X: 1, Y: 1, Z: 1, R: 156, G: 167, B: 204, Intensity: 7, Classification: 3},
			{X: -1, Y: -1, Z: 1, R: 107, G: 114, B: 156, Intensity: 7, Classification: 3},
			{X: 1, Y: 1, Z: -1, R: 165, G: 176, B: 206, Intensity: 7, Classification: 3},
			{X: -1, Y: 1, Z: 1, R: 90, G: 136, B: 213, Intensity: 7, Classification: 3},
			{X: -1, Y: 1, Z: -1, R: 51, G: 66, B: 87, Intensity: 7, Classification: 3},
			{X: 1, Y: -1, Z: 1, R: 80, G: 97, B: 123, Intensity: 7, Classification: 3},
			{X: 1, Y: -1, Z: -1, R: 160, G: 169, B: 202, Intensity: 7, Classification: 3},
			{X: 0.5, Y: 0.5, Z: 0.5, R: 73, G: 78, B: 110, Intensity: 7, Classification: 3},
		},
	}

	baseline := geom.Point64{X: 0, Y: 0, Z: 0, R: 160, G: 166, B: 203, Intensity: 7, Classification: 3}
	expectedAbsolute := []geom.Point64{
		{X: 0, Y: 0, Z: 0, R: 160, G: 166, B: 203, Intensity: 7, Classification: 3}, // baseline
		{X: -1, Y: -1, Z: -1, R: 186, G: 200, B: 237, Intensity: 7, Classification: 3},
		{X: 1, Y: 1, Z: 1, R: 156, G: 167, B: 204, Intensity: 7, Classification: 3},
		{X: -1, Y: -1, Z: 1, R: 107, G: 114, B: 156, Intensity: 7, Classification: 3},
		{X: 1, Y: 1, Z: -1, R: 165, G: 176, B: 206, Intensity: 7, Classification: 3},
		{X: -1, Y: 1, Z: 1, R: 90, G: 136, B: 213, Intensity: 7, Classification: 3},
		{X: -1, Y: 1, Z: -1, R: 51, G: 66, B: 87, Intensity: 7, Classification: 3},
		{X: 1, Y: -1, Z: 1, R: 80, G: 97, B: 123, Intensity: 7, Classification: 3},
		{X: 1, Y: -1, Z: -1, R: 160, G: 169, B: 202, Intensity: 7, Classification: 3},
		{X: 0.5, Y: 0.5, Z: 0.5, R: 73, G: 78, B: 110, Intensity: 7, Classification: 3},
	}
	expected := make([]geom.Point32, len(expectedAbsolute))
	for i := range expectedAbsolute {
		expected[i] = expectedAbsolute[i].ToPointFromBaseline(baseline)
	}
	conv := test.GetTestCoordinateConverterFactory()

	tree.Load(reader, conv, nil, context.TODO())
	// verify the points are stored
	cur := tree.pts
	for i := 0; i < len(reader.Pts); i++ {
		if cur.Pt != expected[i] {
			t.Errorf("expected pt %v got %v", expected[i], cur.Pt)
		}
		cur = cur.Next
	}

	// build
	err := tree.Build()
	if err != nil {
		t.Fatalf("unexpected error during tree build: %v", err)
	}

	root := tree.GetRootNode()
	if actual := root.NumberOfPoints(); actual != 1 {
		t.Errorf("expected 1 points, got %d", actual)
	}
	if actual := root.TotalNumberOfPoints(); actual != 10 {
		t.Errorf("expected 10 points, got %d", actual)
	}
	pt, err := root.GetPoints(nil).Next()
	if err != nil {
		t.Fatalf("unexpected error during point retrieval: %v", err)
	}
	if pt != expected[0] {
		t.Errorf("unexpected point returned, expected %v, got %v", expected[0], pt)
	}

	childExpectedMap := []geom.Point32{
		expected[1],
		expected[8],
		expected[6],
		expected[4],
		expected[3],
		expected[7],
		expected[5],
		expected[9],
	}
	for i, c := range root.GetChildren() {
		if actual := c.NumberOfPoints(); actual != 1 {
			t.Errorf("expected 1 points, got %d", actual)
		}
		if i != 7 {
			if actual := c.TotalNumberOfPoints(); actual != 1 {
				t.Errorf("expected 1 point, got %d", actual)
			}
			if actual := c.IsLeaf(); actual != true {
				t.Errorf("expected point to be leaf, got %v", actual)
			}

		} else {
			if actual := c.TotalNumberOfPoints(); actual != 2 {
				t.Errorf("expected 2 points, got %d", actual)
			}
			if actual := c.IsLeaf(); actual != false {
				t.Errorf("expected point to NOT be leaf, got %v", actual)
			}
		}
		if actual := c.IsRoot(); actual != false {
			t.Errorf("expected point to NOT be root, got %v", actual)
		}
		pt, err := c.GetPoints(nil).Next()
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}
		if pt != childExpectedMap[i] {
			t.Errorf("unexpected point returned for children %d, expected %v, got %v", i, childExpectedMap[i], pt)
		}
		children := c.GetChildren()
		if i == 7 {
			for i := 0; i < 8; i++ {
				if i == 7 {
					if n := children[i].NumberOfPoints(); n != 1 {
						t.Errorf("expected 1 point but got %d", n)
					}
					pt, err := children[i].GetPoints(nil).Next()
					if err != nil {
						t.Fatalf("unexpected error %v", err)
					}
					if pt != expected[2] {
						t.Errorf("unexpected point returned for children %d, expected %v, got %v", i, expected[2], pt)
					}
				} else {
					if children[i] != nil {
						t.Errorf("expected no child, got one: %v", children[i])
					}
				}
			}
		} else {
			for i := 0; i < 8; i++ {
				if children[i] != nil {
					t.Errorf("expected no child, got one: %v", children[i])
				}
			}
		}
	}
}

func TestGetBoundingBoxRegion(t *testing.T) {
	// the grid size is kept big intentionally so that we have at most 1 point per octant during the tests
	tree := NewGridTree(WithGridSize(1000000), WithMaxDepth(3), WithMinPointsPerChildren(1))
	reader := &las.MockLasReader{
		CRS: "EPSG:32633",
		Pts: []geom.Point64{
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
		},
	}

	conv := test.GetTestCoordinateConverterFactory()
	coorConv, err := conv()
	if err != nil {
		t.Errorf("unexpected error while initializing the converter: %v", err)
	}
	err = tree.Load(reader, conv, nil, context.TODO())
	if err != nil {
		t.Fatalf("unexpected error during tree load: %v", err)
	}
	err = tree.Build()
	if err != nil {
		t.Fatalf("unexpected error during tree build: %v", err)
	}
	bbox, err := tree.GetBoundingBoxRegion(coorConv)
	if err != nil {
		t.Fatalf("unexpected error during tree build: %v", err)
	}
	expected := geom.BoundingBox{
		Xmin: 0.2474500490710998,
		Xmax: 0.24745887140813697,
		Ymin: 0.7417700889108481,
		Ymax: 0.7417758833896815,
		Zmin: -13.032904711551964,
		Zmax: 24.624959161505103,
		Xmid: 0.24745446023961837,
		Ymid: 0.7417729861502648,
		Zmid: 5.796027224976569,
	}
	if diff, err := utils.CompareWithTolerance(bbox.Xmin, expected.Xmin, 1e-8); err != nil {
		t.Errorf("Xmin diff above threshold: %f", diff)
	}
	if diff, err := utils.CompareWithTolerance(bbox.Xmax, expected.Xmax, 1e-8); err != nil {
		t.Errorf("Xmax diff above threshold: %f", diff)
	}
	if diff, err := utils.CompareWithTolerance(bbox.Xmid, expected.Xmid, 1e-8); err != nil {
		t.Errorf("Xmid diff above threshold: %f", diff)
	}
	if diff, err := utils.CompareWithTolerance(bbox.Ymin, expected.Ymin, 1e-8); err != nil {
		t.Errorf("Ymin diff above threshold: %f", diff)
	}
	if diff, err := utils.CompareWithTolerance(bbox.Ymax, expected.Ymax, 1e-8); err != nil {
		t.Errorf("Ymax diff above threshold: %f", diff)
	}
	if diff, err := utils.CompareWithTolerance(bbox.Ymid, expected.Ymid, 1e-8); err != nil {
		t.Errorf("Ymid diff above threshold: %f", diff)
	}
	if diff, err := utils.CompareWithTolerance(bbox.Zmin, expected.Zmin, 1e-8); err != nil {
		t.Errorf("Zmin diff above threshold: %f", diff)
	}
	if diff, err := utils.CompareWithTolerance(bbox.Zmax, expected.Zmax, 1e-8); err != nil {
		t.Errorf("Zmax diff above threshold: %f", diff)
	}
	if diff, err := utils.CompareWithTolerance(bbox.Zmid, expected.Zmid, 1e-8); err != nil {
		t.Errorf("Zmid diff above threshold: %f", diff)
	}
}
