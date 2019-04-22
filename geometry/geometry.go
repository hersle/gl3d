package geometry

import (
	gomath "math"
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

type Frustum struct {
	Org math.Vec3
	Dir math.Vec3
	Up math.Vec3
	Right math.Vec3
	NearDist float32
	FarDist float32
	NearWidth float32
	NearHeight float32
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

func (s *Sphere) Geometry(n int) *object.Geometry {
	// TODO: fix tangent/bitangent artifacts on top and bottom

	var geo object.Geometry

	var v object.Vertex

	// top
	v.Position = s.Center.Add(math.NewVec3(0, 0, +s.Radius))
	v.TexCoord = math.NewVec2(0, 1)
	geo.Verts = append(geo.Verts, v)

	// middle
	for i := 1; i < n; i++ {
		ang1 := float64(i) / float64(n) * (gomath.Pi)
		z := s.Center.Z() + s.Radius * float32(gomath.Cos(ang1))
		u := 1 - float32(i) / float32(n)
		for j := 0; j < 2 * n; j++ {
			ang2 := float64(j) / float64(2 * n) * (2 * gomath.Pi)
			x := s.Center.X() + s.Radius * float32(gomath.Sin(ang1) * gomath.Cos(ang2))
			y := s.Center.Y() + s.Radius * float32(gomath.Sin(ang1) * gomath.Sin(ang2))
			vv := float32(j) / float32(2 * n)
			v.Position = math.NewVec3(x, y, z)
			v.TexCoord = math.NewVec2(u, vv)
			geo.Verts = append(geo.Verts, v)
		}
	}

	// bottom
	v.Position = s.Center.Add(math.NewVec3(0, 0, -s.Radius))
	v.TexCoord = math.NewVec2(0, 0)
	geo.Verts = append(geo.Verts, v)

	var i1, i2, i3, i4 int

	// top
	i1 = 0
	for i := 0; i < 2 * n; i++ {
		i2 = 1 + (i + 0) % (2 * n)
		i3 = 1 + (i + 1) % (2 * n)
		geo.Faces = append(geo.Faces, int32(i1), int32(i2), int32(i3))
		geo.Inds += 3
	}

	// middle
	for i := 2; i < n; i++ {
		for j := 0; j < 2 * n; j++ {
			i1 = 1 + (i - 1) * 2 * n + (j + 0) % (2 * n)
			i2 = 1 + (i - 1) * 2 * n + (j + 1) % (2 * n)
			i3 = 1 + (i - 2) * 2 * n + (j + 1) % (2 * n)
			i4 = 1 + (i - 2) * 2 * n + (j + 0) % (2 * n)
			geo.Faces = append(geo.Faces, int32(i1), int32(i2), int32(i3))
			geo.Faces = append(geo.Faces, int32(i3), int32(i4), int32(i1))
			geo.Inds += 6
		}
	}

	// bottom
	i1 = len(geo.Verts) - 1 // bottom
	for i := 0; i < 2 * n; i++ {
		i3 = 1 + (n - 2) * (2 * n) + (i + 0) % (2 * n)
		i2 = 1 + (n - 2) * (2 * n) + (i + 1) % (2 * n)
		geo.Faces = append(geo.Faces, int32(i1), int32(i2), int32(i3))
		geo.Inds += 3
	}

	geo.CalculateNormals()
	geo.CalculateTangents()

	return &geo
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

func NewFrustum(org, dir, up math.Vec3, nearDist, farDist, nearWidth, nearHeight float32) *Frustum {
	var f Frustum
	f.Org = org
	f.Dir = dir.Norm()
	f.Up = up.Norm()
	f.Right = f.Dir.Cross(f.Up).Norm()
	f.NearDist = nearDist
	f.FarDist = farDist
	f.NearWidth = nearWidth
	f.NearHeight = nearHeight
	return &f
}

func (f *Frustum) Geometry() *object.Geometry {
	var geo object.Geometry

	nearCenter := f.Org.Add(f.Dir.Scale(f.NearDist))
	nearBottomLeft := nearCenter.Add(f.Right.Scale(-f.NearWidth / 2)).Add(f.Up.Scale(-f.NearHeight / 2))
	nearBottomRight := nearCenter.Add(f.Right.Scale(+f.NearWidth / 2)).Add(f.Up.Scale(-f.NearHeight / 2))
	nearTopRight := nearCenter.Add(f.Right.Scale(+f.NearWidth / 2)).Add(f.Up.Scale(+f.NearHeight / 2))
	nearTopLeft := nearCenter.Add(f.Right.Scale(-f.NearWidth / 2)).Add(f.Up.Scale(+f.NearHeight / 2))

	farWidth := (f.FarDist / f.NearDist) * f.NearWidth
	farHeight := (f.FarDist / f.NearDist) * f.NearHeight

	farCenter := f.Org.Add(f.Dir.Scale(f.FarDist))
	farBottomLeft := farCenter.Add(f.Right.Scale(-farWidth / 2)).Add(f.Up.Scale(-farHeight / 2))
	farBottomRight := farCenter.Add(f.Right.Scale(+farWidth / 2)).Add(f.Up.Scale(-farHeight / 2))
	farTopRight := farCenter.Add(f.Right.Scale(+farWidth / 2)).Add(f.Up.Scale(+farHeight / 2))
	farTopLeft := farCenter.Add(f.Right.Scale(-farWidth / 2)).Add(f.Up.Scale(+farHeight / 2))

	p := []math.Vec3{}
	p = append(p, farBottomLeft) // p1
	p = append(p, farBottomRight) // p2
	p = append(p, farTopRight) // p3
	p = append(p, farTopLeft) // p4
	p = append(p, nearBottomLeft) // p5
	p = append(p, nearBottomRight) // p6
	p = append(p, nearTopRight) // p7
	p = append(p, nearTopLeft) // p8

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

	return &geo
}
