package writer

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"sync"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/tree"
)

type Producer interface {
	Produce(work chan *WorkUnit, errchan chan error, wg *sync.WaitGroup, node tree.Node, ctx context.Context)
}

type StandardProducer struct {
	basePath string
}

func NewStandardProducer(basepath string, subfolder string) Producer {
	return &StandardProducer{
		basePath: path.Join(basepath, subfolder),
	}
}

// Parses a tree node and submits WorkUnits the the provided workchannel. Should be called only on the tree root node.
// Closes the channel when all work is submitted.
func (p *StandardProducer) Produce(work chan *WorkUnit, errchan chan error, wg *sync.WaitGroup, node tree.Node, ctx context.Context) {
	defer close(work)
	p.produce(errchan, p.basePath, node, work, wg, ctx)
	wg.Done()
}

// Parses a tree node and submits WorkUnits the the provided workchannel.
func (p *StandardProducer) produce(errchan chan error, basePath string, node tree.Node, work chan *WorkUnit, wg *sync.WaitGroup, ctx context.Context) {
	// if node contains points (it should always be the case), then submit work
	if err := ctx.Err(); err != nil {
		errchan <- fmt.Errorf("context closed: %v", err)
		return
	}
	if node.NumberOfPoints() > 0 {
		work <- &WorkUnit{
			Node:     node,
			BasePath: basePath,
		}
	} else {
		errchan <- fmt.Errorf("unexpected error: found tile without points: %v", node)
	}

	// iterate all non nil children and recursively submit all work units
	for i, child := range node.GetChildren() {
		if child != nil {
			p.produce(errchan, path.Join(basePath, strconv.Itoa(i)), child, work, wg, ctx)
		}
	}
}
