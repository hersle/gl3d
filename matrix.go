package main

import (
	"math"
	"fmt"
)

type Mat4 [4*4]float64

func NewMat4Zero() *Mat4 {
	var a Mat4
	a.Zero()
	return &a
}

func NewMat4Identity() *Mat4 {
	return NewMat4Zero().Identity()
}

func (a *Mat4) index(i, j int) int {
	return i * 4 + j
}

func (a *Mat4) At(i, j int) float64 {
	return a[a.index(i, j)]
}

func (a *Mat4) Set(i, j int, aij float64) {
	a[a.index(i, j)] = aij
}

func (a *Mat4) Col(j int) Vec4 {
	return NewVec4(a.At(0, j), a.At(1, j), a.At(2, j), a.At(3, j))
}

func (a *Mat4) SetCol(j int, col Vec4) {
	a.Set(0, j, col.X)
	a.Set(1, j, col.Y)
	a.Set(2, j, col.Z)
	a.Set(3, j, col.W)
}

func (a *Mat4) Row(i int) Vec4 {
	return NewVec4(a.At(i, 0), a.At(i, 1), a.At(i, 2), a.At(i, 3))
}

func (a *Mat4) SetRow(i int, row Vec4) {
	a.Set(i, 0, row.X)
	a.Set(i, 1, row.Y)
	a.Set(i, 2, row.Z)
	a.Set(i, 3, row.W)
}

func (a *Mat4) Add(b *Mat4) *Mat4 {
	for i := 0; i < 4; i++ {
		a.SetRow(i, a.Row(i).Add(b.Row(i)))
	}
	return a
}

func (a *Mat4) Scale(factor float64) *Mat4 {
	for i := 0; i < 4; i++ {
		a.SetRow(i, a.Row(i).Scale(factor))
	}
	return a
}

func (a *Mat4) Sub(b *Mat4) *Mat4 {
	a.Add(b.Scale(-1))
	b.Scale(-1) // leave b unchanged
	return a
}

func (a *Mat4) Mult(b *Mat4) *Mat4 {
	for i := 0; i < 4; i++ {
		aRow := a.Row(i)
		for j := 0; j < 4; j++ {
			bCol := b.Col(j)
			a.Set(i, j, aRow.Dot(bCol))
		}
	}
	return a
}

// TODO: replace with func (a Vec3/4) Transform(m *Mat4)?
func (a *Mat4) MultVec(v Vec4) Vec4 {
	x := a.Row(0).Dot(v)
	y := a.Row(1).Dot(v)
	z := a.Row(2).Dot(v)
	w := a.Row(3).Dot(v)
	return NewVec4(x, y, z, w)
}

func (a *Mat4) Transpose() *Mat4 {
	r0, r1, r2, r3 := a.Row(0), a.Row(1), a.Row(2), a.Row(3)
	a.SetCol(0, r0)
	a.SetCol(1, r1)
	a.SetCol(2, r2)
	a.SetCol(3, r3)
	return a
}

func (a *Mat4) Zero() *Mat4 {
	a.SetRow(0, NewVec4(0, 0, 0, 0))
	a.SetRow(1, NewVec4(0, 0, 0, 0))
	a.SetRow(2, NewVec4(0, 0, 0, 0))
	a.SetRow(3, NewVec4(0, 0, 0, 0))
	return a
}

func (a *Mat4) Identity() *Mat4 {
	a.SetRow(0, NewVec4(1, 0, 0, 0))
	a.SetRow(1, NewVec4(0, 1, 0, 0))
	a.SetRow(2, NewVec4(0, 0, 1, 0))
	a.SetRow(3, NewVec4(0, 0, 0, 1))
	return a
}

func (a *Mat4) Translation(d Vec3) *Mat4 {
	a.Identity()
	a.SetCol(3, d.Vec4(1))
	return a
}

func (a *Mat4) Scaling(factorX, factorY, factorZ float64) *Mat4 {
	a.SetRow(0, NewVec4(factorX, 0, 0, 0))
	a.SetRow(1, NewVec4(0, factorY, 0, 0))
	a.SetRow(2, NewVec4(0, 0, factorZ, 0))
	a.SetRow(3, NewVec4(0, 0, 0, 1))
	return a
}

func (a *Mat4) RotationX(ang float64) *Mat4 {
	a.SetCol(0, NewVec4(1, 0, 0, 0))
	a.SetCol(1, NewVec4(0, math.Cos(ang), math.Sin(ang), 0))
	a.SetCol(2, NewVec4(0, -math.Sin(ang), math.Cos(ang), 0))
	a.SetCol(3, NewVec4(0, 0, 0, 1))
	return a
}

func (a *Mat4) RotationY(ang float64) *Mat4 {
	a.SetCol(0, NewVec4(math.Cos(ang), 0, -math.Sin(ang), 0))
	a.SetCol(1, NewVec4(0, 1, 0, 0))
	a.SetCol(2, NewVec4(math.Sin(ang), 0, math.Cos(ang), 0))
	a.SetCol(3, NewVec4(0, 0, 0, 1))
	return a
}

func (a *Mat4) RotationZ(ang float64) *Mat4 {
	a.SetCol(0, NewVec4(math.Cos(ang), math.Sin(ang), 0, 0))
	a.SetCol(1, NewVec4(-math.Sin(ang), math.Cos(ang), 0, 0))
	a.SetCol(2, NewVec4(0, 0, 1, 0))
	a.SetCol(3, NewVec4(0, 0, 0, 1))
	return a
}

func (a *Mat4) OrthoCentered(size Vec3) *Mat4 {
	return a.Scaling(2 / size.X, 2 / size.Y, -2 / size.Z)
}

func (a *Mat4) Frustum(l, b, r, t, n, f float64) *Mat4 {
	a.Zero()
	a.Set(0, 0, 2 * n / (r - l))
	a.Set(1, 1, 2 * n / (t - b))
	a.Set(0, 2, (r + l) / (r - l))
	a.Set(1, 2, (t + b) / (t - b))
	a.Set(2, 2, -(f + n) / (f - n))
	a.Set(2, 3, -2 * f * n / (f - n))
	a.Set(3, 2, -1)
	return a
}

func (a *Mat4) FrustumCentered(w, h, n, f float64) *Mat4 {
	return a.Frustum(-w / 2, -h / 2, +w / 2, +h / 2, n, f)
}

func (a *Mat4) Perspective(fovY, aspect, n, f float64) *Mat4 {
	h := 2 * n * math.Tan(fovY / 2)
	w := aspect * h
	return a.FrustumCentered(w, h, n, f)
}

func (a *Mat4) LookAt(eye, target, up Vec3) *Mat4 {
	fwd := target.Sub(eye).Norm()
	up = up.Norm()
	right := fwd.Cross(up).Norm()
	a.SetRow(0, right.Vec4(-right.Dot(eye)))
	a.SetRow(1, up.Vec4(-up.Dot(eye)))
	a.SetRow(2, fwd.Scale(-1).Vec4(+fwd.Dot(eye)))
	a.SetRow(3, NewVec4(0, 0, 0, 1))
	return a
}

func (a *Mat4) String() string {
	return fmt.Sprintf("%v\n%v\n%v\n%v\n", a.Row(0), a.Row(1), a.Row(2), a.Row(3))
}
