> 格式参考：https://blog.csdn.net/qq_22642239/article/details/129752911
# Scene
# Scenes
Scenes[Scene].Nodes
# Nodes
Nodes[Children] 递归结构<br>
节点类型：<br>
空节点：children 内部取元素进行再次递归<br>
Mesh节点：按下标去 Meshs 中查询<br>
Skin节点：按下标去 Skins 中查询<br>
Camera节点：按下标去 Cameras 中查询<br>
Matrix 代表当前节点的局部变换，Rotation，Scale，Translation 组合变换<br>
# Meshes
Primitives 可以包含多个子网格<br>
Attributes：存储对应数据下标，根据下标到 Accessors 中查询<br>
Indices：存储对应数据下标，根据下标到 Accessors 中查询<br>
Material：存储下标到 Materials 中取值<br>
Mode：三角形绘制模式<br>
# Accessors
BufferView：存储下标到对应 BufferViews 中查找<br>
ByteOffset，Count：对应数据在 BufferView 下的字节偏移与字节数目<br>
ComponentType，type：最小单位数据类型与数据格式(例如 vec2)<br>
# BufferViews
Buffer：存储下标到对应的 Buffers 中查找<br>
ByteOffset,ByteLength：对应数据偏移啥的<br>
# Buffers
具体字节数据
# Materials
Xxx Texture：其下面的 Index 存储对应 Textures 的下标<br>
# Textures
Sampler：存储 Samplers 的下标<br>
Source：存储 Images 的下标<br>
# Samplers
一些贴图纹理属性
# Images
一些图片信息，URI：具体图片位置
