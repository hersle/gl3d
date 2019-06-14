#version 450

in vec2 texCoordF;

out vec4 fragColor;

#if defined(POINT)
in vec3 worldPosition;
in vec3 tanLightToVertex;
in vec3 tanCameraToVertex;
#endif

#if defined(SPOT)
in vec3 worldPosition;
in vec3 tanLightToVertex;
in vec3 tanCameraToVertex;
in vec3 tanLightDirection;
in vec4 lightSpacePosition;
#endif

#if defined(DIR)
in vec3 tanLightToVertex;
in vec3 tanCameraToVertex;
in vec4 lightSpacePosition;
#endif

uniform struct Material {
	float alpha; // TODO: let textures modify alpha
	sampler2D alphaMap;

	#if defined(AMBIENT)
	vec3 ambient;
	sampler2D ambientMap;
	#endif

	#if defined(POINT) || defined(SPOT) || defined(DIR)
	vec3 diffuse;
	vec3 specular;
	float shine;
	sampler2D diffuseMap;
	sampler2D specularMap;
	sampler2D bumpMap;
	#endif
} material;

uniform struct Light {
	#if defined(AMBIENT)
	vec3 ambient;
	#endif

	#if defined(POINT)
	vec3 position;
	vec3 diffuse;
	vec3 specular;
	float far;
	float attenuationQuadratic;
	#endif

	#if defined(SPOT)
	vec3 position;
	vec3 direction;
	vec3 diffuse;
	vec3 specular;
	float far;
	float attenuationQuadratic;
	#endif

	#if defined(DIR)
	vec3 direction;
	vec3 diffuse;
	vec3 specular;
	float attenuationQuadratic;
	#endif
} light;

#if defined(SHADOW) && defined(POINT)
uniform samplerCube cubeShadowMap;
#endif

#if defined(SHADOW) && defined(SPOT)
uniform sampler2D spotShadowMap;
#endif

#if defined(SHADOW) && defined(DIR)
uniform sampler2D dirShadowMap;
#endif

void main() {
	float alpha = material.alpha * texture(material.alphaMap, texCoordF).r;
	// TODO: add proper transparency?
	if (alpha < 1) {
		discard;
	}

	#if defined(AMBIENT)
	vec4 tex;
	tex = texture(material.ambientMap, texCoordF);
	vec3 ambient = ((1 - tex.a) * material.ambient + tex.a * tex.rgb)
				 * light.ambient;
	fragColor = vec4(ambient, 1);
	#endif

	#if defined(POINT) || defined(SPOT) || defined(DIR)
	vec3 tanNormal = vec3(-1, -1, -1) + 2 * texture(material.bumpMap, texCoordF).rgb;
	tanNormal = normalize(tanNormal);

	vec4 tex;
	vec3 tanReflection = normalize(reflect(tanLightToVertex, tanNormal));
	bool facing = dot(tanNormal, tanLightToVertex) < 0;

	float attenuation = 1 / (1.0 + light.attenuationQuadratic * dot(tanLightToVertex, tanLightToVertex));

	tex = texture(material.diffuseMap, texCoordF);
	vec3 diffuse = ((1 - tex.a) * material.diffuse + tex.a * tex.rgb)
				 * max(dot(tanNormal, normalize(-tanLightToVertex)), 0)
				 * light.diffuse
				 * attenuation;

	tex = texture(material.specularMap, texCoordF);
	vec3 specular = ((1 - tex.a) * material.specular + tex.a * tex.rgb)
				  * pow(max(dot(tanReflection, -normalize(tanCameraToVertex)), 0), material.shine)
				  * light.specular
				  * (facing ? 1 : 0)
				  * attenuation;

	fragColor = vec4(diffuse + specular, 1);
	#endif

	#if defined(SHADOW)

	#if defined(POINT)
	float depth = length(worldPosition - light.position);
	float depthFront = textureCube(cubeShadowMap, worldPosition - light.position).r * light.far;
	bool inShadow = depth > depthFront + 1.0;
	#endif

	#if defined(SPOT)
	if (dot(normalize(tanLightDirection), normalize(tanLightToVertex)) < 0.75)  {
		diffuse = vec3(0, 0, 0);
		specular = vec3(0, 0, 0);
	}
	vec3 ndcCoords = lightSpacePosition.xyz / lightSpacePosition.w;
	vec2 texCoordS = vec2(0.5, 0.5) + 0.5 * ndcCoords.xy;
	float depth = length(worldPosition - light.position);
	float depthFront = texture(spotShadowMap, texCoordS).r * light.far;
	bool inShadow = depth > depthFront + 1.0;
	#endif

	#if defined(DIR)
	vec3 ndcCoords = lightSpacePosition.xyz / lightSpacePosition.w;
	vec2 texCoordS = vec2(0.5, 0.5) + 0.5 * ndcCoords.xy;
	float depth = 0.5 + 0.5 * ndcCoords.z; // make into [0, 1]
	if (texCoordS.x < 0 || texCoordS.y < 0 || texCoordS.x > 1 || texCoordS.y > 1 || depth < 0 || depth > 1) {
		return 1.0;
	}
	float depthFront = texture(dirShadowMap, texCoordS).r;
	bool inShadow = depth > depthFront + 0.05;
	#endif

	float factor = 1.0 - 0.5 * float(inShadow);
	fragColor = vec4(factor * vec3(fragColor), 1);
	#endif
}
