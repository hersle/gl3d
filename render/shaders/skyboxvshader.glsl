#version 450

in vec3 positionV;

out vec3 positionF;

uniform mat4 modelMatrix;
uniform mat4 viewMatrix;
uniform mat4 projectionMatrix;

void main() {
	positionF = positionV;
	gl_Position = projectionMatrix * mat4(mat3(viewMatrix)) * vec4(positionV, 1.0);
}
