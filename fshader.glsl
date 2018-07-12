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
uniform sampler2D ambientMap;
uniform sampler2D diffuseMap;
uniform sampler2D specularMap;

// light
uniform vec3 lightPosition;
uniform vec3 ambientLight;
uniform vec3 diffuseLight;
uniform vec3 specularLight;

uniform mat4 viewMatrix;

void main() {
	vec3 ambientFactor = ambientLight;
	vec4 ambientMapRGBA = texture(ambientMap, texCoordF);
	vec3 ambientColor = ambientFactor * ((1 - ambientMapRGBA.a) * ambient + ambientMapRGBA.a * ambientMapRGBA.rgb);

	vec3 lightDirection = normalize(worldPosition - lightPosition);
	lightDirection = vec3(viewMatrix * vec4(lightDirection, 0));

	vec3 diffuseFactor = max(dot(normalF, -lightDirection), 0) * diffuseLight;
	vec4 diffuseMapRGBA = texture(diffuseMap, texCoordF);
	vec3 diffuseColor = diffuseFactor * ((1 - diffuseMapRGBA.a) * diffuse + diffuseMapRGBA.a * diffuseMapRGBA.rgb);

	vec3 reflection = reflect(lightDirection, normalF);
	vec3 fragDirection = normalize(viewPosition) - vec3(0, 0, 0);
	float facing = dot(normalF, lightDirection) < 0 ? 1 : 0;
	vec3 specularFactor = max(pow(dot(reflection, -fragDirection), shine), 0) *
	                      facing * specularLight;
	vec4 specularMapRGBA = texture(specularMap, texCoordF);
	vec3 specularColor = specularFactor * ((1 - specularMapRGBA.a) * specular + specularMapRGBA.a * specularMapRGBA.rgb);

	fragColor = vec4(ambientColor + diffuseColor + specularColor, 1.0);
}
