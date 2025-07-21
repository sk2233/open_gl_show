#version 330 core
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in vec2 aTexCoord;

uniform mat4 View;
uniform mat4 Model;
uniform mat4 Projection;
uniform float EdgeSize;

void main() {
    vec4 pos=Projection * View * Model * vec4(aPos, 1);
    vec3 nor = mat3(View*Model)*aNormal;
    vec2 screenNor = normalize(vec2(nor));
    pos.xy += screenNor / vec2(16*10,9*10) * EdgeSize * pos.w; // 法线外延
    gl_Position = pos;
}