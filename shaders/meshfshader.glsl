#version 450

in vec3 worldPosition;
in vec3 viewPosition;
in vec4 colorF;
in vec2 texCoordF;
in vec3 viewLightToVertex;
in vec3 tanCameraToVertex;
in vec3 viewLightDirection;
in mat3 viewToTan;

out vec4 fragColor;

// material
uniform vec3 materialAmbient;
uniform vec3 materialDiffuse;
uniform vec3 materialSpecular;
uniform float materialShine;
uniform sampler2D materialAmbientMap;
uniform sampler2D materialDiffuseMap;
uniform sampler2D materialSpecularMap;
uniform float materialAlpha; // TODO: let textures modify alpha
uniform bool materialHasBumpMap;
uniform sampler2D materialBumpMap;
uniform bool materialHasAlphaMap;
uniform sampler2D materialAlphaMap;

const int MAX_LIGHTS = 10;
uniform vec3 lightPositions[MAX_LIGHTS];
uniform vec3 lightAmbients[MAX_LIGHTS];
uniform vec3 lightDiffuses[MAX_LIGHTS];
uniform vec3 lightSpeculars[MAX_LIGHTS];
uniform mat4 lightViewMatrices[MAX_LIGHTS];
uniform mat4 lightProjectionMatrices[MAX_LIGHTS];

// for spotlight
uniform sampler2D lightSpotShadowMaps[MAX_LIGHTS];

uniform samplerCube lightCubeShadowMaps[MAX_LIGHTS];

float CalcShadowFactorSpotLight(vec4 lightSpacePos) {
	vec3 ndcCoords = lightSpacePos.xyz / lightSpacePos.w;
	vec2 texCoordS = vec2(0.5, 0.5) + 0.5 * ndcCoords.xy;
	float depth = length(worldPosition - lightPositions[0]);
	float depthFront = texture(lightSpotShadowMaps[0], texCoordS).r * 50;
	bool inShadow = depth > depthFront + 1.0;
	if (inShadow) {
		return 0.5;
	} else {
		return 1.0;
	}
}

float CalcShadowFactorPointLight() {
	float depth = length(worldPosition - lightPositions[0]);
	float depthFront = textureCube(lightCubeShadowMaps[0], worldPosition - lightPositions[0]).r * 50;
	bool inShadow = depth > depthFront + 1.0;
	if (inShadow) {
		return 0.5;
	} else {
		return 1.0;
	}
}

void main() {
	vec4 lightSpacePosition = lightProjectionMatrices[0] * lightViewMatrices[0] * vec4(worldPosition, 1);

	vec3 tanNormal = vec3(0, 0, 1);
	if (materialHasBumpMap) {
		tanNormal += -1 + 2 * normalize(texture(materialBumpMap, texCoordF).rgb);
		tanNormal = normalize(tanNormal);
	}

	vec3 tanLightToVertex = viewToTan * viewLightToVertex;
	vec3 tanLightDirection = viewToTan * viewLightDirection;

	vec4 tex;
	vec3 tanReflection = reflect(normalize(tanLightToVertex), normalize(tanNormal));
	bool facing = dot(normalize(tanNormal), normalize(tanLightToVertex)) < 0;

	tex = texture(materialAmbientMap, texCoordF);
	vec3 ambient = ((1 - tex.a) * materialAmbient + tex.a * tex.rgb)
				 * lightAmbients[0];

	tex = texture(materialDiffuseMap, texCoordF);
	vec3 diffuse = ((1 - tex.a) * materialDiffuse + tex.a * tex.rgb)
				 * max(dot(normalize(tanNormal), normalize(-tanLightToVertex)), 0)
				 * lightDiffuses[0];

	tex = texture(materialSpecularMap, texCoordF);
	vec3 specular = ((1 - tex.a) * materialSpecular + tex.a * tex.rgb)
				  * pow(max(dot(normalize(tanReflection), -normalize(tanCameraToVertex)), 0), materialShine)
				  * lightSpeculars[0]
				  * (facing ? 1 : 0);

	if (dot(normalize(tanLightDirection), normalize(tanLightToVertex)) < 0.75)  {
		// add/remove + to enable/disable spotlight
		diffuse = vec3(0, 0, 0);
		specular = vec3(0, 0, 0);
	}
	// change to enable/disable spotlight
	float factor = CalcShadowFactorPointLight();
	factor /= CalcShadowFactorPointLight();
	factor *= CalcShadowFactorSpotLight(lightSpacePosition);

	float alpha;
	if (materialHasAlphaMap) {
		alpha = materialAlpha * texture(materialAlphaMap, texCoordF).r;
	} else {
		alpha = materialAlpha;
	}

	// TODO: add proper transparency?
	if (alpha < 1) {
		discard;
	}

	fragColor = vec4(ambient + factor * (diffuse + specular), 1);
}
