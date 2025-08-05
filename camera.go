package main

import (
	"fmt"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Camera struct {
	Pos, Dir   mgl32.Vec3
	dirX, dirY float32 // 不能绕Z旋转
}

func NewCamera() *Camera {
	return &Camera{Pos: mgl32.Vec3{-0.9229493, 19.44317, 11.283254}, Dir: mgl32.Vec3{0.041262392, -0.52491015, -1.6500704}}
}

func (c *Camera) GetView() mgl32.Mat4 {
	return mgl32.LookAtV(c.Pos, c.Pos.Add(c.Dir), VecUp)
}

func (c *Camera) TranslateX(value float32) {
	c.Pos = c.Pos.Add(c.Dir.Cross(VecUp).Normalize().Mul(value))
}

func (c *Camera) TranslateY(value float32) {
	c.Pos = c.Pos.Add(c.Dir.Cross(VecRight).Normalize().Mul(value))
}

func (c *Camera) TranslateZ(value float32) {
	c.Pos = c.Pos.Add(c.Dir.Normalize().Mul(value))
}

func (c *Camera) RotateX(value float32) { // 左右看 沿着 Y轴
	c.Dir = mgl32.Rotate3DY(value).Mul3x1(c.Dir)
}

func (c *Camera) RotateY(value float32) { // 上下看 沿着  X轴
	c.Dir = mgl32.Rotate3DX(value).Mul3x1(c.Dir)
}

func (c *Camera) Update(window *glfw.Window) {
	offsetX := GetAxis(window, glfw.KeyA, glfw.KeyD)
	if offsetX != 0 {
		c.TranslateX(offsetX * 0.1)
	}
	offsetY := GetAxis(window, glfw.KeyE, glfw.KeyQ)
	if offsetY != 0 {
		c.TranslateY(offsetY * 0.1)
	}
	offsetZ := GetAxis(window, glfw.KeyS, glfw.KeyW)
	if offsetZ != 0 {
		c.TranslateZ(offsetZ * 0.1)
	}
	rotateX := GetAxis(window, glfw.KeyRight, glfw.KeyLeft)
	if rotateX != 0 {
		c.RotateX(rotateX * 0.01)
	}
	rotateY := GetAxis(window, glfw.KeyDown, glfw.KeyUp)
	if rotateY != 0 {
		c.RotateY(rotateY * 0.01)
	}
	if PressKey(window, glfw.KeyEnter) {
		fmt.Println(c.Pos, c.Dir)
	}
}
