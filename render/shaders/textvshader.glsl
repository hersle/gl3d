#version 450

in vec2 position;
in vec2 texCoordV;
in vec3 colorV;

out vec2 texCoordF;
out vec3 colorF;

void main() {
	gl_Position = vec4(position, 0, 1);
	texCoordF = texCoordV;
	colorF = colorV;
}
