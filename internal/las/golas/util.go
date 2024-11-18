package golas

var scanAngleRankFormats map[uint8]any = map[uint8]any{
	0: nil,
	1: nil,
	2: nil,
	3: nil,
	4: nil,
	5: nil,
}

var scanAngleFormats map[uint8]any = map[uint8]any{
	6:  nil,
	7:  nil,
	8:  nil,
	9:  nil,
	10: nil,
}

var gpsTimeFormats map[uint8]any = map[uint8]any{
	1:  nil,
	3:  nil,
	4:  nil,
	5:  nil,
	6:  nil,
	7:  nil,
	8:  nil,
	9:  nil,
	10: nil,
}

var rgbTimeFormats map[uint8]any = map[uint8]any{
	2:  nil,
	3:  nil,
	5:  nil,
	7:  nil,
	8:  nil,
	10: nil,
}

var wavePacketsFormats map[uint8]any = map[uint8]any{
	4:  nil,
	5:  nil,
	9:  nil,
	10: nil,
}

var extraFlagByteFormats map[uint8]any = map[uint8]any{
	6:  nil,
	7:  nil,
	8:  nil,
	9:  nil,
	10: nil,
}

var nirFormats map[uint8]any = map[uint8]any{
	8:  nil,
	10: nil,
}

func formatHasGpsTime(f uint8) bool {
	_, ok := gpsTimeFormats[f]
	return ok
}

func formatHasRgbColors(f uint8) bool {
	_, ok := rgbTimeFormats[f]
	return ok
}

func formatHasWavePackets(f uint8) bool {
	_, ok := wavePacketsFormats[f]
	return ok
}

func formatHasScanAngleRank(f uint8) bool {
	_, ok := scanAngleRankFormats[f]
	return ok
}

func formatHasScanAngle(f uint8) bool {
	_, ok := scanAngleFormats[f]
	return ok
}

func formatHasExtraFlagByte(f uint8) bool {
	_, ok := extraFlagByteFormats[f]
	return ok
}

func formatHasNir(f uint8) bool {
	_, ok := nirFormats[f]
	return ok
}
