#version 450

in vec3 worldPosition;
in vec3 viewPosition;
in vec4 colorF;
in vec2 texCoordF;
in vec3 normalF;

out vec4 fragColor;

// material
uniform struct Material {
	vec3 ambient;
	vec3 diffuse;
	vec3 specular;
	float shine;
	sampler2D ambientMap;
	sampler2D diffuseMap;
	sampler2D specularMap;
	float alpha; // TODO: let textures modify alpha
} material;

uniform struct Light {
	vec3 position;
	vec3 ambient;
	vec3 diffuse;
	vec3 specular;
} light;

uniform mat4 viewMatrix;

void main() {
	vec3 ambientFactor = light.ambient;
	vec4 ambientMapRGBA = texture(material.ambientMap, texCoordF);
	vec3 ambientColor = ambientFactor * ((1 - ambientMapRGBA.a) * material.ambient + ambientMapRGBA.a * ambientMapRGBA.rgb);

	vec3 lightDirection = normalize(worldPosition - light.position);
	lightDirection = vec3(viewMatrix * vec4(lightDirection, 0));

	vec3 diffuseFactor = max(dot(normalF, -lightDirection), 0) * light.diffuse;
	vec4 diffuseMapRGBA = texture(material.diffuseMap, texCoordF);
	vec3 diffuseColor = diffuseFactor * ((1 - diffuseMapRGBA.a) * material.diffuse + diffuseMapRGBA.a * diffuseMapRGBA.rgb);

	vec3 reflection = reflect(lightDirection, normalF);
	vec3 fragDirection = normalize(viewPosition) - vec3(0, 0, 0);
	float facing = dot(normalF, lightDirection) < 0 ? 1 : 0;
	vec3 specularFactor = max(pow(dot(reflection, -fragDirection), material.shine), 0) *
	                      facing * light.specular;
	vec4 specularMapRGBA = texture(material.specularMap, texCoordF);
	vec3 specularColor = specularFactor * ((1 - specularMapRGBA.a) * material.specular + specularMapRGBA.a * specularMapRGBA.rgb);

	fragColor = vec4(ambientColor + diffuseColor + specularColor, material.alpha);
}
