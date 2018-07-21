#version 450

in vec3 worldPosition;
in vec3 viewPosition;
in vec4 colorF;
in vec2 texCoordF;
in vec3 normalF;

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
} material;

uniform struct Light {
	vec3 position;
	vec3 direction;
	vec3 ambient;
	vec3 diffuse;
	vec3 specular;
} light;

uniform int mode;

uniform mat4 viewMatrix;
uniform mat4 projectionMatrix;

// for spotlight
//uniform sampler2D shadowMap;

uniform samplerCube shadowMap;

// for spotlight
/*
float CalcShadowFactor(vec4 lightSpacePos) {
	vec3 ndcCoords = lightSpacePos.xyz / lightSpacePos.w;
	vec2 texCoordS = vec2(0.5, 0.5) + 0.5 * ndcCoords.xy;
	float depth = 0.5 + 0.5 * ndcCoords.z;
	float depthFront = texture(shadowMap, texCoordS).r;
	bool inShadow = depth > depthFront + 0.001;
	if (inShadow) {
		return 0.5;
	} else {
		return 1.0;
	}
}
*/

float CalcShadowFactorPointLight() {
	float depth = length(worldPosition - light.position);
	float depthFront = textureCube(shadowMap, worldPosition - light.position).r * 50;
	bool inShadow = depth > depthFront;
	if (inShadow) {
		return 0.5;
	} else {
		return 1.0;
	}
}

void main() {
	if (mode == 1) {
		// TODO: must write always!
		gl_FragDepth = length(worldPosition - light.position) / 50;
	} else if (mode == 2) {
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

		// for spotlight
		/*
		if (dot((viewMatrix * vec4(light.direction, 0)).xyz, lightDirection) < 0.75)  {
			diffuse = vec3(0, 0, 0);
			specular = vec3(0, 0, 0);
		}
		*/

		float factor = CalcShadowFactorPointLight();
		fragColor = vec4(ambient + factor * (diffuse + specular), material.alpha);
		//fragColor -= vec4(ambient + factor * (diffuse + specular), material.alpha);

		float depth = length(worldPosition - light.position) / 50;
		float depthFront = texture(shadowMap, worldPosition - light.position).r;
		//fragColor += vec4(vec3(depth), 1);
		//fragColor += vec4(vec3(depthFront), 1);
		//fragColor += vec4(vec3(depthFront / depth), 1);
	}
}
