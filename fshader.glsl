#version 450

in vec4 colorF;
in vec2 texCoordF;

out vec4 fragColor;

uniform vec3 ambient;

void main() {
	fragColor = vec4(ambient, 1.0);
}
