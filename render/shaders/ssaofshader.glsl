#version 450

out vec4 fragColor;
in vec2 texCoord;
uniform sampler2D depthMap;
uniform int depthMapWidth;
uniform int depthMapHeight;
uniform mat4 projectionMatrix;
uniform mat4 invProjectionMatrix;
uniform vec3 [16]directions;

const float DIRECTION_LENGTH = 0.2;

void main() {
	float x = -1.0 + 2.0 * texCoord.x;
	float y = -1.0 + 2.0 * texCoord.y;
	float z = texture(depthMap, texCoord).r;
	vec4 viewPosition = invProjectionMatrix * vec4(x, y, z, 1.0);
	viewPosition /= viewPosition.w; // perspective divide

	float fraction = 0.0;
	for (int i = 0; i < 16; i++) {
		vec4 viewPos = viewPosition + vec4(directions[i] * DIRECTION_LENGTH, 0);
		vec4 projPos = projectionMatrix * viewPos;
		projPos /= projPos.w; // perspective divide
		float depth = texture(depthMap, vec2(0.5) + 0.5 * projPos.xy).r;

		float x2 = projPos.x;
		float y2 = projPos.y;
		float z2 = depth;
		vec4 viewPosition2 = invProjectionMatrix * vec4(x2, y2, z2, 1.0);
		viewPosition2 /= viewPosition2.w; // perspective divide

		float dist = length(viewPosition2.xyz - viewPosition.xyz);

		fraction += dist > 0.30 ? 1.0 / 16.0 : 0.0;
	}

	fraction = fraction > 0.5 ? fraction : 0.0;
	fraction = 1 - fraction;

	fragColor = vec4(vec3(fraction), 1.0);
}
