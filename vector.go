package main

import (
	"math"
	"fmt"
)

type Vec4 struct {
	X, Y, Z, W float64
}

func NewVec4(x, y, z, w float64) Vec4 {
	var a Vec4
	a.X = x
	a.Y = y
	a.Z = z
	a.W = w
	return a
}

func (a Vec4) Vec3() Vec3 {
	return NewVec3(a.X, a.Y, a.Z)
}

func (a Vec4) Add(b Vec4) Vec4 {
	return NewVec4(a.X + b.X, a.Y + b.Y, a.Z + b.Z, a.W + b.W)
}

func (a Vec4) Scale(factor float64) Vec4 {
	return NewVec4(factor * a.X, factor * a.Y, factor * a.Z, factor * a.W)
}

func (a Vec4) Sub(b Vec4) Vec4 {
	return a.Add(b.Scale(-1))
}

func (a Vec4) Dot(b Vec4) float64 {
	return a.X * b.X + a.Y * b.Y + a.Z * b.Z + a.W * b.W
}

func (a Vec4) Length() float64 {
	return math.Sqrt(a.Dot(a))
}

func (a Vec4) Norm() Vec4 {
	return a.Scale(1 / a.Length())
}

func (a Vec4) String() string {
	return fmt.Sprintf("(%v, %v, %v, %v)", a.X, a.Y, a.Z, a.W)
}



type Vec3 struct {
	X, Y, Z float64
}

func NewVec3(x, y, z float64) Vec3 {
	var a Vec3
	a.X = x
	a.Y = y
	a.Z = z
	return a
}

func (a Vec3) Vec4(w float64) Vec4 {
	return NewVec4(a.X, a.Y, a.Z, w)
}

func (a Vec3) Add(b Vec3) Vec3 {
	return NewVec3(a.X + b.X, a.Y + b.Y, a.Z + b.Z)
}

func (a Vec3) Scale(factor float64) Vec3 {
	return NewVec3(factor * a.X, factor * a.Y, factor * a.Z)
}

func (a Vec3) Sub(b Vec3) Vec3 {
	return a.Add(b.Scale(-1))
}

func (a Vec3) Dot(b Vec3) float64 {
	return a.X * b.X + a.Y * b.Y + a.Z * b.Z
}

func (a Vec3) Length() float64 {
	return math.Sqrt(a.Dot(a))
}

func (a Vec3) Norm() Vec3 {
	return a.Scale(1 / a.Length())
}

func (a Vec3) Cross(b Vec3) Vec3 {
	x := a.Y * b.Z - a.Z * b.Y
	y := a.Z * b.X - a.X * b.Z
	z := a.X * b.Y - a.Y * b.X
	return NewVec3(x, y, z)
}

func (a Vec3) Rotate(axis Vec3, ang float64) Vec3 {
	axis = axis.Norm()
	v1 := axis.Scale((1 - math.Cos(ang)) * a.Dot(axis))
	v2 := a.Scale(math.Cos(ang))
	v3 := axis.Cross(a).Scale(math.Sin(ang))
	return v1.Add(v2).Add(v3)
}

func (a Vec3) String() string {
	return fmt.Sprintf("(%v, %v, %v)", a.X, a.Y, a.Z)
}
