package geoid2ellipsoid

// this geoid to ellipsoid converter caches the offset of the first point passed to it and
// returns it for all subsequent calls. It is very efficient but unsuitable for very large spatial
// point clouds (several km2).
type CachedCalculator struct {
	cachedOffset *float64
	calc         Calculator
}

func NewCachedCalculator(calc Calculator) *CachedCalculator {
	return &CachedCalculator{
		calc: calc,
	}
}

func (s *CachedCalculator) GetEllipsoidToGeoidOffset(lon, lat float64, srid int) (float64, error) {
	if s.cachedOffset == nil {
		offset, err := s.getEllipsoidToGeoidOffset(lon, lat, srid)
		if err != nil {
			return 0, nil
		}
		s.cachedOffset = &offset
	}

	return *s.cachedOffset, nil
}

func (spc *CachedCalculator) getEllipsoidToGeoidOffset(lon, lat float64, srid int) (float64, error) {
	return spc.calc.GetEllipsoidToGeoidOffset(lon, lat, srid)
}
