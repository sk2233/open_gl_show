#version 330 core
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in vec2 aTexCoords;

uniform mat4 View;
uniform mat4 Model;
uniform mat4 Projection;

void main() {
    vec4 pos=Projection * View * Model * vec4(aPos, 1);
    // 沿法线方向偏移，为了使偏移的量在屏幕上固定宽度不随远近变化使用 w 作为参数
    vec3 offset=normalize(aNormal)*pos.w/800;
    gl_Position = Projection * View * Model * vec4(aPos+offset, 1);
}