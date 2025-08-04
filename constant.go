package main

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	BasePath   = "/Users/wepie/Documents/github/open_gl_show/"
	ResPath    = "res/"
	ShaderPath = "shader/"
	VertName   = ".vert"
	FragName   = ".frag"
)

var (
	VecFront = mgl32.Vec3{0, 0, -1}
	VecUp    = mgl32.Vec3{0, 1, 0}
	VecRight = mgl32.Vec3{1, 0, 0}
	VecZero  = mgl32.Vec3{0, 0, 0}
)

type DataType string

func (d DataType) GetSize() int {
	switch d {
	case DataVec4:
		return 4
	case DataVec3:
		return 3
	case DataVec2:
		return 2
	case DataScalar:
		return 1
	default:
		panic(fmt.Sprintf("unknown DataType %v", d))
	}
}

const (
	DataVec4   DataType = "VEC4"
	DataVec3   DataType = "VEC3"
	DataVec2   DataType = "VEC2"
	DataScalar DataType = "SCALAR"
)
