#version 330 core
out vec4 FragColor;

in vec2 TexCoords; // 纹理坐标

uniform sampler2D Texture;

void main() {
    FragColor = texture(Texture, TexCoords);
}