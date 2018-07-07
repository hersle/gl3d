#version 450

in vec3 position;

in vec4 colorV;
out vec4 colorF;

in vec2 texCoordV;
out vec2 texCoordF;

in vec3 normalV;
out vec3 normalF;

out vec3 fragWorldPosition;
out vec3 fragViewPosition;

uniform mat4 modelMatrix;
uniform mat4 viewMatrix;
uniform mat4 projectionMatrix;

void main() {
	mat4 projectionViewModelMatrix = projectionMatrix * viewMatrix * modelMatrix;
	gl_Position = projectionViewModelMatrix * vec4(position, 1.0);
	colorF = colorV;
	texCoordF = texCoordV;
	mat3 normalMatrix = transpose(inverse(mat3(modelMatrix)));
	normalF = normalMatrix * normalV;
	fragWorldPosition = vec3(modelMatrix * vec4(position, 1.0));
	fragViewPosition = vec3(viewMatrix * vec4(fragWorldPosition, 1.0));
}
