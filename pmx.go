package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"os"
	"strings"
)

func LoadPMX(name string) ([]*Mesh, *PMX) {
	file, err := os.Open(BasePath + ResPath + name)
	HandleErr(err)
	pmx, err := DecodePMX(file)
	HandleErr(err)
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
		res = append(res, temp.CurrPos[0]+temp.PosOffset[0], temp.CurrPos[1]+temp.PosOffset[1], temp.CurrPos[2]+temp.PosOffset[2])
		res = append(res, temp.Normal[0], temp.Normal[1], temp.Normal[2])
		res = append(res, temp.UV[0], temp.UV[1])
	}
	return res
}
