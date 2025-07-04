package main

import "github.com/go-gl/gl/v4.1-core/gl"

type Frame struct {
	FrameBuff  uint32
	Texture    *Texture
	RenderBuff uint32
}

func (f *Frame) Use() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, f.FrameBuff)
}

func CreateFrame(width, height int32) *Frame {
	var frameBuff uint32
	gl.GenFramebuffers(1, &frameBuff)
	gl.BindFramebuffer(gl.FRAMEBUFFER, frameBuff)
	// 设置纹理信息
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, width, height, 0, gl.RGBA, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE) // we clamp to the edge as the blur filter would otherwise sample repeated texture values!
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0) // 将它附加到当前绑定的帧缓冲对象
	// 深度信息 与模版信息 附件
	var renderBuff uint32
	gl.GenRenderbuffers(1, &renderBuff)
	gl.BindRenderbuffer(gl.RENDERBUFFER, renderBuff)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH24_STENCIL8, width, height)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.RENDERBUFFER, renderBuff)
	// 还原绘制目标
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("CreateFrame err")
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	return &Frame{
		FrameBuff:  frameBuff,
		Texture:    &Texture{Texture: texture},
		RenderBuff: renderBuff,
	}
}
