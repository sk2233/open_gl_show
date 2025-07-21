package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"testing"
)

// TODO 学习项目 加载大喜模型  与  骨骼，物理动画

func TestPmx(t *testing.T) {
	window := NewWindow(1280, 720, "Test")

	shader := LoadShader("pmx")
	shader.Use()
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(16)/9, 0.1, 30.0)
	shader.SetMat4("Projection", projection)
	shader.SetMat4("Model", mgl32.Ident4())
	shader.SetI1("BaseTex", 0)
	shader.SetI1("ToonTex", 1)
	shader.SetI1("SpeTex", 2)
	outlineShader := LoadShader("outline")
	outlineShader.Use()
	outlineShader.SetMat4("Projection", projection)
	outlineShader.SetMat4("Model", mgl32.Ident4())

	meshes, pmx := LoadPMX("星穹铁道—流萤/星穹铁道—流萤.pmx")
	vmd := LoadVMD("ikuyo/ikuyo.vmd") // 只有骨骼动画与表情动画
	//morphCalculator := NewMorphCalculator(vmd.MorphFrames, pmx.Morphs)
	boneCalculator := NewBoneCalculator(vmd.BoneFrames, pmx.Bones)
	camera := NewCamera()

	time := uint32(0)
	lastTime := glfw.GetTime()
	gl.Enable(gl.DEPTH_TEST)
	for !window.ShouldClose() {
		gl.ClearColor(0.3, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		camera.Update(window)

		nowTime := glfw.GetTime()
		if nowTime-lastTime > 1.0/30 {
			time++
			lastTime += 1.0 / 30
		}
		// 更新节点
		pmx.ResetVertex()
		//morphWeights := morphCalculator.Calculate(time)
		//for idx, weight := range morphWeights {
		//	pmx.ApplyMorph(idx, weight)
		//}
		bonePosAndRotates := boneCalculator.Calculate(time)
		pmx.ApplyBone(bonePosAndRotates)
		shader.Use()

		shader.SetF3("LightPos", mgl32.Vec3{30 * float32(math.Sin(glfw.GetTime())), 30, 30 * float32(math.Cos(glfw.GetTime()))})
		for _, mesh := range meshes {
			mesh.UpdateVertex() // 更新节点
			material := mesh.Material
			if material.BaseTexture == nil {
				continue
			}
			// 先绘制对象
			gl.Disable(gl.CULL_FACE)
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
			if material.SpeTexture != nil {
				material.SpeTexture.Bind(gl.TEXTURE2)
				shader.SetI1("UseSpe", gl.TRUE)
			} else {
				shader.SetI1("UseSpe", gl.FALSE)
			}
			mesh.Vao.Bind()
			mesh.Vao.Draw()
			// 再绘制描边
			if material.Flags&MATERIAL_FLAG_DRAWEDGE == 0 {
				continue // 没有描边
			}
			gl.Enable(gl.CULL_FACE)
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
