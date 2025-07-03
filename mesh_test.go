package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"testing"
)

func TestMesh(t *testing.T) {
	window := NewWindow(1280, 720, "Test")

	shader := LoadShader("color")
	shader.Use()
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(16)/9, 0.1, 30.0)
	shader.SetMat4("projection", projection)
	shader.SetF3("lightPos", mgl32.Vec3{3, 3, 3})

	meshes := LoadMeshes("nina/scene.gltf")

	camera := NewCamera()

	gl.Enable(gl.DEPTH_TEST)
	for !window.ShouldClose() {
		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		camera.Update(window)

		shader.Use()
		shader.SetMat4("view", camera.GetView())
		shader.SetF3("viewPos", camera.Pos)
		for _, mesh := range meshes {
			shader.SetMat4("model", mesh.Model)
			mesh.Texture.Bind(gl.TEXTURE0)
			mesh.Vao.Bind()
			mesh.Vao.DrawIndic(mesh.Mode)
		}

		window.SwapBuffers()
		glfw.PollEvents()
	}
	glfw.Terminate()
}
