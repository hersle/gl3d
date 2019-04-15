#version 450

in vec2 position;
in vec2 texCoordV;

out vec2 texCoordF;

void main() {
	gl_Position = vec4(position, 0, 1);
	texCoordF = texCoordV;
}
