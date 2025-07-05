#version 330 core
out vec4 FragColor;

in vec3 TexCoords;

uniform samplerCube SkyBox;

const float PI=3.141592;

float RadicalInverse(uint idx) {
    idx = (idx << 16u) | (idx >> 16u);
    idx = ((idx & 0x55555555u) << 1u) | ((idx & 0xAAAAAAAAu) >> 1u);
    idx = ((idx & 0x33333333u) << 2u) | ((idx & 0xCCCCCCCCu) >> 2u);
    idx = ((idx & 0x0F0F0F0Fu) << 4u) | ((idx & 0xF0F0F0F0u) >> 4u);
    idx = ((idx & 0x00FF00FFu) << 8u) | ((idx & 0xFF00FF00u) >> 8u);
    return float(idx) * 2.3283064365386963e-10; // / 0x100000000
}

vec2 HammersleyPoint(uint idx, uint allCount){ // 随机均匀采样点函数
    return vec2(float(idx)/float(allCount), RadicalInverse(idx));
}

vec3 ImportanceSample(vec2 Xi, vec3 N, float r){
    // 转化为 local 采样点
    float r4=pow(r, 4);
    float phi=2.0*PI* Xi.x;//0->1
    float cosTheta=sqrt((1.0- Xi.y)/(1.0+(r4-1.0)* Xi.y));//D
    float sinTheta=sqrt(1.0-cosTheta*cosTheta);
    return vec3(sinTheta*cos(phi),sinTheta*sin(phi),cosTheta);
}

void main()
{
    // 计算以法线方向为 Z 的基坐标
    vec3 N=normalize(TexCoords);//Z
    vec3 Y=vec3(0.0,1.0,0.0); // Y 先随机一个用于计算 X
    vec3 X=normalize(cross(Y,N)); // X
    Y=normalize(cross(N,X)); // 纠正 Y

    vec3 res=vec3(0.0);
    float r =0.1; // 粗糙度 这里维持 0.1 即可
    vec3 V=N; // 高光采样时 视线方向与法线一致
    float w =0.0; // 权重
    for(uint i=0u;i<1600u;i++){
        vec2 Xi=HammersleyPoint(i,1600u);
        vec3 LH=ImportanceSample(Xi, N, r);
        vec3 H = LH.x*X+LH.y*Y+LH.z*N; // 转化为全局的
        vec3 L=normalize(2.0*dot(V,H)*H-V);

        float NDotL=max(dot(N,L),0.0);
        res +=texture(SkyBox,L).rgb*NDotL;
        w +=NDotL;
    }
    res=res/ w;
    FragColor =vec4(res,1.0);
}
