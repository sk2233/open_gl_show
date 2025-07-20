package main

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"runtime"
)

func HandleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func NewWindow(width, height int, title string) *glfw.Window {
	runtime.LockOSThread()
	// 初始化 glfw 辅助窗口
	err := glfw.Init()
	HandleErr(err)
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(width, height, title, nil, nil)
	HandleErr(err)
	window.MakeContextCurrent()
	// 初始化 gl
	err = gl.Init()
	HandleErr(err)
	return window
}

func Elem[T any](ptr *T, def T) T {
	if ptr != nil {
		return *ptr
	} else {
		return def
	}
}

func Ptr[T any](val T) *T {
	return &val
}

func PressKey(window *glfw.Window, key glfw.Key) bool {
	return window.GetKey(key) == glfw.Press
}

func GetAxis(window *glfw.Window, min, max glfw.Key) float32 {
	if PressKey(window, min) {
		return -1
	}
	if PressKey(window, max) {
		return 1
	}
	return 0
}

func GetDataSize(dataType uint32) int {
	switch dataType {
	case gl.FLOAT, gl.UNSIGNED_INT:
		return 4
	case gl.UNSIGNED_SHORT:
		return 2
	default:
		panic(fmt.Sprintf("unknown dataType %v", dataType))
	}
}

func Lerp(start, end, rate float32) float32 {
	return start + (end-start)*rate
}
