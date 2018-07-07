#version 450

in vec4 colorF;
in vec2 texCoordF;
in vec3 normalF; // assumed to be normalized TODO: transform with world matrix
in vec3 fragWorldPosition;
in vec3 fragViewPosition;

out vec4 fragColor;

uniform vec3 lightPosition;
uniform vec3 ambientLight;
uniform vec3 ambient;
uniform vec3 diffuseLight;
uniform vec3 diffuse;
uniform vec3 specularLight;
uniform vec3 specular;
uniform float shininess;

void main() {
	vec3 lightDirection = normalize(fragWorldPosition - lightPosition);

	vec3 ambientFactor = ambientLight;
	vec3 ambientColor = ambientFactor * ambient;

	vec3 diffuseFactor = max(dot(normalF, -lightDirection), 0) * diffuseLight;
	vec3 diffuseColor = diffuseFactor * diffuse;

	vec3 reflection = normalize(reflect(lightDirection, normalF));
	vec3 fragDirection = normalize(fragViewPosition);
	float facing = dot(normalF, lightDirection) < 0 ? 1 : 0;
	vec3 specularFactor = max(pow(dot(reflection, -fragDirection), shininess), 0) * facing * specularLight;
	vec3 specularColor = specularFactor * specular;

	fragColor = vec4(ambientColor + diffuseColor + specularColor, 1.0);
}
