#version 450

in vec4 colorF;
in vec2 texCoordF;

out vec4 fragColor;

uniform sampler2D tex;

void main() {
	fragColor = vec4(texture(tex, texCoordF).rgb, colorF.a);
}
