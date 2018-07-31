#version 450

in vec3 position;

uniform mat4 modelMatrix;
uniform mat4 viewMatrix;
uniform mat4 projectionMatrix;

void main() {
	vec3 worldPosition = vec3(modelMatrix * vec4(position, 1));
	vec3 viewPosition = vec3(viewMatrix * vec4(worldPosition, 1));
	gl_Position = projectionMatrix * vec4(viewPosition, 1);
}
