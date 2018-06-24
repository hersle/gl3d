#version 450

in vec4 colorF;
out vec4 fragColor;

void main() {
	fragColor = colorF;
	//fragColor = vec4(1.0, 0.0, 0.0, 1.0);
}
