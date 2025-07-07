#version 330 core
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in vec2 aTexCoord;

uniform mat4 View;
uniform mat4 Model;
uniform mat4 Projection;
uniform float EdgeSize;

void main() {
    mat4 PVM = Projection * View * Model;
    vec4 pos=PVM * vec4(aPos, 1);
    // 沿法线方向偏移，为了使偏移的量在屏幕上固定宽度不随远近变化使用 w 作为参数
    vec3 offset=normalize(aNormal)*pos.w/150*EdgeSize;
    gl_Position = PVM * vec4(aPos+offset, 1);
}