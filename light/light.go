package light

import (
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
)

type AmbientLight struct {
	Color math.Vec3
}

type PointLight struct {
	object.Object
	Diffuse              math.Vec3
	Specular             math.Vec3
	ShadowFar            float32
	Attenuation          float32
	CastShadows          bool
}

type SpotLight struct {
	camera.PerspectiveCamera
	Diffuse              math.Vec3
	Specular             math.Vec3
	Attenuation          float32
	CastShadows          bool
}

type DirectionalLight struct {
	camera.OrthoCamera
	Diffuse     math.Vec3
	Specular    math.Vec3
	CastShadows bool
}

func NewAmbientLight(color math.Vec3) *AmbientLight {
	var l AmbientLight
	l.Color = color
	return &l
}

func NewPointLight(diffuse, specular math.Vec3) *PointLight {
	var l PointLight
	l.Object = *object.NewObject()
	l.Diffuse = diffuse
	l.Specular = specular
	l.ShadowFar = 50
	l.Attenuation = 0
	l.CastShadows = false
	return &l
}

func (l *PointLight) Place(position math.Vec3) {
	l.Object.Place(position)
}

func NewSpotLight(diffuse, specular math.Vec3) *SpotLight {
	var l SpotLight
	l.Diffuse = diffuse
	l.Specular = specular
	l.PerspectiveCamera = *camera.NewPerspectiveCamera(90, 1, 0.1, 50)
	l.Attenuation = 0
	l.CastShadows = false
	return &l
}

func (l *SpotLight) Place(position math.Vec3) {
	l.Object.Place(position)
}

func (l *SpotLight) Orient(unitX, unitY math.Vec3) {
	l.Object.Orient(unitX, unitY)
}

func NewDirectionalLight(diffuse, specular math.Vec3) *DirectionalLight {
	var l DirectionalLight
	l.Diffuse = diffuse
	l.Specular = specular
	l.OrthoCamera = *camera.NewOrthoCamera(30, 1, 0, 25)
	l.OrthoCamera.Object = *object.NewObject()
	l.CastShadows = false
	return &l
}

func (l *DirectionalLight) Place(position math.Vec3) {
	l.Object.Place(position)
}

func (l *DirectionalLight) Orient(unitX, unitY math.Vec3) {
	l.Object.Orient(unitX, unitY)
}
