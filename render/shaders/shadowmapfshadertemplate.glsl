#version 450

in vec3 worldPosition;

#if defined(POINT)
uniform vec3 lightPosition;
uniform float far;
#endif

#if defined(DIR)
uniform vec3 lightDir; // normalized
#endif

void main() {
	#if defined(POINT)
	gl_FragDepth = length(worldPosition - lightPosition) / far;
	#endif

	#if defined(DIR)
	gl_FragDepth = abs(dot(worldPosition - lightPosition, lightDir)) / far;
	#endif
}
