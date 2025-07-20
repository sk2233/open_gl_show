package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"os"
)

func LoadVMD(name string) *VMD {
	file, err := os.Open(BasePath + ResPath + name)
	HandleErr(err)
	vmd, err := DecodeVMD(file)
	HandleErr(err)
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
		if idx+1 < len(frames) && time >= frames[idx+1].Time {
			c.MorphIdx[key]++
		}
	}
	return res
}

func (c *MorphCalculator) calculate(time uint32, frames []*MorphFrame, idx int) float32 {
	if idx < 0 || idx+1 >= len(frames) { // 需要保证 idx 与 idx+1 都在范围内
		return 0
	}
	return frames[idx].Weight + (frames[idx+1].Weight-frames[idx].Weight)*float32(time-frames[idx].Time)/float32(frames[idx+1].Time-frames[idx].Time)
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
	for key, temp := range morphMap {
		if len(temp) == 1 { // 只有一个的不需要动画
			delete(morphMap, key)
		} else {
			morphIdx[key] = -1
		}
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
		frames := c.BoneMap[key]
		res[key] = c.calculate(time, frames, idx)
		if idx+1 < len(frames) && time >= frames[idx+1].Time {
			c.BoneIdx[key]++
		}
	}
	return res
}

func (c *BoneCalculator) calculate(time uint32, frames []*BoneFrame, idx int) *BonePosAndRotate {
	if idx < 0 || idx+1 >= len(frames) { // 超出范围了就恢复原状
		return &BonePosAndRotate{
			Translate: VecZero,
			Rotate:    mgl32.QuatIdent(),
		}
	}
	start := frames[idx]
	end := frames[idx+1]
	rate := float32(time-start.Time) / float32(end.Time-start.Time)
	return &BonePosAndRotate{
		Translate: mgl32.Vec3{
			Lerp(start.Translate[0], end.Translate[0], bezierVal(start.XCurve, rate)),
			Lerp(start.Translate[1], end.Translate[1], bezierVal(start.YCurve, rate)),
			Lerp(start.Translate[2], end.Translate[2], bezierVal(start.ZCurve, rate)),
		},
		Rotate: mgl32.QuatSlerp(start.RotateQuat, end.RotateQuat, bezierVal(start.RCurve, rate)),
	}
}

// curve  0~1
func bezierVal(curve [2]mgl32.Vec2, rate float32) float32 {
	return rate // 先按线性的来
	res := mgl32.CubicBezierCurve2D(rate, mgl32.Vec2{}, curve[0], curve[1], mgl32.Vec2{1, 1})
	return res[1] // TODO 这里的 rate 是不对的 后面可能需要调整
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
		}
	}
	boneIdx := make(map[int]int)
	for name, temp := range boneMap {
		if len(temp) == 1 { // 只有一个的实际就是没有动画
			delete(boneMap, name)
		} else {
			boneIdx[name] = -1
		}
	}
	return &BoneCalculator{BoneMap: boneMap, BoneIdx: boneIdx}
}
