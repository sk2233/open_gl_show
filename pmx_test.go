package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"testing"
)

// mmd 默认不是在 openGL 坐标系下，需要按 z 轴反转
func TestPmx(t *testing.T) {
	window := NewWindow(1280, 720, "Test")

	shader := LoadShader("mmd")
	shader.Use()
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(16)/9, 0.1, 30.0)
	shader.SetMat4("uProj", projection)
	shader.SetMat4("uModel", mgl32.Ident4())
	shader.SetF3("uLightColor", mgl32.Vec3{1, 1, 1})
	shader.SetI1("uTex", 0)
	shader.SetI1("uSphereTex", 1)
	shader.SetI1("uToonTex", 2)
	edgeShader := LoadShader("mmd_edge")
	edgeShader.Use()
	edgeShader.SetMat4("uProj", projection)
	edgeShader.SetMat4("uModel", mgl32.Ident4())
	edgeShader.SetF2("uScreenSize", mgl32.Vec2{1280, 720}) // 屏幕大小

	meshes, pmx := LoadPMX("星穹铁道—流萤/星穹铁道—流萤.pmx")
	vmd := LoadVMD("ikuyo/ikuyo.vmd") // 只有骨骼动画
	boneCalculator := NewBoneCalculator(vmd.BoneFrames, pmx.Bones)
	vmd = LoadVMD("ikuyo/表情.vmd") // 只有表情动画
	morphCalculator := NewMorphCalculator(vmd.MorphFrames, pmx.Morphs)
	camera := NewCamera()

	time := uint32(0)
	lastTime := glfw.GetTime()
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
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
		pmx.ResetBoneAndVertex()
		morphWeights := morphCalculator.Calculate(time)
		for idx, weight := range morphWeights {
			pmx.ApplyMorph(idx, weight)
		}
		bonePosAndRotates := boneCalculator.Calculate(time)
		pmx.ApplyBones(bonePosAndRotates)
		shader.Use()
		shader.SetF3("uLightPos", mgl32.Vec3{30 * float32(math.Sin(glfw.GetTime())), 30, 30 * float32(math.Cos(glfw.GetTime()))})
		shader.SetF3("uViewPos", camera.Pos)
		shader.SetMat4("uView", camera.GetView())
		edgeShader.Use()
		edgeShader.SetMat4("uView", camera.GetView())
		for _, mesh := range meshes {
			mesh.UpdateVertex() // 更新节点
			material := mesh.Material
			if material.Alpha == 0 {
				continue
			}
			// 先绘制对象
			if material.Flags&MATERIAL_FLAG_DOUBLESIDE == 0 {
				gl.Disable(gl.CULL_FACE)
			} else {
				gl.Enable(gl.CULL_FACE)
				gl.CullFace(gl.FRONT)
			}
			shader.Use()
			// 设置基础参数
			shader.SetF3("uDiffuse", material.Diffuse)
			shader.SetF1("uAlpha", material.Alpha)
			shader.SetF3("uSpecular", material.Specular)
			shader.SetF1("uSpecularPower", material.SpecularPower)
			shader.SetF3("uAmbient", material.Ambient)
			// 设置纹理图片
			if material.BaseTexture != nil {
				shader.SetI1("uTexMode", 1)
				material.BaseTexture.Bind(gl.TEXTURE0)
			} else {
				shader.SetI1("uTexMode", 0)
			}
			// 设置高光贴图
			if material.SpeTexture != nil {
				shader.SetI1("uSphereTexMode", material.SpeMode)
				material.SpeTexture.Bind(gl.TEXTURE1)
			} else {
				shader.SetI1("uSphereTexMode", 0)
			}
			// 设置卡通查找表
			if material.ToonTexture != nil {
				shader.SetI1("uToonTexMode", 1)
				material.ToonTexture.Bind(gl.TEXTURE2)
			} else {
				shader.SetI1("uToonTexMode", 0)
			}
			mesh.Vao.Bind()
			mesh.Vao.Draw()
			// 再绘制描边
			gl.Enable(gl.CULL_FACE)
			gl.CullFace(gl.BACK)
			if material.Flags&MATERIAL_FLAG_DRAWEDGE == 0 {
				continue // 没有描边
			}
			edgeShader.Use()
			edgeShader.SetF1("uEdgeSize", material.EdgeSize)
			edgeShader.SetF4("uEdgeColor", material.EdgeColor)
			mesh.Vao.Bind()
			mesh.Vao.Draw()
		}

		window.SwapBuffers()
		glfw.PollEvents()
	}
	glfw.Terminate()
}
