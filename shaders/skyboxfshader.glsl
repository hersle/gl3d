#version 450

in vec3 positionF;

out vec4 fragColor;

uniform samplerCube cubeMap;

void main() {
	fragColor = vec4(texture(cubeMap, positionF).rgb, 1);
}
