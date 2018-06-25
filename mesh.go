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
	p0 := NewVec3(-0.5, -0.5, -0.5)
	p1 := NewVec3(+0.5, -0.5, -0.5)
	p2 := NewVec3(+0.5, +0.5, -0.5)
	p3 := NewVec3(-0.5, +0.5, -0.5)
	p4 := NewVec3(-0.5, -0.5, +0.5)
	p5 := NewVec3(+0.5, -0.5, +0.5)
	p6 := NewVec3(+0.5, +0.5, +0.5)
	p7 := NewVec3(-0.5, +0.5, +0.5)

	c0 := NewColor(0x20, 0x20, 0x20, 0xff)
	c1 := NewColor(0xff, 0x00, 0x00, 0xff)
	c2 := NewColor(0x00, 0xff, 0x00, 0xff)
	c3 := NewColor(0x00, 0x00, 0xff, 0xff)
	c4 := NewColor(0xff, 0xff, 0x00, 0xff)
	c5 := NewColor(0xff, 0x00, 0xff, 0xff)
	c6 := NewColor(0x00, 0xff, 0xff, 0xff)
	c7 := NewColor(0xff, 0xff, 0xff, 0xff)

	v0 := Vertex{p0, c0}
	v1 := Vertex{p1, c1}
	v2 := Vertex{p2, c2}
	v3 := Vertex{p3, c3}
	v4 := Vertex{p4, c4}
	v5 := Vertex{p5, c5}
	v6 := Vertex{p6, c6}
	v7 := Vertex{p7, c7}

	mesh1.verts = []Vertex{v0, v1, v2, v3, v4, v5, v6, v7}
	mesh1.faces = []int{
		0, 1, 2,
		0, 2, 3,
		1, 2, 5,
		2, 5, 6,
		4, 5, 6,
		4, 6, 7,
		0, 4, 3,
		3, 4, 7,
		0, 1, 5,
		0, 5, 4,
		2, 3, 6,
		3, 6, 7,
	}
}
