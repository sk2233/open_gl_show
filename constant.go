package main

import "github.com/go-gl/mathgl/mgl32"

const (
	BasePath   = "/Users/sky/Documents/go/open_gl_show/"
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
