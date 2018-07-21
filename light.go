package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

type BasicLight struct {
	ambient Vec3
	diffuse Vec3
	specular Vec3
}

type PointLight struct {
	Object
	BasicLight
	shadowMap *CubeMap
}

type SpotLight struct {
	Camera
	BasicLight
	shadowMap *Texture2D
}

func NewBasicLight(ambient, diffuse, specular Vec3) *BasicLight {
	var l BasicLight
	l.ambient = ambient
	l.diffuse = diffuse
	l.specular = specular
	return &l
}

func NewPointLight(ambient, diffuse, specular Vec3) *PointLight {
	var l PointLight
	l.BasicLight = *NewBasicLight(ambient, diffuse, specular)
	l.shadowMap = NewCubeMap(gl.NEAREST, gl.DEPTH_COMPONENT16, 512, 512)
	return &l
}

func NewSpotLight(ambient, diffuse, specular Vec3) *SpotLight {
	var l SpotLight
	l.BasicLight = *NewBasicLight(ambient, diffuse, specular)
	l.Camera.Object.Init()
	l.shadowMap = NewTexture2D(gl.NEAREST, gl.CLAMP_TO_BORDER, gl.DEPTH_COMPONENT16, 512, 512)
	l.shadowMap.SetBorderColor(NewVec4(1, 1, 1, 1))
	return &l
}
