#version 330 core
layout (location = 0) in vec3 aPos;

out vec3 TexCoords;

uniform mat4 Projection; // model 无需变化
uniform mat4 View;
uniform mat4 Model;

void main()
{
    TexCoords = vec3(Model* vec4(aPos, 1.0)); // 使用世界坐标
    gl_Position = Projection*View*Model* vec4(aPos, 1.0);
}