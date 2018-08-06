package camera

import (
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
)

type Camera struct {
	object.Object
	fovY         float32
	aspect       float32
	near, Far    float32
	viewMat      math.Mat4
	projMat      math.Mat4
	DirtyViewMat bool
	DirtyProjMat bool
}

func NewCamera(fovYDeg, aspect, near, Far float32) *Camera {
	var c Camera
	c.Object.Reset()
	c.fovY = math.Radians(fovYDeg)
	c.SetAspect(aspect)
	c.near = near
	c.Far = Far
	c.DirtyViewMat = true
	c.updateViewMatrix()
	c.updateProjectionMatrix()
	return &c
}

func (c *Camera) Right() math.Vec3 {
	return c.UnitX
}

func (c *Camera) Up() math.Vec3 {
	return c.UnitY
}

func (c *Camera) Forward() math.Vec3 {
	return c.UnitZ.Scale(-1)
}

func (c *Camera) SetForwardUp(forward, up math.Vec3) {
	right := forward.Cross(up).Norm()
	c.Orient(right, up) // since unitX == right and unitY == up
}

func (c *Camera) SetAspect(aspect float32) {
	c.aspect = aspect
	c.DirtyProjMat = true
}

func (c *Camera) updateViewMatrix() {
	// TODO: elegantly make the inverse of the underlying object world matrix
	//c.viewMat.Copy(c.WorldMatrix()).MultScaling(NewVec3(1, 1, -1)).Invert()
	c.viewMat.LookAt(c.Position, c.Position.Add(c.Forward()), c.Up())
	c.DirtyViewMat = false
}

func (c *Camera) updateProjectionMatrix() {
	c.projMat.Perspective(c.fovY, c.aspect, c.near, c.Far)
	c.DirtyProjMat = false
}

func (c *Camera) ViewMatrix() *math.Mat4 {
	// camera view matrix dirty iff object world matrix (its inverse) is dirty
	if c.DirtyWorldMatrix {
		c.updateViewMatrix()
	}
	c.updateViewMatrix()
	return &c.viewMat
}

func (c *Camera) ProjectionMatrix() *math.Mat4 {
	if c.DirtyProjMat {
		c.updateProjectionMatrix()
	}
	return &c.projMat
}
