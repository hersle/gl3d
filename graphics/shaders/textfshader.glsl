#version 450

in vec2 texCoordF;

out vec4 fragColor;

uniform sampler2D fontAtlas;

void main() {
	fragColor = vec4(texture(fontAtlas, texCoordF).rrr, 1);
}
