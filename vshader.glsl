#version 450

in vec3 position;

in vec4 colorV;
out vec4 colorF;

uniform mat4 projectionViewModelMatrix;

void main() {
	gl_Position = projectionViewModelMatrix * vec4(position, 1.0);
	colorF = colorV;
}
