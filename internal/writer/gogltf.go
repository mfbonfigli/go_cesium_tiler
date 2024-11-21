package writer

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path"
	"strings"

	"github.com/mfbonfigli/gocesiumtiler/v2/internal/tree"
	"github.com/mfbonfigli/gocesiumtiler/v2/version"
)

type bufferTarget int

const (
	arrayBuffer        bufferTarget = 34962
	elementArrayBuffer bufferTarget = 34963
)

type accessorType string

const (
	accessorTypeScalar accessorType = "SCALAR"
	accessorTypeVec2   accessorType = "VEC2"
	accessorTypeVec3   accessorType = "VEC3"
	accessorTypeVec4   accessorType = "VEC4"
	accessorTypeMat2   accessorType = "MAT2"
	accessorTypeMat3   accessorType = "MAT3"
	accessorTypeMat4   accessorType = "MAT4"
)

type componentType int

const (
	componentTypeByte          componentType = 5120
	componentTypeUnsignedByte  componentType = 5121
	componentTypeShort         componentType = 5122
	componentTypeUnsignedShort componentType = 5123
	componentTypeUnsignedInt   componentType = 5125
	componentTypeFloat         componentType = 5126
)

type GltfPointCloud struct {
}

func NewGltfPointCloudEncoder() *GltfPointCloud {
	return &GltfPointCloud{}
}
func (g *GltfPointCloud) TilesetVersion() version.TilesetVersion {
	return version.TilesetVersion_1_1
}

func (g *GltfPointCloud) Filename() string {
	return "content.glb"
}

func (g *GltfPointCloud) Write(node tree.Node, folderPath string) error {
	filePath := path.Join(folderPath, g.Filename())
	f, err := os.Create(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()

	binaryData, err := binaryData(node)
	paddedBinaryData := padByteArrayTo4byte(binaryData.data)
	if err != nil {
		return err
	}
	jsonString, err := jsonNode(node, binaryData, len(paddedBinaryData))
	if err != nil {
		return err
	}
	paddedJsonData := []byte(padStringTo4byte(jsonString))
	fmt.Println(len(jsonString), len(paddedJsonData))
	jsonLength := len(paddedJsonData)
	binaryLength := len(paddedBinaryData)
	bufLen := EncodeVertexBufferBound(node.NumberOfPoints(), 12)
	d := make([]byte, bufLen)
	n, err := EncodeVertexBuffer(d, paddedBinaryData[0:node.NumberOfPoints()*12], node.NumberOfPoints(), 12)
	fmt.Println(err)
	fmt.Println(n, node.NumberOfPoints()*12)
	//fmt.Println(d)
	totalLength := 12 + 4 + 4 + jsonLength + 4 + 4 + binaryLength

	// 12 BYTE HEADER
	// write magic
	if _, err := w.WriteString("glTF"); err != nil {
		return err
	}
	// write version
	if err := binary.Write(w, binary.LittleEndian, uint32(2)); err != nil {
		return err
	}
	// TODO DETERMINE LENGTH
	// write length
	if err := binary.Write(w, binary.LittleEndian, uint32(totalLength)); err != nil {
		return err
	}

	// CHUNK 0 JSON
	// write chunk 0 length TODO
	if err := binary.Write(w, binary.LittleEndian, uint32(jsonLength)); err != nil {
		return err
	}
	// write chunk type JSON
	if _, err := w.WriteString("JSON"); err != nil {
		return err
	}
	// write chunk JSON TODO
	if _, err := w.Write(paddedJsonData); err != nil {
		return err
	}

	// CHUNK 1 BINARY BUFFER
	// write chunk 0 length TODO
	if err := binary.Write(w, binary.LittleEndian, uint32(binaryLength)); err != nil {
		return err
	}
	// write chunk type BIN
	if _, err := w.WriteString("BIN\u0000"); err != nil {
		return err
	}
	// write chunk DATA
	if _, err := w.Write(paddedBinaryData); err != nil {
		return err
	}

	return nil
}

func jsonNode(n tree.Node, b *binaryOutput, bufLen int) (string, error) {
	j := glTFJson{
		Buffers: []buffer{
			{ByteLength: bufLen},
		},
		Extensions: map[string]map[string]any{
			"EXT_structural_metadata": {
				"schema": map[string]any{
					"id":          "pts_schema",
					"name":        "pts_schema",
					"description": "point cloud point attribute schema",
					"version":     "1.0.0",
					"classes": map[string]any{
						"point": map[string]any{
							"name":        "point",
							"description": "Properties of point cloud points",
							"properties": map[string]any{
								"INTENSITY": map[string]any{
									"description":   "Laser intensity",
									"type":          "SCALAR",
									"componentType": "UINT16",
									"required":      true,
								},
								"CLASSIFICATION": map[string]any{
									"description":   "Point classification",
									"type":          "SCALAR",
									"componentType": "UINT8",
									"required":      true,
								},
							},
						},
					},
				},
				"propertyAttributes": []any{
					map[string]any{
						"class": "point",
						"properties": map[string]any{
							"INTENSITY": map[string]any{
								"attribute": "_INTENSITY",
							},
							"CLASSIFICATION": map[string]any{
								"attribute": "_CLASSIFICATION",
							},
						},
					},
				},
			},
		},
		ExtensionsUsed: []string{
			"EXT_structural_metadata",
		},
		Accessors: []accessor{
			{
				BufferView:    0,
				ComponentType: componentTypeFloat,
				Count:         n.Points().Len(),
				Type:          accessorTypeVec3,
				Max:           b.max,
				Min:           b.min,
			},
			{
				BufferView:    1,
				ComponentType: componentTypeUnsignedByte,
				Count:         n.Points().Len(),
				Type:          accessorTypeVec3,
				Normalized:    true,
			},
			{
				BufferView:    2,
				ComponentType: componentTypeUnsignedShort,
				Count:         n.Points().Len(),
				Type:          accessorTypeScalar,
			},
			{
				BufferView:    3,
				ComponentType: componentTypeUnsignedByte,
				Count:         n.Points().Len(),
				Type:          accessorTypeScalar,
			},
		},
		Asset: asset{
			Generator: "gocesiumtiler",
			Version:   "2.0",
		},
		BufferViews: []bufferView{
			{
				Buffer:     0,
				ByteLength: b.coordinatesEnd - b.coordinatesStart,
				Target:     arrayBuffer,
			},
			{
				Buffer:     0,
				ByteLength: b.colorEnd - b.colorStart,
				ByteOffset: b.colorStart,
				ByteStride: 4,
				Target:     arrayBuffer,
			},
			{
				Buffer:     0,
				ByteLength: b.intensityEnd - b.intensityStart,
				ByteOffset: b.intensityStart,
				ByteStride: 4,
				Target:     arrayBuffer,
			},
			{
				Buffer:     0,
				ByteLength: b.classificationEnd - b.classificationStart,
				ByteOffset: b.classificationStart,
				ByteStride: 4,
				Target:     arrayBuffer,
			},
		},
		Meshes: []mesh{
			{
				Name: "PointCloud",
				Primitives: []primitive{
					{
						Extensions: map[string]map[string]any{
							"EXT_structural_metadata": {
								"propertyAttributes": []int{0},
							},
						},
						Attributes: map[string]int{
							"POSITION":        0,
							"COLOR_0":         1,
							"_INTENSITY":      2,
							"_CLASSIFICATION": 3,
						},
						Mode: 0,
					},
				},
			},
		},
		Nodes: []node{
			{
				Matrix: []float64{1, 0, 0, 0, 0, 0, -1, 0, 0, 1, 0, 0, 0, 0, 0, 1},
				Name:   "PointCloud",
				Mesh:   0,
			},
		},
		Scene: 0,
		Scenes: []scene{
			{
				Name:  "Root Scene",
				Nodes: []int{0},
			},
		},
	}
	jBytes, err := json.Marshal(j)
	if err != nil {
		return "", err
	}
	return string(jBytes), nil
}

type binaryOutput struct {
	coordinatesStart    int
	coordinatesEnd      int
	colorStart          int
	colorEnd            int
	intensityStart      int
	intensityEnd        int
	classificationStart int
	classificationEnd   int
	max                 []float32
	min                 []float32
	data                []byte
}

func binaryData(n tree.Node) (*binaryOutput, error) {
	// x,y,z are VEC3 floats, hence offset should be aligned to 4 bytes
	// color are VEC3 of bytes, hence offset should be aligned to 1 byte
	// intensity and classifications are unsigned shorts, hence offset should be aligned to 2 bytes
	pts := n.Points()
	b := &bytes.Buffer{}
	out := &binaryOutput{
		max: []float32{math.MinInt32, math.MinInt32, math.MinInt32},
		min: []float32{math.MaxInt32, math.MaxInt32, math.MaxInt32},
	}
	// Coordinates bufferView
	out.coordinatesStart = 0
	for i := 0; i < pts.Len(); i++ {
		p, err := pts.Next()
		if err != nil {
			return nil, err
		}
		// update bounds
		if p.X > out.max[0] {
			out.max[0] = p.X
		}
		if p.Y > out.max[1] {
			out.max[1] = p.Y
		}
		if p.Z > out.max[2] {
			out.max[2] = p.Z
		}
		if p.X < out.min[0] {
			out.min[0] = p.X
		}
		if p.Y < out.min[1] {
			out.min[1] = p.Y
		}
		if p.Z < out.min[2] {
			out.min[2] = p.Z
		}
		// X, Y, Z are 4x3 bytes in a VEC3 and offset is 0 hence aligned automatically
		if err := binary.Write(b, binary.LittleEndian, p.X); err != nil {
			return nil, err
		}
		if err := binary.Write(b, binary.LittleEndian, p.Y); err != nil {
			return nil, err
		}
		if err := binary.Write(b, binary.LittleEndian, p.Z); err != nil {
			return nil, err
		}
		out.coordinatesEnd += 12
	}

	// colors start is automatically aligned, no need to add padding

	pts.Reset()
	// Colors bufferView
	out.colorStart = out.coordinatesEnd
	out.colorEnd = out.colorStart
	for i := 0; i < pts.Len(); i++ {
		p, err := pts.Next()
		if err != nil {
			return nil, err
		}
		// R, G, B are 1x3 bytes in a VEC3: Offset: N*12
		if err := binary.Write(b, binary.LittleEndian, p.R); err != nil {
			return nil, err
		}
		if err := binary.Write(b, binary.LittleEndian, p.G); err != nil {
			return nil, err
		}
		if err := binary.Write(b, binary.LittleEndian, p.B); err != nil {
			return nil, err
		}
		if _, err := b.Write(make([]byte, 1)); err != nil {
			return nil, err
		}
		out.colorEnd += 4
	}

	pts.Reset()
	// Intensity view: Offset: N*12 + N*3 = N*15. This must be aligned to 4 bytes
	out.intensityStart = out.colorEnd
	out.intensityEnd = out.intensityStart
	for i := 0; i < pts.Len(); i++ {
		p, err := pts.Next()
		if err != nil {
			return nil, err
		}
		// Intensity, 2 bytes
		if err := binary.Write(b, binary.LittleEndian, uint16(p.Intensity)); err != nil {
			return nil, err
		}
		if _, err := b.Write(make([]byte, 2)); err != nil {
			return nil, err
		}
		out.intensityEnd += 4
	}

	pts.Reset()
	// Classification view: Offset: N*12+N*3+N*2+intPadding=N*17+intPadding, type is 1 byte so it's aligned
	out.classificationStart = out.intensityEnd
	out.classificationEnd = out.classificationStart
	for i := 0; i < pts.Len(); i++ {
		p, err := pts.Next()
		if err != nil {
			return nil, err
		}
		// Classification is 1 byte
		if err := binary.Write(b, binary.LittleEndian, p.Classification); err != nil {
			return nil, err
		}
		if _, err := b.Write(make([]byte, 3)); err != nil {
			return nil, err
		}
		out.classificationEnd += 4
	}
	out.data = b.Bytes()
	return out, nil
}

func padStringTo4byte(str string) string {
	n := len(str) % 4
	if n == 0 {
		return str
	}
	padding := 4 - n
	return fmt.Sprintf("%s%s", str, strings.Repeat(" ", padding))
}

func padByteArrayTo4byte(arr []byte) []byte {
	n := len(arr) % 4
	if n == 0 {
		return arr
	}
	padding := 4 - n
	return append(arr, make([]byte, padding)...)
}

type glTFJson struct {
	Buffers        []buffer                  `json:"buffers"`
	Extensions     map[string]map[string]any `json:"extensions"`
	ExtensionsUsed []string                  `json:"extensionsUsed"`
	Accessors      []accessor                `json:"accessors"`
	Asset          asset                     `json:"asset"`
	BufferViews    []bufferView              `json:"bufferViews"`
	Meshes         []mesh                    `json:"meshes"`
	Nodes          []node                    `json:"nodes"`
	Scene          int                       `json:"scene"`
	Scenes         []scene                   `json:"scenes"`
}

type buffer struct {
	ByteLength int `json:"byteLength"`
}

type accessor struct {
	BufferView    int           `json:"bufferView"`
	ByteOffset    int           `json:"byteOffset,omitempty"`
	ComponentType componentType `json:"componentType"`
	Count         int           `json:"count"`
	Type          accessorType  `json:"type"`
	Max           []float32     `json:"max,omitempty"`
	Min           []float32     `json:"min,omitempty"`
	Normalized    bool          `json:"normalized,omitempty"`
}

type asset struct {
	Generator string `json:"generator"`
	Version   string `json:"version"`
}

type bufferView struct {
	Buffer     int          `json:"buffer"`
	ByteLength int          `json:"byteLength"`
	ByteStride int          `json:"byteStride,omitempty"`
	ByteOffset int          `json:"byteOffset,omitempty"`
	Target     bufferTarget `json:"target"`
}

type mesh struct {
	Name       string      `json:"name"`
	Primitives []primitive `json:"primitives"`
}

type primitive struct {
	Extensions map[string]map[string]any `json:"extensions"`
	Attributes map[string]int            `json:"attributes"`
	Mode       int                       `json:"mode"`
}

type node struct {
	Matrix []float64 `json:"matrix"`
	Name   string    `json:"name"`
	Mesh   int       `json:"mesh"`
}

type scene struct {
	Name  string `json:"name"`
	Nodes []int  `json:"nodes"`
}
