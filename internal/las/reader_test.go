package las

import (
	"fmt"
	"os"
	"testing"
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
