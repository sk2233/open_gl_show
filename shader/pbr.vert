#version 330 core
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in vec2 aTexCoord;
layout (location = 3) in vec4 aTangent;

out vec3 FragPos;
out vec2 TexCoord;
out mat3 TBN; // 切线空间转换矩阵

uniform mat4 View;
uniform mat4 Model;
uniform mat4 Projection;

void main() {
    FragPos = vec3(Model * vec4(aPos, 1));
    // 使用切线空间的三个基向量组成基线空间矩阵
    vec3 t=normalize(vec3(Model*vec4(aTangent.xyz,0.0)));//world space
    vec3 n=normalize(mat3(transpose(inverse(Model))) * aNormal);
    vec3 b=cross(n,t);
    TBN=mat3(t,b,n);
    TexCoord = aTexCoord;
    gl_Position = Projection * View * Model * vec4(aPos, 1);
}