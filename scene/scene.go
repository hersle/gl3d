package scene

import (
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/graphics"
)

type Scene struct {
	Meshes       []*object.Mesh
	AmbientLight *light.AmbientLight
	SpotLights   []*light.SpotLight
	PointLights  []*light.PointLight
	DirectionalLights  []*light.DirectionalLight
	Skybox       *graphics.CubeMap
}

func NewScene() *Scene {
	var s Scene
	var err error
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

func (s *Scene) AddDirectionalLight(l *light.DirectionalLight) {
	s.DirectionalLights = append(s.DirectionalLights, l)
}

func (s *Scene) AddSkybox(skybox *graphics.CubeMap) {
	s.Skybox = skybox
}
