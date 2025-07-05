package main

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"os"
	"strings"
)

type Shader struct {
	Program    uint32
	UniformMap map[string]int32
}

func (s *Shader) Use() {
	gl.UseProgram(s.Program)
}

func (s *Shader) getUniformLoc(name string) int32 {
	if _, ok := s.UniformMap[name]; !ok {
		s.UniformMap[name] = gl.GetUniformLocation(s.Program, gl.Str(name+"\x00")) // c 字符串需要这个结束标识
	}
	return s.UniformMap[name]
}

func (s *Shader) SetMat4(name string, mat4 mgl32.Mat4) {
	uniformLoc := s.getUniformLoc(name)
	gl.UniformMatrix4fv(uniformLoc, 1, false, &mat4[0])
}

func (s *Shader) SetF4(name string, val mgl32.Vec4) {
	uniformLoc := s.getUniformLoc(name)
	gl.Uniform4f(uniformLoc, val[0], val[1], val[2], val[3])
}

func (s *Shader) SetF3(name string, val mgl32.Vec3) {
	uniformLoc := s.getUniformLoc(name)
	gl.Uniform3f(uniformLoc, val[0], val[1], val[2])
}

func (s *Shader) SetF1(name string, val float32) {
	uniformLoc := s.getUniformLoc(name)
	gl.Uniform1f(uniformLoc, val)
}

func (s *Shader) SetI1(name string, val int32) {
	uniformLoc := s.getUniformLoc(name)
	gl.Uniform1i(uniformLoc, val)
}

func loadShader(path string, shaderType uint32) uint32 {
	bs, err := os.ReadFile(path)
	HandleErr(err)
	shader := gl.CreateShader(shaderType)
	cStr, free := gl.Strs(string(bs) + "\x00") // c 字符串需要这个结束标识
	gl.ShaderSource(shader, 1, cStr, nil)
	free()
	gl.CompileShader(shader)
	// 校验错误
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		panic(fmt.Sprintf("loadShader path %s shaderType %v err %s", path, shaderType, log))
	}
	return shader
}

func LoadShader(name string) *Shader {
	// 顶点着色器与片源着色器一定是要有的
	vertShader := loadShader(BasePath+ShaderPath+name+VertName, gl.VERTEX_SHADER)
	fragShader := loadShader(BasePath+ShaderPath+name+FragName, gl.FRAGMENT_SHADER)
	// 链接着色器
	program := gl.CreateProgram()
	gl.AttachShader(program, vertShader)
	gl.AttachShader(program, fragShader)
	gl.LinkProgram(program)
	gl.DeleteShader(vertShader)
	gl.DeleteShader(fragShader)
	// 错误检查
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		panic(fmt.Sprintf("LoadShader name %s err %s", name, log))
	}
	return &Shader{
		Program:    program,
		UniformMap: make(map[string]int32),
	}
}
