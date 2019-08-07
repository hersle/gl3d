#version 450

in vec2 position;
out vec2 texCoord;

void main() {
	texCoord = vec2(0.5, 0.5) + 0.5 * position;
	gl_Position = vec4(position, 0, 1.0);
}
