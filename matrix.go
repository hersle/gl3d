package main

import (
	"math"
	"fmt"
	"errors"
)

type Mat4 [4*4]float32

func NewMat4Zero() *Mat4 {
	var a Mat4
	a.Zero()
	return &a
}

func NewMat4Identity() *Mat4 {
	var a Mat4
	a.Identity()
	return &a
}

func (a *Mat4) index(i, j int) int {
	return i * 4 + j
}

func (a *Mat4) At(i, j int) float32 {
	return a[a.index(i, j)]
}

func (a *Mat4) Set(i, j int, aij float32) {
	a[a.index(i, j)] = aij
}

func (a *Mat4) Col(j int) Vec4 {
	return NewVec4(a.At(0, j), a.At(1, j), a.At(2, j), a.At(3, j))
}

func (a *Mat4) SetCol(j int, col Vec4) {
	a.Set(0, j, col.X())
	a.Set(1, j, col.Y())
	a.Set(2, j, col.Z())
	a.Set(3, j, col.W())
}

func (a *Mat4) Row(i int) Vec4 {
	return NewVec4(a.At(i, 0), a.At(i, 1), a.At(i, 2), a.At(i, 3))
}

func (a *Mat4) SetRow(i int, row Vec4) {
	a.Set(i, 0, row.X())
	a.Set(i, 1, row.Y())
	a.Set(i, 2, row.Z())
	a.Set(i, 3, row.W())
}

func (a *Mat4) Copy(b *Mat4) {
	for i := 0; i < 4; i++ {
		a.SetRow(i, b.Row(i))
	}
}

func (a *Mat4) Add(b *Mat4) {
	for i := 0; i < 4; i++ {
		a.SetRow(i, a.Row(i).Add(b.Row(i)))
	}
}

func (a *Mat4) Scale(factor float32) {
	for i := 0; i < 4; i++ {
		a.SetRow(i, a.Row(i).Scale(factor))
	}
}

func (a *Mat4) Sub(b *Mat4) {
	b.Scale(-1)
	a.Add(b)
	b.Scale(-1) // leave b unchanged
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

func (a *Mat4) MultRight(b *Mat4) *Mat4 {
	return a.Mult(b)
}

func (a *Mat4) MultLeft(b *Mat4) *Mat4 {
	b.MultRight(a)
	a.Copy(b)
	return a
}

func (a *Mat4) Transpose() {
	r0, r1, r2, r3 := a.Row(0), a.Row(1), a.Row(2), a.Row(3)
	a.SetCol(0, r0)
	a.SetCol(1, r1)
	a.SetCol(2, r2)
	a.SetCol(3, r3)
}

func (a *Mat4) Zero() {
	a.SetRow(0, NewVec4(0, 0, 0, 0))
	a.SetRow(1, NewVec4(0, 0, 0, 0))
	a.SetRow(2, NewVec4(0, 0, 0, 0))
	a.SetRow(3, NewVec4(0, 0, 0, 0))
}

func (a *Mat4) Identity() {
	a.SetRow(0, NewVec4(1, 0, 0, 0))
	a.SetRow(1, NewVec4(0, 1, 0, 0))
	a.SetRow(2, NewVec4(0, 0, 1, 0))
	a.SetRow(3, NewVec4(0, 0, 0, 1))
}

func (a *Mat4) Translation(d Vec3) {
	a.Identity()
	a.SetCol(3, d.Vec4(1))
}

func (a *Mat4) Scaling(factor Vec3) {
	a.SetRow(0, NewVec4(factor.X(), 0, 0, 0))
	a.SetRow(1, NewVec4(0, factor.Y(), 0, 0))
	a.SetRow(2, NewVec4(0, 0, factor.Z(), 0))
	a.SetRow(3, NewVec4(0, 0, 0, 1))
}

func (a *Mat4) RotationX(ang float32) {
	cos := float32(math.Cos(float64(ang)))
	sin := float32(math.Sin(float64(ang)))
	a.SetCol(0, NewVec4(1, 0, 0, 0))
	a.SetCol(1, NewVec4(0, cos, sin, 0))
	a.SetCol(2, NewVec4(0, -sin, cos, 0))
	a.SetCol(3, NewVec4(0, 0, 0, 1))
}

func (a *Mat4) RotationY(ang float32) {
	cos := float32(math.Cos(float64(ang)))
	sin := float32(math.Sin(float64(ang)))
	a.SetCol(0, NewVec4(cos, 0, -sin, 0))
	a.SetCol(1, NewVec4(0, 1, 0, 0))
	a.SetCol(2, NewVec4(sin, 0, cos, 0))
	a.SetCol(3, NewVec4(0, 0, 0, 1))
}

func (a *Mat4) RotationZ(ang float32) {
	cos := float32(math.Cos(float64(ang)))
	sin := float32(math.Sin(float64(ang)))
	a.SetCol(0, NewVec4(cos, sin, 0, 0))
	a.SetCol(1, NewVec4(-sin, cos, 0, 0))
	a.SetCol(2, NewVec4(0, 0, 1, 0))
	a.SetCol(3, NewVec4(0, 0, 0, 1))
}

func (a *Mat4) OrthoCentered(size Vec3) {
	a.Scaling(NewVec3(2 / size.X(), 2 / size.Y(), -2 / size.Z()))
}

func (a *Mat4) Frustum(l, b, r, t, n, f float32) {
	a.SetRow(0, NewVec4(2 * n / (r - l), 0, (r + l) / (r - l), 0))
	a.SetRow(1, NewVec4(0, 2 * n / (t - b), (t + b) / (t - b), 0))
	a.SetRow(2, NewVec4(0, 0, -(f + n) / (f - n), -2 * f * n / (f - n)))
	a.SetRow(3, NewVec4(0, 0, -1, 0))
}

func (a *Mat4) FrustumCentered(w, h, n, f float32) {
	a.Frustum(-w / 2, -h / 2, +w / 2, +h / 2, n, f)
}

func (a *Mat4) Perspective(fovY, aspect, n, f float32) {
	h := 2 * n * float32(math.Tan(float64(fovY / 2)))
	w := aspect * h
	a.FrustumCentered(w, h, n, f)
}

func (a *Mat4) LookAt(eye, target, up Vec3) {
	fwd := target.Sub(eye).Norm()
	up = up.Norm()
	right := fwd.Cross(up).Norm()
	a.SetRow(0, right.Vec4(-right.Dot(eye)))
	a.SetRow(1, up.Vec4(-up.Dot(eye)))
	a.SetRow(2, fwd.Scale(-1).Vec4(+fwd.Dot(eye)))
	a.SetRow(3, NewVec4(0, 0, 0, 1))
}

func (a *Mat4) Determinant() float32 {
	a11, a12, a13, a14 := a.At(0, 0), a.At(0, 1), a.At(0, 2), a.At(0, 3)
	a21, a22, a23, a24 := a.At(1, 0), a.At(1, 1), a.At(1, 2), a.At(1, 3)
	a31, a32, a33, a34 := a.At(2, 0), a.At(2, 1), a.At(2, 2), a.At(2, 3)
	a41, a42, a43, a44 := a.At(3, 0), a.At(3, 1), a.At(3, 2), a.At(3, 3)

	// 2D determinans (aijkl = aij * akl - ail * akj)
	a3142 := a31 * a42 - a32 * a41
	a3143 := a31 * a43 - a33 * a41
	a3144 := a31 * a44 - a34 * a41
	a3243 := a32 * a43 - a33 * a42
	a3244 := a32 * a44 - a34 * a42
	a3344 := a33 * a44 - a34 * a43

	t1 := +1 * a11 * (a22 * a3344 - a23 * a3244 + a24 * a3243)
	t2 := -1 * a12 * (a21 * a3344 - a23 * a3144 + a24 * a3143)
	t3 := +1 * a13 * (a21 * a3244 - a22 * a3144 + a24 * a3142)
	t4 := -1 * a14 * (a21 * a3243 - a22 * a3143 + a23 * a3142)

	return t1 + t2 + t3 + t4
}

func (a *Mat4) Invert() error {
	det := a.Determinant()

	if det == 0 {
		return errors.New("cannot invert singular matrix")
	}

	// 1-indexed variable names
	a11, a12, a13, a14 := a.At(0, 0), a.At(0, 1), a.At(0, 2), a.At(0, 3)
	a21, a22, a23, a24 := a.At(1, 0), a.At(1, 1), a.At(1, 2), a.At(1, 3)
	a31, a32, a33, a34 := a.At(2, 0), a.At(2, 1), a.At(2, 2), a.At(2, 3)
	a41, a42, a43, a44 := a.At(3, 0), a.At(3, 1), a.At(3, 2), a.At(3, 3)

	b11 := (a22 * a33 * a44) + (a23 * a34 * a42) + (a24 * a32 * a43)
	b11 -= (a22 * a34 * a43) + (a23 * a32 * a44) + (a24 * a33 * a42)

	b12 := (a12 * a34 * a43) + (a13 * a32 * a44) + (a14 * a33 * a42)
	b12 -= (a12 * a33 * a44) + (a13 * a34 * a42) + (a14 * a32 * a43)

	b13 := (a12 * a23 * a44) + (a13 * a24 * a42) + (a14 * a22 * a43)
	b13 -= (a12 * a24 * a43) + (a13 * a22 * a44) + (a14 * a23 * a42)

	b14 := (a12 * a24 * a33) + (a13 * a22 * a34) + (a14 * a23 * a32)
	b14 -= (a12 * a23 * a34) + (a13 * a24 * a32) + (a14 * a22 * a33)

	b21 := (a21 * a34 * a43) + (a23 * a31 * a44) + (a24 * a33 * a41)
	b21 -= (a21 * a33 * a44) + (a23 * a34 * a41) + (a24 * a31 * a43)

	b22 := (a11 * a33 * a44) + (a13 * a34 * a41) + (a14 * a31 * a43)
	b22 -= (a11 * a34 * a43) + (a13 * a31 * a44) + (a14 * a33 * a41)

	b23 := (a11 * a24 * a43) + (a13 * a21 * a44) + (a14 * a23 * a41)
	b23 -= (a11 * a23 * a44) + (a13 * a24 * a41) + (a14 * a21 * a43)

	b24 := (a11 * a23 * a34) + (a13 * a24 * a31) + (a14 * a21 * a33)
	b24 -= (a11 * a24 * a33) + (a13 * a21 * a34) + (a14 * a23 * a31)

	b31 := (a21 * a32 * a44) + (a22 * a34 * a41) + (a24 * a31 * a42)
	b31 -= (a21 * a34 * a42) + (a22 * a31 * a44) + (a24 * a32 * a41)

	b32 := (a11 * a34 * a42) + (a12 * a31 * a44) + (a14 * a32 * a41)
	b32 -= (a11 * a32 * a44) + (a12 * a34 * a41) + (a14 * a31 * a42)

	b33 := (a11 * a22 * a44) + (a12 * a24 * a41) + (a14 * a21 * a42)
	b33 -= (a11 * a24 * a42) + (a12 * a21 * a44) + (a14 * a22 * a41)

	b34 := (a11 * a24 * a32) + (a12 * a21 * a34) + (a14 * a22 * a31)
	b34 -= (a11 * a22 * a34) + (a12 * a24 * a31) + (a14 * a21 * a32)

	b41 := (a21 * a33 * a42) + (a22 * a31 * a43) + (a23 * a32 * a41)
	b41 -= (a21 * a32 * a43) + (a22 * a33 * a41) + (a23 * a31 * a42)

	b42 := (a11 * a32 * a43) + (a12 * a33 * a41) + (a13 * a31 * a42)
	b42 -= (a11 * a33 * a42) + (a12 * a31 * a43) + (a13 * a32 * a41)

	b43 := (a11 * a23 * a42) + (a12 * a21 * a43) + (a13 * a22 * a41)
	b43 -= (a11 * a22 * a43) + (a12 * a23 * a41) + (a13 * a21 * a42)

	b44 := (a11 * a22 * a33) + (a12 * a23 * a31) + (a13 * a21 * a32)
	b44 -= (a11 * a23 * a32) + (a12 * a21 * a33) + (a13 * a22 * a31)

	a.SetRow(0, NewVec4(b11, b12, b13, b14))
	a.SetRow(1, NewVec4(b21, b22, b23, b24))
	a.SetRow(2, NewVec4(b31, b32, b33, b34))
	a.SetRow(3, NewVec4(b41, b42, b43, b44))
	a.Scale(1 / det)

	return nil
}

func (a *Mat4) String() string {
	return fmt.Sprintf("%v\n%v\n%v\n%v\n", a.Row(0), a.Row(1), a.Row(2), a.Row(3))
}
