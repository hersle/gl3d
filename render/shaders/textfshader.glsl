#version 450

in vec2 texCoordF;
in vec3 colorF;

out vec4 fragColor;

uniform sampler2D fontAtlas;

void main() {
	vec4 tex = texture(fontAtlas, texCoordF);
	fragColor = vec4(colorF, tex.r);
}
