#version 450

in vec3 worldPosition;
in vec3 viewPosition;
in vec4 colorF;
in vec2 texCoordF;
in vec3 tanLightToVertex;
in vec3 tanCameraToVertex;
in vec3 tanLightDirection;

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
	int type;
	vec3 position;
	vec3 direction;
	vec3 ambient;
	vec3 diffuse;
	vec3 specular;
	float far;
} light;

uniform mat4 viewMatrix;

uniform mat4 shadowViewMatrix;
uniform mat4 shadowProjectionMatrix;

// for spotlight
uniform sampler2D spotShadowMap;

uniform samplerCube cubeShadowMap;

uniform sampler2D dirShadowMap;

uniform mat4 normalMatrix;

float CalcShadowFactorSpotLight(vec4 lightSpacePos) {
	vec3 ndcCoords = lightSpacePos.xyz / lightSpacePos.w;
	vec2 texCoordS = vec2(0.5, 0.5) + 0.5 * ndcCoords.xy;
	float depth = length(worldPosition - light.position);
	float depthFront = texture(spotShadowMap, texCoordS).r * light.far;
	bool inShadow = depth > depthFront + 1.0;
	if (inShadow) {
		return 0.5;
	} else {
		return 1.0;
	}
}

float CalcShadowFactorPointLight() {
	float depth = length(worldPosition - light.position);
	float depthFront = textureCube(cubeShadowMap, worldPosition - light.position).r * light.far;
	bool inShadow = depth > depthFront + 1.0;
	if (inShadow) {
		return 0.5;
	} else {
		return 1.0;
	}
}

float CalcShadowFactorDirLight(vec4 lightSpacePos) {
	vec3 ndcCoords = lightSpacePos.xyz / lightSpacePos.w;
	vec2 texCoordS = vec2(0.5, 0.5) + 0.5 * ndcCoords.xy;
	float depth = 0.5 + 0.5 * ndcCoords.z; // make into [0, 1]
	if (texCoordS.x < 0 || texCoordS.y < 0 || texCoordS.x > 1 || texCoordS.y > 1 || depth < 0 || depth > 1) {
		return 1.0;
	}
	float depthFront = texture(dirShadowMap, texCoordS).r;
	bool inShadow = depth > depthFront + 0.05;
	if (inShadow) {
		return 0.5;
	} else {
		return 1.0;
	}
}

void main() {
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

	vec3 tanNormal = vec3(0, 0, 1);
	if (material.hasBumpMap) {
		tanNormal = -1 + 2 * normalize(texture(material.bumpMap, texCoordF).rgb);
		tanNormal = normalize(tanNormal);
	}

	vec4 tex;
	vec3 tanReflection = normalize(reflect(tanLightToVertex, tanNormal));
	bool facing = dot(tanNormal, tanLightToVertex) < 0;

	switch (light.type) {
	case 0: // ambient light
		tex = texture(material.ambientMap, texCoordF);
		vec3 ambient = ((1 - tex.a) * material.ambient + tex.a * tex.rgb)
					 * light.ambient;
		fragColor = vec4(ambient, 1);
		return;
	}

	tex = texture(material.diffuseMap, texCoordF);
	vec3 diffuse = ((1 - tex.a) * material.diffuse + tex.a * tex.rgb)
				 * max(dot(tanNormal, normalize(-tanLightToVertex)), 0)
				 * light.diffuse;

	tex = texture(material.specularMap, texCoordF);
	vec3 specular = ((1 - tex.a) * material.specular + tex.a * tex.rgb)
				  * pow(max(dot(tanReflection, -normalize(tanCameraToVertex)), 0), material.shine)
				  * light.specular
				  * (facing ? 1 : 0);

	float factor;
	switch (light.type) {
	case 1: // point light
		factor = CalcShadowFactorPointLight();
		break;
	case 2: // spot light
		if (dot(normalize(tanLightDirection), normalize(tanLightToVertex)) < 0.75)  {
			diffuse = vec3(0, 0, 0);
			specular = vec3(0, 0, 0);
		}
		factor = CalcShadowFactorSpotLight(lightSpacePosition);
		break;
	case 3: // directional light
		factor = CalcShadowFactorDirLight(lightSpacePosition);
		//factor /= CalcShadowFactorDirLight(lightSpacePosition);
		//fragColor = vec4(factor, 0, 0, 1);
		//return;
		break;
	}

	fragColor = vec4(factor * (diffuse + specular), 1);
}
