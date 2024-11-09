package elev

type PipelineElevationConverter struct {
	Converters []Converter
}

func NewPipelineElevationCorrector(elevationConverters ...Converter) *PipelineElevationConverter {
	return &PipelineElevationConverter{
		Converters: elevationConverters,
	}
}

func (c *PipelineElevationConverter) ConvertElevation(x, y, z float64) (outZ float64, err error) {
	outZ = z
	for _, elevationConverter := range c.Converters {
		outZ, err = elevationConverter.ConvertElevation(x, y, outZ)
		if err != nil {
			return z, err
		}
	}

	return outZ, nil
}
