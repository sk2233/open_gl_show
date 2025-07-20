# BoneFrames 骨骼动画
Bone 骨骼名称<br>
Time 动作所在的帧<br>
Translate 位移动画<br>
RotateQuat 旋转四元数<br>
XCurve YCurve ZCurve x,y,z 位移曲线<br>
RCurve 旋转曲线<br>
曲线数据为 16 byte 每 4 个读取第一个就行，得到的是下图中的两个控制点的位置(每个控制点范围 0~127)<br>
![img_14.png](img_14.png)<br>
加上默认的点 (0,0) (127,127) 组成 三阶贝塞尔曲线
![img_15.png](img_15.png)
# MorphFrames 表情帧
Morph 变形名称<br>
Time 变形帧<br>
Weight 变形比例(线性插值)<br>
# CameraFrames 相机帧
Time 时间帧<br>
Translate 相机位置<br>
RotateXYZ 相机旋转<br>
Distance ViewAngle Ortho 相机设置<br>
Curve 变化曲线 只看前 4 个即可<br>
# LightFrames 光源帧
Time 时间帧<br>
Color 光源颜色<br>
Direction 光源方向<br>
线性插值