#version 450

in vec3 position;

void main() {
	gl_Position = vec4(position, 1.0);
	//gl_Position = vec4(0.0, 0.0, 0.0, 1.0);
}
