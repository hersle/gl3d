package main

type Scene struct {
	meshes []*Mesh
}

func NewScene() *Scene {
	var s Scene
	return &s
}

func (s *Scene) AddMesh(m *Mesh) {
	s.meshes = append(s.meshes, m)
}
