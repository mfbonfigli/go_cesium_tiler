package writer

import (
	"encoding/json"
	"path"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/tree"
	"github.com/mfbonfigli/gocesiumtiler/v2/version"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

// Intensity and Classifications are stored using the EXT_structural_metadata
// GLTF extension. The following is the static schema that defines such properties and
// links them the the _INTENSITY and _CLASSIFICATION point
var extJson = `
{
	"schema": {
	  "id": "pts_schema",
	  "name": "pts_schema",
	  "description": "point cloud point attribute schema",
	  "version": "1.0.0",
	  "classes": {
		"point": {
		  "name": "point",
		  "description": "Properties of point cloud points",
		  "properties": {
			"INTENSITY": {
			  "description": "Laser intensity",
			  "type": "SCALAR",
			  "componentType": "UINT16",
			  "required": true
			},
			"CLASSIFICATION": {
			  "description": "Point classification",
			  "type": "SCALAR",
			  "componentType": "UINT16",
			  "required": true
			}
		  }
		}
	  }
	},
	"propertyAttributes": [
	  {
		"class": "point",
		"properties": {
		  "INTENSITY": {
			"attribute": "_INTENSITY"
		  },
		  "CLASSIFICATION": {
			"attribute": "_CLASSIFICATION"
		  }
		}
	  }
	]
  }
  `

// GltfEncoder writes a node data as Gltf/Glb binary file (3D Tiles 1.1 specs)
// Encodes intensity and classification using the EXT_structural_metadata GLTF extension
type GltfEncoder struct{}

func (e *GltfEncoder) TilesetVersion() version.TilesetVersion {
	return version.TilesetVersion_1_1
}

func (e *GltfEncoder) Filename() string {
	return "content.glb"
}

func NewGltfEncoder() *GltfEncoder {
	return &GltfEncoder{}
}

func (e *GltfEncoder) Write(node tree.Node, folderPath string) error {
	pts := node.Points()

	doc := gltf.NewDocument()
	doc.Asset = gltf.Asset{
		Generator: "gocesiumtiler",
		Version:   "2.0",
	}

	coords := make([][3]float32, pts.Len())
	colors := make([][3]uint8, pts.Len())

	// Note: for some reason uint8 results in an invalid GLTF being generated
	intensities := make([]uint16, pts.Len())
	classifications := make([]uint16, pts.Len())
	for i := 0; i < pts.Len(); i++ {
		pt, err := pts.Next()
		if err != nil {
			return err
		}
		coords[i][0] = pt.X
		coords[i][1] = pt.Y
		coords[i][2] = pt.Z
		colors[i][0] = pt.R
		colors[i][1] = pt.G
		colors[i][2] = pt.B
		intensities[i] = uint16(pt.Intensity)
		classifications[i] = uint16(pt.Classification)
	}

	attrs, err := modeler.WriteAttributesInterleaved(doc, modeler.Attributes{
		Position: coords,
		Color:    colors,
		CustomAttributes: []modeler.CustomAttribute{
			{Name: "_INTENSITY", Data: intensities},
			{Name: "_CLASSIFICATION", Data: classifications},
		},
	})
	if err != nil {
		return err
	}

	// When both featureId.attribute and featureId.texture are undefined, then the feature ID value
	// for each vertex is given implicitly, via the index of the vertex.
	// In this case, the featureCount must match the number of vertices of the mesh primitive.
	doc.Meshes = []*gltf.Mesh{{
		Name: "PointCloud",
		Primitives: []*gltf.Primitive{{
			Mode:       gltf.PrimitivePoints,
			Attributes: attrs,
			Extensions: gltf.Extensions{
				"EXT_structural_metadata": json.RawMessage(`{"propertyAttributes": [0]}`),
			},
		}},
	}}
	// gltf is Y up, however Cesium is Z up. This means that a rotation transform needs to be applied.
	doc.Nodes = []*gltf.Node{
		{
			Name:   "PointCloud",
			Mesh:   gltf.Index(0),
			Matrix: [16]float64{1, 0, 0, 0, 0, 0, -1, 0, 0, 1, 0, 0, 0, 0, 0, 1},
		},
	}
	doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, 0)
	doc.Extensions = gltf.Extensions{
		"EXT_structural_metadata": json.RawMessage(extJson),
	}
	doc.ExtensionsUsed = []string{
		"EXT_structural_metadata",
	}

	pntsFilePath := path.Join(folderPath, e.Filename())
	return gltf.SaveBinary(doc, pntsFilePath)
}
