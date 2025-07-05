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

void main() {
    vec3 color = texture(Texture, TexCoords).rgb;
    if(UseColor){
        color=Color.rgb;
    }

    // ambient
    vec3 ambient = 0.5 * color;

    // diffuse
    vec3 lightDir = normalize(LightPos - FragPos);
    vec3 normal = normalize(Normal);
    float diff = max(dot(lightDir, normal), 0.0);
    vec3 diffuse = color*floor(diff*3)/3*0.5;

    // specular
    vec3 viewDir = normalize(ViewPos - FragPos);
    vec3 halfwayDir = normalize(lightDir + viewDir);
    float spec = pow(max(dot(normal, halfwayDir), 0.0), 64.0);
    vec3 specular = vec3(0.2) * floor(spec*2)/2; // assuming bright white light color

    // 边缘光
    float rim = 1.0 - max(dot(normal, viewDir), 0.0);
    rim = smoothstep(0.8, 1.0, rim); // 平滑边缘
    vec3 rimColor=vec3(0.5)*rim;

    FragColor = vec4(ambient + diffuse + specular+rimColor, 1.0);
}
