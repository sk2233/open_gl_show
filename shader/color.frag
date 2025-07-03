#version 330 core
out vec4 FragColor;

in vec3 Normal;  // 法线
in vec3 FragPos;  // 世界空间坐标
in vec2 TexCoords; // 纹理坐标

uniform sampler2D Texture;
uniform vec3 lightPos;
uniform vec3 viewPos;

void main() {
    vec3 color = texture(Texture, TexCoords).rgb;

    // ambient
    vec3 ambient = 0.3 * color;

    // diffuse
    vec3 lightDir = normalize(lightPos - FragPos);
    vec3 normal = normalize(Normal);
    float diff = max(dot(lightDir, normal), 0.0);
    vec3 diffuse = diff * color;

    // specular
    vec3 viewDir = normalize(viewPos - FragPos);
    vec3 halfwayDir = normalize(lightDir + viewDir);
    float spec = pow(max(dot(normal, halfwayDir), 0.0), 64.0);
    vec3 specular = vec3(0.3) * spec; // assuming bright white light color

    FragColor = vec4(ambient + diffuse + specular, 1.0);
//    FragColor = vec4(color, 1.0);
}