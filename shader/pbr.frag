#version 330 core
out vec4 FragColor;

in mat3 TBN;  // 法线
in vec3 FragPos;  // 世界空间坐标
in vec2 TexCoord; // 纹理坐标

uniform sampler2D BaseTex;
uniform samplerCube DiffuseTex; // 预先生成的漫反射与高光 cube_map
uniform samplerCube SpecularTex;
uniform sampler2D MetallicRoughnessTex;
uniform sampler2D EmissiveTex;
uniform sampler2D OcclusionTex;
uniform sampler2D NormalTex;
uniform vec3 LightPos;
uniform vec3 ViewPos;

vec3 Fresnel(vec3 F,float HDotV){ // 高亮比例，看物体反射的光 F 是物质特性，对光的反射率
    return F+(vec3(1.0)-F)*pow(1.0-HDotV,5.0);
}

vec3 FresnelRoughness(vec3 F,float NDotV,float r){
    return F+(max(vec3(1.0-r),F)-F)*pow(1.0-NDotV,5.0);
}

const float PI=3.1415926;

float NDF(vec3 N,vec3 H,float r){ // 根据微表面模型对法线方向进行调整
    float NDotH=max(dot(N,H),0.0);
    float r4=pow(r,4.0); // r 是粗糙度
    return r4/(PI*pow(pow(NDotH,2.0)*(r4-1.0)+1.0,2.0));
}

float Geometry(float NDotV,float r){ // 计算物体对光的反射度 类似 OA
    float k=pow(r+1.0,2.0)/8.0; // r 是粗糙度
    return NDotV/(NDotV*(1.0-k)+k);
}

void main() {
    // 提前预计算
    vec4 normalTS=texture(NormalTex,TexCoord); // 获取的是切线空间下的法线需要使用 TBN 转换到世界坐标系
    normalTS=normalize(normalTS*2.0-vec4(1.0));
    vec3 normal =normalize(TBN*vec3(normalTS)); // 转换到正常空间下的法线
    vec3 N = normalize(normal); // 法线
    vec3 L = normalize(LightPos - FragPos); // 光照方向
    vec3 V = normalize(ViewPos - FragPos); // 视线方向
    vec3 H = normalize(L + V); // 光照+视线的半程向量
    vec3 R=normalize(2.0*dot(V,N)*N-V); // 反射光线
    vec3 lightColor = vec3(1.0, 1.0, 1.0);
    float lightIntensity = 4.0;
    float NDotL = max(dot(N, L), 0.0);
    float HDotV = max(dot(H, V), 0.0);
    float NDotV = max(dot(N, V), 0.0);
    vec3 color = vec3(0.0);
    vec3 F = vec3(0.04); // 物质参数
    vec3 albedo = texture(BaseTex, TexCoord).rgb;
    vec2 mr= texture(MetallicRoughnessTex,TexCoord).bg;
    float metallic =mr.x;
    float roughness =mr.y;
    F = mix(F, albedo, metallic); // 金属度越大原来的颜色占比越小
    // 直接光照
    vec3 Ks = Fresnel(F, HDotV);
    float D = NDF(N, H, roughness);
    float G = Geometry(NDotV, roughness);
    vec3 Kd = vec3(1.0) - Ks;
    Kd *= (1.0 - metallic); // 漫反射受金属度的影响
    vec3 diffuse = Kd * albedo / PI;
    vec3 specular = (D * Ks * G) / (4.0 * NDotV * NDotL + 0.0001); // 防止除0
    color += (diffuse + specular) * lightColor * lightIntensity * NDotL;
    // 间接光照  可以使用 AO 对间接光照进行加权
    float ao=texture(OcclusionTex,TexCoord).r;
    Ks=FresnelRoughness(F, NDotV, roughness);  // 间接光照材质需要考虑粗糙度
    Kd=vec3(1.0)-Ks;
    Kd*=1.0- metallic;
    vec3 diffuseLight=texture(DiffuseTex, N).rgb;
    color+=Kd*diffuseLight*albedo*ao; // 环境光漫反射
    vec3 specularLight =texture(SpecularTex, R).rgb;
    color+= specularLight *F*ao;
    // 输出最终结果
    vec3 emissive = texture(EmissiveTex,TexCoord).rgb;
    FragColor = vec4(color+emissive, 1.0); // emissive 贴图直接叠加到最后即可无需额外计算
}