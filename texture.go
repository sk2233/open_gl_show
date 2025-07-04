package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

type Texture struct {
	Texture   uint32
	IsCubeMap bool
}

func (t *Texture) Bind(texture uint32) {
	gl.ActiveTexture(texture)
	if t.IsCubeMap {
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, t.Texture)
	} else {
		gl.BindTexture(gl.TEXTURE_2D, t.Texture)
	}
}

var (
	imageCache = make(map[string]*image.RGBA)
)

func loadImage(name string) *image.RGBA {
	if _, ok := imageCache[name]; !ok {
		file, err := os.Open(BasePath + ResPath + name)
		HandleErr(err)
		defer file.Close()
		img, _, err := image.Decode(file)
		HandleErr(err)
		rgba := image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)
		imageCache[name] = rgba
	}
	return imageCache[name]
}

func LoadTexture(name string) *Texture {
	rgba := loadImage(name)
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(rgba.Rect.Size().X), int32(rgba.Rect.Size().Y), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
	gl.GenerateMipmap(gl.TEXTURE_2D)
	return &Texture{
		Texture:   texture,
		IsCubeMap: false,
	}
}

func LoadTextureWithSampler(name string, minFilter int32, magFilter int32, wrapS int32, wrapT int32) *Texture {
	rgba := loadImage(name)
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, minFilter)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, magFilter)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, wrapS)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, wrapT)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(rgba.Rect.Size().X), int32(rgba.Rect.Size().Y), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
	gl.GenerateMipmap(gl.TEXTURE_2D)
	return &Texture{
		Texture:   texture,
		IsCubeMap: false,
	}
}

func LoadCubeMap(names ...string) *Texture { // 必须是 6 张
	if len(names) != 6 {
		panic("cube map must 6 img")
	}
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, texture)
	// 添加图片
	for i := uint32(0); i < 6; i++ {
		rgba := loadImage(names[i])
		gl.TexImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_X+i, 0, gl.RGBA, int32(rgba.Rect.Size().X), int32(rgba.Rect.Size().Y), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
	}
	// 设置选项
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	return &Texture{
		Texture:   texture,
		IsCubeMap: true,
	}
}
