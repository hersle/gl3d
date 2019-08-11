#version 450

in vec2 texCoord;

uniform sampler2D aoMap;
uniform int aoMapWidth;
uniform int aoMapHeight;

uniform sampler2D depthMap;
uniform int depthWidth;
uniform int depthHeight;

out vec4 fragColor;

const int SAMPLE_SIZE = 8;

void main() {
	float dx = 1.0 / aoMapWidth;
	float dy = 1.0 / aoMapHeight;

	float result = 0.0;

	vec2 tl = texCoord - vec2(float(SAMPLE_SIZE) / 2.0 * dx, float(SAMPLE_SIZE) / 2.0 * dy);

	for (int x = 0; x < SAMPLE_SIZE; x++) {
		for (int y = 0; y < SAMPLE_SIZE; y++) {
			result += texture(aoMap, tl + vec2(x * dx, y * dy)).r;
		}
	}
	result /= (SAMPLE_SIZE * SAMPLE_SIZE);

	fragColor = vec4(result, 0, 0, 0);
}
