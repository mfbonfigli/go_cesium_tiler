package tiler

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/las"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/tree"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/tree/grid"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/utils"
	"github.com/mfbonfigli/gocesiumtiler/v2/internal/writer"
)

func TestTilerDefaults(t *testing.T) {
	tiler, err := NewGoCesiumTiler()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tr := tiler.treeProvider(NewDefaultTilerOptions())
	switch tr.(type) {
	case *grid.Node:
	default:
		t.Errorf("unexpected tree type returned")
	}
	// this returns an error due to a non-esitant path
	// but we ignore it on purpose for the sake of this test
	l, _ := tiler.lasReaderProvider([]string{""}, "EPSG:123", true)
	switch l.(type) {
	case *las.CombinedFileLasReader:
	default:
		t.Errorf("unexpected las reader type returned")
	}
	// this returns an error due to a non-esitant path
	// but we ignore it on purpose for the sake of this test
	w, err := tiler.writerProvider("", NewDefaultTilerOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	switch w.(type) {
	case *writer.StandardWriter:
	default:
		t.Errorf("unexpected writer type returned")
	}
}

func TestTilerProcessFile(t *testing.T) {
	tiler, err := NewGoCesiumTiler()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w := &writer.MockWriter{}
	tr := &tree.MockNode{}
	l := &las.MockLasReader{}
	opts := NewDefaultTilerOptions()
	c := context.TODO()
	tiler.writerProvider = func(folder string, opts *TilerOptions) (writer.Writer, error) {
		return w, nil
	}
	tiler.treeProvider = func(opts *TilerOptions) tree.Tree {
		return tr
	}
	tiler.lasReaderProvider = func(inputLasFiles []string, sourceCRS string, eightbit bool) (las.LasReader, error) {
		return l, nil
	}

	tiler.ProcessFiles([]string{"abc.las"}, "out", "EPSG:123", opts, c)
	if !tr.LoadCalled {
		t.Errorf("Load was not called on the tree")
	}
	if actual := tr.Las; actual != l {
		t.Errorf("expected las reader %v got %v", l, actual)
	}
	if actual := tr.ConvFactory; actual == nil {
		t.Errorf("expected non-nil coordinate converter factory")
	}
	if actual := tr.Mut; actual == nil {
		t.Errorf("expected non-nil mutator")
	}
	if actual := tr.Ctx; actual != c {
		t.Errorf("expected different context")
	}
	if !tr.BuildCalled {
		t.Errorf("Build was not called on the tree")
	}
	if !w.WriteCalled {
		t.Errorf("Write was not called on the writer")
	}
	if actual := w.Tr; actual != tr {
		t.Errorf("expected tree %v got %v", tr, actual)
	}
	if actual := w.FolderName; actual != "" {
		t.Errorf("expected folder name '%v' got %v", "", actual)
	}
	if actual := w.Ctx; actual != c {
		t.Errorf("expected different context")
	}
}

func TestTilerProcessFolder(t *testing.T) {
	tiler, err := NewGoCesiumTiler()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w := &writer.MockWriter{}
	tr := &tree.MockNode{}
	l := &las.MockLasReader{}
	opts := NewDefaultTilerOptions()
	c := context.TODO()
	tiler.writerProvider = func(folder string, opts *TilerOptions) (writer.Writer, error) {
		return w, nil
	}
	tiler.treeProvider = func(opts *TilerOptions) tree.Tree {
		return tr
	}
	files := []string{}
	tiler.lasReaderProvider = func(inputLasFiles []string, sourceCRS string, eightbit bool) (las.LasReader, error) {
		files = append(files, inputLasFiles...)
		return l, nil
	}

	tmp, err := os.MkdirTemp(os.TempDir(), "tst")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmp)
	})
	utils.TouchFile(filepath.Join(tmp, "abc.las"))
	utils.TouchFile(filepath.Join(tmp, "def.xyz"))
	utils.TouchFile(filepath.Join(tmp, "ghi.las"))
	tiler.ProcessFolder(tmp, "out", "EPSG:123", opts, c)
	if !tr.LoadCalled {
		t.Errorf("Load was not called on the tree")
	}
	if actual := tr.Las; actual != l {
		t.Errorf("expected las reader %v got %v", l, actual)
	}
	if actual := tr.ConvFactory; actual == nil {
		t.Errorf("expected non-nil coordinate converter factory")
	}
	if actual := tr.Mut; actual == nil {
		t.Errorf("expected non-nil mutator")
	}
	if actual := tr.Ctx; actual != c {
		t.Errorf("expected different context")
	}
	if !tr.BuildCalled {
		t.Errorf("Build was not called on the tree")
	}
	if !w.WriteCalled {
		t.Errorf("Write was not called on the writer")
	}
	if actual := w.Tr; actual != tr {
		t.Errorf("expected tree %v got %v", tr, actual)
	}
	if actual := w.FolderName; actual != "" {
		t.Errorf("expected folder name '%v' got %v", "", actual)
	}
	if actual := w.Ctx; actual != c {
		t.Errorf("expected different context")
	}
	expected := []string{
		filepath.Join(tmp, "abc.las"),
		filepath.Join(tmp, "ghi.las"),
	}
	if !reflect.DeepEqual(files, expected) {
		t.Errorf("expected files processed %v, got %v", files, expected)
	}
}
