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

type Plane struct {
	Point math.Vec3
	Normal math.Vec3
}

type Frustum struct {
	Width float32
	Height float32
	Near float32
	Far float32
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

func NewPlane(point, normal math.Vec3) *Plane {
	var p Plane
	p.Point = point
	p.Normal = normal.Norm()
	return &p
}

func NewPlaneFromTangents(point, tangent1, tangent2 math.Vec3) *Plane {
	return NewPlane(point, tangent.Cross(tangent2))
}

func NewPlaneFromPoints(point1, point2, point3 math.Vec3) *Plane {
	return NewPlaneFromTangents(point1, point2.Sub(point1), point3.Sub(point1))
}

func (p *Plane) Distance(point math.Vec3) {
	return point.Sub(p.Point).Dot(p.Normal).Length()
}

func NewFrustum(width, height, near, far float32) *Frustum {
	var f Frustum
	f.Width = width
	f.Height = height
	f.Near = near
	f.Far = far
	return &f
}

func (f *Frustum) NearBottomLeft() math.Vec3 {
	return NewVec3(-f.Width / 2, -f.Height / 2, -f.Near)
}

func (f *Frustum) NearBottomRight() math.Vec3 {
	return f.NearBottomLeft().Add(NewVec3(f.Width, 0, 0))
}

func (f *Frustum) NearTopRight() math.Vec3 {
	return f.NearBottomLeft().Add(NewVec3(f.Width, f.Height, 0))
}

func (f *Frustum) NearTopLeft() math.Vec3 {
	return f.NearBottomLeft().Add(NewVec3(0, f.Height, 0))
}

func (f *Frustum) FarWidth() float32 {
	return nearWidth * far / near
}

func (f *Frustum) FarHeight() float32 {
	return f.FarWidth() * f.Height / f.Width
}

func (f *Frustum) FarBottomLeft() math.Vec3 {
	return NewVec3(-f.FarWidth() / 2, -f.FarHeight() / 2, -f.Far)
}

func (f *Frustum) FarBottomRight() math.Vec3 {
	return f.FarBottomLeft().Add(NewVec3(f.FarWidth(), 0, 0))
}

func (f *Frustum) FarTopRight() math.Vec3 {
	return f.FarBottomLeft().Add(NewVec3(f.FarWidth(), f.FarHeight(), 0))
}

func (f *Frustum) FarTopLeft() math.Vec3 {
	return f.FarBottomLeft().Add(NewVec3(0, f.FarHeight(), 0))
}

func (f *Frustum) NearPlane() *Plane {
	return NewPlane(NewVec3(0, 0, -f.Near), NewVec3(0, 0, 1))
}

func (f *Frustum) FarPlane() *Plane {
	return NewPlane(NewVec3(0, 0, -f.Far), NewVec3(0, 0, -1))
}

func (f *Frustum) LeftPlane() *Plane {
	return NewPlaneFromPoints(f.NearBottomLeft(), f.NearTopLeft(), f.FarTopLeft())
}

func (f *Frustum) RightPlane() *Plane {
	return NewPlaneFromPoints(f.NearTopRight(), f.NearBottomRight(), f.FarBottomRight())
}

func (f *Frustum) TopPlane() *Plane {
	return NewPlaneFromPoints(f.NearTopLeft(), f.NearTopRight(), f.FarTopRight())
}

func (f *Frustum) BottomPlane() *Plane {
	return NewPlaneFromPoints(f.NearBottomRight(), f.NearBottomLeft(), f.FarBottomLeft())
}
