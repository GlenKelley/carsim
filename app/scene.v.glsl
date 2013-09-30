#version 150

uniform mat4 projection;
uniform mat4 cameraview;
uniform mat4 worldview;

in vec3 position;
out vec4 worldCoord;

void main() {
    vec4 p = vec4(position, 1);
    worldCoord = worldview * p;
    gl_Position = projection * cameraview * worldCoord;
}
