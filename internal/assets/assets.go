package assets

import "embed"

//go:embed *
var assets embed.FS

func GetAssets() embed.FS {
	return assets
}
