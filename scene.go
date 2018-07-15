package main

type Scene struct {
	meshes []*Mesh
	Light *Light
	quad *Mesh
}

func NewScene() *Scene {
	var s Scene
	var err error
	s.quad, err = ReadMesh("objects/quad.obj")
	if err != nil {
		panic(err)
	}
	return &s
}

func (s *Scene) AddMesh(m *Mesh) {
	s.meshes = append(s.meshes, m)
}
