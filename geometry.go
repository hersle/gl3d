package main

import (
	"math"
)

type Box struct {
	Min Vec3
	Max Vec3
}

type Sphere struct {
	Center Vec3
	Radius float32
}

func NewBox(point1, point2 Vec3) *Box {
	var b Box

	minX := float32(math.Min(float64(point1.X()), float64(point2.X())))
	minY := float32(math.Min(float64(point1.Y()), float64(point2.Y())))
	minZ := float32(math.Min(float64(point1.Z()), float64(point2.Z())))
	b.Min = NewVec3(minX, minY, minZ)

	maxX := float32(math.Max(float64(point1.X()), float64(point2.X())))
	maxY := float32(math.Max(float64(point1.Y()), float64(point2.Y())))
	maxZ := float32(math.Max(float64(point1.Z()), float64(point2.Z())))
	b.Max = NewVec3(maxX, maxY, maxZ)

	return &b
}

func (b *Box) Dx() float32 {
	return b.Max.X() - b.Min.X()
}

func (b *Box) Dy() float32 {
	return b.Max.Y() - b.Min.Y()
}

func (b *Box) Dz() float32 {
	return b.Max.Z() - b.Min.Z()
}

func (b *Box) Center() Vec3 {
	return b.Min.Add(b.Max).Scale(0.5)
}

func NewSphere(center Vec3, radius float32) *Sphere {
	var s Sphere

	s.Center = center
	s.Radius = radius

	return &s
}
