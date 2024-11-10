package grid

import "testing"

func TestBBoxBuilderMergeWith(t *testing.T) {
	u := newBoundingBoxBuilder()
	v := newBoundingBoxBuilder()
	u.processPoint(1, 2, 3)
	u.processPoint(4, 5, 6)
	v.processPoint(2, -1, 2)
	v.processPoint(5, 4, 7)

	u.mergeWith(v)
	u.build()

	if actual := u.minX; actual != 1 {
		t.Errorf("expected minx %v, got %v", 1, actual)
	}
	if actual := u.minY; actual != -1 {
		t.Errorf("expected miny %v, got %v", -1, actual)
	}
	if actual := u.minZ; actual != 2 {
		t.Errorf("expected minz %v, got %v", 2, actual)
	}
	if actual := u.maxX; actual != 5 {
		t.Errorf("expected maxx %v, got %v", 5, actual)
	}
	if actual := u.maxY; actual != 5 {
		t.Errorf("expected maxy %v, got %v", 5, actual)
	}
	if actual := u.maxZ; actual != 7 {
		t.Errorf("expected maxz %v, got %v", 7, actual)
	}
	v.processPoint(-1, 10, 6)
	u.mergeWith(v)
	if actual := u.minX; actual != -1 {
		t.Errorf("expected minx %v, got %v", -1, actual)
	}
	if actual := u.minY; actual != -1 {
		t.Errorf("expected miny %v, got %v", -1, actual)
	}
	if actual := u.minZ; actual != 2 {
		t.Errorf("expected minz %v, got %v", 2, actual)
	}
	if actual := u.maxX; actual != 5 {
		t.Errorf("expected maxx %v, got %v", 5, actual)
	}
	if actual := u.maxY; actual != 10 {
		t.Errorf("expected maxy %v, got %v", 10, actual)
	}
	if actual := u.maxZ; actual != 7 {
		t.Errorf("expected maxz %v, got %v", 7, actual)
	}
}
