package main

type Light struct {
	position Vec3
	ambient Vec3
	diffuse Vec3
	specular Vec3
}

func NewLight(position, ambient, diffuse, specular Vec3) *Light {
	var l Light
	l.position = position
	l.ambient = ambient
	l.diffuse = diffuse
	l.specular = specular
	return &l
}
