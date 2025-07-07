package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"testing"
)

func TestPmx(t *testing.T) {
	window := NewWindow(1280, 720, "Test")

	shader := LoadShader("pmx")
	shader.Use()
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(16)/9, 0.1, 30.0)
	shader.SetMat4("Projection", projection)
	shader.SetMat4("Model", mgl32.Ident4())
	shader.SetF3("LightPos", mgl32.Vec3{-30, 30, -30})
	shader.SetI1("BaseTex", 0)
	shader.SetI1("ToonTex", 1)
	outlineShader := LoadShader("outline")
	outlineShader.Use()
	outlineShader.SetMat4("Projection", projection)
	outlineShader.SetMat4("Model", mgl32.Ident4())

	meshes := LoadPMX("坎特蕾拉_鸣潮/坎特蕾拉.pmx")
	camera := NewCamera()

	for !window.ShouldClose() {
		gl.ClearColor(0.3, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		camera.Update(window)

		// 正常绘制对象
		gl.Enable(gl.DEPTH_TEST)
		gl.Enable(gl.CULL_FACE)
		for i, mesh := range meshes {
			material := mesh.Material
			// 先绘制对象
			gl.CullFace(gl.BACK) // 只绘制正面
			shader.Use()
			shader.SetMat4("View", camera.GetView())
			shader.SetF3("ViewPos", camera.Pos)
			material.BaseTexture.Bind(gl.TEXTURE0)
			if material.ToonTexture != nil {
				material.ToonTexture.Bind(gl.TEXTURE1)
				shader.SetI1("UseToon", gl.TRUE)
			} else {
				shader.SetI1("UseToon", gl.FALSE)
			}
			mesh.Vao.Bind()
			mesh.Vao.Draw()
			// 再绘制描边
			if i == 10 || i == 5 { // 嘴就不要描边了，不好看
				continue
			}
			gl.CullFace(gl.FRONT) // 放大一点且只绘制反面
			outlineShader.Use()
			outlineShader.SetMat4("View", camera.GetView())
			outlineShader.SetF1("EdgeSize", *material.EdgeSize)
			outlineShader.SetF4("EdgeColor", *material.EdgeColor)
			mesh.Vao.Bind()
			mesh.Vao.Draw()
		}

		window.SwapBuffers()
		glfw.PollEvents()
	}
	glfw.Terminate()
}
