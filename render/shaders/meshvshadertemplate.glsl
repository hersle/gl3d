#version 450

in vec3 position;
in vec2 texCoordV;
in vec3 normalV;
in vec3 tangentV;
in vec3 bitangentV;

out vec2 texCoordF;
out vec3 worldPosition;

uniform mat4 modelMatrix;
uniform mat4 viewMatrix;
uniform mat4 projectionMatrix;

#if defined(POINT)
out vec3 tanLightToVertex;
out vec3 tanCameraToVertex;
#endif

#if defined(SPOT)
out vec3 tanLightToVertex;
out vec3 tanCameraToVertex;
out vec3 tanLightDirection;
out vec4 lightSpacePosition;
#endif

#if defined(DIR)
out vec3 tanLightToVertex;
out vec3 tanCameraToVertex;
out vec4 lightSpacePosition;
#endif

#if defined(POINT) || defined(SPOT) || defined(DIR)
uniform mat4 normalMatrix;
#endif

#if defined(SHADOW) && (defined(SPOT) || defined(DIR))
uniform mat4 shadowProjectionViewMatrix;
#endif

#if defined(DEPTH)
uniform float materialAlpha; // TODO: let textures modify alpha
uniform sampler2D materialAlphaMap;
#endif

#if defined(AMBIENT)
uniform vec3 materialAmbient;
uniform sampler2D materialAmbientMap;
#endif

#if defined(POINT) || defined(SPOT) || defined(DIR)
uniform vec3 materialDiffuse;
uniform vec3 materialSpecular;
uniform float materialShine;
uniform sampler2D materialDiffuseMap;
uniform sampler2D materialSpecularMap;
uniform sampler2D materialBumpMap;
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
#endif

#if defined(DIR)
uniform vec3 lightDirection;
uniform vec3 lightColor;
uniform float lightAttenuation;
#endif

void main() {
	worldPosition = vec3(modelMatrix * vec4(position, 1));
	vec3 viewPosition = vec3(viewMatrix * vec4(worldPosition, 1));
	gl_Position = projectionMatrix * vec4(viewPosition, 1);

	texCoordF = texCoordV;

	#if defined(POINT) || defined(SPOT) || defined(DIR)
	vec3 viewNormal = normalize(vec3(normalMatrix * vec4(normalV, 0)));
	vec3 viewTangent = normalize(vec3(normalMatrix * vec4(tangentV, 0)));
	vec3 viewBitangent = normalize(cross(viewNormal, viewTangent));
	mat3 tanToView = mat3(viewTangent, viewBitangent, viewNormal);
	mat3 viewToTan = transpose(tanToView); // orthonormal

	#if defined(DIR)
	vec3 worldLightToVertex = lightDirection;
	#else
	vec3 worldLightToVertex = worldPosition - lightPosition;
	#endif
	vec3 viewLightToVertex = vec3(viewMatrix * vec4(worldLightToVertex, 0));
	tanLightToVertex = viewToTan * viewLightToVertex;
	tanCameraToVertex = viewToTan * (viewPosition - vec3(0, 0, 0));
	#endif

	#if defined(SPOT)
	vec3 viewLightDirection = vec3(viewMatrix * vec4(lightDirection, 0));
	tanLightDirection = viewToTan * viewLightDirection;
	#endif

	#if defined(SHADOW) && (defined(SPOT) || defined(DIR))
	lightSpacePosition = shadowProjectionViewMatrix * modelMatrix * vec4(position, 1);
	#endif
}
