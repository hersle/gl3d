#version 450

in vec3 worldPosition;
in vec3 viewPosition;
in vec4 colorF;
in vec2 texCoordF;
in vec3 normalF;

out vec4 fragColor;

// material
uniform vec3 ambient;
uniform vec3 diffuse;
uniform vec3 specular;
uniform float shine;

// light
uniform vec3 lightPosition;
uniform vec3 ambientLight;
uniform vec3 diffuseLight;
uniform vec3 specularLight;

uniform mat4 viewMatrix;

void main() {
	vec3 ambientFactor = ambientLight;
	vec3 ambientColor = ambientFactor * ambient;

	vec3 lightDirection = normalize(worldPosition - lightPosition);
	lightDirection = vec3(viewMatrix * vec4(lightDirection, 0));

	vec3 diffuseFactor = max(dot(normalF, -lightDirection), 0) * diffuseLight;
	vec3 diffuseColor = diffuseFactor * diffuse;

	vec3 reflection = reflect(lightDirection, normalF);
	vec3 fragDirection = normalize(viewPosition) - vec3(0, 0, 0);
	float facing = dot(normalF, lightDirection) < 0 ? 1 : 0;
	vec3 specularFactor = max(pow(dot(reflection, -fragDirection), shine), 0) *
	                      facing * specularLight;
	vec3 specularColor = specularFactor * specular;

	fragColor = vec4(ambientColor + diffuseColor + specularColor, 1.0);
}
