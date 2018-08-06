package main

import (
	stdmath "math"
	"github.com/hersle/gl3d/math"
)

type Camera struct {
	Object
	fovY float32
	aspect float32
	near, far float32
	viewMat math.Mat4
	projMat math.Mat4
	dirtyViewMat bool
	dirtyProjMat bool
}

func NewCamera(fovYDeg, aspect, near, far float32) *Camera {
	var c Camera
	c.Object.Reset()
	c.fovY = fovYDeg / 360.0 * 2.0 * stdmath.Pi
	c.SetAspect(aspect)
	c.near = near
	c.far = far
	c.dirtyViewMat = true
	c.updateViewMatrix()
	c.updateProjectionMatrix()
	return &c
}

func (c *Camera) Right() math.Vec3 {
	return c.unitX
}

func (c *Camera) Up() math.Vec3 {
	return c.unitY
}

func (c *Camera) Forward() math.Vec3 {
	return c.unitZ.Scale(-1)
}

func (c *Camera) SetForwardUp(forward, up math.Vec3) {
	right := forward.Cross(up).Norm()
	c.Orient(right, up) // since unitX == right and unitY == up
}

func (c *Camera) SetAspect(aspect float32) {
	c.aspect = aspect
	c.dirtyProjMat = true
}

func (c *Camera) updateViewMatrix() {
	// TODO: elegantly make the inverse of the underlying object world matrix
	//c.viewMat.Copy(c.WorldMatrix()).MultScaling(NewVec3(1, 1, -1)).Invert()
	c.viewMat.LookAt(c.position, c.position.Add(c.Forward()), c.Up())
	c.dirtyViewMat = false
}

func (c *Camera) updateProjectionMatrix() {
	c.projMat.Perspective(c.fovY, c.aspect, c.near, c.far)
	c.dirtyProjMat = false
}

func (c *Camera) ViewMatrix() *math.Mat4 {
	// camera view matrix dirty iff object world matrix (its inverse) is dirty
	if c.dirtyWorldMatrix {
		c.updateViewMatrix()
	}
	c.updateViewMatrix()
	return &c.viewMat
}

func (c *Camera) ProjectionMatrix() *math.Mat4 {
	if c.dirtyProjMat {
		c.updateProjectionMatrix()
	}
	return &c.projMat
}
