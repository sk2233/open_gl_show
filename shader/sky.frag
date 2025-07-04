#version 330 core
out vec4 FragColor;

in vec3 TexCoords;

uniform samplerCube SkyBox;

void main()
{
    FragColor = texture(SkyBox, normalize(TexCoords));
}