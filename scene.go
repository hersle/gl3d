package main

import (
	"github.com/hersle/gl3d/object"
)

type Scene struct {
	meshes []*object.Mesh
	ambientLight *AmbientLight
	spotLights []*SpotLight
	pointLights []*PointLight
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

func (s *Scene) AddPointLight(l *PointLight) {
	s.pointLights = append(s.pointLights, l)
}

func (s *Scene) AddSpotLight(l *SpotLight) {
	s.spotLights = append(s.spotLights, l)
}
