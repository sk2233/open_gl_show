#version 330 core
out vec4 FragColor;

in vec3 Normal;  // 法线
in vec3 FragPos;  // 世界空间坐标
in vec2 TexCoords; // 纹理坐标

uniform sampler2D Texture;
uniform vec3 LightPos;
uniform vec3 ViewPos;
uniform vec4 Color;
uniform bool UseColor;
uniform vec3 LightColor;

const int NumLevels = 2;

void main() {
    // 计算法线和光照方向与物体颜色
    vec3 norm = normalize(Normal);
    vec3 lightDir = normalize(LightPos - FragPos);
    vec3 color = texture(Texture, TexCoords).rgb;
    if(UseColor){
        color=Color.rgb;
    }

    // 漫反射计算
    float diff = max(dot(norm, lightDir), 0.0);

    // 阶梯式漫反射（卡通效果核心）
    float level = floor(diff * NumLevels) / NumLevels;
    vec3 diffuse = level * LightColor;

    // 边缘光效果
    vec3 viewDir = normalize(ViewPos - FragPos);
    float rim = 1.0 - max(dot(norm, viewDir), 0.0);
    rim = smoothstep(0.6, 1.0, rim);
    vec3 rimColor = vec3(1.0, 1.0, 1.0) * rim * 0.5;

    // 最终颜色
    vec3 result = (diffuse + rimColor) * color;
    FragColor = vec4(result, 1.0);
}
