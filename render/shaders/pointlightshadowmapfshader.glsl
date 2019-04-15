#version 450

uniform vec3 lightPosition;
uniform float far;

in vec3 worldPosition;

void main() {
	gl_FragDepth = length(worldPosition - lightPosition) / far;
}
