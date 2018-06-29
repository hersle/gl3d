#version 450

in vec3 position;

in vec4 colorV;
out vec4 colorF;

in vec2 texCoordV;
out vec2 texCoordF;

uniform mat4 projectionViewModelMatrix;

void main() {
	gl_Position = projectionViewModelMatrix * vec4(position, 1.0);
	colorF = colorV;
	texCoordF = texCoordV;
}
