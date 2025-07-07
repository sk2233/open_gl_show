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
    vec3 lightColor=vec3(1);
    vec3 objectColor=texture(BaseTex,TexCoord).rgb;

    //ambient环境光
    vec3 ambient = 0.2f * lightColor;

    //diffuse漫反射光
    vec3 norm =	normalize(Normal);
    vec3 lightDir = normalize(LightPos - FragPos); //光的方向向量是光的位置向量与片段的位置向量之间的向量差
    float diff = max(dot(norm, lightDir), 0.0f);
    vec3 diffuse = diff * lightColor;
    if(UseToon){
        diffuse=texture(ToonTex,vec2(1-diff)).rgb;
    }

    //specular镜面反射光
    vec3 viewDir = normalize(ViewPos - FragPos);
    vec3 reflectDir = reflect(-lightDir, norm);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0f), 32);
    vec3 specular = 0.5 * spec * lightColor;

    vec3 result = (ambient + diffuse+specular) * objectColor;
    FragColor = vec4(result, 1.0f);
}
