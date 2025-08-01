package main

// 用来把PMX格式的3D模型读进内存.
import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"io"
	"unicode/utf16"
)

var autoDescriptionForEncode = "encoded by test"

// PMX文件头
type Header struct {
	Magic              uint32  // "PMX " 0x20584d50
	Version            float32 // 2.0 2.1
	NumBytes           uint8   // 后续字节数, PMX2.0固定为8
	TextEncoding       uint8   // 0:UTF16 1:UTF8
	NumExtraUV         uint8   // 0 ~ 4
	SizeVertexIndex    uint8   // 1,2 或 4
	SizeTextureIndex   uint8   // 1,2 或 4
	SizeMaterialIndex  uint8   // 1,2 或 4
	SizeBoneIndex      uint8   // 1,2 或 4
	SizeMorphIndex     uint8   // 1,2 或 4
	SizeRigidBodyIndex uint8   // 1,2 或 4
}

type BoneMethod uint8

const (
	BDEF1 BoneMethod = iota
	BDEF2
	BDEF4
	SDEF
	QDEF // 数据结构和BDEF4通用
)

// 顶点
type Vertex struct {
	OldPos    mgl32.Vec3
	CurrPos   mgl32.Vec3
	PosOffset mgl32.Vec3
	Normal    [3]float32
	UV        [2]float32
	UV1       [4]float32
	UV2       [4]float32
	UV3       [4]float32
	UV4       [4]float32

	BoneMethod BoneMethod // BDEF1, BDEF2, BDEF4, SDEF
	Bones      [4]int32
	Weights    [4]float32

	SDEF_C  [3]float32
	SDEF_R0 [3]float32
	SDEF_R1 [3]float32

	EdgeFrac float32 // 材质描边倍率
}

type MaterialFlags uint8

const (
	MATERIAL_FLAG_DOUBLESIDE MaterialFlags = 1 << iota
	MATERIAL_FLAG_GROUNDSHADOW
	MATERIAL_FLAG_SELFSHADOWMAP
	MATERIAL_FLAG_SELFSHADOW
	MATERIAL_FLAG_DRAWEDGE
	MATERIAL_FLAG_VERTEXCOLOR // 2.1
	MATERIAL_FLAG_DRAWPOINT   // 2.1
	MATERIAL_FLAG_DRAWLINE    // 2.1

	MATERIAL_FLAG_MASK_2_1 = MATERIAL_FLAG_VERTEXCOLOR | MATERIAL_FLAG_DRAWPOINT | MATERIAL_FLAG_DRAWLINE
)

// 材质
type PmxMaterial struct {
	Name     string
	NameEN   string
	Diffuse  mgl32.Vec4 // RGBA
	Specular [4]float32 // RGB + 系数
	Ambient  [3]float32 // RGB

	Flags MaterialFlags

	EdgeColor mgl32.Vec4
	EdgeSize  float32

	Texture   int32
	SpTexture int32 // 环境高光纹理 (应该是2次元模型那种头发高光)
	SpMode    uint8 // 0:无效 1:乘法(sph) 2:加法(spa) 3:子纹理(UV参照追加UV1的x,y进行通常纹理绘制)

	ShareToon   uint8 // 0:独立 1:共享
	ToonTexture int32 // 共享模式的 0~9 对应 toon01~toon10, 独立模型用模型里的纹理

	Comment  string
	NumVerts int32 // 材质对应的顶点数, 一定是3的倍数. 所有材质的顶点数加起来等于模型的顶点数 也就是 Faces 数
}

type IKJoint struct {
	Bone        int32
	AngleLimit  uint8
	MinAngleXyz [3]float32 // 最小角度(弧度角), 适用于 AngleLimit=true
	MaxAngleXyz [3]float32 // 最大角度(弧度角), 适用于 AngleLimit=true
}

type IKLink struct {
	EndBone      int32     // end-effector 适用于 IK
	NumLoop      int32     // 循环次数, 适用于 IK
	MaxAngleStep float32   // IK单步角度限制(弧度角), 适用于 IK
	Joints       []IKJoint // IK链表, 适用于 IK
}

type BoneFlags uint16

const (
	BONE_FLAG_TAIL_BONE             BoneFlags = 1 << iota // 骨骼尾部连接到另一骨骼
	BONE_FLAG_ROTATION_ENABLED                            // 支持旋转
	BONE_FLAG_TRANSLATION_ENABLED                         // 支持移动
	BONE_FLAG_VISIBLE                                     // 可见
	BONE_FLAG_ENABLED                                     // 允许操作
	BONE_FLAG_INVERSE_KINEMATICS                          // 反向动力学
	BONE_FLAG_0X0040                                      //
	BONE_FLAG_0X0080                                      // 本地付与. 0=用户变形/IK/多重付与, 1=父骨骼的本地变形 ???
	BONE_FLAG_BLEND_ROTATION                              // 旋转付与. 随着付与骨旋转.
	BONE_FLAG_BLEND_TRANSLATION                           // 移动付与. 随着付与骨移动.
	BONE_FLAG_TWIST_AXIS                                  // 固定轴. 限制只能绕着特定轴旋转.
	BONE_FLAG_LOCAL_AXIS                                  // 本地XZ轴指向
	BONE_FLAG_PHYSICAL_AFTER_DEFORM                       // 先计算变形, 后计算物理
	BONE_FLAG_EXTERNAL_PARENT                             // 外部父骨骼
)

// 骨骼
type Bone struct {
	Name   string
	NameEN string

	Position   mgl32.Vec3
	Rotate     mgl32.Quat
	WorldMat   mgl32.Mat4
	LocalMat   mgl32.Mat4
	Parent     int32
	MorphLevel int32 // 应该是用来控制变形的顺序

	Flags BoneFlags

	// 骨骼尖端显示为指向另一根骨骼.  适用于 BONE_FLAG_TAIL_BONE==1
	TailBone int32

	// 骨骼尖端显示为指向相对于骨骼的自身的偏移量. BONE_FLAG_TAIL_BONE==0
	TailOffset [3]float32

	// 骨间数值调制的数值来源骨骼序号
	BlendTransformSourceBone int32

	// 骨间数值调制的比例 DST' = SRC * frac + DST * (1 - frac) ???
	BlendTransformFrac float32

	TwistAxis [3]float32 // 轴向旋转坐标轴

	LocalXAxis mgl32.Vec3 // 适用于 BONE_FLAG_LOCAL_AXIS==1
	LocalZAxis mgl32.Vec3 // 适用于 BONE_FLAG_LOCAL_AXIS==1

	ExternalParent int32 // 适用于 BONE_FLAG_EXTERNAL_PARENT=1

	IKLink IKLink // IK链 适用于 BONE_FLAG_IK=1
}

// Morph在mmd软件中的分组. 主要是便于界面操作, 对模型本身意义不大.
type MorphPanel uint8

const (
	MORPH_PANEL_0        MorphPanel = iota //
	MORPH_PANEL_1_BROW                     // 1:眉(左下)
	MORPH_PANEL_2_EYE                      // 2:目(左上)
	MORPH_PANEL_3_MOUTH                    // 3:口(右上)
	MORPH_PANEL_4_OTHERS                   // 4:其他(右下)
)

type MorphType uint8

const (
	MORPH_TYPE_PROXY MorphType = iota // MMD 里的组合变形
	MORPH_TYPE_POSITION
	MORPH_TYPE_BONE
	MORPH_TYPE_UV
	MORPH_TYPE_UV1
	MORPH_TYPE_UV2
	MORPH_TYPE_UV3
	MORPH_TYPE_UV4
	MORPH_TYPE_MATERIAL
	MORPH_TYPE_FLIP    // 2.1
	MORPH_TYPE_IMPULSE // 2.1
)

type PositionMorphOffset struct {
	Vertex int32
	Offset [3]float32
}

type UVMorphOffset struct {
	Vertex int32
	Offset [4]float32 // MORPH_TYPE_UV 只用到x和y
}

type BoneMorphOffset struct {
	Bone        int32
	Translation [3]float32
	Rotation    [4]float32 // Quaternion 四元组 (x, y, z, w)
}

type MaterialMorphOffset struct {
	Material    int32 // -1 表示全部材质
	Addition    uint8 // true用加法. false用乘法 (计算方式: 绘制值=材质值x乘法值+加法值)
	Diffuse     [4]float32
	Specular    [4]float32
	Ambient     [3]float32
	EdgeColor   [4]float32
	EdgeSize    float32
	Texture     [4]float32
	SpTexture   [4]float32
	ToonTexture [4]float32
}

type ProxyMorphOffset struct {
	Morph int32
	Frac  float32
}

type FlipMorphOffset ProxyMorphOffset

type ImpulseMorphOffset struct {
	RigidBody   int32
	Local       uint8      // 0:OFF 1:ON
	Translation [3]float32 // (x, y, z) 速度
	Rotation    [3]float32 // (x, y, z) 扭矩
}

// 变形动画
type Morph struct {
	Name   string
	NameEN string

	Panel MorphPanel // MORPH_PANEL_*
	Type  MorphType  // MORPH_TYPE_*

	// 为了便于使用, 我们为每种类型单独定义其中一个数组. 实际只有其中一个有数据, 不能混用.
	PositionMorphOffsets []PositionMorphOffset
	UVMorphOffsets       []UVMorphOffset
	BoneMorphOffsets     []BoneMorphOffset
	MaterialMorphOffsets []MaterialMorphOffset
	ProxyMorphOffsets    []ProxyMorphOffset
	FlipMorphOffsets     []FlipMorphOffset
	ImpulseMorphOffsets  []ImpulseMorphOffset
}

type DisplayFrameElem struct {
	Type  uint8 // 0:骨骼 1:变形
	Index int32 // 骨骼或变形的索引
}

// DisplayFrame 动作分组, 主要用于界面显示, 把同组的骨骼/动画放在一起.
type DisplayFrame struct {
	Name   string
	NameEN string

	SpecialFrame uint8 // 0: 普通 1: 特殊

	Elements []DisplayFrameElem
}

type RigidShape uint8

const (
	RIGID_SHAPE_SPHERE  RigidShape = iota // 球
	RIGID_SHAPE_BOX                       // 盒
	RIGID_SHAPE_CAPSULE                   // 胶囊
)

type RigidPhysical uint8

const (
	RIGID_PHYSICAL_BONE         RigidPhysical = iota // 静态绑到骨骼(仅碰撞)
	RIGID_PHYSICAL_DYNAMIC                           // 动态物理演算(重力)
	RIGID_PHYSICAL_DYNAMIC_BONE                      // 动态物理演算(重力) + 绑骨骼
)

// 刚体
type RigidBody struct {
	Name   string
	NameEN string

	Bone int32 // 关联的骨骼. -1表示不关联

	Group             uint8  // 分组
	NonCollisionGroup uint16 // PE里的"非冲突group"掩码

	Shape    RigidShape // RIGID_SHAPE_*
	Size     [3]float32 // (x,y,z) 尺寸
	Position [3]float32 // (x,y,z) 位置
	Rotation [3]float32 // (x,y,z) 旋转 (弧度角)

	Mass               float32 // 物理量: 质量
	TranslationDamping float32 // 物理量: 移动衰减 attenuation
	RotationDamping    float32 // 物理量: 旋转衰减
	Repulsion          float32 // 物理量: 排斥力
	Friction           float32 // 物理量: 摩檫力

	Physical RigidPhysical // RIGID_PHYSICAL_*

}

type JointType uint8

const (
	JOINT_TYPE_SPRING_6DOF JointType = iota // 2.0 带弹簧的 6DOF (Bullet 的 btGeneric6DofSpringConstraint)
	JOINT_TYPE_6DOF                         // 2.1 带弹簧的 6DOF，禁用弹簧常数 (Bullet 的 btGeneric6DofConstraint)
	JOINT_TYPE_P2P                          // 2.1 (Bullet 的 btPoint2PointConstraint)
	JOINT_TYPE_CONETWIST                    // 2.1 (Bullet 的 btConeTwistConstraint)
	JOINT_TYPE_SLIDER                       // 2.1 (Bullet 的 btSliderConstraint)
	JOINT_TYPE_HINGE                        // 2.1 (Bullet 的 btHingeConstraint)
)

// 刚体物理的连接点
type Joint struct {
	Name   string
	NameEN string

	Type JointType // JOINT_TYPE_*

	// 到2.1版为止, 数据格式都是一样的

	RigidBodyA int32 // 相关刚体A, -1为不相关
	RigidBodyB int32 // 相关刚体B, -1为不相关

	Position [3]float32 // (x,y,z) 位置
	Rotation [3]float32 // (x,y,z) 旋转 (弧度角)

	MinPosition [3]float32 // (x,y,z) 移动下限
	MaxPosition [3]float32 // (x,y,z) 移动上限
	MinRotation [3]float32 // (x,y,z) 旋转下限 (弧度角)
	MaxRotation [3]float32 // (x,y,z) 旋转上限 (弧度角)

	K1 [3]float32 // (x,y,z) 弹簧移动常数
	K2 [3]float32 // (x,y,z) 弹簧旋转常数
}

type SoftBodyShape uint8

const (
	SOFTBODY_SHAPE_TRIMESH SoftBodyShape = iota
	SOFTBODY_SHAPE_ROPE
)

type SoftBodyFlags uint8

const (
	SOFTBODY_FLAG_B_LINK SoftBodyFlags = iota
	SOFTBODY_FLAG_CLUSTER
	SOFTBODY_FLAG_LINK_CROSS
)

type SoftBodyAeroModel int32

const (
	SOFTBODY_AERO_MODEL_V_POINT SoftBodyAeroModel = iota
	SOFTBODY_AERO_MODEL_V_TWOSIDED
	SOFTBODY_AERO_MODEL_V_ONESIDED
	SOFTBODY_AERO_MODEL_F_TWOSIDED
	SOFTBODY_AERO_MODEL_F_ONESIDED
)

type AnchorRigidBody struct {
	RigidBody int32
	Vertex    int32
	NearMode  uint8 // 0:OFF 1:ON
}

// 软体. PMX 2.1
type SoftBody struct {
	Name              string
	NameEN            string
	Shape             SoftBodyShape
	Material          int32  // 材质
	Group             uint8  // 分组
	NonCollisionGroup uint16 // PE里的"非冲突group"掩码
	Flags             SoftBodyFlags
	BLinkDistance     int32
	NumCluster        int32
	TotalMass         float32
	CollisionMargin   float32
	AeroModel         SoftBodyAeroModel

	Config struct {
		VCF float32
		DP  float32
		DG  float32
		LF  float32
		PR  float32
		VC  float32
		DF  float32
		MT  float32
		CHR float32
		KHR float32
		SHR float32
		AHR float32
	}

	Cluster struct {
		SRHR_CL    float32
		SKHR_CL    float32
		SSHR_CL    float32
		SR_SPLT_CL float32
		SK_SPLT_CL float32
		SS_SPLT_CL float32
	}

	Iteration struct {
		V_IT int32
		P_IT int32
		D_IT int32
		C_IT int32
	}

	// 物理材料系数
	PhyMaterial struct {
		LST float32
		AST float32
		VST float32
	}

	AnchorRigidBodies []AnchorRigidBody
	PinVertices       []int32
}

// PMX模型.
// 默认左手坐标系, Y轴朝上.
type PMX struct {
	Header Header

	Name          string
	NameEN        string
	Description   string
	DescriptionEN string

	Vertices      []*Vertex
	Faces         []uint32 // 3点1面
	Textures      []string
	Materials     []PmxMaterial
	Bones         []*Bone // 骨骼
	Morphs        []*Morph
	DisplayFrames []DisplayFrame
	RigidBodies   []RigidBody
	Joints        []Joint // 连接两个刚体的关节 (注意Joint不是骨骼的关节)
	SoftBodies    []SoftBody
}

func isValidSize(n uint8) bool {
	return n == 1 || n == 2 || n == 4
}

func decodeString(r io.Reader, encoding uint8) (s string, err error) {
	var numBytes int32
	if err = binary.Read(r, binary.LittleEndian, &numBytes); err != nil {
		return
	}

	if numBytes < 0 {
		return "", errors.New("corrupted data")
	}

	if numBytes == 0 {
		return "", nil
	}

	// encoding!=0 都当作UTF8处理
	buf := make([]byte, int(numBytes))
	if _, err = io.ReadFull(r, buf); err != nil {
		return
	}

	if encoding == 0 {
		// 0 是 UTF16, 应该是偶数字节
		buf16 := make([]uint16, int(numBytes/2))
		for i := range buf16 {
			buf16[i] = uint16(buf[i*2]) | (uint16(buf[i*2+1]) << 8)
		}
		s = string(utf16.Decode(buf16))
		return
	}

	s = string(buf)
	return
}

func decodeInt(r io.Reader, size uint8) (n int32, err error) {
	switch size {
	case 1:
		var tmp int8
		if err = binary.Read(r, binary.LittleEndian, &tmp); err != nil {
			return
		}
		n = int32(tmp)
		return
	case 2:
		var tmp int16
		if err = binary.Read(r, binary.LittleEndian, &tmp); err != nil {
			return
		}
		n = int32(tmp)
		return
	case 4:
		if err = binary.Read(r, binary.LittleEndian, &n); err != nil {
			return
		}
		return
	default:
		return 0, fmt.Errorf("unsupported integer size %d, want 1,2 or 4", size)
	}
}

func decodeUint(r io.Reader, size uint8) (n uint32, err error) {
	switch size {
	case 1:
		var tmp uint8
		if err = binary.Read(r, binary.LittleEndian, &tmp); err != nil {
			return
		}
		n = uint32(tmp)
		return
	case 2:
		var tmp uint16
		if err = binary.Read(r, binary.LittleEndian, &tmp); err != nil {
			return
		}
		n = uint32(tmp)
		return
	case 4:
		if err = binary.Read(r, binary.LittleEndian, &n); err != nil {
			return
		}
		return
	default:
		return 0, fmt.Errorf("unsupported integer size %d, want 1,2 or 4", size)
	}
}

func (pm *PMX) decodeHeader(r io.Reader) (err error) {
	// 2.0和2.1版文件头尺寸固定, 整块读进来
	if err = binary.Read(r, binary.LittleEndian, &pm.Header); err != nil {
		return
	}

	// PMX格式标志
	if pm.Header.Magic != 0x20584d50 {
		return errors.New("not PMX format")
	}

	// PMX版本
	if pm.Header.Version < 2.0 || pm.Header.Version > 2.1 {
		return fmt.Errorf("unsupported PMX version %f", pm.Header.Version)
	}

	// 文件头剩下的字节数, 2.0和2.1版都是整好8个字节
	if pm.Header.NumBytes != 8 {
		return errors.New("corrupted PMX header")
	}

	// 验证8个参数的正确性
	if pm.Header.TextEncoding > 1 {
		return fmt.Errorf("unsupported text encoding: %d", pm.Header.TextEncoding)
	}
	if pm.Header.NumExtraUV > 4 {
		return fmt.Errorf("unsupported number extra UV: %d", pm.Header.NumExtraUV)
	}
	if !isValidSize(pm.Header.SizeVertexIndex) {
		return fmt.Errorf("unsupported vertex index size: %d", pm.Header.SizeVertexIndex)
	}
	if !isValidSize(pm.Header.SizeTextureIndex) {
		return fmt.Errorf("unsupported texture index size: %d", pm.Header.SizeTextureIndex)
	}
	if !isValidSize(pm.Header.SizeMaterialIndex) {
		return fmt.Errorf("unsupported material index size: %d", pm.Header.SizeMaterialIndex)
	}
	if !isValidSize(pm.Header.SizeBoneIndex) {
		return fmt.Errorf("unsupported bone index size: %d", pm.Header.SizeBoneIndex)
	}
	if !isValidSize(pm.Header.SizeMorphIndex) {
		return fmt.Errorf("unsupported morph index size: %d", pm.Header.SizeMorphIndex)
	}
	if !isValidSize(pm.Header.SizeRigidBodyIndex) {
		return fmt.Errorf("unsupported rigid body index size: %d", pm.Header.SizeRigidBodyIndex)
	}
	return
}

func (pm *PMX) decodeTextInfo(r io.Reader) (err error) {
	if pm.Name, err = decodeString(r, pm.Header.TextEncoding); err != nil {
		return
	}
	if pm.NameEN, err = decodeString(r, pm.Header.TextEncoding); err != nil {
		return
	}
	if pm.Description, err = decodeString(r, pm.Header.TextEncoding); err != nil {
		return
	}
	if pm.DescriptionEN, err = decodeString(r, pm.Header.TextEncoding); err != nil {
		return
	}
	return
}

func (pm *PMX) decodeVertices(r io.Reader) (err error) {
	var n int32
	if err = binary.Read(r, binary.LittleEndian, &n); err != nil {
		return
	}
	if n == 0 {
		return
	}
	pm.Vertices = make([]*Vertex, n)
	for i := range pm.Vertices {
		pm.Vertices[i] = &Vertex{}
		pm.Vertices[i].Bones = [4]int32{-1, -1, -1, -1}
		if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].OldPos); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].Normal); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].UV); err != nil {
			return
		}
		if pm.Header.NumExtraUV > 0 {
			if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].UV1); err != nil {
				return
			}
		}
		if pm.Header.NumExtraUV > 1 {
			if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].UV2); err != nil {
				return
			}
		}
		if pm.Header.NumExtraUV > 2 {
			if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].UV3); err != nil {
				return
			}
		}
		if pm.Header.NumExtraUV > 3 {
			if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].UV4); err != nil {
				return
			}
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].BoneMethod); err != nil {
			return
		}
		switch pm.Vertices[i].BoneMethod {
		case BDEF1:
			if pm.Vertices[i].Bones[0], err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
				return
			}
		case BDEF2:
			if pm.Vertices[i].Bones[0], err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if pm.Vertices[i].Bones[1], err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].Weights[0]); err != nil {
				return
			}
		case BDEF4, QDEF:
			if pm.Vertices[i].Bones[0], err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if pm.Vertices[i].Bones[1], err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if pm.Vertices[i].Bones[2], err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if pm.Vertices[i].Bones[3], err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].Weights[0]); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].Weights[1]); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].Weights[2]); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].Weights[3]); err != nil {
				return
			}
		case SDEF:
			if pm.Vertices[i].Bones[0], err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if pm.Vertices[i].Bones[1], err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].Weights[0]); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].SDEF_C); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].SDEF_R0); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].SDEF_R1); err != nil {
				return
			}
		default:
			return fmt.Errorf("unsupported bone deform method: %d", pm.Vertices[i].BoneMethod)
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Vertices[i].EdgeFrac); err != nil {
			return
		}
	}
	return
}

func (pm *PMX) decodeFaces(r io.Reader) (err error) {
	var n int32
	if err = binary.Read(r, binary.LittleEndian, &n); err != nil {
		return
	}
	if n == 0 {
		return
	}
	pm.Faces = make([]uint32, n) // n 是顶点数, 不是面数
	for i := range pm.Faces {
		if pm.Faces[i], err = decodeUint(r, pm.Header.SizeVertexIndex); err != nil {
			return
		}
	}
	return
}

func (pm *PMX) decodeTextures(r io.Reader) (err error) {
	var n int32
	if err = binary.Read(r, binary.LittleEndian, &n); err != nil {
		return
	}
	if n == 0 {
		return
	}
	pm.Textures = make([]string, n)
	for i := range pm.Textures {
		if pm.Textures[i], err = decodeString(r, pm.Header.TextEncoding); err != nil {
			return
		}
	}
	return
}

func (pm *PMX) decodeMaterials(r io.Reader) (err error) {
	var n int32
	if err = binary.Read(r, binary.LittleEndian, &n); err != nil {
		return
	}
	if n == 0 {
		return
	}
	pm.Materials = make([]PmxMaterial, n)
	for i := range pm.Materials {
		if pm.Materials[i].Name, err = decodeString(r, pm.Header.TextEncoding); err != nil {
			return
		}
		if pm.Materials[i].NameEN, err = decodeString(r, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Materials[i].Diffuse); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Materials[i].Specular); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Materials[i].Ambient); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Materials[i].Flags); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Materials[i].EdgeColor); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Materials[i].EdgeSize); err != nil {
			return
		}
		if pm.Materials[i].Texture, err = decodeInt(r, pm.Header.SizeTextureIndex); err != nil {
			return
		}
		if pm.Materials[i].SpTexture, err = decodeInt(r, pm.Header.SizeTextureIndex); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Materials[i].SpMode); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Materials[i].ShareToon); err != nil {
			return
		}
		if pm.Materials[i].ShareToon == 0 {
			if pm.Materials[i].ToonTexture, err = decodeInt(r, 1); err != nil {
				return
			}
		} else {
			if pm.Materials[i].ToonTexture, err = decodeInt(r, pm.Header.SizeTextureIndex); err != nil {
				return
			}
		}
		if pm.Materials[i].Comment, err = decodeString(r, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Materials[i].NumVerts); err != nil {
			return
		}
	}
	return
}

func (pm *PMX) decodeBones(r io.Reader) (err error) {
	var n int32
	if err = binary.Read(r, binary.LittleEndian, &n); err != nil {
		return
	}
	if n == 0 {
		return
	}
	pm.Bones = make([]*Bone, n)
	for i := range pm.Bones {
		pm.Bones[i] = &Bone{}
		pm.Bones[i].TailBone = -1
		pm.Bones[i].BlendTransformSourceBone = -1
		pm.Bones[i].IKLink.EndBone = -1
		pm.Bones[i].ExternalParent = -1

		if pm.Bones[i].Name, err = decodeString(r, pm.Header.TextEncoding); err != nil {
			return
		}
		if pm.Bones[i].NameEN, err = decodeString(r, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Bones[i].Position); err != nil {
			return
		}
		if pm.Bones[i].Parent, err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Bones[i].MorphLevel); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Bones[i].Flags); err != nil {
			return
		}
		if pm.Bones[i].Flags&BONE_FLAG_TAIL_BONE != 0 {
			if pm.Bones[i].TailBone, err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
				return
			}
		} else {
			if err = binary.Read(r, binary.LittleEndian, &pm.Bones[i].TailOffset); err != nil {
				return
			}
		}
		if pm.Bones[i].Flags&BONE_FLAG_BLEND_ROTATION != 0 || pm.Bones[i].Flags&BONE_FLAG_BLEND_TRANSLATION != 0 {
			if pm.Bones[i].BlendTransformSourceBone, err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &pm.Bones[i].BlendTransformFrac); err != nil {
				return
			}
		}
		if pm.Bones[i].Flags&BONE_FLAG_TWIST_AXIS != 0 {
			if err = binary.Read(r, binary.LittleEndian, &pm.Bones[i].TwistAxis); err != nil {
				return
			}
		}
		if pm.Bones[i].Flags&BONE_FLAG_LOCAL_AXIS != 0 {
			if err = binary.Read(r, binary.LittleEndian, &pm.Bones[i].LocalXAxis); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &pm.Bones[i].LocalZAxis); err != nil {
				return
			}
		}
		if pm.Bones[i].Flags&BONE_FLAG_EXTERNAL_PARENT != 0 {
			if err = binary.Read(r, binary.LittleEndian, &pm.Bones[i].ExternalParent); err != nil {
				return
			}
		}
		if pm.Bones[i].Flags&BONE_FLAG_INVERSE_KINEMATICS != 0 {
			if pm.Bones[i].IKLink.EndBone, err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &pm.Bones[i].IKLink.NumLoop); err != nil {
				return
			}
			if err = binary.Read(r, binary.LittleEndian, &pm.Bones[i].IKLink.MaxAngleStep); err != nil {
				return
			}
			var numIKJoints int32
			if err = binary.Read(r, binary.LittleEndian, &numIKJoints); err != nil {
				return
			}
			if numIKJoints > 0 {
				pm.Bones[i].IKLink.Joints = make([]IKJoint, numIKJoints)
				for j := range pm.Bones[i].IKLink.Joints {
					if pm.Bones[i].IKLink.Joints[j].Bone, err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Bones[i].IKLink.Joints[j].AngleLimit); err != nil {
						return
					}
					if pm.Bones[i].IKLink.Joints[j].AngleLimit != 0 {
						if err = binary.Read(r, binary.LittleEndian, &pm.Bones[i].IKLink.Joints[j].MinAngleXyz); err != nil {
							return
						}
						if err = binary.Read(r, binary.LittleEndian, &pm.Bones[i].IKLink.Joints[j].MaxAngleXyz); err != nil {
							return
						}
					}
				}
			}
		}
		bone := pm.Bones[i]
		bone.Rotate = mgl32.QuatIdent()
		if bone.Flags&BONE_FLAG_LOCAL_AXIS > 0 {
			xAxis := bone.LocalXAxis.Normalize()
			zAxis := bone.LocalZAxis.Normalize()
			// 计算局部Y轴（Z × X）
			yAxis := zAxis.Cross(xAxis).Normalize()
			// 重新计算Z轴确保正交性（X × Y）
			zAxis = xAxis.Cross(yAxis)
			bone.Rotate = mgl32.Mat4ToQuat(mgl32.Mat3{
				xAxis.X(), yAxis.X(), zAxis.X(),
				xAxis.Y(), yAxis.Y(), zAxis.Y(),
				xAxis.Z(), yAxis.Z(), zAxis.Z(),
			}.Mat4())
		}
	}
	return
}

func (pm *PMX) decodeMorphs(r io.Reader) (err error) {
	var n int32
	if err = binary.Read(r, binary.LittleEndian, &n); err != nil {
		return
	}
	if n == 0 {
		return
	}
	pm.Morphs = make([]*Morph, n)
	for i := range pm.Morphs {
		pm.Morphs[i] = &Morph{}
		if pm.Morphs[i].Name, err = decodeString(r, pm.Header.TextEncoding); err != nil {
			return
		}
		if pm.Morphs[i].NameEN, err = decodeString(r, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].Panel); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].Type); err != nil {
			return
		}
		var numOffsets int32
		if err = binary.Read(r, binary.LittleEndian, &numOffsets); err != nil {
			return
		}
		if numOffsets > 0 {
			switch pm.Morphs[i].Type {
			case MORPH_TYPE_PROXY:
				pm.Morphs[i].ProxyMorphOffsets = make([]ProxyMorphOffset, numOffsets)
				for j := range pm.Morphs[i].ProxyMorphOffsets {
					if pm.Morphs[i].ProxyMorphOffsets[j].Morph, err = decodeInt(r, pm.Header.SizeMorphIndex); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].ProxyMorphOffsets[j].Frac); err != nil {
						return
					}
				}
			case MORPH_TYPE_POSITION:
				pm.Morphs[i].PositionMorphOffsets = make([]PositionMorphOffset, numOffsets)
				for j := range pm.Morphs[i].PositionMorphOffsets {
					if pm.Morphs[i].PositionMorphOffsets[j].Vertex, err = decodeInt(r, pm.Header.SizeVertexIndex); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].PositionMorphOffsets[j].Offset); err != nil {
						return
					}
				}
			case MORPH_TYPE_BONE:
				pm.Morphs[i].BoneMorphOffsets = make([]BoneMorphOffset, numOffsets)
				for j := range pm.Morphs[i].BoneMorphOffsets {
					if pm.Morphs[i].BoneMorphOffsets[j].Bone, err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].BoneMorphOffsets[j].Translation); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].BoneMorphOffsets[j].Rotation); err != nil {
						return
					}
				}
			case MORPH_TYPE_UV, MORPH_TYPE_UV1, MORPH_TYPE_UV2, MORPH_TYPE_UV3, MORPH_TYPE_UV4:
				pm.Morphs[i].UVMorphOffsets = make([]UVMorphOffset, numOffsets)
				for j := range pm.Morphs[i].UVMorphOffsets {
					if pm.Morphs[i].UVMorphOffsets[j].Vertex, err = decodeInt(r, pm.Header.SizeVertexIndex); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].UVMorphOffsets[j].Offset); err != nil {
						return
					}
				}
			case MORPH_TYPE_MATERIAL:
				pm.Morphs[i].MaterialMorphOffsets = make([]MaterialMorphOffset, numOffsets)
				for j := range pm.Morphs[i].MaterialMorphOffsets {
					if pm.Morphs[i].MaterialMorphOffsets[j].Material, err = decodeInt(r, pm.Header.SizeMaterialIndex); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].Addition); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].Diffuse); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].Specular); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].Ambient); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].EdgeColor); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].EdgeSize); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].Texture); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].SpTexture); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].ToonTexture); err != nil {
						return
					}
				}
			case MORPH_TYPE_FLIP:
				pm.Morphs[i].FlipMorphOffsets = make([]FlipMorphOffset, numOffsets)
				for j := range pm.Morphs[i].FlipMorphOffsets {
					if pm.Morphs[i].FlipMorphOffsets[j].Morph, err = decodeInt(r, pm.Header.SizeMorphIndex); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].FlipMorphOffsets[j].Frac); err != nil {
						return
					}
				}
			case MORPH_TYPE_IMPULSE:
				pm.Morphs[i].ImpulseMorphOffsets = make([]ImpulseMorphOffset, numOffsets)
				for j := range pm.Morphs[i].ImpulseMorphOffsets {
					if pm.Morphs[i].ImpulseMorphOffsets[j].RigidBody, err = decodeInt(r, pm.Header.SizeRigidBodyIndex); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].ImpulseMorphOffsets[j].Local); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].ImpulseMorphOffsets[j].Translation); err != nil {
						return
					}
					if err = binary.Read(r, binary.LittleEndian, &pm.Morphs[i].ImpulseMorphOffsets[j].Rotation); err != nil {
						return
					}
				}
			default:
				return fmt.Errorf("unkown morph type %d", pm.Morphs[i].Type)
			}
		}
	}
	return
}

func (pm *PMX) decodeDisplayFrames(r io.Reader) (err error) {
	var n int32
	if err = binary.Read(r, binary.LittleEndian, &n); err != nil {
		return
	}
	if n == 0 {
		return
	}
	pm.DisplayFrames = make([]DisplayFrame, n)
	for i := range pm.DisplayFrames {
		if pm.DisplayFrames[i].Name, err = decodeString(r, pm.Header.TextEncoding); err != nil {
			return
		}
		if pm.DisplayFrames[i].NameEN, err = decodeString(r, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.DisplayFrames[i].SpecialFrame); err != nil {
			return
		}
		var numElems int32
		if err = binary.Read(r, binary.LittleEndian, &numElems); err != nil {
			return
		}
		if numElems > 0 {
			pm.DisplayFrames[i].Elements = make([]DisplayFrameElem, numElems)
			for j := range pm.DisplayFrames[i].Elements {
				if err = binary.Read(r, binary.LittleEndian, &pm.DisplayFrames[i].Elements[j].Type); err != nil {
					return
				}
				if pm.DisplayFrames[i].Elements[j].Type == 0 {
					if pm.DisplayFrames[i].Elements[j].Index, err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
						return
					}
				} else {
					if pm.DisplayFrames[i].Elements[j].Index, err = decodeInt(r, pm.Header.SizeMorphIndex); err != nil {
						return
					}
				}
			}
		}
	}
	return
}

func (pm *PMX) decodeRigidBodies(r io.Reader) (err error) {
	var n int32
	if err = binary.Read(r, binary.LittleEndian, &n); err != nil {
		return
	}
	if n == 0 {
		return
	}
	pm.RigidBodies = make([]RigidBody, n)
	for i := range pm.RigidBodies {
		pm.RigidBodies[i].Bone = -1
		if pm.RigidBodies[i].Name, err = decodeString(r, pm.Header.TextEncoding); err != nil {
			return
		}
		if pm.RigidBodies[i].NameEN, err = decodeString(r, pm.Header.TextEncoding); err != nil {
			return
		}
		if pm.RigidBodies[i].Bone, err = decodeInt(r, pm.Header.SizeBoneIndex); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.RigidBodies[i].Group); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.RigidBodies[i].NonCollisionGroup); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.RigidBodies[i].Shape); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.RigidBodies[i].Size); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.RigidBodies[i].Position); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.RigidBodies[i].Rotation); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.RigidBodies[i].Mass); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.RigidBodies[i].TranslationDamping); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.RigidBodies[i].RotationDamping); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.RigidBodies[i].Repulsion); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.RigidBodies[i].Friction); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.RigidBodies[i].Physical); err != nil {
			return
		}
	}
	return
}

func (pm *PMX) decodeJoints(r io.Reader) (err error) {
	var n int32
	if err = binary.Read(r, binary.LittleEndian, &n); err != nil {
		return
	}
	if n == 0 {
		return
	}
	pm.Joints = make([]Joint, n)
	for i := range pm.Joints {
		if pm.Joints[i].Name, err = decodeString(r, pm.Header.TextEncoding); err != nil {
			return
		}
		if pm.Joints[i].NameEN, err = decodeString(r, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Joints[i].Type); err != nil {
			return
		}

		if pm.Joints[i].RigidBodyA, err = decodeInt(r, pm.Header.SizeRigidBodyIndex); err != nil {
			return
		}
		if pm.Joints[i].RigidBodyB, err = decodeInt(r, pm.Header.SizeRigidBodyIndex); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Joints[i].Position); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Joints[i].Rotation); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Joints[i].MinPosition); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Joints[i].MaxPosition); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Joints[i].MinRotation); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Joints[i].MaxRotation); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Joints[i].K1); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.Joints[i].K2); err != nil {
			return
		}

	}
	return
}

func (pm *PMX) decodeSoftBodies(r io.Reader) (err error) {
	var n int32
	if err = binary.Read(r, binary.LittleEndian, &n); err != nil {
		return
	}
	if n == 0 {
		return
	}
	pm.SoftBodies = make([]SoftBody, n)
	for i := range pm.SoftBodies {
		if pm.SoftBodies[i].Name, err = decodeString(r, pm.Header.TextEncoding); err != nil {
			return
		}
		if pm.SoftBodies[i].NameEN, err = decodeString(r, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Shape); err != nil {
			return
		}

		if pm.SoftBodies[i].Material, err = decodeInt(r, pm.Header.SizeMaterialIndex); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Group); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].NonCollisionGroup); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Flags); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].BLinkDistance); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].NumCluster); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].TotalMass); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].CollisionMargin); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].AeroModel); err != nil {
			return
		}

		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Config.VCF); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Config.DP); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Config.DG); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Config.LF); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Config.PR); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Config.VC); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Config.DF); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Config.MT); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Config.CHR); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Config.KHR); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Config.SHR); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Config.AHR); err != nil {
			return
		}

		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Cluster.SRHR_CL); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Cluster.SKHR_CL); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Cluster.SSHR_CL); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Cluster.SR_SPLT_CL); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Cluster.SK_SPLT_CL); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Cluster.SS_SPLT_CL); err != nil {
			return
		}

		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Iteration.V_IT); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Iteration.P_IT); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Iteration.D_IT); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].Iteration.C_IT); err != nil {
			return
		}

		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].PhyMaterial.LST); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].PhyMaterial.AST); err != nil {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].PhyMaterial.VST); err != nil {
			return
		}

		var numAnchor int32
		if err = binary.Read(r, binary.LittleEndian, &numAnchor); err != nil {
			return
		}
		if numAnchor > 0 {
			pm.SoftBodies[i].AnchorRigidBodies = make([]AnchorRigidBody, numAnchor)
			for j := range pm.SoftBodies[i].AnchorRigidBodies {
				if pm.SoftBodies[i].AnchorRigidBodies[j].RigidBody, err = decodeInt(r, pm.Header.SizeRigidBodyIndex); err != nil {
					return
				}
				if pm.SoftBodies[i].AnchorRigidBodies[j].Vertex, err = decodeInt(r, pm.Header.SizeVertexIndex); err != nil {
					return
				}
				if err = binary.Read(r, binary.LittleEndian, &pm.SoftBodies[i].AnchorRigidBodies[j].NearMode); err != nil {
					return
				}
			}
		}

		var numPin int32
		if err = binary.Read(r, binary.LittleEndian, &numPin); err != nil {
			return
		}
		if numPin > 0 {
			pm.SoftBodies[i].PinVertices = make([]int32, numAnchor)
			for j := range pm.SoftBodies[i].PinVertices {
				if pm.SoftBodies[i].PinVertices[j], err = decodeInt(r, pm.Header.SizeVertexIndex); err != nil {
					return
				}
			}
		}
	}
	return
}

func (pm *PMX) ResetVertex() {
	for _, vertex := range pm.Vertices {
		vertex.PosOffset = VecZero
		vertex.CurrPos = vertex.OldPos
	}
}

func (pm *PMX) getMorph(name string) *Morph {
	for _, morph := range pm.Morphs {
		if morph.Name == name {
			return morph
		}
	}
	return nil
}

func (pm *PMX) ApplyMorph(idx int, weight float32) {
	morph := pm.Morphs[idx]
	// 先只考虑位移 这里是累加的注意 Reset 恢复
	for _, offset := range morph.PositionMorphOffsets {
		temp := pm.Vertices[offset.Vertex]
		temp.PosOffset[0] += offset.Offset[0] * weight
		temp.PosOffset[1] += offset.Offset[1] * weight
		temp.PosOffset[2] += offset.Offset[2] * weight
	}
}

func (pm *PMX) ApplyBone(posAndRotates map[int]*BonePosAndRotate) {
	childMap := make(map[int][]int)
	roots := make([]int, 0)
	for i, bone := range pm.Bones {
		if bone.Parent >= 0 {
			parent := int(bone.Parent)
			childMap[parent] = append(childMap[parent], i)
		} else {
			roots = append(roots, i) // 根节点
		}
	}
	for _, root := range roots {
		pm.dfs(root, mgl32.Ident4(), childMap, posAndRotates)
	}
	for _, vertex := range pm.Vertices {
		bones := vertex.Bones
		weights := vertex.Weights
		pos := vertex.OldPos.Add(vertex.PosOffset).Vec4(1)
		switch vertex.BoneMethod {
		case BDEF1: // 只有一个 bone 100% 权重
			bone := pm.Bones[bones[0]]
			vertex.CurrPos = bone.WorldMat.Mul4x1(bone.LocalMat.Mul4x1(pos)).Vec3()
		case BDEF2:
			weight := weights[0]
			bone1 := pm.Bones[bones[0]]
			bone2 := pm.Bones[bones[1]]
			vertex.CurrPos = bone1.WorldMat.Mul4x1(bone1.LocalMat.Mul4x1(pos)).Mul(weight).
				Add(bone2.WorldMat.Mul4x1(bone2.LocalMat.Mul4x1(pos)).Mul(1 - weight)).
				Vec3()
		case BDEF4:
			temp := mgl32.Vec4{}
			for i := 0; i < 4; i++ {
				if weights[i] == 0 {
					continue
				}
				bone := pm.Bones[bones[i]]
				temp = temp.Add(bone.WorldMat.Mul4x1(bone.LocalMat.Mul4x1(pos)).Mul(weights[i]))
			}
			vertex.CurrPos = temp.Vec3()
		default:
			panic(fmt.Sprintf("unknown boneMethod %v", vertex.BoneMethod))
		}
		//if vertex.CurrPos[0] > 100 || vertex.CurrPos[1] > 100 || vertex.CurrPos[2] > 100 {
		//	fmt.Println(vertex.CurrPos)
		//}
	}
}

func (pm *PMX) dfs(curr int, mat mgl32.Mat4, childMap map[int][]int, posAndRotates map[int]*BonePosAndRotate) {
	bone := pm.Bones[curr]
	pos := bone.Position // 初始位置
	rota := bone.Rotate
	if val, ok := posAndRotates[curr]; ok { // 先旋转，再平移
		if bone.Flags&BONE_FLAG_ROTATION_ENABLED > 0 {
			rota = rota.Mul(val.Rotate)
			//mat = val.Rotate.Mat4().Mul4(mat)
		}
		if bone.Flags&BONE_FLAG_TRANSLATION_ENABLED > 0 {
			pos = pos.Add(val.Translate) // 附加位置
			//mat = mgl32.Translate3D(val.Translate[0], val.Translate[1], val.Translate[2]).Mul4(mat)
		}
	}
	mat = mat.Mul4(mgl32.Translate3D(pos[0], pos[1], pos[2]))
	mat = mat.Mul4(rota.Mat4())

	bone.WorldMat = mat
	for _, next := range childMap[curr] {
		pm.dfs(next, mat, childMap, posAndRotates)
	}
}

// DecodePMX PMX 2.0/2.1 format from reader r
func DecodePMX(r io.Reader) (p *PMX, err error) {
	p = new(PMX)
	defer func() {
		if err != nil {
			p = nil
		}
	}()

	pm := p
	if err = pm.decodeHeader(r); err != nil {
		err = fmt.Errorf("pmx: error decoding header: %w", err)
		return
	}
	if err = pm.decodeTextInfo(r); err != nil {
		err = fmt.Errorf("pmx: error decoding text info: %w", err)
		return
	}
	if err = pm.decodeVertices(r); err != nil {
		err = fmt.Errorf("pmx: error decoding vertices: %w", err)
		return
	}
	if err = pm.decodeFaces(r); err != nil {
		err = fmt.Errorf("pmx: error decoding faces: %w", err)
		return
	}
	if err = pm.decodeTextures(r); err != nil {
		err = fmt.Errorf("pmx: error decoding textures: %w", err)
		return
	}
	if err = pm.decodeMaterials(r); err != nil {
		err = fmt.Errorf("pmx: error decoding materials: %w", err)
		return
	}
	if err = pm.decodeBones(r); err != nil {
		err = fmt.Errorf("pmx: error decoding bones: %w", err)
		return
	}
	if err = pm.decodeMorphs(r); err != nil {
		err = fmt.Errorf("pmx: error decoding morphs: %w", err)
		return
	}
	if err = pm.decodeDisplayFrames(r); err != nil {
		err = fmt.Errorf("pmx: error decoding display frames: %w", err)
		return
	}
	if err = pm.decodeRigidBodies(r); err != nil {
		err = fmt.Errorf("pmx: error decoding rigid bodies: %w", err)
		return
	}
	if err = pm.decodeJoints(r); err != nil {
		err = fmt.Errorf("pmx: error decoding joints: %w", err)
		return
	}
	if pm.Header.Version > 2.0 {
		if err = pm.decodeSoftBodies(r); err != nil {
			err = fmt.Errorf("pmx: error decoding soft bodies: %w", err)
			return
		}
	}
	return
}

type encPMX struct {
	*PMX
	Header                 Header
	haveVersion201Contents bool
}

func encodeString(w io.Writer, s string, encoding uint8) (err error) {
	var buf []byte
	if encoding == 0 {
		// 0 是 UTF16, 应该是偶数字节
		buf16 := utf16.Encode([]rune(s))
		for _, r := range buf16 {
			buf = append(buf, byte(r&0xFF), byte((r>>8)&0xFF))
		}
	} else {
		// encoding!=0 都当作UTF8处理
		buf = []byte(s)
	}

	numBytes := int32(len(buf))
	if err = binary.Write(w, binary.LittleEndian, &numBytes); err != nil {
		return
	}

	if numBytes > 0 {
		if _, err = w.Write(buf); err != nil {
			return
		}
	}
	return
}

func encodeInt(w io.Writer, n int32, size uint8) (err error) {
	switch size {
	case 1:
		tmp := int8(n)
		if err = binary.Write(w, binary.LittleEndian, &tmp); err != nil {
			return
		}
		n = int32(tmp)
		return
	case 2:
		tmp := int16(n)
		if err = binary.Write(w, binary.LittleEndian, &tmp); err != nil {
			return
		}
		n = int32(tmp)
		return
	case 4:
		if err = binary.Write(w, binary.LittleEndian, &n); err != nil {
			return
		}
		return
	default:
		return fmt.Errorf("unsupported integer size %d, want 1,2 or 4", size)
	}
}

func encodeUint(w io.Writer, n uint32, size uint8) (err error) {
	switch size {
	case 1:
		tmp := uint8(n)
		if err = binary.Write(w, binary.LittleEndian, &tmp); err != nil {
			return
		}
		n = uint32(tmp)
		return
	case 2:
		tmp := uint16(n)
		if err = binary.Write(w, binary.LittleEndian, &tmp); err != nil {
			return
		}
		n = uint32(tmp)
		return
	case 4:
		if err = binary.Write(w, binary.LittleEndian, &n); err != nil {
			return
		}
		return
	default:
		return fmt.Errorf("unsupported integer size %d, want 1,2 or 4", size)
	}
}

func lenToSizeI(n int) uint8 {
	if n > 0x00008000 {
		return 4
	}
	if n > 0x00000080 {
		return 2
	}
	return 1
}

func lenToSizeU(n int) uint8 {
	if n > 0x00010000 {
		return 4
	}
	if n > 0x00000100 {
		return 2
	}
	return 1
}

func (pm *encPMX) detectVersion201Contents() bool {
	if len(pm.SoftBodies) != 0 {
		return true
	}

	for _, m := range pm.Materials {
		if m.Flags&MATERIAL_FLAG_VERTEXCOLOR != 0 ||
			m.Flags&MATERIAL_FLAG_DRAWPOINT != 0 ||
			m.Flags&MATERIAL_FLAG_DRAWLINE != 0 {
			return true
		}
	}

	for _, m := range pm.Morphs {
		if m.Type == MORPH_TYPE_FLIP || m.Type == MORPH_TYPE_IMPULSE {
			return true
		}
	}

	for _, j := range pm.Joints {
		if j.Type > JOINT_TYPE_SPRING_6DOF {
			return true
		}
	}
	return false

}

func (pm *encPMX) prepareForEncode() (err error) {

	pm.Header = pm.PMX.Header

	// 额外的UV需要手动设置, 检查它的范围
	if pm.Header.NumExtraUV > 4 {
		return fmt.Errorf("unsupported number extra UV: %d", pm.Header.NumExtraUV)
	}

	// PMX格式标志
	pm.Header.Magic = 0x20584d50

	// PMX版本
	pm.haveVersion201Contents = pm.detectVersion201Contents()
	if pm.Header.Version == 0 {
		if pm.haveVersion201Contents {
			pm.Header.Version = 2.1
		} else {
			pm.Header.Version = 2.0
		}
	}

	// 文件头剩下的字节数, 2.0和2.1版都是整好8个字节
	pm.Header.NumBytes = 8

	// 根据数据量构造8个参数的正确性
	pm.Header.TextEncoding = 0

	// 根据数据数量自动确定各项索引的长度
	pm.Header.SizeVertexIndex = lenToSizeI(len(pm.Vertices))
	pm.Header.SizeTextureIndex = lenToSizeI(len(pm.Textures))
	pm.Header.SizeMaterialIndex = lenToSizeI(len(pm.Materials))
	pm.Header.SizeBoneIndex = lenToSizeI(len(pm.Bones))
	pm.Header.SizeMorphIndex = lenToSizeI(len(pm.Morphs))
	pm.Header.SizeRigidBodyIndex = lenToSizeI(len(pm.RigidBodies))
	return nil
}

func (pm *encPMX) encodeHeader(w io.Writer) (err error) {
	// 2.0和2.1版文件头尺寸固定, 整块写
	return binary.Write(w, binary.LittleEndian, &pm.Header)
}

func (pm *encPMX) encodeTextInfo(w io.Writer) (err error) {
	if err = encodeString(w, pm.Name, pm.Header.TextEncoding); err != nil {
		return
	}
	if err = encodeString(w, pm.NameEN, pm.Header.TextEncoding); err != nil {
		return
	}
	if pm.Description == "" && pm.DescriptionEN == "" {
		if err = encodeString(w, autoDescriptionForEncode, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = encodeString(w, autoDescriptionForEncode, pm.Header.TextEncoding); err != nil {
			return
		}
		return
	}
	if err = encodeString(w, pm.Description, pm.Header.TextEncoding); err != nil {
		return
	}
	if err = encodeString(w, pm.DescriptionEN, pm.Header.TextEncoding); err != nil {
		return
	}
	return
}

func (pm *encPMX) encodeVertices(w io.Writer) (err error) {
	n := int32(len(pm.Vertices))
	if err = binary.Write(w, binary.LittleEndian, &n); err != nil {
		return
	}
	for i := range pm.Vertices {
		if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].OldPos); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].Normal); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].UV); err != nil {
			return
		}
		if pm.Header.NumExtraUV > 0 {
			if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].UV1); err != nil {
				return
			}
		}
		if pm.Header.NumExtraUV > 1 {
			if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].UV2); err != nil {
				return
			}
		}
		if pm.Header.NumExtraUV > 2 {
			if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].UV3); err != nil {
				return
			}
		}
		if pm.Header.NumExtraUV > 3 {
			if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].UV4); err != nil {
				return
			}
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].BoneMethod); err != nil {
			return
		}
		switch pm.Vertices[i].BoneMethod {
		case BDEF1:
			if err = encodeInt(w, pm.Vertices[i].Bones[0], pm.Header.SizeBoneIndex); err != nil {
				return
			}
		case BDEF2:
			if err = encodeInt(w, pm.Vertices[i].Bones[0], pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if err = encodeInt(w, pm.Vertices[i].Bones[1], pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].Weights[0]); err != nil {
				return
			}
		case BDEF4, QDEF:
			if err = encodeInt(w, pm.Vertices[i].Bones[0], pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if err = encodeInt(w, pm.Vertices[i].Bones[1], pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if err = encodeInt(w, pm.Vertices[i].Bones[2], pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if err = encodeInt(w, pm.Vertices[i].Bones[3], pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].Weights[0]); err != nil {
				return
			}
			if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].Weights[1]); err != nil {
				return
			}
			if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].Weights[2]); err != nil {
				return
			}
			if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].Weights[3]); err != nil {
				return
			}
		case SDEF:
			if err = encodeInt(w, pm.Vertices[i].Bones[0], pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if err = encodeInt(w, pm.Vertices[i].Bones[1], pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].Weights[0]); err != nil {
				return
			}
			if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].SDEF_C); err != nil {
				return
			}
			if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].SDEF_R0); err != nil {
				return
			}
			if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].SDEF_R1); err != nil {
				return
			}
		default:
			return fmt.Errorf("unsupported bone deform method: %d", pm.Vertices[i].BoneMethod)
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Vertices[i].EdgeFrac); err != nil {
			return
		}
	}
	return
}

func (pm *encPMX) encodeFaces(w io.Writer) (err error) {
	n := int32(len(pm.Faces))
	if err = binary.Write(w, binary.LittleEndian, &n); err != nil {
		return
	}
	for i := range pm.Faces {
		if err = encodeUint(w, pm.Faces[i], pm.Header.SizeVertexIndex); err != nil {
			return
		}
	}
	return
}

func (pm *encPMX) encodeTextures(w io.Writer) (err error) {
	n := int32(len(pm.Textures))
	if err = binary.Write(w, binary.LittleEndian, &n); err != nil {
		return
	}
	for i := range pm.Textures {
		if err = encodeString(w, pm.Textures[i], pm.Header.TextEncoding); err != nil {
			return
		}
	}
	return
}

func (pm *encPMX) encodeMaterials(w io.Writer) (err error) {
	n := int32(len(pm.Materials))
	if err = binary.Write(w, binary.LittleEndian, &n); err != nil {
		return
	}
	for i := range pm.Materials {
		if err = encodeString(w, pm.Materials[i].Name, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = encodeString(w, pm.Materials[i].NameEN, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Materials[i].Diffuse); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Materials[i].Specular); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Materials[i].Ambient); err != nil {
			return
		}
		flags := pm.Materials[i].Flags
		if pm.Header.Version < 2.1 {
			flags &= ^(MATERIAL_FLAG_MASK_2_1)
		}
		if err = binary.Write(w, binary.LittleEndian, &flags); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Materials[i].EdgeColor); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Materials[i].EdgeSize); err != nil {
			return
		}
		if err = encodeInt(w, pm.Materials[i].Texture, pm.Header.SizeTextureIndex); err != nil {
			return
		}
		if err = encodeInt(w, pm.Materials[i].SpTexture, pm.Header.SizeTextureIndex); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Materials[i].SpMode); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Materials[i].ShareToon); err != nil {
			return
		}
		if pm.Materials[i].ShareToon == 0 {
			if err = encodeInt(w, pm.Materials[i].ToonTexture, 1); err != nil {
				return
			}
		} else {
			if err = encodeInt(w, pm.Materials[i].ToonTexture, pm.Header.SizeTextureIndex); err != nil {
				return
			}
		}
		if err = encodeString(w, pm.Materials[i].Comment, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Materials[i].NumVerts); err != nil {
			return
		}
	}
	return
}

func (pm *encPMX) encodeBones(w io.Writer) (err error) {
	n := int32(len(pm.Bones))
	if err = binary.Write(w, binary.LittleEndian, &n); err != nil {
		return
	}
	for i := range pm.Bones {
		if err = encodeString(w, pm.Bones[i].Name, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = encodeString(w, pm.Bones[i].NameEN, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Bones[i].Position); err != nil {
			return
		}
		if err = encodeInt(w, pm.Bones[i].Parent, pm.Header.SizeBoneIndex); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Bones[i].MorphLevel); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Bones[i].Flags); err != nil {
			return
		}
		if pm.Bones[i].Flags&BONE_FLAG_TAIL_BONE != 0 {
			if err = encodeInt(w, pm.Bones[i].TailBone, pm.Header.SizeBoneIndex); err != nil {
				return
			}
		} else {
			if err = binary.Write(w, binary.LittleEndian, &pm.Bones[i].TailOffset); err != nil {
				return
			}
		}
		if pm.Bones[i].Flags&BONE_FLAG_BLEND_ROTATION != 0 || pm.Bones[i].Flags&BONE_FLAG_BLEND_TRANSLATION != 0 {
			if err = encodeInt(w, pm.Bones[i].BlendTransformSourceBone, pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if err = binary.Write(w, binary.LittleEndian, &pm.Bones[i].BlendTransformFrac); err != nil {
				return
			}
		}
		if pm.Bones[i].Flags&BONE_FLAG_TWIST_AXIS != 0 {
			if err = binary.Write(w, binary.LittleEndian, &pm.Bones[i].TwistAxis); err != nil {
				return
			}
		}
		if pm.Bones[i].Flags&BONE_FLAG_LOCAL_AXIS != 0 {
			if err = binary.Write(w, binary.LittleEndian, &pm.Bones[i].LocalXAxis); err != nil {
				return
			}
			if err = binary.Write(w, binary.LittleEndian, &pm.Bones[i].LocalZAxis); err != nil {
				return
			}
		}
		if pm.Bones[i].Flags&BONE_FLAG_EXTERNAL_PARENT != 0 {
			if err = binary.Write(w, binary.LittleEndian, &pm.Bones[i].ExternalParent); err != nil {
				return
			}
		}
		if pm.Bones[i].Flags&BONE_FLAG_INVERSE_KINEMATICS != 0 {
			if err = encodeInt(w, pm.Bones[i].IKLink.EndBone, pm.Header.SizeBoneIndex); err != nil {
				return
			}
			if err = binary.Write(w, binary.LittleEndian, &pm.Bones[i].IKLink.NumLoop); err != nil {
				return
			}
			if err = binary.Write(w, binary.LittleEndian, &pm.Bones[i].IKLink.MaxAngleStep); err != nil {
				return
			}
			numIKJoints := int32(len(pm.Bones[i].IKLink.Joints))
			if err = binary.Write(w, binary.LittleEndian, &numIKJoints); err != nil {
				return
			}
			for j := range pm.Bones[i].IKLink.Joints {
				if err = encodeInt(w, pm.Bones[i].IKLink.Joints[j].Bone, pm.Header.SizeBoneIndex); err != nil {
					return
				}
				if err = binary.Write(w, binary.LittleEndian, &pm.Bones[i].IKLink.Joints[j].AngleLimit); err != nil {
					return
				}
				if pm.Bones[i].IKLink.Joints[j].AngleLimit != 0 {
					if err = binary.Write(w, binary.LittleEndian, &pm.Bones[i].IKLink.Joints[j].MinAngleXyz); err != nil {
						return
					}
					if err = binary.Write(w, binary.LittleEndian, &pm.Bones[i].IKLink.Joints[j].MaxAngleXyz); err != nil {
						return
					}
				}
			}
		}
	}
	return
}

func (pm *encPMX) encodeMorphs(w io.Writer) (err error) {
	n := int32(len(pm.Morphs))
	if err = binary.Write(w, binary.LittleEndian, &n); err != nil {
		return
	}
	for i := range pm.Morphs {
		if err = encodeString(w, pm.Morphs[i].Name, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = encodeString(w, pm.Morphs[i].NameEN, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].Panel); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].Type); err != nil {
			return
		}

		switch pm.Morphs[i].Type {
		case MORPH_TYPE_PROXY:
			numOffsets := int32(len(pm.Morphs[i].ProxyMorphOffsets))
			if err = binary.Write(w, binary.LittleEndian, &numOffsets); err != nil {
				return
			}
			for j := range pm.Morphs[i].ProxyMorphOffsets {
				if err = encodeInt(w, pm.Morphs[i].ProxyMorphOffsets[j].Morph, pm.Header.SizeMorphIndex); err != nil {
					return
				}
				if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].ProxyMorphOffsets[j].Frac); err != nil {
					return
				}
			}
		case MORPH_TYPE_POSITION:
			numOffsets := int32(len(pm.Morphs[i].PositionMorphOffsets))
			if err = binary.Write(w, binary.LittleEndian, &numOffsets); err != nil {
				return
			}
			for j := range pm.Morphs[i].PositionMorphOffsets {
				if err = encodeInt(w, pm.Morphs[i].PositionMorphOffsets[j].Vertex, pm.Header.SizeVertexIndex); err != nil {
					return
				}
				if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].PositionMorphOffsets[j].Offset); err != nil {
					return
				}
			}
		case MORPH_TYPE_BONE:
			numOffsets := int32(len(pm.Morphs[i].BoneMorphOffsets))
			if err = binary.Write(w, binary.LittleEndian, &numOffsets); err != nil {
				return
			}
			for j := range pm.Morphs[i].BoneMorphOffsets {
				if err = encodeInt(w, pm.Morphs[i].BoneMorphOffsets[j].Bone, pm.Header.SizeBoneIndex); err != nil {
					return
				}
				if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].BoneMorphOffsets[j].Translation); err != nil {
					return
				}
				if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].BoneMorphOffsets[j].Rotation); err != nil {
					return
				}
			}
		case MORPH_TYPE_UV, MORPH_TYPE_UV1, MORPH_TYPE_UV2, MORPH_TYPE_UV3, MORPH_TYPE_UV4:
			numOffsets := int32(len(pm.Morphs[i].UVMorphOffsets))
			if err = binary.Write(w, binary.LittleEndian, &numOffsets); err != nil {
				return
			}
			for j := range pm.Morphs[i].UVMorphOffsets {
				if err = encodeInt(w, pm.Morphs[i].UVMorphOffsets[j].Vertex, pm.Header.SizeVertexIndex); err != nil {
					return
				}
				if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].UVMorphOffsets[j].Offset); err != nil {
					return
				}
			}
		case MORPH_TYPE_MATERIAL:
			numOffsets := int32(len(pm.Morphs[i].MaterialMorphOffsets))
			if err = binary.Write(w, binary.LittleEndian, &numOffsets); err != nil {
				return
			}
			for j := range pm.Morphs[i].MaterialMorphOffsets {
				if err = encodeInt(w, pm.Morphs[i].MaterialMorphOffsets[j].Material, pm.Header.SizeMaterialIndex); err != nil {
					return
				}
				if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].Addition); err != nil {
					return
				}
				if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].Diffuse); err != nil {
					return
				}
				if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].Specular); err != nil {
					return
				}
				if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].Ambient); err != nil {
					return
				}
				if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].EdgeColor); err != nil {
					return
				}
				if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].EdgeSize); err != nil {
					return
				}
				if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].Texture); err != nil {
					return
				}
				if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].SpTexture); err != nil {
					return
				}
				if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].MaterialMorphOffsets[j].ToonTexture); err != nil {
					return
				}
			}
		case MORPH_TYPE_FLIP:
			if pm.Header.Version >= 2.1 {
				numOffsets := int32(len(pm.Morphs[i].FlipMorphOffsets))
				if err = binary.Write(w, binary.LittleEndian, &numOffsets); err != nil {
					return
				}
				for j := range pm.Morphs[i].FlipMorphOffsets {
					if err = encodeInt(w, pm.Morphs[i].FlipMorphOffsets[j].Morph, pm.Header.SizeMorphIndex); err != nil {
						return
					}
					if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].FlipMorphOffsets[j].Frac); err != nil {
						return
					}
				}
			} else {
				if err = binary.Write(w, binary.LittleEndian, int32(0)); err != nil {
					return
				}
			}
		case MORPH_TYPE_IMPULSE:
			if pm.Header.Version >= 2.1 {
				numOffsets := int32(len(pm.Morphs[i].ImpulseMorphOffsets))
				if err = binary.Write(w, binary.LittleEndian, &numOffsets); err != nil {
					return
				}
				for j := range pm.Morphs[i].ImpulseMorphOffsets {
					if err = encodeInt(w, pm.Morphs[i].ImpulseMorphOffsets[j].RigidBody, pm.Header.SizeRigidBodyIndex); err != nil {
						return
					}
					if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].ImpulseMorphOffsets[j].Local); err != nil {
						return
					}
					if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].ImpulseMorphOffsets[j].Translation); err != nil {
						return
					}
					if err = binary.Write(w, binary.LittleEndian, &pm.Morphs[i].ImpulseMorphOffsets[j].Rotation); err != nil {
						return
					}
				}
			} else {
				if err = binary.Write(w, binary.LittleEndian, int32(0)); err != nil {
					return
				}
			}
		default:
			return fmt.Errorf("unkown morph type %d", pm.Morphs[i].Type)
		}

	}
	return
}

func (pm *encPMX) encodeDisplayFrames(w io.Writer) (err error) {
	n := int32(len(pm.DisplayFrames))
	if err = binary.Write(w, binary.LittleEndian, &n); err != nil {
		return
	}
	for i := range pm.DisplayFrames {
		if err = encodeString(w, pm.DisplayFrames[i].Name, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = encodeString(w, pm.DisplayFrames[i].NameEN, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.DisplayFrames[i].SpecialFrame); err != nil {
			return
		}
		numElems := int32(len(pm.DisplayFrames[i].Elements))
		if err = binary.Write(w, binary.LittleEndian, &numElems); err != nil {
			return
		}
		for j := range pm.DisplayFrames[i].Elements {
			if err = binary.Write(w, binary.LittleEndian, &pm.DisplayFrames[i].Elements[j].Type); err != nil {
				return
			}
			if pm.DisplayFrames[i].Elements[j].Type == 0 {
				if err = encodeInt(w, pm.DisplayFrames[i].Elements[j].Index, pm.Header.SizeBoneIndex); err != nil {
					return
				}
			} else {
				if err = encodeInt(w, pm.DisplayFrames[i].Elements[j].Index, pm.Header.SizeMorphIndex); err != nil {
					return
				}
			}
		}
	}
	return
}

func (pm *encPMX) encodeRigidBodies(w io.Writer) (err error) {
	n := int32(len(pm.RigidBodies))
	if err = binary.Write(w, binary.LittleEndian, &n); err != nil {
		return
	}
	for i := range pm.RigidBodies {
		if err = encodeString(w, pm.RigidBodies[i].Name, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = encodeString(w, pm.RigidBodies[i].NameEN, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = encodeInt(w, pm.RigidBodies[i].Bone, pm.Header.SizeBoneIndex); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.RigidBodies[i].Group); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.RigidBodies[i].NonCollisionGroup); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.RigidBodies[i].Shape); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.RigidBodies[i].Size); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.RigidBodies[i].Position); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.RigidBodies[i].Rotation); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.RigidBodies[i].Mass); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.RigidBodies[i].TranslationDamping); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.RigidBodies[i].RotationDamping); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.RigidBodies[i].Repulsion); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.RigidBodies[i].Friction); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.RigidBodies[i].Physical); err != nil {
			return
		}
	}
	return
}

func (pm *encPMX) encodeJoints(w io.Writer) (err error) {
	n := int32(len(pm.Joints))
	if err = binary.Write(w, binary.LittleEndian, &n); err != nil {
		return
	}
	for i := range pm.Joints {
		if err = encodeString(w, pm.Joints[i].Name, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = encodeString(w, pm.Joints[i].NameEN, pm.Header.TextEncoding); err != nil {
			return
		}
		jointType := pm.Joints[i].Type
		if pm.Header.Version < 2.1 && jointType > JOINT_TYPE_SPRING_6DOF {
			jointType = JOINT_TYPE_SPRING_6DOF // 无论如何处理都会有问题, 这里直接改为2.0支持的格式
		}
		if err = binary.Write(w, binary.LittleEndian, &jointType); err != nil {
			return
		}
		if err = encodeInt(w, pm.Joints[i].RigidBodyA, pm.Header.SizeRigidBodyIndex); err != nil {
			return
		}
		if err = encodeInt(w, pm.Joints[i].RigidBodyB, pm.Header.SizeRigidBodyIndex); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Joints[i].Position); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Joints[i].Rotation); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Joints[i].MinPosition); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Joints[i].MaxPosition); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Joints[i].MinRotation); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Joints[i].MaxRotation); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Joints[i].K1); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.Joints[i].K2); err != nil {
			return
		}

	}
	return
}

func (pm *encPMX) encodeSoftBodies(w io.Writer) (err error) {
	n := int32(len(pm.SoftBodies))
	if err = binary.Write(w, binary.LittleEndian, &n); err != nil {
		return
	}
	for i := range pm.SoftBodies {
		if err = encodeString(w, pm.SoftBodies[i].Name, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = encodeString(w, pm.SoftBodies[i].NameEN, pm.Header.TextEncoding); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Shape); err != nil {
			return
		}

		if err = encodeInt(w, pm.SoftBodies[i].Material, pm.Header.SizeMaterialIndex); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Group); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].NonCollisionGroup); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Flags); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].BLinkDistance); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].NumCluster); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].TotalMass); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].CollisionMargin); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].AeroModel); err != nil {
			return
		}

		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Config.VCF); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Config.DP); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Config.DG); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Config.LF); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Config.PR); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Config.VC); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Config.DF); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Config.MT); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Config.CHR); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Config.KHR); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Config.SHR); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Config.AHR); err != nil {
			return
		}

		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Cluster.SRHR_CL); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Cluster.SKHR_CL); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Cluster.SSHR_CL); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Cluster.SR_SPLT_CL); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Cluster.SK_SPLT_CL); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Cluster.SS_SPLT_CL); err != nil {
			return
		}

		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Iteration.V_IT); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Iteration.P_IT); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Iteration.D_IT); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].Iteration.C_IT); err != nil {
			return
		}

		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].PhyMaterial.LST); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].PhyMaterial.AST); err != nil {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].PhyMaterial.VST); err != nil {
			return
		}

		numAnchor := int32(len(pm.SoftBodies[i].AnchorRigidBodies))
		if err = binary.Write(w, binary.LittleEndian, &numAnchor); err != nil {
			return
		}
		for j := range pm.SoftBodies[i].AnchorRigidBodies {
			if err = encodeInt(w, pm.SoftBodies[i].AnchorRigidBodies[j].RigidBody, pm.Header.SizeRigidBodyIndex); err != nil {
				return
			}
			if err = encodeInt(w, pm.SoftBodies[i].AnchorRigidBodies[j].Vertex, pm.Header.SizeVertexIndex); err != nil {
				return
			}
			if err = binary.Write(w, binary.LittleEndian, &pm.SoftBodies[i].AnchorRigidBodies[j].NearMode); err != nil {
				return
			}
		}

		numPin := int32(len(pm.SoftBodies[i].PinVertices))
		if err = binary.Write(w, binary.LittleEndian, &numPin); err != nil {
			return
		}
		for j := range pm.SoftBodies[i].PinVertices {
			if err = encodeInt(w, pm.SoftBodies[i].PinVertices[j], pm.Header.SizeVertexIndex); err != nil {
				return
			}
		}
	}
	return
}

// Encode as PMX 2.0/2.1 format into writer w
func Encode(w io.Writer, p *PMX) (err error) {
	pm := &encPMX{PMX: p}
	if err = pm.prepareForEncode(); err != nil {
		err = fmt.Errorf("pmx: error preparing for encode: %w", err)
		return
	}
	if err = pm.encodeHeader(w); err != nil {
		err = fmt.Errorf("pmx: error encoding header: %w", err)
		return
	}
	if err = pm.encodeTextInfo(w); err != nil {
		err = fmt.Errorf("pmx: error encoding text info: %w", err)
		return
	}
	if err = pm.encodeVertices(w); err != nil {
		err = fmt.Errorf("pmx: error encoding vertices: %w", err)
		return
	}
	if err = pm.encodeFaces(w); err != nil {
		err = fmt.Errorf("pmx: error encoding faces: %w", err)
		return
	}
	if err = pm.encodeTextures(w); err != nil {
		err = fmt.Errorf("pmx: error encoding textures: %w", err)
		return
	}
	if err = pm.encodeMaterials(w); err != nil {
		err = fmt.Errorf("pmx: error encoding materials: %w", err)
		return
	}
	if err = pm.encodeBones(w); err != nil {
		err = fmt.Errorf("pmx: error encoding bones: %w", err)
		return
	}
	if err = pm.encodeMorphs(w); err != nil {
		err = fmt.Errorf("pmx: error encoding morphs: %w", err)
		return
	}
	if err = pm.encodeDisplayFrames(w); err != nil {
		err = fmt.Errorf("pmx: error encoding display frames: %w", err)
		return
	}
	if err = pm.encodeRigidBodies(w); err != nil {
		err = fmt.Errorf("pmx: error encoding rigid bodies: %w", err)
		return
	}
	if err = pm.encodeJoints(w); err != nil {
		err = fmt.Errorf("pmx: error encoding joints: %w", err)
		return
	}
	if pm.Header.Version >= 2.1 {
		if err = pm.encodeSoftBodies(w); err != nil {
			err = fmt.Errorf("pmx: error encoding soft bodies: %w", err)
			return
		}
	}
	return
}
