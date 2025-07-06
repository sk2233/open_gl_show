package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"os"
	"strings"
)

type Material struct {
	EmissiveColor            *mgl32.Vec3
	EmissiveTexture          *Texture
	NormalTexture            *Texture
	OcclusionTexture         *Texture
	BaseColor                *mgl32.Vec4
	BaseTexture              *Texture
	MetallicRoughnessTexture *Texture
	Metallic                 *float32
	Roughness                *float32
	AlphaMode                *string
	AlphaCutoff              *float32
}

type Mesh struct {
	Name     string
	Model    mgl32.Mat4 // 位置
	Vao      *Vao       // 模型
	Material *Material  // 材质
}

func getRootNode(gltf *GlTFData) *NodeData {
	nodes := gltf.Scenes[gltf.Scene].Nodes
	if len(nodes) != 1 {
		panic("scenes node not one")
	}
	return gltf.Nodes[nodes[0]]
}

var (
	meshRes = make([]*Mesh, 0) // 有使用全局变量注意不能并发
)

func LoadMeshes(path string) []*Mesh {
	meshRes = make([]*Mesh, 0)
	bs, err := os.ReadFile(BasePath + ResPath + path)
	HandleErr(err)
	gltf := &GlTFData{}
	err = json.Unmarshal(bs, gltf)
	HandleErr(err)
	root := getRootNode(gltf)
	subPath := path[:strings.LastIndexByte(path, '/')+1]
	dfsMesh(root, gltf, subPath, mgl32.Ident4())
	return meshRes
}

func dfsMesh(node *NodeData, gltf *GlTFData, subPath string, model mgl32.Mat4) {
	if node.Matrix != nil {
		model = model.Mul4(*node.Matrix)
	}
	if node.Mesh != nil { // 一个 mesh 可能有多个子 mesh 尽可能细的拆分
		temp := gltf.Meshes[*node.Mesh]
		for _, mesh := range temp.Primitives {
			loadMesh(mesh, temp.Name, gltf, subPath, model)
		}
	}
	for _, child := range node.Children {
		dfsMesh(gltf.Nodes[child], gltf, subPath, model)
	}
}

func mergeData(temp [][]float32, sizes []int32) []float32 {
	idx := 0
	res := make([]float32, 0)
	for {
		for i := 0; i < len(sizes); i++ {
			start := idx * int(sizes[i])
			end := (idx + 1) * int(sizes[i])
			if end > len(temp[i]) {
				return res
			}
			res = append(res, temp[i][start:end]...)
		}
		idx++
	}
}

func loadMesh(mesh *PrimitiveData, name string, gltf *GlTFData, subPath string, model mgl32.Mat4) {
	temp := make([][]float32, 0)
	sizes := []int32{3, 3, 2}
	temp = append(temp, load4FData(mesh.Attributes["POSITION"], gltf, subPath))
	temp = append(temp, load4FData(mesh.Attributes["NORMAL"], gltf, subPath))
	temp = append(temp, load4FData(mesh.Attributes["TEXCOORD_0"], gltf, subPath))
	if _, ok := mesh.Attributes["TANGENT"]; ok {
		temp = append(temp, load4FData(mesh.Attributes["TANGENT"], gltf, subPath))
		sizes = append(sizes, 4)
	}
	indData := loadU4IData(mesh.Indices, gltf, subPath)
	data := mergeData(temp, sizes)
	material := gltf.Materials[mesh.Material]
	meshRes = append(meshRes, &Mesh{
		Name:  name,
		Vao:   NewVaoWithIndic(data, indData, Elem(mesh.Mode, gl.TRIANGLES), sizes...),
		Model: model,
		Material: &Material{
			EmissiveColor:            material.EmissiveFactor,
			EmissiveTexture:          loadTexture(material.EmissiveTexture, gltf, subPath),
			NormalTexture:            loadTexture(material.NormalTexture, gltf, subPath),
			OcclusionTexture:         loadTexture(material.OcclusionTexture, gltf, subPath),
			BaseColor:                material.PbrMetallicRoughness.BaseColorFactor,
			BaseTexture:              loadTexture(material.PbrMetallicRoughness.BaseColorTexture, gltf, subPath),
			MetallicRoughnessTexture: loadTexture(material.PbrMetallicRoughness.MetallicRoughnessTexture, gltf, subPath),
			Metallic:                 material.PbrMetallicRoughness.MetallicFactor,
			Roughness:                material.PbrMetallicRoughness.RoughnessFactor,
			AlphaMode:                material.AlphaMode,
			AlphaCutoff:              material.AlphaCutoff,
		},
	})
}

func loadTexture(data *textureData, gltf *GlTFData, subPath string) *Texture {
	if data == nil {
		return nil
	}
	texture := gltf.Textures[data.Index]
	path := subPath + gltf.Images[texture.Source].Uri
	if texture.Sampler != nil {
		sampler := gltf.Samplers[*texture.Sampler]
		return LoadTextureWithSampler(path, sampler)
	} else {
		return LoadTexture(path)
	}
}

var (
	buffCache = make(map[string][]byte)
)

func loadData(acc *AccessorData, gltf *GlTFData, subPath string) []byte {
	view := gltf.BufferViews[acc.BufferView]
	buff := gltf.Buffers[view.Buffer]
	offset := view.ByteOffset + acc.ByteOffset
	count := acc.Count * GetDataSize(acc.ComponentType) * acc.Type.GetSize()
	if count == 0 { // 极有可能是类型没有映射
		panic(fmt.Sprintf("loadData count is zero componentType %v , type %v", acc.ComponentType, acc.Type))
	}
	if _, ok := buffCache[buff.Uri]; !ok {
		bs, err := os.ReadFile(BasePath + ResPath + subPath + buff.Uri)
		HandleErr(err)
		buffCache[buff.Uri] = bs
	}
	return buffCache[buff.Uri][offset : offset+count]
}

func loadU4IData(idx int, gltf *GlTFData, subPath string) []uint32 {
	acc := gltf.Accessors[idx]
	data := loadData(acc, gltf, subPath)
	res := make([]uint32, 0)
	switch acc.ComponentType {
	case gl.UNSIGNED_INT:
		for i := 0; i < len(data); i += 4 {
			res = append(res, binary.LittleEndian.Uint32(data[i:]))
		}
	case gl.UNSIGNED_SHORT:
		for i := 0; i < len(data); i += 2 {
			res = append(res, uint32(binary.LittleEndian.Uint16(data[i:])))
		}
	default:
		panic(fmt.Sprintf("componentType %v not support", acc.ComponentType))
	}
	return res
}

func load4FData(idx int, gltf *GlTFData, subPath string) []float32 {
	acc := gltf.Accessors[idx]
	data := loadData(acc, gltf, subPath)
	res := make([]float32, 0)
	for i := 0; i < len(data); i += 4 {
		temp := binary.LittleEndian.Uint32(data[i:])
		res = append(res, math.Float32frombits(temp))
	}
	return res
}
