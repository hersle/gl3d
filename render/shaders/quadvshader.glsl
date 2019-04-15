#version 450

in vec2 position;

out vec2 texCoord;

void main() {
	texCoord = 0.5 * (position - vec2(-1, -1));
	gl_Position = vec4(position, 0, 1);
}
