#version 450

in vec3 position;
in vec2 texCoordV;
in vec3 normalV;
in vec3 tangentV;
in vec3 bitangentV;

out vec3 worldPosition;
out vec3 viewPosition;
out vec2 texCoordF;
out vec3 normalF;
out vec3 tanLightToVertex;
out vec3 tanCameraToVertex;
out vec3 tanLightDirection;

uniform mat4 modelMatrix;
uniform mat4 viewMatrix;
uniform mat4 projectionMatrix;

uniform mat4 shadowViewMatrix;
uniform mat4 shadowProjectionMatrix;

uniform struct Light {
	int type;
	vec3 position;
	vec3 direction;
	vec3 ambient;
	vec3 diffuse;
	vec3 specular;
	float far;
	float attenuationQuadratic;
} light;

out vec4 lightSpacePosition;

void main() {
	worldPosition = vec3(modelMatrix * vec4(position, 1));
	viewPosition = vec3(viewMatrix * vec4(worldPosition, 1));
	gl_Position = projectionMatrix * vec4(viewPosition, 1);

	texCoordF = texCoordV;

	mat4 normalMatrix = transpose(inverse(viewMatrix * modelMatrix));
	vec3 viewNormal = normalize(vec3(normalMatrix * vec4(normalV, 0)));
	vec3 viewTangent = normalize(vec3(normalMatrix * vec4(tangentV, 0)));
	vec3 viewBitangent = normalize(cross(viewNormal, viewTangent));
	mat3 tanToView = mat3(viewTangent, viewBitangent, viewNormal);
	mat3 viewToTan = transpose(tanToView); // orthonormal

	vec3 worldLightToVertex;
	if (light.type == 3) {
		worldLightToVertex = light.direction;
	} else {
		worldLightToVertex = worldPosition - light.position;
	}
	vec3 viewLightToVertex = vec3(viewMatrix * vec4(worldLightToVertex, 0));
	tanLightToVertex = viewToTan * viewLightToVertex;

	tanCameraToVertex = viewToTan * (viewPosition - vec3(0, 0, 0));

	vec3 viewLightDirection = vec3(viewMatrix * vec4(light.direction, 0));
	tanLightDirection = viewToTan * viewLightDirection;

	switch (light.type) {
	case 2: // spotlight
	case 3:
		mat4 shadowProjectionViewModelMatrix = shadowProjectionMatrix * shadowViewMatrix * modelMatrix;
		lightSpacePosition = shadowProjectionViewModelMatrix * vec4(position, 1);
	}
}
