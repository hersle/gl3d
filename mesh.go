package main

type Mesh struct {
	verts []Vertex
	faces []int
	// TODO: transformation matrix, etc.
}

func NewMesh(verts []Vertex, faces []int) *Mesh {
	var m Mesh
	m.verts = verts
	m.faces = faces
	return &m
}

var mesh1 Mesh

func init() {
	color := NewColor(0xff, 0xff, 0xff, 0x80)
	p0 := Vec3{0.5, 0.5, 0}
	p1 := Vec3{0.0, 0.0, 0}
	p2 := Vec3{0.5, 0.0, 0}
	p3 := Vec3{0.25, 0.75, 0}
	p4 := Vec3{-0.1, -0.3, 0}
	v0 := Vertex{p0, color}
	v1 := Vertex{p1, color}
	v2 := Vertex{p2, color}
	v3 := Vertex{p3, color}
	v4 := Vertex{p4, color}
	mesh1.verts = []Vertex{v0, v1, v2, v3, v4}
	mesh1.faces = []int{
		0, 1, 2,
		2, 3, 4,
		1, 3, 4,
	}
}
