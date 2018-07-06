#version 450

in vec4 colorF;
in vec2 texCoordF;
in vec3 normalF; // assumed to be normalized TODO: transform with world matrix
in vec3 fragWorldPosition;

out vec4 fragColor;

uniform vec3 lightPosition;
uniform vec3 ambientLight;
uniform vec3 ambient;
uniform vec3 diffuseLight;
uniform vec3 diffuse;

void main() {
	vec3 lightDirection = normalize(fragWorldPosition - lightPosition);

	vec3 ambientFactor = ambientLight;
	vec3 ambientColor = ambientFactor * ambient;

	vec3 diffuseFactor = max(dot(normalF, -lightDirection), 0) * diffuseLight;
	vec3 diffuseColor = diffuseFactor * diffuse;

	fragColor = vec4(ambientColor + diffuseColor, 1.0);
}
