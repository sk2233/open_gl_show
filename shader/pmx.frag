#version 330 core
out vec4 FragColor;

in vec3 Normal;  // 法线
in vec3 FragPos;  // 世界空间坐标
in vec2 TexCoord; // 纹理坐标

uniform sampler2D BaseTex;
uniform sampler2D ToonTex; // 可能为 nil
uniform bool UseToon;
uniform vec3 LightPos;
uniform vec3 ViewPos;

void main() {
    vec3 objectColor=texture(BaseTex,TexCoord).rgb;

    //ambient环境光
    vec3 ambient = 0.5f * objectColor;

    //diffuse漫反射光
    vec3 norm =	normalize(Normal);
    vec3 lightDir = normalize(LightPos - FragPos); //光的方向向量是光的位置向量与片段的位置向量之间的向量差
    float diff = max(dot(norm, lightDir), 0.0f);
    vec3 diffuse = diff * objectColor*0.5;
//    if(UseToon){
//        diffuse=texture(ToonTex,vec2(diff,0.5)).r*objectColor*0.5;
//    }

    //specular镜面反射光
    vec3 viewDir = normalize(ViewPos - FragPos);
    vec3 reflectDir = reflect(-lightDir, norm);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0f), 32);
    vec3 specular = 0.2 * spec * vec3(1);

    float f =  1.0 - max(dot(viewDir, norm),0);
    float rim = smoothstep(0.6, 1, f);
    vec3 rimColor = rim * vec3(1);

    FragColor = vec4(objectColor+rimColor, 1.0f);
}
