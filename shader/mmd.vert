#version 330 core
layout (location = 0) in vec3 inPos;
layout (location = 1) in vec3 inNor;
layout (location = 2) in vec2 inUV;

out vec3 vsPos;
out vec3 vsNor;
out vec2 vsUV;

uniform mat4 uModel;
uniform mat4 uView;
uniform mat4 uProj;

void main() {
    mat4 mv= uView *uModel;
    mat4 mvp= uProj * uView *uModel;
    gl_Position = mvp * vec4(inPos, 1.0);
    vsPos = (uModel * vec4(inPos, 1.0)).xyz;
    vsNor = mat3(mvp) * inNor;
    vsUV = inUV;
}
