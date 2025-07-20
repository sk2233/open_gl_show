package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"io"
	"sort"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

var shiftJisDecoder = japanese.ShiftJIS.NewDecoder()

type VMDHeader struct {
	Version int    // 1 or 2
	Model   string // 模型里应该是Shift_JIS编码, 我们转为utf-8
}

type BoneFrame struct {
	Bone       string        // 骨骼名字
	Time       uint32        // 帧序号
	Translate  mgl32.Vec3    // 移动
	RotateQuat mgl32.Quat    // 旋转 四元数
	XCurve     [2]mgl32.Vec2 // X 曲线 控制点  0~1
	YCurve     [2]mgl32.Vec2 // Y 曲线
	ZCurve     [2]mgl32.Vec2 // Z 曲线
	RCurve     [2]mgl32.Vec2 // 旋转曲线
}

type MorphFrame struct {
	Morph  string  // morph动画名字
	Time   uint32  // 帧序号
	Weight float32 // 权重

}

type CameraFrame struct {
	Time      uint32     // 帧序号
	Distance  float32    // 距离
	Translate [3]float32 // 移动
	RotateXyz [3]float32 // 旋转xyz
	Curve     [24]byte
	ViewAngle float32
	Ortho     byte
}

type LightFrame struct {
	Time      uint32 // 帧序号
	Color     [3]float32
	Direction [3]float32
}

type VMD struct {
	VMDHeader

	BoneFrames   []*BoneFrame
	MorphFrames  []*MorphFrame
	CameraFrames []*CameraFrame
	LightFrames  []*LightFrame
}

func decodeString2(r io.Reader, n int) (s string, err error) {
	if n == 0 {
		return
	}
	buf := make([]byte, n)
	if _, err = io.ReadFull(r, buf); err != nil {
		return
	}
	if i := bytes.IndexByte(buf, 0); i != -1 {
		buf = buf[:i]
	}
	src := string(buf)
	if s, _, err = transform.String(shiftJisDecoder, src); err != nil {
		s = src
		err = nil
	}
	return
}

func (vm *VMD) decodeHeader(r io.Reader) (err error) {
	// 30 bytes magic and version
	magicStr, err := decodeString2(r, 30)
	if err != nil {
		return
	}
	magicStr = magicStr[:25]
	if strings.HasPrefix(magicStr, "Vocaloid Motion Data") {
		if magicStr == "Vocaloid Motion Data file" {
			vm.Version = 1
		} else if magicStr == "Vocaloid Motion Data 0002" {
			vm.Version = 2
		} else {
			err = errors.New("unsupported vmd version")
			return
		}
	} else {
		err = errors.New("not a vmd format file")
		return
	}
	vm.Model, err = decodeString2(r, vm.Version*10)
	return
}

func (vm *VMD) decodeBoneFrames(r io.Reader) (err error) {
	var numFrames uint32
	if err = binary.Read(r, binary.LittleEndian, &numFrames); err != nil {
		return
	}
	if numFrames > 0 {
		vm.BoneFrames = make([]*BoneFrame, numFrames)
		for i := range vm.BoneFrames {
			vm.BoneFrames[i] = &BoneFrame{}
			if vm.BoneFrames[i].Bone, err = decodeString2(r, 15); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &vm.BoneFrames[i].Time); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &vm.BoneFrames[i].Translate); err != nil {
				return
			}
			temp1 := [4]float32{}
			if err = binary.Read(r, binary.LittleEndian, &temp1); err != nil {
				return
			}
			vm.BoneFrames[i].RotateQuat = mgl32.Quat{
				W: temp1[3],
				V: mgl32.Vec3{temp1[0], temp1[1], temp1[2]},
			}
			temp2 := [16]byte{}
			if err = binary.Read(r, binary.LittleEndian, &temp2); err != nil {
				return
			}
			vm.BoneFrames[i].XCurve = toCPoint(temp2)
			if err = binary.Read(r, binary.LittleEndian, &temp2); err != nil {
				return
			}
			vm.BoneFrames[i].YCurve = toCPoint(temp2)
			if err = binary.Read(r, binary.LittleEndian, &temp2); err != nil {
				return
			}
			vm.BoneFrames[i].ZCurve = toCPoint(temp2)
			if err = binary.Read(r, binary.LittleEndian, &temp2); err != nil {
				return
			}
			vm.BoneFrames[i].RCurve = toCPoint(temp2)
		}
	}
	sort.Slice(vm.BoneFrames, func(i, j int) bool {
		return vm.BoneFrames[i].Time < vm.BoneFrames[j].Time
	})
	return
}

func toCPoint(temp2 [16]byte) [2]mgl32.Vec2 {
	return [2]mgl32.Vec2{{float32(temp2[0]) / 127, float32(temp2[4]) / 127}, {float32(temp2[8]) / 127, float32(temp2[12]) / 127}}
}

func (vm *VMD) decodeMorphFrames(r io.Reader) (err error) {
	var numFrames uint32
	if err = binary.Read(r, binary.LittleEndian, &numFrames); err != nil {
		return
	}
	if numFrames > 0 {
		vm.MorphFrames = make([]*MorphFrame, numFrames)
		for i := range vm.MorphFrames {
			vm.MorphFrames[i] = &MorphFrame{}
			if vm.MorphFrames[i].Morph, err = decodeString2(r, 15); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &vm.MorphFrames[i].Time); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &vm.MorphFrames[i].Weight); err != nil {
				return
			}
		}
	}
	sort.Slice(vm.MorphFrames, func(i, j int) bool {
		return vm.MorphFrames[i].Time < vm.MorphFrames[i].Time
	})
	return
}

func (vm *VMD) decodeCameraFrames(r io.Reader) (err error) {
	var numFrames uint32
	if err = binary.Read(r, binary.LittleEndian, &numFrames); err != nil {
		return
	}
	if numFrames > 0 {
		vm.CameraFrames = make([]*CameraFrame, numFrames)
		for i := range vm.CameraFrames {
			vm.CameraFrames[i] = &CameraFrame{}
			if err = binary.Read(r, binary.LittleEndian, &vm.CameraFrames[i].Time); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &vm.CameraFrames[i].Distance); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &vm.CameraFrames[i].Translate); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &vm.CameraFrames[i].RotateXyz); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &vm.CameraFrames[i].Curve); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &vm.CameraFrames[i].ViewAngle); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &vm.CameraFrames[i].Ortho); err != nil {
				return
			}
		}
	}
	sort.Slice(vm.CameraFrames, func(i, j int) bool {
		return vm.CameraFrames[i].Time < vm.CameraFrames[j].Time
	})
	return
}

func (vm *VMD) decodeLightFrames(r io.Reader) (err error) {
	var numFrames uint32
	if err = binary.Read(r, binary.LittleEndian, &numFrames); err != nil {
		return
	}
	if numFrames > 0 {
		vm.LightFrames = make([]*LightFrame, numFrames)
		for i := range vm.LightFrames {
			vm.LightFrames[i] = &LightFrame{}
			if err = binary.Read(r, binary.LittleEndian, &vm.LightFrames[i].Time); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &vm.LightFrames[i].Color); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &vm.LightFrames[i].Direction); err != nil {
				return
			}
		}
	}
	sort.Slice(vm.LightFrames, func(i, j int) bool {
		return vm.LightFrames[i].Time < vm.LightFrames[j].Time
	})
	return
}

func DecodeVMD(r io.Reader) (vm *VMD, err error) {
	vm = new(VMD)

	defer func() {
		if err != nil {
			vm = nil
		}
	}()

	if err = vm.decodeHeader(r); err != nil {
		err = fmt.Errorf("vmd: error decoding header: %w", err)
		return
	}
	if err = vm.decodeBoneFrames(r); err != nil {
		err = fmt.Errorf("vmd: error decoding bone frames: %w", err)
		return
	}
	if err = vm.decodeMorphFrames(r); err != nil {
		err = fmt.Errorf("vmd: error decoding morph frames: %w", err)
		return
	}
	if err = vm.decodeCameraFrames(r); err != nil {
		err = fmt.Errorf("vmd: error decoding camera frames: %w", err)
		return
	}
	if err = vm.decodeLightFrames(r); err != nil {
		err = fmt.Errorf("vmd: error decoding light frames: %w", err)
		return
	}

	return
}
