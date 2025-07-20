#version 330 core
out vec4 FragColor;

in vec3 Normal;  // 法线
in vec3 FragPos;  // 世界空间坐标
in vec2 TexCoord; // 纹理坐标
in float LightIntensity;

uniform sampler2D BaseTex;
uniform sampler2D SpeTex; // 要看 UseSpe
uniform sampler2D ToonTex; // 要看 UseToon
uniform bool UseSpe;
uniform bool UseToon;
uniform vec3 ViewPos;
uniform vec3 LightPos;

void main() {
    vec3 baseColor =texture(BaseTex, TexCoord).rgb;
    // 应用阴影
    if(UseToon){
        baseColor=baseColor*texture(ToonTex, vec2(LightIntensity)).rgb;
    }
    // 计算高光
    vec3 viewDir = normalize(ViewPos - FragPos);
    vec3 normal =normalize(Normal);
    vec3 speColor = vec3(0);
    if(UseSpe){
        vec3 lightDir = normalize(LightPos - FragPos);
        vec3 halfwayDir = normalize(lightDir + viewDir);
        float spec = pow(max(dot(normal, halfwayDir), 0.0), 32);
        speColor = vec3(spec);
        speColor *= texture(SpeTex, TexCoord).r;
    }
    // 边缘光
    float f =  1.0 - max(dot(viewDir, normal),0);
    vec3 rimColor = vec3(smoothstep(0.6, 1, f));

    FragColor = vec4(baseColor+speColor*0.5+rimColor*0.3, 1.0f);
}
