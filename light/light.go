package light

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
)

type AmbientLight struct {
	Color math.Vec3
}

type PointLight struct {
	object.Object
	Diffuse        math.Vec3
	Specular       math.Vec3
	ShadowMap      *graphics.CubeMap
	DirtyShadowMap bool
	ShadowFar      float32
}

type SpotLight struct {
	camera.PerspectiveCamera
	Diffuse        math.Vec3
	Specular       math.Vec3
	ShadowMap      *graphics.Texture2D
	DirtyShadowMap bool
}

func NewAmbientLight(color math.Vec3) *AmbientLight {
	var l AmbientLight
	l.Color = color
	return &l
}

func NewPointLight(diffuse, specular math.Vec3) *PointLight {
	var l PointLight
	l.Diffuse = diffuse
	l.Specular = specular
	l.ShadowMap = graphics.NewCubeMap(graphics.NearestFilter, gl.DEPTH_COMPONENT16, 512, 512)
	l.DirtyShadowMap = true
	l.ShadowFar = 50
	return &l
}

func (l *PointLight) Place(position math.Vec3) {
	l.Object.Place(position)
	l.DirtyShadowMap = true
}

func NewSpotLight(diffuse, specular math.Vec3) *SpotLight {
	var l SpotLight
	l.Diffuse = diffuse
	l.Specular = specular
	l.PerspectiveCamera.Object.Reset()
	l.ShadowMap = graphics.NewTexture2D(graphics.NearestFilter, graphics.BorderClampWrap, gl.DEPTH_COMPONENT16, 512, 512)
	l.ShadowMap.SetBorderColor(math.NewVec4(1, 1, 1, 1))
	l.DirtyShadowMap = true
	return &l
}

func (l *SpotLight) Place(position math.Vec3) {
	l.Object.Place(position)
	l.DirtyShadowMap = true
}

func (l *SpotLight) Orient(unitX, unitY math.Vec3) {
	l.Object.Orient(unitX, unitY)
	l.DirtyShadowMap = true
}
