#version 450

in vec3 worldPosition;

#if defined(POINT) || defined(SPOT)
uniform vec3 lightPosition;
uniform float far;
#endif

void main() {
	#if defined(POINT) || defined(SPOT)
	gl_FragDepth = length(worldPosition - lightPosition) / far;
	#endif
}
