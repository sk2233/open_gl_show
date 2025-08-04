package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"os"
	"sort"
	"strings"
)

func LoadPMX(name string) ([]*Mesh, *PMX) {
	// 加载基本信息
	file, err := os.Open(BasePath + ResPath + name)
	HandleErr(err)
	pmx, err := DecodePMX(file)
	HandleErr(err)
	// 模型格式非 OpenGL 坐标系 各种坐标系都需要转换
	for _, item := range pmx.Vertices {
		item.Position[2] *= -1
		item.Normal[2] *= -1
	}
	for _, morph := range pmx.Morphs {
		for _, item := range morph.PositionMorphOffsets {
			item.Offset[2] *= -1
		}
	}
	// 初始化骨骼信息
	for _, bone := range pmx.Bones {
		if bone.ParentIndex >= 0 {
			parent := pmx.Bones[bone.ParentIndex]
			bone.Parent = parent
			parent.Children = append(parent.Children, bone)
			bone.Translate = bone.Position.Sub(parent.Position)
			// 原坐标系不是 OpenGL 需要进行调整
			bone.Translate[2] *= -1
		} else {
			bone.Translate = bone.Position
			bone.Translate[2] *= -1
		}
		bone.Rotate = mgl32.QuatIdent()
		bone.Global = mgl32.Translate3D(bone.Position[0], bone.Position[1], -bone.Position[2])
		bone.GlobalInverse = bone.Global.Inv()
		bone.IsAppendRotate = (bone.Flags & BONE_FLAG_BLEND_ROTATION) > 0
		bone.IsAppendTranslate = (bone.Flags & BONE_FLAG_BLEND_TRANSLATION) > 0
		bone.IsAppendLocal = (bone.Flags & BONE_FLAG_BLEND_LOCAL) > 0
	}
	// 先按计算顺序排序 不要破坏原顺序
	pmx.SortBones = make([]*Bone, 0)
	for _, bone := range pmx.Bones {
		pmx.SortBones = append(pmx.SortBones, bone)
	}
	sort.SliceStable(pmx.SortBones, func(i, j int) bool { // 排序孩子
		return pmx.SortBones[i].TransformOrder < pmx.SortBones[j].TransformOrder
	})
	// 组装 mesh
	subPath := name[:strings.LastIndexByte(name, '/')+1]
	meshes := make([]*Mesh, 0)
	start := 0
	for _, material := range pmx.Materials {
		end := start + int(material.NumVerts)
		face := pmx.Faces[start:end]
		data := loadVec(pmx.Vertices, face)
		start = end
		meshes = append(meshes, &Mesh{
			Name: material.Name,
			Vao:  NewVao(data, gl.TRIANGLES, 3, 3, 2),
			Material: &Material{
				BaseTexture: loadPMXTexture(material.Texture, pmx, subPath),
				ToonTexture: loadPMXTexture(material.ToonTexture, pmx, subPath),
				SpeTexture:  loadPMXTexture(material.SpTexture, pmx, subPath),
				EdgeColor:   Ptr(material.EdgeColor),
				EdgeSize:    Ptr(material.EdgeSize),
				Flags:       material.Flags,
			},
			Faces:    face,
			Vertices: pmx.Vertices,
		})
	}
	return meshes, pmx
}

func loadPMXTexture(idx int32, pmx *PMX, subPath string) *Texture {
	if idx < 0 || idx >= int32(len(pmx.Textures)) {
		return nil
	}
	name := strings.ReplaceAll(pmx.Textures[idx], "\\", "/")
	return LoadTexture(subPath + name)
}

func loadVec(vs []*Vertex, faces []uint32) []float32 {
	res := make([]float32, 0)
	for _, face := range faces {
		temp := vs[face]
		res = append(res, temp.UpdatePosition[0], temp.UpdatePosition[1], temp.UpdatePosition[2])
		res = append(res, temp.UpdateNormal[0], temp.UpdateNormal[1], temp.UpdateNormal[2])
		res = append(res, temp.UV[0], temp.UV[1])
	}
	return res
}
