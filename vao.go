package main

import (
	"bufio"
	"github.com/go-gl/gl/v4.1-core/gl"
	"os"
	"strconv"
	"strings"
)

type Vao struct {
	Vao        uint32
	IndicSize  int32
	PointCount int32
}

func (v *Vao) Bind() {
	gl.BindVertexArray(v.Vao)
}

func (v *Vao) Draw(mode uint32) {
	if v.PointCount <= 0 {
		panic("vao not has point")
	}
	gl.DrawArrays(mode, 0, v.PointCount)
}

func (v *Vao) DrawIndic(mode uint32) {
	if v.IndicSize <= 0 {
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
		Vao:        vao,
		IndicSize:  0,
		PointCount: int32(len(data)) / sum,
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
		Vao:        vao,
		IndicSize:  int32(len(indices)),
		PointCount: 0,
	}
}

func LoadObj(path string) *Vao {
	// 打开文件
	file, err := os.Open(BasePath + ResPath + path)
	HandleErr(err)
	defer file.Close()
	// 逐行处理
	vs := make([]float32, 0)
	vns := make([]float32, 0)
	vts := make([]float32, 0)
	data := make([]float32, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "v ") {
			vs = append(vs, parseFloat(line[2:])...)
		} else if strings.HasPrefix(line, "vn ") {
			vns = append(vns, parseFloat(line[3:])...)
		} else if strings.HasPrefix(line, "vt ") {
			vts = append(vts, parseFloat(line[3:])...)
		} else if strings.HasPrefix(line, "f ") {
			points := strings.Split(line[2:], " ")
			for _, point := range points {
				items := strings.Split(point, "/")
				vIdx, err := strconv.ParseInt(items[0], 10, 64)
				HandleErr(err)
				tIdx, err := strconv.ParseInt(items[1], 10, 64)
				HandleErr(err)
				nIdx, err := strconv.ParseInt(items[2], 10, 64)
				HandleErr(err)
				data = append(data, vs[(vIdx-1)*3:vIdx*3]...)
				data = append(data, vns[(nIdx-1)*3:nIdx*3]...)
				data = append(data, vts[(tIdx-1)*2:tIdx*2]...)
			}
		}
	}
	return NewVao(data, 3, 3, 2)
}

func parseFloat(line string) []float32 {
	items := strings.Split(line, " ")
	res := make([]float32, 0)
	for _, item := range items {
		val, err := strconv.ParseFloat(strings.TrimSpace(item), 64)
		HandleErr(err)
		res = append(res, float32(val))
	}
	return res
}
