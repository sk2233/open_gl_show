package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"testing"
)

func TestFrame(t *testing.T) {
	window := NewWindow(1280, 720, "Test")

	shader := LoadShader("pbr")
	shader.Use()
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(16)/9, 0.1, 100.0)
	shader.SetMat4("Projection", projection)
	quadShader := LoadShader("quad")
	skyShader := LoadShader("sky")
	skyShader.Use()
	skyShader.SetMat4("Projection", projection)
	skyShader.SetMat4("Model", mgl32.Scale3D(50, 50, 50))

	meshes := LoadMeshes("nina/scene.gltf")
	quadVao := NewVao(quadVT, 3, 2)
	skyVao := NewVao(cubeV, 3)

	frame := CreateFrame(1280*2, 720*2)
	camera := NewCamera()
	skyCube := LoadCubeMap("cube/right.jpg", "cube/left.jpg", "cube/top.jpg", "cube/bottom.jpg", "cube/front.jpg", "cube/back.jpg")

	gl.Enable(gl.DEPTH_TEST)
	for !window.ShouldClose() {
		camera.Update(window)

		frame.Use()
		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		skyShader.Use()
		skyShader.SetMat4("View", camera.GetView())
		skyCube.Bind(gl.TEXTURE0)
		skyVao.Bind()
		skyVao.Draw(gl.TRIANGLES)
		shader.Use()
		curr := glfw.GetTime()
		shader.SetF3("LightPos", mgl32.Vec3{float32(math.Cos(curr) * 5), float32(math.Sin(curr) * 5), 5})
		shader.SetMat4("View", camera.GetView())
		shader.SetF3("ViewPos", camera.Pos)
		for _, mesh := range meshes {
			shader.SetMat4("Model", mesh.Model)
			mesh.Texture.Bind(gl.TEXTURE0)
			mesh.Vao.Bind()
			mesh.Vao.DrawIndic(mesh.Mode)
		}

		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		quadShader.Use()
		frame.Texture.Bind(gl.TEXTURE0)
		quadVao.Bind()
		quadVao.Draw(gl.TRIANGLE_STRIP)

		window.SwapBuffers()
		glfw.PollEvents()
	}
	glfw.Terminate()
}
