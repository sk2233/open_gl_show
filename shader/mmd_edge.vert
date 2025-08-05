#version 330 core
layout (location = 0) in vec3 inPos;
layout (location = 1) in vec3 inNor;
layout (location = 2) in vec2 inUV;

uniform mat4 uModel;
uniform mat4 uView;
uniform mat4 uProj;
uniform vec2 uScreenSize;
uniform float uEdgeSize;

void main() {
    mat4 mv = uView * uModel;
    mat4 mvp = uProj * uView * uModel;
    vec3 nor = mat3(mv) * inNor;
    vec4 pos = mvp * vec4(inPos, 1.0);
    vec2 screenNor = normalize(vec2(nor));
    pos.xy += screenNor / vec2(16*10,9*10)  * uEdgeSize * pos.w; // 法线外延
    gl_Position = pos;
}
