package geometry

import (
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
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
	Point  math.Vec3
	Normal math.Vec3
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

func (b *Box) Geometry() *object.Geometry {
	var geo object.Geometry

	p1 := math.NewVec3(b.Min.X(), b.Min.Y(), b.Min.Z())
	p2 := math.NewVec3(b.Max.X(), b.Min.Y(), b.Min.Z())
	p3 := math.NewVec3(b.Max.X(), b.Max.Y(), b.Min.Z())
	p4 := math.NewVec3(b.Min.X(), b.Max.Y(), b.Min.Z())
	p5 := math.NewVec3(b.Min.X(), b.Min.Y(), b.Max.Z())
	p6 := math.NewVec3(b.Max.X(), b.Min.Y(), b.Max.Z())
	p7 := math.NewVec3(b.Max.X(), b.Max.Y(), b.Max.Z())
	p8 := math.NewVec3(b.Min.X(), b.Max.Y(), b.Max.Z())
	p := []math.Vec3{p1, p2, p3, p4, p5, p6, p7, p8}

	pi := [][]int{
		{5, 6, 7, 8},
		{6, 2, 3, 7},
		{2, 1, 4, 3},
		{1, 5, 8, 4},
		{8, 7, 3, 4},
		{6, 5, 1, 2},
	}

	var v1, v2, v3, v4 object.Vertex
	for i := 0; i < 6; i++ {
		v1.Position = p[pi[i][0]-1]
		v2.Position = p[pi[i][1]-1]
		v3.Position = p[pi[i][2]-1]
		v4.Position = p[pi[i][3]-1]
		geo.AddTriangle(v1, v2, v3)
		geo.AddTriangle(v1, v3, v4)
	}

	geo.CalculateNormals()
	geo.CalculateTangents()

	return &geo
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
	return NewPlane(point, tangent1.Cross(tangent2))
}

func NewPlaneFromPoints(point1, point2, point3 math.Vec3) *Plane {
	return NewPlaneFromTangents(point1, point2.Sub(point1), point3.Sub(point1))
}

func (p *Plane) Distance(point math.Vec3) float32 {
	return point.Sub(p.Point).Dot(p.Normal)
}
