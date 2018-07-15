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
uniform mat4 projectionMatrix;

void main() {
	vec4 tex;
	vec3 lightDirection = worldPosition - light.position;
	lightDirection = normalize((viewMatrix * vec4(lightDirection, 0)).xyz);
	vec3 reflection = reflect(lightDirection, normalF);
	vec3 fragDirection = normalize(viewPosition) - vec3(0, 0, 0);
	bool facing = dot(normalF, lightDirection) < 0;

	tex = texture(material.ambientMap, texCoordF);
	vec3 ambient = ((1 - tex.a) * material.ambient + tex.a * tex.rgb)
	             * light.ambient;

	tex = texture(material.diffuseMap, texCoordF);
	vec3 diffuse = ((1 - tex.a) * material.diffuse + tex.a * tex.rgb)
	             * max(dot(normalF, -lightDirection), 0)
			     * light.diffuse;

	tex = texture(material.specularMap, texCoordF);
	vec3 specular = ((1 - tex.a) * material.specular + tex.a * tex.rgb)
	              * max(pow(dot(reflection, -fragDirection), material.shine), 0)
			      * light.specular
			      * (facing ? 1 : 0);

	fragColor = vec4(ambient + diffuse + specular, material.alpha);
}
