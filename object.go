package main

import (
	"github.com/hersle/gl3d/math"
)

type Object struct {
	position math.Vec3 // translation
	unitX, unitY, unitZ math.Vec3 // orientation
	scale math.Vec3 // scale

	dirtyWorldMatrix bool
	worldMatrix math.Mat4
}

func (o *Object) Reset() {
	o.Place(math.NewVec3(0, 0, 0))
	o.Orient(math.NewVec3(1, 0, 0), math.NewVec3(0, 1, 0))
	o.SetScale(math.NewVec3(1, 1, 1))
}

func (o *Object) updateUnitZVector() {
	o.unitZ = o.unitX.Cross(o.unitY).Norm()
}

func (o *Object) updateWorldMatrix() {
	o.worldMatrix.Identity()
	o.worldMatrix.MultTranslation(o.position)
	o.worldMatrix.MultOrientation(o.unitX, o.unitY, o.unitZ)
	o.worldMatrix.MultScaling(o.scale)
	o.dirtyWorldMatrix = false
}

func (o *Object) WorldMatrix() *math.Mat4 {
	if o.dirtyWorldMatrix {
		o.updateWorldMatrix()
	}
	return &o.worldMatrix
}

func (o *Object) Place(position math.Vec3) {
	o.position = position
	o.dirtyWorldMatrix = true
}

func (o *Object) Translate(displacement math.Vec3) {
	o.Place(o.position.Add(displacement))
}

func (o *Object) Orient(unitX, unitY math.Vec3) {
	o.unitX = unitX.Norm()
	o.unitY = unitY.Norm()
	o.updateUnitZVector()
	o.dirtyWorldMatrix = true
}

func (o *Object) Rotate(axis math.Vec3, ang float32) {
	o.Orient(o.unitX.Rotate(axis, ang), o.unitY.Rotate(axis, ang))
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

func (o *Object) SetScale(scale math.Vec3) {
	o.scale = scale
	o.dirtyWorldMatrix = true
}

func (o *Object) Scale(factor math.Vec3) {
	o.SetScale(o.scale.Mult(factor))
}
