package object

import (
	"github.com/hersle/gl3d/math"
)

type Object struct {
	Position            math.Vec3 // translation
	UnitX, UnitY, UnitZ math.Vec3 // orientation
	Scaling             math.Vec3 // scale

	DirtyWorldMatrix bool
	worldMatrix      math.Mat4
}

func (o *Object) Reset() {
	o.Place(math.NewVec3(0, 0, 0))
	o.Orient(math.NewVec3(1, 0, 0), math.NewVec3(0, 1, 0))
	o.SetScale(math.NewVec3(1, 1, 1))
}

func (o *Object) updateUnitZVector() {
	o.UnitZ = o.UnitX.Cross(o.UnitY).Norm()
}

func (o *Object) updateWorldMatrix() {
	o.worldMatrix.Identity()
	o.worldMatrix.MultTranslation(o.Position)
	o.worldMatrix.MultOrientation(o.UnitX, o.UnitY, o.UnitZ)
	o.worldMatrix.MultScaling(o.Scaling)
	o.DirtyWorldMatrix = false
}

func (o *Object) WorldMatrix() *math.Mat4 {
	if o.DirtyWorldMatrix {
		o.updateWorldMatrix()
	}
	return &o.worldMatrix
}

func (o *Object) Place(position math.Vec3) {
	o.Position = position
	o.DirtyWorldMatrix = true
}

func (o *Object) Translate(displacement math.Vec3) {
	o.Place(o.Position.Add(displacement))
}

func (o *Object) Orient(unitX, unitY math.Vec3) {
	o.UnitX = unitX.Norm()
	o.UnitY = unitY.Norm()
	o.updateUnitZVector()
	o.DirtyWorldMatrix = true
}

func (o *Object) Rotate(axis math.Vec3, ang float32) {
	o.Orient(o.UnitX.Rotate(axis, ang), o.UnitY.Rotate(axis, ang))
}

func (o *Object) RotateX(ang float32) {
	o.Rotate(math.NewVec3(1, 0, 0), ang)
}

func (o *Object) RotateY(ang float32) {
	o.Rotate(math.NewVec3(0, 1, 0), ang)
}

func (o *Object) RotateZ(ang float32) {
	o.Rotate(math.NewVec3(0, 0, 1), ang)
}

func (o *Object) SetScale(scaling math.Vec3) {
	o.Scaling = scaling
	o.DirtyWorldMatrix = true
}

func (o *Object) Scale(factor math.Vec3) {
	o.SetScale(o.Scaling.Mult(factor))
}