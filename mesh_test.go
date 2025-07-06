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
	shader.SetMat4("Projection", projection)
	shader.SetF3("LightPos", mgl32.Vec3{5, 5, 5})
	outlineShader := LoadShader("outline")
	outlineShader.Use()
	outlineShader.SetF3("Color", mgl32.Vec3{1, 0, 0})
	outlineShader.SetMat4("Projection", projection)

	meshes := LoadMeshes("nina/scene.gltf")

	camera := NewCamera()

	for !window.ShouldClose() {
		gl.ClearColor(0.3, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		camera.Update(window)

		// 绘制描边
		gl.Disable(gl.DEPTH_TEST)
		outlineShader.Use()
		outlineShader.SetMat4("View", camera.GetView())
		for _, mesh := range meshes {
			outlineShader.SetMat4("Model", mesh.Model)
			mesh.Vao.Bind()
			mesh.Vao.DrawIndic()
		}
		// 正常绘制对象
		gl.Enable(gl.DEPTH_TEST)
		shader.Use()
		shader.SetMat4("View", camera.GetView())
		shader.SetF3("ViewPos", camera.Pos)
		for _, mesh := range meshes {
			shader.SetMat4("Model", mesh.Model)
			if mesh.Material.BaseTexture == nil {
				if mesh.Name != "Line_Line_0" {
					continue
				}
				shader.SetF4("Color", *mesh.Material.BaseColor)
				shader.SetI1("UseColor", gl.TRUE)
			} else {
				mesh.Material.BaseTexture.Bind(gl.TEXTURE0)
				shader.SetI1("UseColor", gl.FALSE)
			}
			mesh.Vao.Bind()
			mesh.Vao.DrawIndic()
		}

		window.SwapBuffers()
		glfw.PollEvents()
	}
	glfw.Terminate()
}
