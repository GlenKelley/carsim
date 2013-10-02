#version 150

in vec4 worldCoord;
in vec4 textureCoord;
out vec4 fragColor;

void main()
{
    vec4 t = textureCoord;
    vec3 c = vec3(1,1,1) * clamp(20*sin(4*atan(t.y,t.z))-19,0,1)+vec3(0.2) * length(t.yz) / (t.x*t.x*10+1);
    fragColor = vec4(c, t.w);
}