#version 150

in vec4 worldCoord;
out vec4 fragColor;

void main()
{
    fragColor = vec4(worldCoord.xyz, 1.0);
}