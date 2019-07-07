package light

import (
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	gomath "math"
)

type AmbientLight struct {
	Color math.Vec3
	Intensity float32
}

type PointLight struct {
	object.Object
	Color                math.Vec3
	Intensity            float32
	ShadowFar            float32
	Attenuation          float32
	CastShadows          bool
}

type SpotLight struct {
	camera.PerspectiveCamera
	Color                math.Vec3
	Intensity            float32
	Attenuation          float32
	CastShadows          bool
	FOV                  float32
}

type DirectionalLight struct {
	camera.OrthoCamera
	Color       math.Vec3
	Intensity   float32
	CastShadows bool
}

func NewAmbientLight(color math.Vec3) *AmbientLight {
	var l AmbientLight
	l.Color = color
	l.Intensity = 1
	return &l
}

func NewPointLight(color math.Vec3) *PointLight {
	var l PointLight
	l.Object = *object.NewObject()
	l.Color = color
	l.Intensity = 1
	l.ShadowFar = 50
	l.Attenuation = 0
	l.CastShadows = false
	return &l
}

func (l *PointLight) Place(position math.Vec3) {
	l.Object.Place(position)
}

func NewSpotLight(color math.Vec3) *SpotLight {
	var l SpotLight
	l.Color = color
	l.Intensity = 1
	l.PerspectiveCamera = *camera.NewPerspectiveCamera(90, 1, 0.1, 50)
	l.Attenuation = 0
	l.CastShadows = false
	l.FOV = gomath.Pi / 2
	return &l
}

func (l *SpotLight) Place(position math.Vec3) {
	l.Object.Place(position)
}

func (l *SpotLight) Orient(unitX, unitY math.Vec3) {
	l.Object.Orient(unitX, unitY)
}

func NewDirectionalLight(color math.Vec3) *DirectionalLight {
	var l DirectionalLight
	l.Color = color
	l.Intensity = 1
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
