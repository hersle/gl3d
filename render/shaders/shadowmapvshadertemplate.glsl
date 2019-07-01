#version 450

uniform mat4 modelMatrix;
uniform mat4 viewMatrix;
uniform mat4 projectionMatrix;

in vec3 position;

out vec3 worldPositionG;

void main() {
	worldPositionG = vec3(modelMatrix * vec4(position, 1));
	gl_Position = projectionMatrix * viewMatrix * vec4(worldPositionG, 1);
}
