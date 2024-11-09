package elev

type Converter interface {
	ConvertElevation(x, y, z float64) (float64, error)
}
