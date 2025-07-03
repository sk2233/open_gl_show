# Objects
## Geometry
Attributes[0]  UID
Vertices 3个一组的顶点数据    PolygonVertexIndex  对应三角形的下标  n(最后一个值无效)
LayerElementNormal.Normals  3个一组的法线数据    LayerElementNormal.NormalsIndex 对应的下标 n
LayerElementUV.UV   uv 数据           LayerElementUV.UVIndex  uv下标  n
## Material
Attributes[0]  UID
Properties70  材质设置
## Texture
Attributes[0]  UID
RelativeFilename  有贴图
# Connections
引用关系：<br>
0 -> Model(0根节点)<br>
Model -> Geometry,Material,Model(父子节点)<br>
Material -> Texture