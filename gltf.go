package main

import "github.com/go-gl/mathgl/mgl32"

// 不一定是最全的，有需要再加字段
type GlTFData struct {
	Accessors   []*AccessorData   `json:"accessors"`
	BufferViews []*BufferViewData `json:"bufferViews"`
	Buffers     []*BufferData     `json:"buffers"`
	Images      []*ImageData      `json:"images"`
	Materials   []*MaterialData   `json:"materials"`
	Meshes      []*MeshData       `json:"meshes"`
	Nodes       []*NodeData       `json:"nodes"`
	Samplers    []*SamplerData    `json:"samplers"`
	Scene       int               `json:"scene"`
	Scenes      []*SceneData      `json:"scenes"`
	Textures    []*TextureData    `json:"textures"`
}

type TextureData struct {
	Sampler int `json:"sampler"`
	Source  int `json:"source"`
}

type SceneData struct {
	Nodes []int `json:"nodes"`
}

type SamplerData struct {
	MagFilter int32 `json:"magFilter"`
	MinFilter int32 `json:"minFilter"`
	WrapS     int32 `json:"wrapS"`
	WrapT     int32 `json:"wrapT"`
}

type NodeData struct {
	Children []int       `json:"children"`
	Matrix   *mgl32.Mat4 `json:"matrix"` // 可能不存在
	Mesh     *int        `json:"mesh"`   // 可能没有
}

type PrimitiveData struct {
	Attributes map[string]int `json:"attributes"`
	Indices    int            `json:"indices"`
	Material   int            `json:"material"`
	Mode       uint32         `json:"mode"`
}

type MeshData struct {
	Name       string           `json:"name"`
	Primitives []*PrimitiveData `json:"primitives"`
}

type textureData struct {
	Index int `json:"index"`
}

type MaterialData struct {
	EmissiveFactor       *mgl32.Vec3  `json:"emissiveFactor"`
	EmissiveTexture      *textureData `json:"emissiveTexture"`  // 自发光贴图
	NormalTexture        *textureData `json:"normalTexture"`    // 法线贴图
	OcclusionTexture     *textureData `json:"occlusionTexture"` // AO 贴图 R 通道
	PbrMetallicRoughness struct {
		BaseColorFactor          *mgl32.Vec4  `json:"baseColorFactor"`
		BaseColorTexture         *textureData `json:"baseColorTexture"`         // 基本色贴图
		MetallicRoughnessTexture *textureData `json:"metallicRoughnessTexture"` // 金属(B)&粗糙度(G)贴图
		MetallicFactor           *float32     `json:"metallicFactor"`
		RoughnessFactor          *float32     `json:"roughnessFactor"`
	} `json:"pbrMetallicRoughness"`
	AlphaMode   *string  `json:"alphaMode"` // 透明模式 MASK 镂空，根据 alphaCutoff 与 a 值进行判断是非展示
	AlphaCutoff *float32 `json:"alphaCutoff"`
}

type ImageData struct {
	Uri string `json:"uri"`
}

type BufferData struct {
	ByteLength int    `json:"byteLength"`
	Uri        string `json:"uri"`
}

type BufferViewData struct {
	Buffer     int `json:"buffer"`
	ByteLength int `json:"byteLength"`
	ByteOffset int `json:"byteOffset"` // 默认为 0
}

type AccessorData struct {
	BufferView    int      `json:"bufferView"`
	ByteOffset    int      `json:"byteOffset"` // 默认为 0
	ComponentType uint32   `json:"componentType"`
	Count         int      `json:"count"`
	Type          DataType `json:"type"`
}
