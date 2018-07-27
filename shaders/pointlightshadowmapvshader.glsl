#version 450

uniform mat4 modelMatrix;
uniform mat4 viewMatrix;
uniform mat4 projectionMatrix;

in vec3 position;

out vec3 worldPosition;

void main() {
	worldPosition = vec3(modelMatrix * vec4(position, 1));
	gl_Position = projectionMatrix * viewMatrix * vec4(worldPosition, 1);
}
