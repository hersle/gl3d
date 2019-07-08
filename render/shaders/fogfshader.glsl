#version 450

in vec2 fragPosition;

uniform sampler2D depthTexture;

uniform mat4 invProjectionMatrix;

uniform float cameraFar;

out vec4 fragColor;

void main() {
	vec3 fogColor = vec3(1, 1, 1);

	float depth = texture(depthTexture, 0.5 + 0.5 * fragPosition).r; // [0, 1]
	vec4 ndcPosition = vec4(fragPosition, -1.0 + 2.0 * depth, 1); // [-1, +1]
	vec4 clipPosition = invProjectionMatrix * ndcPosition;
	vec3 camPosition = clipPosition.xyz / clipPosition.w;

	float fogFactor = 1 - length(camPosition) / (cameraFar * 0.5);
	fogFactor = pow(2.71828, -(fogFactor));

	fragColor = vec4(fogColor, fogFactor);
}
