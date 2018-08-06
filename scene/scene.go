package scene

import (
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/light"
)

type Scene struct {
	Meshes []*object.Mesh
	AmbientLight *light.AmbientLight
	SpotLights []*light.SpotLight
	PointLights []*light.PointLight
	Quad *object.Mesh
}

func NewScene() *Scene {
	var s Scene
	var err error
	s.Quad, err = object.ReadMesh("assets/objects/quad/quad.obj")
	if err != nil {
		panic(err)
	}
	return &s
}

func (s *Scene) AddMesh(m *object.Mesh) {
	s.Meshes = append(s.Meshes, m)
}

func (s *Scene) AddPointLight(l *light.PointLight) {
	s.PointLights = append(s.PointLights, l)
}

func (s *Scene) AddSpotLight(l *light.SpotLight) {
	s.SpotLights = append(s.SpotLights, l)
}
