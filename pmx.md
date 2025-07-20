> 参考资料：https://www.cnblogs.com/ifwz/p/17544729.html
> 参考文档：https://gist.github.com/felixjones/f8a06bd48f9da9a4539f
# Header
NumExtraUV 额外 UV 数，不用管<br>
SizeXxxIndex 对应的顶点下标占用几个 Byte<br>
# Vertices 顶点数据
Position 位置信息<br>
Normal 法线信息<br>
UV 贴图信息<br>
UVn 第n个信息，额外的UV数，最多 4 个有效数看 NumExtraUV（可以忽略）<br>
BoneMethod 骨骼权重方式<br>
Bones Weights 对应骨骼与权重，最多 4 个<br>
# Faces 顶点数据
3个索引一个面
# Textures 纹理数据
纹理图片
#  Materials 材质数据
Diffuse Specular Ambient 材质渲染颜色<br>
Flags 渲染的一些特性（DRAWEDGE是否绘制描边）<br>
EdgeColor EdgeSize 描边设置<br>
Texture 纹理图片索引<br>
SpTexture SpMode 高光贴图与混合模式<br>
ShareToon 使用内部贴图<br>
ToonTexture 阴影贴图<br>
LightIntensity = max(dot(normal, -lightDir), 0.0) 最好在顶点着色器中计算<br>
NumVerts 按面来算多少个顶点受影响<br>
# Bones 骨骼数据
# Morphs 表情数据
PositionMorphOffsets 每个表情操控索引对应的偏移<br>
其他都是类似的，是对对应元素的插值<br>
# DisplayFrames 
以节点发方式对上面元素进行引用，可以不使用<br>
# RigidBodies Joints SoftBodies 
都先暂时忽略