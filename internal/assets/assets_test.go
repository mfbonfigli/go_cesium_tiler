package assets

import (
	"testing"
)

func TestGetAsset(t *testing.T) {
	r, err := GetAssets().Open("egm180.nor")
	if err != nil {
		t.Errorf("unexpected error returned: %v", err)
	}
	if r == nil {
		t.Errorf("expected file, nil returned")
	}
}
