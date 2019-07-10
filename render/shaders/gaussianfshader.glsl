#version 450

in vec2 fragPosition;
uniform sampler2D inTexture;
uniform vec2 dir;
uniform float texDim;

out vec4 fragColor;

const float pi = 3.1415;
uniform float stddev = 3.0;

void main() {
	vec2 coord0 = 0.5 + 0.5 * fragPosition;

	int size = int(ceil(3.0 * stddev));

	vec3 color = vec3(texture(inTexture, coord0));
	float sum = 1.0;
	for (int i = 1; i <= size; i++) {
		float d = 1.0 / texDim * float(i);
		float coeff = exp(-d*d / (2*stddev*stddev));
		color += coeff * texture(inTexture, coord0 - d*dir).rgb;
		color += coeff * texture(inTexture, coord0 + d*dir).rgb;
		sum += 2 * coeff;
	}
	color /= sum; // normalize

	fragColor = vec4(color, 1.0);
}
