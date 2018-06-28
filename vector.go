package main

import (
	"math"
	"fmt"
)

type Vec2 [2]float64
type Vec3 [3]float64
type Vec4 [4]float64

func NewVec2(x, y float64) Vec2 {
	var a Vec2
	a[0] = x
	a[1] = y
	return a
}

func NewVec3(x, y, z float64) Vec3 {
	var a Vec3
	a[0] = x
	a[1] = y
	a[2] = z
	return a
}

func NewVec4(x, y, z, w float64) Vec4 {
	var a Vec4
	a[0] = x
	a[1] = y
	a[2] = z
	a[3] = w
	return a
}

func (a Vec2) X() float64 {
	return a[0]
}

func (a Vec3) X() float64 {
	return a[0]
}

func (a Vec4) X() float64 {
	return a[0]
}

func (a Vec2) Y() float64 {
	return a[0]
}

func (a Vec3) Y() float64 {
	return a[1]
}

func (a Vec4) Y() float64 {
	return a[1]
}

func (a Vec3) Z() float64 {
	return a[2]
}

func (a Vec4) Z() float64 {
	return a[2]
}

func (a Vec4) W() float64 {
	return a[3]
}

func (a Vec2) Vec3(z float64) Vec3 {
	return NewVec3(a.X(), a.Y(), z)
}

func (a Vec3) Vec4(w float64) Vec4 {
	return NewVec4(a.X(), a.Y(), a.Z(), w)
}

func (a Vec3) Vec2() Vec2 {
	return NewVec2(a.X(), a.Y())
}

func (a Vec4) Vec3() Vec3 {
	return NewVec3(a.X(), a.Y(), a.Z())
}

func (a Vec2) Add(b Vec2) Vec2 {
	return NewVec2(a.X() + b.X(), a.Y() + b.Y())
}

func (a Vec3) Add(b Vec3) Vec3 {
	return NewVec3(a.X() + b.X(), a.Y() + b.Y(), a.Z() + b.Z())
}

func (a Vec4) Add(b Vec4) Vec4 {
	return NewVec4(a.X() + b.X(), a.Y() + b.Y(), a.Z() + b.Z(), a.W() + b.W())
}

func (a Vec2) Scale(factor float64) Vec2 {
	return NewVec2(factor * a.X(), factor * a.Y())
}

func (a Vec3) Scale(factor float64) Vec3 {
	return NewVec3(factor * a.X(), factor * a.Y(), factor * a.Z())
}

func (a Vec4) Scale(factor float64) Vec4 {
	return NewVec4(factor * a.X(), factor * a.Y(), factor * a.Z(), factor * a.W())
}

func (a Vec2) Sub(b Vec2) Vec2 {
	return a.Add(b.Scale(-1))
}

func (a Vec3) Sub(b Vec3) Vec3 {
	return a.Add(b.Scale(-1))
}

func (a Vec4) Sub(b Vec4) Vec4 {
	return a.Add(b.Scale(-1))
}

func (a Vec2) Dot(b Vec2) float64 {
	return a.X() * b.X() + a.Y() * b.Y()
}

func (a Vec3) Dot(b Vec3) float64 {
	return a.X() * b.X() + a.Y() * b.Y() + a.Z() * b.Z()
}

func (a Vec4) Dot(b Vec4) float64 {
	return a.X() * b.X() + a.Y() * b.Y() + a.Z() * b.Z() + a.W() * b.W()
}

func (a Vec2) Length() float64 {
	return math.Sqrt(a.Dot(a))
}

func (a Vec3) Length() float64 {
	return math.Sqrt(a.Dot(a))
}

func (a Vec4) Length() float64 {
	return math.Sqrt(a.Dot(a))
}

func (a Vec2) Norm() Vec2 {
	return a.Scale(1 / a.Length())
}

func (a Vec3) Norm() Vec3 {
	return a.Scale(1 / a.Length())
}

func (a Vec4) Norm() Vec4 {
	return a.Scale(1 / a.Length())
}

func (a Vec3) Cross(b Vec3) Vec3 {
	x := a.Y() * b.Z() - a.Z() * b.Y()
	y := a.Z() * b.X() - a.X() * b.Z()
	z := a.X() * b.Y() - a.Y() * b.X()
	return NewVec3(x, y, z)
}

func (a Vec3) Rotate(axis Vec3, ang float64) Vec3 {
	axis = axis.Norm()
	v1 := axis.Scale((1 - math.Cos(ang)) * a.Dot(axis))
	v2 := a.Scale(math.Cos(ang))
	v3 := axis.Cross(a).Scale(math.Sin(ang))
	return v1.Add(v2).Add(v3)
}

func (a Vec2) String() string {
	return fmt.Sprintf("(%v, %v)", a.X(), a.Y())
}

func (a Vec3) String() string {
	return fmt.Sprintf("(%v, %v, %v)", a.X(), a.Y(), a.Z())
}

func (a Vec4) String() string {
	return fmt.Sprintf("(%v, %v, %v, %v)", a.X(), a.Y(), a.Z(), a.W())
}
