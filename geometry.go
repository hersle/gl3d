package main

type Box struct {
	width float32
	height float32
	depth float32
}

type Sphere struct {
	radius float32
}

func NewBox(width, height, depth float32) *Box {
	var b Box
	b.width = width
	b.height = height
	b.depth = depth
	return &b
}

func NewSphere(radius float32) *Sphere {
	var s Sphere
	s.radius = radius
	return &s
}
