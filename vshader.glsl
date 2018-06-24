#version 450

in vec3 position;

in vec4 colorV;
out vec4 colorF;

uniform mat4 modelMatrix;
uniform mat4 viewProjectionMatrix;

void main() {
	gl_Position = viewProjectionMatrix * modelMatrix * vec4(position, 1.0);
	colorF = colorV;
}
