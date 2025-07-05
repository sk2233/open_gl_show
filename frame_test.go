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
	samplerCube := createSamplerCube(skyVao, skyCube)

	gl.Enable(gl.DEPTH_TEST)
	for !window.ShouldClose() {
		camera.Update(window)

		frame.Use()
		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		skyShader.Use()
		skyShader.SetMat4("View", camera.GetView())
		samplerCube.Bind(gl.TEXTURE0)
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

func createSamplerCube(vao *Vao, cube *Texture) *Texture {
	frame := CreateCubeFrame(512, 512)
	gl.Viewport(0, 0, 512, 512) // 与贴图对齐
	shader := LoadShader("sampler")
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
		vao.Draw(gl.TRIANGLES)
	}
	gl.Viewport(0, 0, 1280*2, 720*2) // 恢复渲染窗口
	return frame.Texture
}
