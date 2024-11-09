package writer

import (
	"context"
	"fmt"
	"testing"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/geom"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/tree"
	"github.com/mfbonfigli/gocesiumtiler/v2/version"
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
		ChildNodes: [8]tree.Node{
			nil,
			child,
		},
	}

	w, err := NewWriter("base",
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
	w.consumerFunc = func(v version.TilesetVersion) Consumer {
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
		ChildNodes: [8]tree.Node{
			nil,
			child,
		},
	}

	w, err := NewWriter("base",
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
	w.consumerFunc = func(v version.TilesetVersion) Consumer {
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
		ChildNodes: [8]tree.Node{
			nil,
			child,
		},
	}

	w, err := NewWriter("base",
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
	w.consumerFunc = func(v version.TilesetVersion) Consumer {
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

func TestWriterTilesetVersion(t *testing.T) {
	w, err := NewWriter("base",
		WithNumWorkers(1),
		WithBufferRatio(10),
		WithTilesetVersion(version.TilesetVersion_1_0),
	)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if w.version != version.TilesetVersion_1_0 {
		t.Errorf("unexpected tileset version")
	}
	c := w.consumerFunc(version.TilesetVersion_1_0)
	if _, success := (c.(*StandardConsumer).encoder).(*PntsEncoder); success != true {
		t.Errorf("unexpected geometry encoder for tileset version 1.0")
	}
	w, err = NewWriter("base",
		WithNumWorkers(1),
		WithBufferRatio(10),
		WithTilesetVersion(version.TilesetVersion_1_1),
	)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if w.version != version.TilesetVersion_1_1 {
		t.Errorf("unexpected tileset version")
	}
	c = w.consumerFunc(version.TilesetVersion_1_1)
	if _, success := (c.(*StandardConsumer).encoder).(*GltfEncoder); success != true {
		t.Errorf("unexpected geometry encoder for tileset version 1.1")
	}
}
