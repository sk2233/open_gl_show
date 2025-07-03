package main

import (
	"encoding/binary"
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/qmuntal/gltf"
	"math"
	"strings"
)

type Mesh struct {
	Vao     *Vao
	Mode    uint32
	Texture *Texture
	Model   mgl32.Mat4
}

func getRootNode(doc *gltf.Document) *gltf.Node {
	scene := Elem(doc.Scene, 0)
	nodes := doc.Scenes[scene].Nodes
	if len(nodes) != 1 {
		panic("scenes node not one")
	}
	return doc.Nodes[nodes[0]]
}

var (
	meshRes = make([]*Mesh, 0) // 有使用全局变量注意不能并发
)

func LoadMeshes(path string) []*Mesh {
	meshRes = make([]*Mesh, 0)
	doc, err := gltf.Open(BasePath + ResPath + path)
	HandleErr(err)
	root := getRootNode(doc)
	subPath := path[:strings.LastIndexByte(path, '/')+1]
	dfsMesh(root, doc, subPath, mgl32.Ident4())
	return meshRes
}

func toMat4(val [16]float64) mgl32.Mat4 {
	mat4 := mgl32.Mat4{}
	for i := 0; i < len(val); i++ {
		mat4[i] = float32(val[i])
	}
	return mat4
}

func dfsMesh(node *gltf.Node, doc *gltf.Document, subPath string, model mgl32.Mat4) {
	model = model.Mul4(toMat4(node.Matrix))
	if node.Mesh != nil { // 一个 mesh 可能有多个子 mesh 尽可能细的拆分
		for _, mesh := range doc.Meshes[*node.Mesh].Primitives {
			loadMesh(mesh, doc, subPath, model)
		}
	}
	for _, child := range node.Children {
		dfsMesh(doc.Nodes[child], doc, subPath, model)
	}
}

var (
	modeMap = map[gltf.PrimitiveMode]uint32{
		gltf.PrimitiveTriangles: gl.TRIANGLES,
	}
)

func loadMesh(mesh *gltf.Primitive, doc *gltf.Document, subPath string, model mgl32.Mat4) {
	mode, ok := modeMap[mesh.Mode]
	if !ok {
		panic(fmt.Sprintf("mode %v not support", mesh.Mode))
	}
	posData := load4FData(mesh.Attributes["POSITION"], doc)
	norData := load4FData(mesh.Attributes["NORMAL"], doc)
	texData := load4FData(mesh.Attributes["TEXCOORD_0"], doc)
	indData := loadU4IData(*mesh.Indices, doc)
	data := make([]float32, 0)
	posIdx, norIdx, texIdx := 0, 0, 0
	for posIdx < len(posData) && norIdx < len(norData) && texIdx < len(texData) {
		data = append(data, posData[posIdx:posIdx+3]...)
		data = append(data, norData[norIdx:norIdx+3]...)
		data = append(data, texData[texIdx:texIdx+2]...)
		posIdx += 3
		norIdx += 3
		texIdx += 2
	}
	temp := doc.Materials[*mesh.Material].PBRMetallicRoughness
	if temp.BaseColorTexture == nil {
		return // 暂定只加载有纹理的
	}
	meshRes = append(meshRes, &Mesh{
		Vao:     NewVaoWithIndic(data, indData, 3, 3, 2), // Pos Nor Tex
		Mode:    mode,
		Texture: loadTexture(doc.Textures[temp.BaseColorTexture.Index], doc, subPath),
		Model:   model,
	})
}

func toVec4(val [4]float64) mgl32.Vec4 {
	res := mgl32.Vec4{}
	for i := 0; i < len(val); i++ {
		res[i] = float32(val[i])
	}
	return res
}

var (
	magMap = map[gltf.MagFilter]int32{
		gltf.MagLinear:  gl.LINEAR,
		gltf.MagNearest: gl.NEAREST,
	}
	minMap = map[gltf.MinFilter]int32{
		gltf.MinLinearMipMapLinear: gl.LINEAR_MIPMAP_LINEAR,
		gltf.MinNearest:            gl.NEAREST,
	}
	wrapMap = map[gltf.WrappingMode]int32{
		gltf.WrapRepeat:      gl.REPEAT,
		gltf.WrapClampToEdge: gl.CLAMP_TO_EDGE,
	}
)

func loadTexture(texture *gltf.Texture, doc *gltf.Document, subPath string) *Texture {
	sampler := doc.Samplers[*texture.Sampler]
	minFilter, ok := minMap[sampler.MinFilter]
	if !ok {
		panic(fmt.Sprintf("minFilter %v not support", sampler.MinFilter))
	}
	magFilter, ok := magMap[sampler.MagFilter]
	if !ok {
		panic(fmt.Sprintf("magFilter %v not support", sampler.MagFilter))
	}
	wrapS, ok := wrapMap[sampler.WrapS]
	if !ok {
		panic(fmt.Sprintf("wrapS %v not support", sampler.WrapS))
	}
	wrapT, ok := wrapMap[sampler.WrapT]
	if !ok {
		panic(fmt.Sprintf("wrapT %v not support", sampler.WrapT))
	}
	path := subPath + doc.Images[*texture.Source].URI
	return LoadTextureWithSampler(path, minFilter, magFilter, wrapS, wrapT)
}

var (
	dataSize = map[gltf.ComponentType]int{
		gltf.ComponentFloat:  4,
		gltf.ComponentUint:   4,
		gltf.ComponentUshort: 2,
	}
	typeSize = map[gltf.AccessorType]int{
		gltf.AccessorScalar: 1,
		gltf.AccessorVec2:   2,
		gltf.AccessorVec3:   3,
	}
)

func loadData(acc *gltf.Accessor, doc *gltf.Document) []byte {
	view := doc.BufferViews[*acc.BufferView]
	buff := doc.Buffers[view.Buffer]
	offset := view.ByteOffset + acc.ByteOffset
	count := acc.Count * dataSize[acc.ComponentType] * typeSize[acc.Type]
	if count == 0 { // 极有可能是类型没有映射
		panic(fmt.Sprintf("loadData count is zero componentType %v , type %v", acc.ComponentType, acc.Type))
	}
	return buff.Data[offset : offset+count]
}

func loadU4IData(idx int, doc *gltf.Document) []uint32 {
	acc := doc.Accessors[idx]
	data := loadData(acc, doc)
	res := make([]uint32, 0)
	switch acc.ComponentType {
	case gltf.ComponentUint:
		for i := 0; i < len(data); i += 4 {
			res = append(res, binary.LittleEndian.Uint32(data[i:]))
		}
	case gltf.ComponentUshort:
		for i := 0; i < len(data); i += 2 {
			res = append(res, uint32(binary.LittleEndian.Uint16(data[i:])))
		}
	default:
		panic(fmt.Sprintf("componentType %v not support", acc.ComponentType))
	}
	return res
}

func load4FData(idx int, doc *gltf.Document) []float32 {
	acc := doc.Accessors[idx]
	data := loadData(acc, doc)
	res := make([]float32, 0)
	for i := 0; i < len(data); i += 4 {
		temp := binary.LittleEndian.Uint32(data[i:])
		res = append(res, math.Float32frombits(temp))
	}
	return res
}
