#version 450

in vec3 worldPosition;
in vec3 viewPosition;
in vec4 colorF;
in vec2 texCoordF;
in vec3 tanLightToVertex;
in vec3 tanCameraToVertex;

out vec4 fragColor;

in vec4 lightSpacePosition;

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
	bool hasBumpMap;
	sampler2D bumpMap;
	bool hasAlphaMap;
	sampler2D alphaMap;
} material;

uniform struct Light {
	vec3 position;
	vec3 direction;
	vec3 ambient;
	vec3 diffuse;
	vec3 specular;
} light;

uniform mat4 viewMatrix;

// for spotlight
uniform sampler2D spotShadowMap;

uniform samplerCube cubeShadowMap;

uniform mat4 normalMatrix;

float CalcShadowFactorSpotLight(vec4 lightSpacePos) {
	vec3 ndcCoords = lightSpacePos.xyz / lightSpacePos.w;
	vec2 texCoordS = vec2(0.5, 0.5) + 0.5 * ndcCoords.xy;
	float depth = length(worldPosition - light.position);
	float depthFront = texture(spotShadowMap, texCoordS).r * 50;
	bool inShadow = depth > depthFront + 1.0;
	if (inShadow) {
		return 0.5;
	} else {
		return 1.0;
	}
}

float CalcShadowFactorPointLight() {
	float depth = length(worldPosition - light.position);
	float depthFront = textureCube(cubeShadowMap, worldPosition - light.position).r * 50;
	bool inShadow = depth > depthFront + 1.0;
	if (inShadow) {
		return 0.5;
	} else {
		return 1.0;
	}
}

void main() {
	vec3 tanNormal = vec3(0, 0, 1);
	if (material.hasBumpMap) {
		tanNormal += -1 + 2 * normalize(texture(material.bumpMap, texCoordF).rgb);
		tanNormal = normalize(tanNormal);
	}

	vec4 tex;
	vec3 tanReflection = reflect(normalize(tanLightToVertex), normalize(tanNormal));
	bool facing = dot(normalize(tanNormal), normalize(tanLightToVertex)) < 0;

	tex = texture(material.ambientMap, texCoordF);
	vec3 ambient = ((1 - tex.a) * material.ambient + tex.a * tex.rgb)
				 * light.ambient;

	tex = texture(material.diffuseMap, texCoordF);
	vec3 diffuse = ((1 - tex.a) * material.diffuse + tex.a * tex.rgb)
				 * max(dot(normalize(tanNormal), normalize(-tanLightToVertex)), 0)
				 * light.diffuse;

	tex = texture(material.specularMap, texCoordF);
	vec3 specular = ((1 - tex.a) * material.specular + tex.a * tex.rgb)
				  * pow(max(dot(normalize(tanReflection), -normalize(tanCameraToVertex)), 0), material.shine)
				  * light.specular
				  * (facing ? 1 : 0);

	// UNCOMMENT THESE LINES FOR SPOT LIGHT
	/*
	if (dot((viewMatrix * vec4(light.direction, 0)).xyz, lightDirection) < 0.75)  {
		diffuse = vec3(0, 0, 0);
		specular = vec3(0, 0, 0);
	}
	float factor = CalcShadowFactorPointLight();
	factor /= CalcShadowFactorPointLight();
	factor *= CalcShadowFactorSpotLight(lightSpacePosition);
	*/

	// UNCOMMENT THESE LINES FOR POINT LIGHT
	float factor = CalcShadowFactorSpotLight(lightSpacePosition);
	factor /= CalcShadowFactorSpotLight(lightSpacePosition);
	factor *= CalcShadowFactorPointLight();

	float alpha;
	if (material.hasAlphaMap) {
		alpha = material.alpha * texture(material.alphaMap, texCoordF).r;
	} else {
		alpha = material.alpha;
	}

	// TODO: add proper transparency?
	if (alpha < 1) {
		discard;
	}

	fragColor = vec4(ambient + factor * (diffuse + specular), 1);
}
