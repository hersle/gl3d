#version 450

in vec3 position;
in vec4 colorV;
in vec2 texCoordV;
in vec3 normalV;

out vec3 worldPosition;
out vec3 viewPosition;
out vec4 colorF;
out vec2 texCoordF;
out vec3 normalF;

uniform mat4 modelMatrix;
uniform mat4 viewMatrix;
uniform mat4 projectionMatrix;
uniform mat4 normalMatrix;

uniform mat4 shadowModelMatrix;
uniform mat4 shadowViewMatrix;
uniform mat4 shadowProjectionMatrix;

out vec4 lightSpacePosition;

void main() {
	worldPosition = vec3(modelMatrix * vec4(position, 1));
	viewPosition = vec3(viewMatrix * vec4(worldPosition, 1));
	gl_Position = projectionMatrix * vec4(viewPosition, 1);

	colorF = colorV;

	texCoordF = texCoordV;

	normalF = normalize(vec3(normalMatrix * vec4(normalV, 0)));

	mat4 shadowProjectionViewModelMatrix = shadowProjectionMatrix * shadowViewMatrix * shadowModelMatrix;
	lightSpacePosition = shadowProjectionViewModelMatrix * vec4(position, 1);
}
