#version 450

in vec2 position;

out vec2 fragPosition;

void main() {
	fragPosition = position;
	gl_Position = vec4(position, 0, 1);
}
