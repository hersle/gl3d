package main

type Scene struct {
	meshes []*Mesh
	Light *Light
}

func NewScene() *Scene {
	var s Scene
	return &s
}

func (s *Scene) AddMesh(m *Mesh) {
	s.meshes = append(s.meshes, m)
}
