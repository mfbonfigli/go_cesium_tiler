package writer

import (
	"context"
	"fmt"
	"testing"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/conv/coor"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/tree"
)

func TestWriter(t *testing.T) {
	pt1 := &geom.LinkedPoint{
		Pt: geom.NewPoint32(1, 2, 3, 4, 5, 6, 7, 8),
	}
	pt2 := &geom.LinkedPoint{
		Pt: geom.NewPoint32(9, 10, 11, 12, 13, 14, 15, 16),
	}
	pt3 := &geom.LinkedPoint{
		Pt: geom.NewPoint32(17, 18, 19, 20, 21, 22, 23, 24),
	}
	pt1.Next = pt2
	pt2.Next = pt3

	stream := geom.NewLinkedPointStream(pt1, 3)
	stream2 := geom.NewLinkedPointStream(pt2, 2)

	child := &tree.MockNode{
		TotalNumPts: 2,
		Pts:         stream2,
	}
	root := &tree.MockNode{
		TotalNumPts: 5,
		Pts:         stream,
		Children: [8]tree.Node{
			nil,
			child,
		},
	}

	w, err := NewWriter("base", nil,
		WithNumWorkers(1),
		WithBufferRatio(10),
	)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p := &MockProducer{}
	c := &MockConsumer{}
	w.producerFunc = func(basepath, folder string) Producer {
		return p
	}
	w.consumerFunc = func(cc coor.CoordinateConverter) Consumer {
		return c
	}
	err = w.Write(root, "base", context.TODO())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if p.Wc == nil {
		t.Errorf("empty work channel passed")
	} else {
		if c.Wc != p.Wc {
			t.Errorf("passed different work channel to consumer")
		}
	}
	if p.Ec == nil {
		t.Errorf("empty error channel passed")
	} else {
		if c.Ec != p.Ec {
			t.Errorf("passed different error channel to consumer")
		}
	}
}

func TestWriterWithProducerError(t *testing.T) {
	pt1 := &geom.LinkedPoint{
		Pt: geom.NewPoint32(1, 2, 3, 4, 5, 6, 7, 8),
	}
	pt2 := &geom.LinkedPoint{
		Pt: geom.NewPoint32(9, 10, 11, 12, 13, 14, 15, 16),
	}
	pt3 := &geom.LinkedPoint{
		Pt: geom.NewPoint32(17, 18, 19, 20, 21, 22, 23, 24),
	}
	pt1.Next = pt2
	pt2.Next = pt3

	stream := geom.NewLinkedPointStream(pt1, 3)
	stream2 := geom.NewLinkedPointStream(pt2, 2)

	child := &tree.MockNode{
		TotalNumPts: 2,
		Pts:         stream2,
	}
	root := &tree.MockNode{
		TotalNumPts: 5,
		Pts:         stream,
		Children: [8]tree.Node{
			nil,
			child,
		},
	}

	w, err := NewWriter("base", nil,
		WithNumWorkers(1),
		WithBufferRatio(10),
	)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p := &MockProducer{
		Err: fmt.Errorf("mock error"),
	}
	c := &MockConsumer{}
	w.producerFunc = func(basepath, folder string) Producer {
		return p
	}
	w.consumerFunc = func(cc coor.CoordinateConverter) Consumer {
		return c
	}
	err = w.Write(root, "base", context.TODO())
	if err == nil {
		t.Errorf("expected error but got none")
	}
	if p.Wc == nil {
		t.Errorf("empty work channel passed")
	} else {
		if c.Wc != p.Wc {
			t.Errorf("passed different work channel to consumer")
		}
	}
	if p.Ec == nil {
		t.Errorf("empty error channel passed")
	} else {
		if c.Ec != p.Ec {
			t.Errorf("passed different error channel to consumer")
		}
	}
}

func TestWriterWithConsumerError(t *testing.T) {
	pt1 := &geom.LinkedPoint{
		Pt: geom.NewPoint32(1, 2, 3, 4, 5, 6, 7, 8),
	}
	pt2 := &geom.LinkedPoint{
		Pt: geom.NewPoint32(9, 10, 11, 12, 13, 14, 15, 16),
	}
	pt3 := &geom.LinkedPoint{
		Pt: geom.NewPoint32(17, 18, 19, 20, 21, 22, 23, 24),
	}
	pt1.Next = pt2
	pt2.Next = pt3

	stream := geom.NewLinkedPointStream(pt1, 3)
	stream2 := geom.NewLinkedPointStream(pt2, 2)

	child := &tree.MockNode{
		TotalNumPts: 2,
		Pts:         stream2,
	}
	root := &tree.MockNode{
		TotalNumPts: 5,
		Pts:         stream,
		Children: [8]tree.Node{
			nil,
			child,
		},
	}

	w, err := NewWriter("base", nil,
		WithNumWorkers(1),
		WithBufferRatio(10),
	)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	p := &MockProducer{}
	c := &MockConsumer{
		Err: fmt.Errorf("mock error"),
	}
	w.producerFunc = func(basepath, folder string) Producer {
		return p
	}
	w.consumerFunc = func(cc coor.CoordinateConverter) Consumer {
		return c
	}
	err = w.Write(root, "base", context.TODO())
	if err == nil {
		t.Errorf("expected error but got none")
	}
	if p.Wc == nil {
		t.Errorf("empty work channel passed")
	} else {
		if c.Wc != p.Wc {
			t.Errorf("passed different work channel to consumer")
		}
	}
	if p.Ec == nil {
		t.Errorf("empty error channel passed")
	} else {
		if c.Ec != p.Ec {
			t.Errorf("passed different error channel to consumer")
		}
	}
}
