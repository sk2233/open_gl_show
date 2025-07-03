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
	Texture uint32
}

func (t *Texture) Bind(texture uint32) {
	gl.ActiveTexture(texture)
	gl.BindTexture(gl.TEXTURE_2D, t.Texture)
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
		Texture: texture,
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
		Texture: texture,
	}
}
