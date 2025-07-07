package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"os"
	"strings"
)

func LoadPMX(name string) []*Mesh {
	file, err := os.Open(BasePath + ResPath + name)
	HandleErr(err)
	pmx, err := DecodePMX(file)
	HandleErr(err)
	subPath := name[:strings.LastIndexByte(name, '/')+1]
	meshes := make([]*Mesh, 0)
	start := 0
	for _, material := range pmx.Materials {
		end := start + int(material.NumVerts)
		data := loadVec(pmx.Vertices, pmx.Faces[start:end])
		start = end
		meshes = append(meshes, &Mesh{
			Name: material.Name,
			Vao:  NewVao(data, gl.TRIANGLES, 3, 3, 2),
			Material: &Material{
				BaseTexture: loadPMXTexture(material.Texture, pmx, subPath),
				ToonTexture: loadPMXTexture(material.ToonTexture, pmx, subPath),
				EdgeColor:   Ptr(material.EdgeColor),
				EdgeSize:    Ptr(material.EdgeSize),
			},
		})
	}
	return meshes
}

func loadPMXTexture(idx int32, pmx *PMX, subPath string) *Texture {
	if idx < 0 || idx >= int32(len(pmx.Textures)) {
		return nil
	}
	name := strings.ReplaceAll(pmx.Textures[idx], "\\", "/")
	return LoadTexture(subPath + name)
}

func loadVec(vs []Vertex, faces []uint32) []float32 {
	res := make([]float32, 0)
	for _, face := range faces {
		res = append(res, vs[face].Position[0], vs[face].Position[1], vs[face].Position[2])
		res = append(res, vs[face].Normal[0], vs[face].Normal[1], vs[face].Normal[2])
		res = append(res, vs[face].UV[0], vs[face].UV[1])
	}
	return res
}
