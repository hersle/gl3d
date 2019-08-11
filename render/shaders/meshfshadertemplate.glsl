#version 450

in vec2 texCoordF;
in vec4 projPosition;

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

#if defined(DEPTH)
uniform float materialAlpha; // TODO: let textures modify alpha
uniform sampler2D materialAlphaMap;
#endif

#if defined(AMBIENT)
uniform vec3 materialAmbient;
uniform sampler2D materialAmbientMap;
uniform sampler2D aoMap;
#endif

#if defined(POINT) || defined(SPOT) || defined(DIR)
uniform vec3 materialDiffuse;
uniform vec3 materialSpecular;
uniform float materialShine;
uniform sampler2D materialDiffuseMap;
uniform sampler2D materialSpecularMap;
uniform sampler2D materialBumpMap;
uniform int materialBumpMapWidth;
uniform int materialBumpMapHeight;
#endif

#if defined(AMBIENT)
uniform vec3 lightColor;
#endif

#if defined(POINT)
uniform vec3 lightPosition;
uniform vec3 lightColor;
uniform float lightFar;
uniform float lightAttenuation;
#endif

#if defined(SPOT)
uniform vec3 lightPosition;
uniform vec3 lightDirection;
uniform vec3 lightColor;
uniform float lightFar;
uniform float lightAttenuation;
uniform float lightCosAng;
#endif

#if defined(DIR)
uniform vec3 lightDirection;
uniform vec3 lightColor;
uniform float lightAttenuation;
#endif

#if defined(SHADOW)
#if defined(POINT)
uniform samplerCube shadowMap;
#elif defined(SPOT) || defined(DIR)
uniform sampler2D shadowMap;
#endif
#if defined(PCF)
uniform int kernelSize;
#endif
#endif

void main() {
	#if defined(DEPTH)
	float alpha = materialAlpha * texture(materialAlphaMap, texCoordF).r;
	// TODO: add proper transparency?
	if (alpha < 1) {
		discard;
	}
	#endif

	#if defined(AMBIENT)
	vec4 tex;
	tex = texture(materialAmbientMap, texCoordF);
	vec2 screenTexCoord = vec2(0.5) + 0.5 * projPosition.xy / projPosition.w;
	float ao = texture(aoMap, screenTexCoord).r;
	vec3 ambient = ((1 - tex.a) * materialAmbient + tex.a * tex.rgb)
				 * ao
				 * lightColor;
	fragColor = vec4(ambient, 1);
	#endif

	#if defined(POINT) || defined(SPOT) || defined(DIR)
	float dx = 1.0 / materialBumpMapWidth;
	float dy = 1.0 / materialBumpMapHeight;
	float z1 = texture(materialBumpMap, vec2(texCoordF.x-dx, texCoordF.y)).r;
	float z2 = texture(materialBumpMap, vec2(texCoordF.x+dx, texCoordF.y)).r;
	float dzdx = (z2-z1) * 10.0;
	z1 = texture(materialBumpMap, vec2(texCoordF.x, texCoordF.y-dy)).r;
	z2 = texture(materialBumpMap, vec2(texCoordF.x, texCoordF.y+dy)).r;
	float dzdy = (z2-z1) * 10.0;
	vec3 tanNormal = normalize(vec3(dzdx, dzdy, 2));

	vec4 tex;
	vec3 tanReflection = normalize(reflect(tanLightToVertex, tanNormal));
	bool facing = dot(tanNormal, tanLightToVertex) < 0;

	float attenuation = 1 / (1.0 + lightAttenuation * dot(tanLightToVertex, tanLightToVertex));

	tex = texture(materialDiffuseMap, texCoordF);
	vec3 diffuse = ((1 - tex.a) * materialDiffuse + tex.a * tex.rgb)
				 * max(dot(tanNormal, normalize(-tanLightToVertex)), 0)
				 * lightColor
				 * attenuation;

	tex = texture(materialSpecularMap, texCoordF);
	vec3 specular = ((1 - tex.a) * materialSpecular + tex.a * tex.rgb)
				  * pow(max(dot(tanReflection, -normalize(tanCameraToVertex)), 0), materialShine)
				  * lightColor
				  * (facing ? 1 : 0)
				  * attenuation;

	#if defined(SPOT)
	if (dot(normalize(tanLightDirection), normalize(tanLightToVertex)) < lightCosAng)  {
		diffuse = vec3(0, 0, 0);
		specular = vec3(0, 0, 0);
	}
	#endif

	fragColor = vec4(diffuse + specular, 1);
	#endif

	#if defined(SHADOW)

	#if defined(POINT)
	#if defined(PCF)
	vec3 coord = worldPosition - lightPosition;
	float depth = length(coord);
	vec4 depthFront = textureGather(shadowMap, coord) * lightFar;
	vec4 inShadow = vec4(greaterThan(vec4(depth), depthFront + 0.1));
	float sum = float(dot(inShadow, inShadow));
	float factor = 1.0 - sum / 4.0;
	#else
	float depth = length(worldPosition - lightPosition);
	float depthFront = textureCube(shadowMap, worldPosition - lightPosition).r * lightFar;
	bool inShadow = depth > depthFront + 1.0;
	float factor = 1.0 - 1.0 * float(inShadow);
	#endif
	#endif

	#if defined(SPOT)
	#if defined(PCF)
	vec3 ndcCoords = lightSpacePosition.xyz / lightSpacePosition.w;
	vec2 texCoords = vec2(0.5, 0.5) + 0.5 * ndcCoords.xy;
	float depth = length(worldPosition - lightPosition);
	vec4 depthFront = textureGather(shadowMap, texCoords) * lightFar;
	vec4 inShadow = vec4(greaterThan(vec4(depth), depthFront + 1.0));
	float sum = float(dot(inShadow, inShadow));
	float factor = 1.0 - sum / 4.0;
	#else
	vec3 ndcCoords = lightSpacePosition.xyz / lightSpacePosition.w;
	vec2 texCoordS = vec2(0.5, 0.5) + 0.5 * ndcCoords.xy;
	float depth = length(worldPosition - lightPosition);
	float depthFront = texture(shadowMap, texCoordS).r * lightFar;
	bool inShadow = depth > depthFront + 1.0;
	float factor = 1.0 - 1.0 * float(inShadow);
	#endif
	#endif

	#if defined(DIR)
	#if defined(PCF)
	vec3 ndcCoords = lightSpacePosition.xyz / lightSpacePosition.w;
	vec2 texCoords = vec2(0.5, 0.5) + 0.5 * ndcCoords.xy;
	float depth = 0.5 + 0.5 * ndcCoords.z; // make into [0, 1]
	vec4 depthFront = textureGather(shadowMap, texCoords);
	vec4 inShadow = vec4(greaterThan(vec4(depth), depthFront + 0.05));
	float sum = float(dot(inShadow, inShadow));
	float factor = 1.0 - sum / 4.0;
	#else
	vec3 ndcCoords = lightSpacePosition.xyz / lightSpacePosition.w;
	vec2 texCoordS = vec2(0.5, 0.5) + 0.5 * ndcCoords.xy;
	float depth = 0.5 + 0.5 * ndcCoords.z; // make into [0, 1]
	float depthFront = texture(shadowMap, texCoordS).r;
	bool inShadow = depth > depthFront + 0.05;
	float factor = 1.0 - 1.0 * float(inShadow);
	#endif
	#endif

	fragColor = vec4(factor * vec3(fragColor), 1);
	#endif
}
