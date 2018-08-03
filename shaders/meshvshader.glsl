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
out mat3 viewToTan;
out vec3 viewLightToVertex;
out vec3 tanCameraToVertex;
out vec3 viewLightDirection;

uniform mat4 modelMatrix;
uniform mat4 viewMatrix;
uniform mat4 projectionMatrix;
uniform mat4 normalMatrix;

uniform mat4 lightViewMatrix;
uniform mat4 lightProjectionMatrix;

const int MAX_LIGHTS = 10;
uniform vec3 lightPositions[MAX_LIGHTS];
uniform vec3 lightDirections[MAX_LIGHTS];

void main() {
	worldPosition = vec3(modelMatrix * vec4(position, 1));
	viewPosition = vec3(viewMatrix * vec4(worldPosition, 1));
	gl_Position = projectionMatrix * vec4(viewPosition, 1);

	texCoordF = texCoordV;

	vec3 viewNormal = normalize(vec3(normalMatrix * vec4(normalV, 0)));
	vec3 viewTangent = normalize(vec3(normalMatrix * vec4(tangentV, 0)));
	vec3 viewBitangent = normalize(cross(viewNormal, viewTangent));
	mat3 tanToView = mat3(viewTangent, viewBitangent, viewNormal);
	viewToTan = transpose(tanToView); // orthonormal

	vec3 worldLightToVertex = worldPosition - lightPositions[0];
	viewLightToVertex = vec3(viewMatrix * vec4(worldLightToVertex, 0));

	tanCameraToVertex = viewToTan * (viewPosition - vec3(0, 0, 0));

	viewLightDirection = vec3(viewMatrix * vec4(lightDirections[0], 0));
}
