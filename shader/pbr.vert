#version 330 core
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in vec2 aTexCoord;
layout (location = 3) in vec4 aTangent;

out vec3 FragPos;
out vec2 TexCoord;
out vec3 Normal; // 切线空间转换矩阵

uniform mat4 View;
uniform mat4 Model;
uniform mat4 Projection;

void main() {
    FragPos = vec3(Model * vec4(aPos, 1));
    Normal=normalize(mat3(transpose(inverse(Model))) * aNormal);
    TexCoord = aTexCoord;
    gl_Position = Projection * View * Model * vec4(aPos, 1);
}