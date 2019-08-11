#version 450

out vec2 texCoord;

void main() {
	float [4]pos = {-1, -1, +1, +1};

	float y = pos[(gl_VertexID + 0) % 4]; // -1, -1, +1, +1
	float x = pos[(gl_VertexID + 1) % 4]; // -1, +1, +1, -1

	gl_Position = vec4(x, y, 0, 1);
	texCoord = vec2(0.5) + 0.5 * gl_Position.xy;
}
