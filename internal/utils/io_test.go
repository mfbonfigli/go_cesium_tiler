package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFindLasFilesInFolder(t *testing.T) {
	tmp, err := os.MkdirTemp(os.TempDir(), "tst")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmp)
	})

	TouchFile(filepath.Join(tmp, "test0.las"))
	TouchFile(filepath.Join(tmp, "test0.xyz"))
	TouchFile(filepath.Join(tmp, "test1.LAS"))
	TouchFile(filepath.Join(tmp, "test2.LAS"))

	files, err := FindLasFilesInFolder(tmp)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := []string{
		filepath.Join(tmp, "test0.las"),
		filepath.Join(tmp, "test1.LAS"),
		filepath.Join(tmp, "test2.LAS"),
	}
	if !reflect.DeepEqual(expected, files) {
		t.Errorf("expected %v got %v", expected, files)
	}
}
