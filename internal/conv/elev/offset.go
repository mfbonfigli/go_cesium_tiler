package elev

type OffsetElevationConverter struct {
	Offset float64
}

func NewOffsetElevationConverter(offset float64) *OffsetElevationConverter {
	return &OffsetElevationConverter{
		Offset: offset,
	}
}

func (c *OffsetElevationConverter) ConvertElevation(x, y, z float64) (float64, error) {
	return z + c.Offset, nil
}
