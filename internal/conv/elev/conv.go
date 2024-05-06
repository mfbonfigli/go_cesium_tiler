package elev

type ElevationConverter interface {
	ConvertElevation(x, y, z float64) (float64, error)
}
