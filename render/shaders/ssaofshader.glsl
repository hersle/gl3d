#version 450

out vec4 fragColor;
in vec2 texCoord;
uniform sampler2D depthMap;
uniform int depthMapWidth;
uniform int depthMapHeight;
uniform mat4 invProjectionMatrix;

void main() {
	float x = -1.0 + 2.0 * texCoord.x;
	float y = -1.0 + 2.0 * texCoord.y;
	float z = texture(depthMap, texCoord).r;
	vec4 viewPosition = invProjectionMatrix * vec4(x, y, z, 1.0);
	viewPosition /= viewPosition.w; // perspective divide

	fragColor = vec4(vec3(length(viewPosition) / 30.0), 1.0);
}
