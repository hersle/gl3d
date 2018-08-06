package geometry

import (
	"github.com/hersle/gl3d/math"
)

type Box struct {
	Min math.Vec3
	Max math.Vec3
}

type Sphere struct {
	Center math.Vec3
	Radius float32
}

func NewBox(point1, point2 math.Vec3) *Box {
	var b Box

	minX := math.Min(point1.X(), point2.X())
	minY := math.Min(point1.Y(), point2.Y())
	minZ := math.Min(point1.Z(), point2.Z())
	b.Min = math.NewVec3(minX, minY, minZ)

	maxX := math.Max(point1.X(), point2.X())
	maxY := math.Max(point1.Y(), point2.Y())
	maxZ := math.Max(point1.Z(), point2.Z())
	b.Max = math.NewVec3(maxX, maxY, maxZ)

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

func (b *Box) Center() math.Vec3 {
	return b.Min.Add(b.Max).Scale(0.5)
}

func NewSphere(center math.Vec3, radius float32) *Sphere {
	var s Sphere

	s.Center = center
	s.Radius = radius

	return &s
}
