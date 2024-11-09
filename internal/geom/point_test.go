package geom

import "testing"

func TestToLocal(t *testing.T) {
	p := &Point64{
		Vector3:        Vector3{X: 100, Y: 200, Z: 300},
		R:              1,
		G:              2,
		B:              3,
		Intensity:      4,
		Classification: 5,
	}
	expected := NewPoint32(-100, 300, 600, 1, 2, 3, 4, 5)

	tr := Transform{
		GlobalToLocal: Quaternion{
			{0, -1, 0, 100},
			{1, 0, 0, 200},
			{0, 0, 1, 300},
			{0, 0, 0, 1},
		},
	}
	pt := p.ToLocal(tr)
	if pt != expected {
		t.Errorf("unexpected point, expected %v got %v", expected, pt)
	}
}

func TestLinkedPointStream(t *testing.T) {
	pt1 := &LinkedPoint{
		Pt: NewPoint32(1, 2, 3, 4, 5, 6, 7, 8),
	}
	pt2 := &LinkedPoint{
		Pt: NewPoint32(9, 10, 11, 12, 13, 14, 15, 16),
	}
	pt3 := &LinkedPoint{
		Pt: NewPoint32(17, 18, 19, 20, 21, 22, 23, 24),
	}
	pt1.Next = pt2
	pt2.Next = pt3

	stream := NewLinkedPointStream(pt1, 3)

	if actual := stream.Len(); actual != 3 {
		t.Errorf("expected Len %d got %d", 3, actual)
	}

	if actual, err := stream.Next(); actual != pt1.Pt || err != nil {
		if err == nil {
			t.Errorf("expected point %v got %v", pt1.Pt, actual)
		} else {
			t.Errorf("unexpected error %v", err)

		}
	}

	if actual, err := stream.Next(); actual != pt2.Pt || err != nil {
		if err == nil {
			t.Errorf("expected point %v got %v", pt2.Pt, actual)
		} else {
			t.Errorf("unexpected error %v", err)

		}
	}

	if actual, err := stream.Next(); actual != pt3.Pt || err != nil {
		if err == nil {
			t.Errorf("expected point %v got %v", pt3.Pt, actual)
		} else {
			t.Errorf("unexpected error %v", err)

		}
	}

	if _, err := stream.Next(); err == nil {
		t.Errorf("expected error but got none error %v", err)
	}

	stream.Reset()

	if actual, err := stream.Next(); actual != pt1.Pt || err != nil {
		if err == nil {
			t.Errorf("expected point %v got %v", pt1.Pt, actual)
		} else {
			t.Errorf("unexpected error %v", err)

		}
	}
}
