#version 450

in vec3 worldPositionG[3];
out vec3 worldPosition;

uniform mat4 projectionViewMatrices[6];

layout(triangles) in;
layout(triangle_strip, max_vertices=6*3) out;

void main() {
	for (int face = 0; face < 6; face++) {
		gl_Layer = face;
		for (int vert = 0; vert < 3; vert++) {
			worldPosition = worldPositionG[vert];
			gl_Position = projectionViewMatrices[face] * vec4(worldPosition, 1);
			EmitVertex();
		}
		EndPrimitive();
	}
}
