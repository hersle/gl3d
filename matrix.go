package main

// TODO: translation matrix, rotation matrix, other transformation matrices
// TODO: Mat3

import (
	// "math"
)

type Mat4 [4*4]float64

var dummyMat Mat4

func NewMat4Zero() *Mat4 {
	return &Mat4{}
}

func NewMat4Identity() *Mat4 {
	return NewMat4Zero().Identity()
}

func (a *Mat4) Zero() *Mat4 {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			a.Set(i, j, 0)
		}
	}
	return a
}

func (a *Mat4) Identity() *Mat4 {
	a.Zero()
	a.Set(0, 0, 1)
	a.Set(1, 1, 1)
	a.Set(2, 2, 1)
	a.Set(3, 3, 1)
	return a
}

func (a *Mat4) Translation(d Vec3) *Mat4 {
	a.Identity()
	a.SetCol(3, d.Vec4(1))
	return a
}

func (a *Mat4) MultTranslation(d Vec3) *Mat4 {
	return a.Mult(dummyMat.Translation(d))
}

func (a *Mat4) Scaling(factorX, factorY, factorZ float64) *Mat4 {
	a.Identity()
	a.Set(0, 0, factorX)
	a.Set(1, 1, factorY)
	a.Set(2, 2, factorZ)
	return a
}

func (a *Mat4) MultScaling(factorX, factorY, factorZ float64) *Mat4 {
	return a.Mult(dummyMat.Scaling(factorX, factorY, factorZ))
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
	a.Set(3, j, col.X)
}

func (a *Mat4) Row(i int) Vec4 {
	return NewVec4(a.At(i, 0), a.At(i, 1), a.At(i, 2), a.At(i, 3))
}

func (a *Mat4) SetRow(i int, row Vec4) {
	a.Set(i, 0, row.X)
	a.Set(i, 1, row.Y)
	a.Set(i, 2, row.Z)
	a.Set(i, 3, row.X)
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
		ai := a.Row(i)
		for j := 0; j < 4; j++ {
			bj := b.Col(j)
			a.Set(i, j, ai.Dot(bj))
		}
	}
	return a
}

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
