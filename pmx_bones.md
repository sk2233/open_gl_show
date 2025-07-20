# Bones
Name   名称<br>
Parent   父骨骼索引<br>
TwistAxis   先对于旋转<br>
Position   相对于父骨骼的位置<br>
LocalXAxis   LocalZAxis   局部坐标轴<br>
大部分字段都要看 Flags 是否生效<br>
https://gist.github.com/felixjones/f8a06bd48f9da9a4539f#bone-flags <br>
## 如何计算某一 BoneFrame 时 Bone 的位置？
先根据两个 BoneFrame 插值得到位移与旋转<br>
先旋转后平移<br>
位移也要加上自己局部坐标下的位移<br>
BoneFrame 位移 + 局部坐标位移 ， BoneFrame 旋转  =  局部变换矩阵<br>
若有父节点还需要计算应用父节点变换矩阵<br>
## 得知 Bone 位置，如何计算顶点位置
根据顶点类型与骨骼权重进行加权<br>
![img.png](img.png)
## IK 控制
根据目标点反向调整整个骨骼，因为骨骼只能旋转，就从目标点到最远骨骼逐个在旋转限制内进行调整，
有点类似伸直胳膊的过程，可以进行多轮调整