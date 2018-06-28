package main

import (
	"math"
)

type Camera struct {
	pos Vec3
	fwd, up, right Vec3
	fovY float64
	aspect float64
	near, far float64
	viewMat, projMat, viewProjMat *Mat4
}

func NewCamera(pos, fwd, up Vec3, fovYDeg, aspect, near, far float64) *Camera {
	var c Camera
	c.MoveTo(pos)
	c.fwd = fwd
	c.up = up
	c.right = c.fwd.Cross(c.up)
	c.fovY = fovYDeg / 360.0 * 2.0 * math.Pi
	c.SetAspect(aspect)
	c.near = near
	c.far = far
	c.viewMat = NewMat4Identity()
	c.projMat = NewMat4Identity()
	c.viewProjMat = NewMat4Identity()
	return &c
}

func (c *Camera) MoveTo(pos Vec3) {
	c.pos = pos
}

func (c *Camera) MoveBy(displacement Vec3) {
	c.MoveTo(c.pos.Add(displacement))
}

func (c *Camera) Rotate(axis Vec3, ang float64) {
	c.fwd = c.fwd.Rotate(axis, ang).Norm()
	c.up = c.up.Rotate(axis, ang).Norm()
	c.right = c.fwd.Cross(c.up).Norm()
}

func (c *Camera) SetAspect(aspect float64) {
	c.aspect = aspect
}

func (c *Camera) ProjectionViewMatrix() *Mat4 {
	c.viewMat.LookAt(c.pos, c.pos.Add(c.fwd), c.up)
	c.projMat.Perspective(c.fovY, c.aspect, c.near, c.far)
	return c.viewProjMat.Identity().Mult(c.projMat).Mult(c.viewMat)
}
