#version 450

in vec3 position;

in vec4 colorV;
out vec4 colorF;

in vec2 texCoordV;
out vec2 texCoordF;

in vec3 normalV;
out vec3 normalF;

out vec3 fragWorldPosition;

uniform mat4 modelMatrix;
uniform mat4 projectionViewMatrix;

void main() {
	mat4 projectionViewModelMatrix = projectionViewMatrix * modelMatrix;
	gl_Position = projectionViewModelMatrix * vec4(position, 1.0);
	colorF = colorV;
	texCoordF = texCoordV;
	normalF = normalV;
	fragWorldPosition = vec3(modelMatrix * vec4(position, 1.0));
}
