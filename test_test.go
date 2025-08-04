package main

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"testing"
)

func TestTest(t *testing.T) {
	window := NewWindow(800, 600, "Test")

	shader := LoadShader("test")
	shader.Use()
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(4)/3, 0.1, 10.0)
	shader.SetMat4("projection", projection)
	view := mgl32.LookAtV(mgl32.Vec3{3, 3, 3}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	shader.SetMat4("view", view)
	shader.SetMat4("model", mgl32.Ident4())

	texture := LoadTexture("square.png")

	vao := NewVao(cubeVT, gl.TRIANGLES, 3, 2)

	gl.Enable(gl.DEPTH_TEST)
	for !window.ShouldClose() {
		gl.ClearColor(1.0, 1.0, 1.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		shader.Use()
		model := mgl32.HomogRotate3D(float32(glfw.GetTime()), mgl32.Vec3{0, 1, 0})
		shader.SetMat4("model", model)
		texture.Bind(gl.TEXTURE0)
		vao.Bind()
		vao.Draw()

		window.SwapBuffers()
		glfw.PollEvents()
	}
	glfw.Terminate()
}

func TestPmxL(t *testing.T) {
	_, pmx := LoadPMX("pmx.pmx")
	fmt.Println(pmx)
}
