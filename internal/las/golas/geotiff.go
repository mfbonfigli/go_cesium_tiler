package golas

type GeoTIFFMetadata struct {
	Keys map[int]*GeoTIFFKey
}

type GeoTIFFKey struct {
	KeyId    int
	Type     GeoTIFFTagType
	RawValue any
}

type GeoTIFFTagType int

const (
	GTTagTypeShort  GeoTIFFTagType = 0
	GTTagTypeDouble GeoTIFFTagType = 1
	GTTagTypeString GeoTIFFTagType = 2
)

func (g GeoTIFFKey) Name() string {
	return geotiffKeys[g.KeyId]
}

func (g GeoTIFFKey) AsShort() uint16 {
	return g.RawValue.(uint16)
}

func (g GeoTIFFKey) AsDouble() float64 {
	return g.RawValue.(float64)
}

func (g GeoTIFFKey) AsString() string {
	return g.RawValue.(string)
}

var geotiffKeys = map[int]string{
	// GeoTiff Configuration Keys
	1024: "GTModelTypeGeoKey",
	1025: "GTRasterTypeGeoKey",
	1026: "GTCitationGeoKey",

	// Geographic CS Parameter Keys
	2048: "GeographicTypeGeoKey",
	2049: "GeogCitationGeoKey",
	2050: "GeogGeodeticDatumGeoKey",
	2051: "GeogPrimeMeridianGeoKey",
	2052: "GeogLinearUnitsGeoKey",
	2053: "GeogLinearUnitSizeGeoKey",
	2054: "GeogAngularUnitsGeoKey",
	2055: "GeogAngularUnitSizeGeoKey",
	2056: "GeogEllipsoidGeoKey",
	2057: "GeogSemiMajorAxisGeoKey",
	2058: "GeogSemiMinorAxisGeoKey",
	2059: "GeogInvFlatteningGeoKey",
	2060: "GeogAzimuthUnitsGeoKey",
	2061: "GeogPrimeMeridianLongGeoKey",

	// Projected CS Parameter Keys
	3072: "ProjectedCSTypeGeoKey",
	3073: "PCSCitationGeoKey",
	3074: "ProjectionGeoKey",
	3075: "ProjCoordTransGeoKey",
	3076: "ProjLinearUnitsGeoKey",
	3077: "ProjLinearUnitSizeGeoKey",
	3078: "ProjStdParallel1GeoKey",
	3079: "ProjStdParallel2GeoKey",
	3080: "ProjNatOriginLongGeoKey",
	3081: "ProjNatOriginLatGeoKey",
	3082: "ProjFalseEastingGeoKey",
	3083: "ProjFalseNorthingGeoKey",
	3084: "ProjFalseOriginLongGeoKey",
	3085: "ProjFalseOriginLatGeoKey",
	3086: "ProjFalseOriginEastingGeoKey",
	3087: "ProjFalseOriginNorthingGeoKey",
	3088: "ProjCenterLongGeoKey",
	3089: "ProjCenterLatGeoKey",
	3090: "ProjCenterEastingGeoKey",
	3091: "ProjFalseOriginNorthingGeoKey",
	3092: "ProjScaleAtNatOriginGeoKey",
	3093: "ProjScaleAtCenterGeoKey",
	3094: "ProjAzimuthAngleGeoKey",
	3095: "ProjStraightVertPoleLongGeoKey",

	// Vertical CS Keys
	4096: "VerticalCSTypeGeoKey",
	4097: "VerticalCitationGeoKey",
	4098: "VerticalDatumGeoKey",
	4099: "VerticalUnitsGeoKey",
}
