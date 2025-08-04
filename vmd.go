package main

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"os"
)

func LoadVMD(name string) *VMD {
	file, err := os.Open(BasePath + ResPath + name)
	HandleErr(err)
	vmd, err := DecodeVMD(file)
	HandleErr(err)
	// 原坐标系不是 OpenGL 需要进行转换
	invZ := mgl32.Ident4()
	invZ[10] *= -1
	for _, frame := range vmd.BoneFrames {
		frame.Translate[2] *= -1
		mat := frame.RotateQuat.Mat4()
		frame.RotateQuat = mgl32.Mat4ToQuat(invZ.Mul4(mat).Mul4(invZ))
	}
	for _, frame := range vmd.CameraFrames {
		frame.Translate[2] *= -1
	}
	return vmd
}

type MorphCalculator struct {
	MorphMap map[int][]*MorphFrame // 每个动作都是按时间顺序来的
	MorphIdx map[int]int
}

// 假设时间是慢慢变化的
func (c *MorphCalculator) Calculate(time uint32) map[int]float32 {
	res := make(map[int]float32)
	for key, idx := range c.MorphIdx {
		frames := c.MorphMap[key]
		res[key] = c.calculate(time, frames, idx)
		if idx+1 < len(frames) && time >= frames[idx+1].Frame {
			c.MorphIdx[key]++
		}
	}
	return res
}

func (c *MorphCalculator) calculate(time uint32, frames []*MorphFrame, idx int) float32 {
	if len(frames) == 1 {
		return frames[0].Weight
	}
	if idx < 0 || idx+1 >= len(frames) { // 需要保证 idx 与 idx+1 都在范围内
		return 0
	}
	return frames[idx].Weight + (frames[idx+1].Weight-frames[idx].Weight)*float32(time-frames[idx].Frame)/float32(frames[idx+1].Frame-frames[idx].Frame)
}

func NewMorphCalculator(frames []*MorphFrame, morphs []*Morph) *MorphCalculator {
	// 只获取存在的即可
	name2Idx := make(map[string]int)
	for i, morph := range morphs {
		name2Idx[morph.Name] = i
	}
	morphMap := make(map[int][]*MorphFrame)
	for _, frame := range frames {
		if idx, ok := name2Idx[frame.Morph]; ok {
			morphMap[idx] = append(morphMap[idx], frame)
		}
	}
	morphIdx := make(map[int]int)
	for key := range morphMap {
		morphIdx[key] = -1
	}
	return &MorphCalculator{MorphMap: morphMap, MorphIdx: morphIdx}
}

type BoneCalculator struct {
	BoneMap map[int][]*BoneFrame
	BoneIdx map[int]int
}

type BonePosAndRotate struct {
	Translate mgl32.Vec3 // 移动
	Rotate    mgl32.Quat // 旋转 四元数
}

func (c *BoneCalculator) Calculate(time uint32) map[int]*BonePosAndRotate {
	res := make(map[int]*BonePosAndRotate)
	for key, idx := range c.BoneIdx {
		frames := c.BoneMap[key] // 立即用上
		if idx+1 < len(frames) && time >= frames[idx+1].Frame {
			c.BoneIdx[key]++
			idx++
		}
		res[key] = c.calculate(time, frames, idx)
	}
	return res
}

func (c *BoneCalculator) calculate(time uint32, frames []*BoneFrame, idx int) *BonePosAndRotate {
	if len(frames) == 1 { // 只有一个直接应用
		return &BonePosAndRotate{
			Translate: frames[0].Translate,
			Rotate:    frames[0].RotateQuat,
		}
	}
	if idx < 0 || idx+1 >= len(frames) { // 超出范围了就恢复原状
		return &BonePosAndRotate{
			Translate: VecZero,
			Rotate:    mgl32.QuatIdent(),
		}
	}
	start := frames[idx]
	end := frames[idx+1]
	rate := float32(time-start.Frame) / float32(end.Frame-start.Frame)
	return &BonePosAndRotate{
		Translate: mgl32.Vec3{
			Lerp(start.Translate[0], end.Translate[0], bezierVal(start.XCurve, rate)),
			Lerp(start.Translate[1], end.Translate[1], bezierVal(start.YCurve, rate)),
			Lerp(start.Translate[2], end.Translate[2], bezierVal(start.ZCurve, rate)),
		},
		Rotate: mgl32.QuatSlerp(start.RotateQuat, end.RotateQuat, bezierVal(start.RCurve, rate)),
	}
}

func evalX(curve [2]mgl32.Vec2, rate float32) float32 {
	rate2 := rate * rate
	rate3 := rate2 * rate
	invRate := 1 - rate
	invRate2 := invRate * invRate
	invRate3 := invRate2 * invRate
	return rate3*1 + 3*rate2*invRate*curve[1].X() + 3*rate*invRate2*curve[0].X() + invRate3*0
}

func findX(curve [2]mgl32.Vec2, rate float32) float32 {
	e := 0.00001
	start := float32(0.0)
	stop := float32(1.0)
	res := float32(0.5)
	x := evalX(curve, res)
	for math.Abs(float64(rate-x)) > e {
		if rate < x {
			stop = res
		} else {
			start = res
		}
		res = (stop + start) * 0.5
		x = evalX(curve, res)
	}
	return res
}

func evalY(curve [2]mgl32.Vec2, rate float32) float32 {
	rate2 := rate * rate
	rate3 := rate2 * rate
	invRate := 1 - rate
	invRate2 := invRate * invRate
	invRate3 := invRate2 * invRate
	return rate3*1 + 3*rate2*invRate*curve[1].Y() + 3*rate*invRate2*curve[0].Y() + invRate3*0
}

// curve  0~1
func bezierVal(curve [2]mgl32.Vec2, rate float32) float32 {
	temp := findX(curve, rate)
	return evalY(curve, temp)
}

func NewBoneCalculator(frames []*BoneFrame, bones []*Bone) *BoneCalculator {
	name2Idx := make(map[string]int)
	for i, bone := range bones {
		name2Idx[bone.Name] = i
	}
	boneMap := make(map[int][]*BoneFrame)
	for _, frame := range frames {
		if idx, ok := name2Idx[frame.Bone]; ok { // 存在对应骨骼才收集
			boneMap[idx] = append(boneMap[idx], frame)
		} else {
			fmt.Printf("Bone %s not found\n", frame.Bone)
		}
	}
	boneIdx := make(map[int]int)
	for name := range boneMap {
		boneIdx[name] = -1
	}
	return &BoneCalculator{BoneMap: boneMap, BoneIdx: boneIdx}
}
