package main

import (
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/light"
)

type Scene struct {
	meshes []*object.Mesh
	ambientLight *light.AmbientLight
	spotLights []*light.SpotLight
	pointLights []*light.PointLight
	quad *object.Mesh
}

func NewScene() *Scene {
	var s Scene
	var err error
	s.quad, err = object.ReadMesh("objects/quad.obj")
	if err != nil {
		panic(err)
	}
	return &s
}

func (s *Scene) AddMesh(m *object.Mesh) {
	s.meshes = append(s.meshes, m)
}

func (s *Scene) AddPointLight(l *light.PointLight) {
	s.pointLights = append(s.pointLights, l)
}

func (s *Scene) AddSpotLight(l *light.SpotLight) {
	s.spotLights = append(s.spotLights, l)
}
