package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

type AmbientLight struct {
	color Vec3
}

type PointLight struct {
	Object
	diffuse Vec3
	specular Vec3
	shadowMap *CubeMap
	dirtyShadowMap bool
}

type SpotLight struct {
	Camera
	diffuse Vec3
	specular Vec3
	shadowMap *Texture2D
	dirtyShadowMap bool
}

func NewAmbientLight(color Vec3) *AmbientLight {
	var l AmbientLight
	l.color = color
	return &l
}

func NewPointLight(diffuse, specular Vec3) *PointLight {
	var l PointLight
	l.diffuse = diffuse
	l.specular = specular
	l.shadowMap = NewCubeMap(gl.NEAREST, gl.DEPTH_COMPONENT16, 512, 512)
	l.dirtyShadowMap = true
	return &l
}

func (l *PointLight) Place(position Vec3) {
	l.Object.Place(position)
	l.dirtyShadowMap = true
}

func NewSpotLight(diffuse, specular Vec3) *SpotLight {
	var l SpotLight
	l.diffuse = diffuse
	l.specular = specular
	l.Camera.Object.Reset()
	l.shadowMap = NewTexture2D(gl.NEAREST, gl.CLAMP_TO_BORDER, gl.DEPTH_COMPONENT16, 512, 512)
	l.shadowMap.SetBorderColor(NewVec4(1, 1, 1, 1))
	l.dirtyShadowMap = true
	return &l
}

func (l *SpotLight) Place(position Vec3) {
	l.Object.Place(position)
	l.dirtyShadowMap = true
}

func (l *SpotLight) Orient(unitX, unitY Vec3) {
	l.Object.Orient(unitX, unitY)
	l.dirtyShadowMap = true
}
