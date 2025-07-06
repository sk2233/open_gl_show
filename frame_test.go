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
	shader.SetI1("BaseTex", 0)
	shader.SetI1("DiffuseTex", 1)
	shader.SetI1("SpecularTex", 2)
	skyShader := LoadShader("sky")
	skyShader.Use()
	skyShader.SetMat4("Projection", projection)
	skyShader.SetMat4("Model", mgl32.Scale3D(50, 50, 50))

	meshes := LoadMeshes("nina/scene.gltf")
	skyVao := NewVao(cubeV, gl.TRIANGLES, 3)

	camera := NewCamera()
	skyCube := LoadCubeMap("cube/right.jpg", "cube/left.jpg", "cube/top.jpg", "cube/bottom.jpg", "cube/front.jpg", "cube/back.jpg")
	diffuseCube := createSamplerCube(skyVao, skyCube, "diffuse")   //漫反射采样
	specularCube := createSamplerCube(skyVao, skyCube, "specular") //高光采样

	gl.Enable(gl.DEPTH_TEST)
	for !window.ShouldClose() {
		camera.Update(window)

		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		skyShader.Use()
		skyShader.SetMat4("View", camera.GetView())
		skyCube.Bind(gl.TEXTURE0)
		skyVao.Bind()
		skyVao.Draw()
		shader.Use()
		curr := glfw.GetTime()
		shader.SetF3("LightPos", mgl32.Vec3{float32(math.Cos(curr) * 5), float32(math.Sin(curr) * 5), 5})
		shader.SetMat4("View", camera.GetView())
		shader.SetF3("ViewPos", camera.Pos)
		diffuseCube.Bind(gl.TEXTURE1)
		specularCube.Bind(gl.TEXTURE2)
		for _, mesh := range meshes {
			material := mesh.Material
			shader.SetMat4("Model", mesh.Model)
			if material.BaseTexture == nil {
				if mesh.Name != "Line_Line_0" {
					continue
				}
				shader.SetF4("Color", *material.BaseColor)
				shader.SetI1("UseColor", gl.TRUE)
			} else {
				material.BaseTexture.Bind(gl.TEXTURE0)
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

func createSamplerCube(vao *Vao, cube *Texture, name string) *Texture {
	frame := CreateCubeFrame(256, 256)
	gl.Viewport(0, 0, 256, 256) // 与贴图对齐
	shader := LoadShader(name)
	shader.Use()
	shader.SetMat4("Projection", mgl32.Perspective(mgl32.DegToRad(90), 1, 0.1, 10.0))
	shader.SetMat4("Model", mgl32.Ident4())
	up := mgl32.Vec3{0, -1, 0} // 渲染会反向，使用特殊 up 方向进行纠正
	views := []mgl32.Mat4{mgl32.LookAtV(VecZero, mgl32.Vec3{1, 0, 0}, up), mgl32.LookAtV(VecZero, mgl32.Vec3{-1, 0, 0}, up), mgl32.LookAtV(VecZero, mgl32.Vec3{0, 1, 0}, mgl32.Vec3{0, 0, 1}),
		mgl32.LookAtV(VecZero, mgl32.Vec3{0, -1, 0}, mgl32.Vec3{0, 0, -1}), mgl32.LookAtV(VecZero, mgl32.Vec3{0, 0, 1}, up), mgl32.LookAtV(VecZero, mgl32.Vec3{0, 0, -1}, up)}
	frame.Use()
	cube.Bind(gl.TEXTURE0)
	for i := uint32(0); i < 6; i++ { // 指定要渲染到那个面上
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_CUBE_MAP_POSITIVE_X+i, frame.Texture.Texture, 0)
		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		shader.Use()
		shader.SetMat4("View", views[i]) // 循环渲染6个面
		vao.Bind()
		vao.Draw()
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(0, 0, 1280*2, 720*2) // 恢复渲染窗口
	return frame.Texture
}
