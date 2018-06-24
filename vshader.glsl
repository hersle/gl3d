#version 450

in vec3 position;
in vec4 colorV;
out vec4 colorF;

void main() {
	gl_Position = vec4(position, 1.0);
	colorF = colorV;
	//gl_Position = vec4(0.0, 0.0, 0.0, 1.0);
}
