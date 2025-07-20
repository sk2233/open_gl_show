package main

import (
	"os"
	"testing"
)

func TestVMD(t *testing.T) {
	f, err := os.Open("/Users/sky/Documents/go/open_gl_show/res/你的笑容.vmd")
	HandleErr(err)
	p, err := DecodeVMD(f)
	HandleErr(err)
	if p.VMDHeader.Version < 1 || p.VMDHeader.Version > 2 {
		panic("err Version")
	}
}
