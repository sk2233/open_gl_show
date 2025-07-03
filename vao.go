package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

type Vao struct {
	Vao       uint32
	IndicSize int32
}

func (v *Vao) Bind() {
	gl.BindVertexArray(v.Vao)
}

func (v *Vao) Draw(mode uint32, count int32) {
	if v.IndicSize > 0 {
		panic("vao has indic")
	}
	gl.DrawArrays(mode, 0, count)
}

func (v *Vao) DrawIndic(mode uint32) {
	if v.IndicSize == 0 {
		panic("vao not has indic")
	}
	gl.DrawElements(mode, v.IndicSize, gl.UNSIGNED_INT, nil)
}

func NewVao(data []float32, sizes ...int32) *Vao {
	// 创建对象&写入数据
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, gl.Ptr(data), gl.STATIC_DRAW)
	// 设置数据
	if len(sizes) == 0 {
		panic("data format not set")
	}
	sum := int32(0)
	for _, size := range sizes {
		sum += size
	}
	curr := uintptr(0)
	for i := 0; i < len(sizes); i++ {
		gl.EnableVertexAttribArray(uint32(i))
		gl.VertexAttribPointerWithOffset(uint32(i), sizes[i], gl.FLOAT, false, sum*4, curr*4)
		curr += uintptr(sizes[i])
	}
	return &Vao{
		Vao:       vao,
		IndicSize: 0,
	}
}

func NewVaoWithIndic(data []float32, indices []uint32, sizes ...int32) *Vao {
	// 创建对象&写入数据
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	var ibo uint32
	gl.GenBuffers(1, &ibo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo) // 放在最后方便后面设置
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, gl.Ptr(data), gl.STATIC_DRAW)
	// 设置数据
	if len(sizes) == 0 {
		panic("data format not set")
	}
	sum := int32(0)
	for _, size := range sizes {
		sum += size
	}
	curr := uintptr(0)
	for i := 0; i < len(sizes); i++ {
		gl.EnableVertexAttribArray(uint32(i))
		gl.VertexAttribPointerWithOffset(uint32(i), sizes[i], gl.FLOAT, false, sum*4, curr*4)
		curr += uintptr(sizes[i])
	}
	return &Vao{
		Vao:       vao,
		IndicSize: int32(len(indices)),
	}
}
