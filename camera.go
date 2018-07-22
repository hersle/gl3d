package main

import (
	"math"
)

type Camera struct {
	Object
	fovY float32
	aspect float32
	near, far float32
	viewMat, projMat *Mat4
	dirtyViewMat, dirtyProjMat bool
}

func NewCamera(fovYDeg, aspect, near, far float32) *Camera {
	var c Camera
	c.Object.Init()
	c.fovY = fovYDeg / 360.0 * 2.0 * math.Pi
	c.SetAspect(aspect)
	c.near = near
	c.far = far
	c.viewMat = NewMat4Identity()
	c.projMat = NewMat4Identity()
	c.dirtyViewMat = true
	c.updateViewMatrix()
	c.updateProjectionMatrix()
	return &c
}

func (c *Camera) SetAspect(aspect float32) {
	c.aspect = aspect
	c.dirtyProjMat = true
}

func (c *Camera) updateViewMatrix() {
	// flip z for openGL
	c.viewMat.Copy(c.WorldMatrix()).MultScaling(NewVec3(1, 1, -1)).Invert()
	c.dirtyViewMat = false
}

func (c *Camera) updateProjectionMatrix() {
	c.projMat.Perspective(c.fovY, c.aspect, c.near, c.far)
	c.dirtyProjMat = false
}

func (c *Camera) ViewMatrix() *Mat4 {
	// camera view matrix dirty iff object world matrix (its inverse) is dirty
	if c.dirtyWorldMatrix {
		c.updateViewMatrix()
	}
	c.updateViewMatrix()
	return c.viewMat
}

func (c *Camera) ProjectionMatrix() *Mat4 {
	if c.dirtyProjMat {
		c.updateProjectionMatrix()
	}
	return c.projMat
}
