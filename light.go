package main

type SpotLight struct {
	Camera
	ambient Vec3
	diffuse Vec3
	specular Vec3
}

func NewSpotLight(ambient, diffuse, specular Vec3) *SpotLight {
	var l SpotLight
	l.Camera.Object.Init()
	l.ambient = ambient
	l.diffuse = diffuse
	l.specular = specular
	return &l
}
