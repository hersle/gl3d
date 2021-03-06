package math

import (
	"fmt"
	"math"
)

type Vec2 [2]float32
type Vec3 [3]float32
type Vec4 [4]float32

func NewVec2(x, y float32) Vec2 {
	var a Vec2
	a[0] = x
	a[1] = y
	return a
}

func NewVec3(x, y, z float32) Vec3 {
	var a Vec3
	a[0] = x
	a[1] = y
	a[2] = z
	return a
}

func NewVec4(x, y, z, w float32) Vec4 {
	var a Vec4
	a[0] = x
	a[1] = y
	a[2] = z
	a[3] = w
	return a
}

func (a Vec2) X() float32 {
	return a[0]
}

func (a Vec3) X() float32 {
	return a[0]
}

func (a Vec4) X() float32 {
	return a[0]
}

func (a Vec2) Y() float32 {
	return a[1]
}

func (a Vec3) Y() float32 {
	return a[1]
}

func (a Vec4) Y() float32 {
	return a[1]
}

func (a Vec3) Z() float32 {
	return a[2]
}

func (a Vec4) Z() float32 {
	return a[2]
}

func (a Vec4) W() float32 {
	return a[3]
}

func (a Vec2) Vec3(z float32) Vec3 {
	return Vec3{a.X(), a.Y(), z}
}

func (a Vec3) Vec4(w float32) Vec4 {
	return Vec4{a.X(), a.Y(), a.Z(), w}
}

func (a Vec3) Vec2() Vec2 {
	return Vec2{a.X(), a.Y()}
}

func (a Vec4) Vec3() Vec3 {
	return Vec3{a.X(), a.Y(), a.Z()}
}

func (a Vec2) Add(b Vec2) Vec2 {
	return Vec2{a.X() + b.X(), a.Y() + b.Y()}
}

func (a Vec3) Add(b Vec3) Vec3 {
	return Vec3{a.X() + b.X(), a.Y() + b.Y(), a.Z() + b.Z()}
}

func (a Vec4) Add(b Vec4) Vec4 {
	return Vec4{a.X() + b.X(), a.Y() + b.Y(), a.Z() + b.Z(), a.W() + b.W()}
}

func (a Vec2) Scale(factor float32) Vec2 {
	return Vec2{factor * a.X(), factor * a.Y()}
}

func (a Vec3) Scale(factor float32) Vec3 {
	return Vec3{factor * a.X(), factor * a.Y(), factor * a.Z()}
}

func (a Vec4) Scale(factor float32) Vec4 {
	return Vec4{factor * a.X(), factor * a.Y(), factor * a.Z(), factor * a.W()}
}

func (a Vec2) Sub(b Vec2) Vec2 {
	return Vec2{a.X() - b.X(), a.Y() - b.Y()}
}

func (a Vec3) Sub(b Vec3) Vec3 {
	return Vec3{a.X() - b.X(), a.Y() - b.Y(), a.Z() - b.Z()}
}

func (a Vec4) Sub(b Vec4) Vec4 {
	return Vec4{a.X() - b.X(), a.Y() - b.Y(), a.Z() - b.Z(), a.W() - b.W()}
}

func (a Vec2) Mult(b Vec2) Vec2 {
	return Vec2{a.X() * b.X(), a.Y() * b.Y()}
}

func (a Vec3) Mult(b Vec3) Vec3 {
	return Vec3{a.X() * b.X(), a.Y() * b.Y(), a.Z() * b.Z()}
}

func (a Vec4) Mult(b Vec4) Vec4 {
	return Vec4{a.X() * b.X(), a.Y() * b.Y(), a.Z() * b.Z(), a.W() * b.W()}
}

func (a Vec2) Dot(b Vec2) float32 {
	return a.X()*b.X() + a.Y()*b.Y()
}

func (a Vec3) Dot(b Vec3) float32 {
	return a.X()*b.X() + a.Y()*b.Y() + a.Z()*b.Z()
}

func (a Vec4) Dot(b Vec4) float32 {
	return a.X()*b.X() + a.Y()*b.Y() + a.Z()*b.Z() + a.W()*b.W()
}

func (a Vec2) Length() float32 {
	return float32(math.Sqrt(float64(a.Dot(a))))
}

func (a Vec3) Length() float32 {
	return float32(math.Sqrt(float64(a.Dot(a))))
}

func (a Vec4) Length() float32 {
	return float32(math.Sqrt(float64(a.Dot(a))))
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
	x := a.Y()*b.Z() - a.Z()*b.Y()
	y := a.Z()*b.X() - a.X()*b.Z()
	z := a.X()*b.Y() - a.Y()*b.X()
	return Vec3{x, y, z}
}

func (a Vec3) Normal() Vec3 {
	if a.Dot(a) == 0 {
		return Vec3{0, 0, 0} // zero vector has no normal
	}

	// one of these must be partly normal to a
	n1 := Vec3{1, 0, 0}
	n2 := Vec3{0, 1, 0}

	dot1 := a.Dot(n1)
	dot2 := a.Dot(n2)

	var n Vec3
	if dot1*dot1 < dot2*dot2 {
		n = n1 // n1 most normal to a
	} else {
		n = n2 // n2 most normal to a
	}

	// subtract part of n that is parallell to a
	a = a.Norm()
	n = n.Sub(a.Scale(n.Dot(a))).Norm()

	return n
}

func (a Vec3) Rotate(axis Vec3, ang float32) Vec3 {
	axis = axis.Norm()
	cos := float32(math.Cos(float64(ang)))
	sin := float32(math.Sin(float64(ang)))
	v1 := axis.Scale((1 - cos) * a.Dot(axis))
	v2 := a.Scale(cos)
	v3 := axis.Cross(a).Scale(sin)
	return v1.Add(v2).Add(v3)
}

func (a Vec4) Transform(m *Mat4) Vec4 {
	x := m.Row(0).Dot(a)
	y := m.Row(1).Dot(a)
	z := m.Row(2).Dot(a)
	w := m.Row(3).Dot(a)
	return Vec4{x, y, z, w}
}

func (a Vec2) String() string {
	return fmt.Sprintf("(%.2f, %.2f)", a.X(), a.Y())
}

func (a Vec3) String() string {
	return fmt.Sprintf("(%.2f, %.2f, %.2f)", a.X(), a.Y(), a.Z())
}

func (a Vec4) String() string {
	return fmt.Sprintf("(%.2f, %.2f, %.2f, %.2f)", a.X(), a.Y(), a.Z(), a.W())
}
