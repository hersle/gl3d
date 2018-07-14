package main

type Light struct {
	Object
	ambient Vec3
	diffuse Vec3
	specular Vec3
}

func NewLight(ambient, diffuse, specular Vec3) *Light {
	var l Light
	l.Object.Init()
	l.ambient = ambient
	l.diffuse = diffuse
	l.specular = specular
	return &l
}
