package camera

import (
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
)

type BasicCamera struct {
	object.Object
	aspect       float32
	near, Far    float32
	viewMat      math.Mat4
	projMat      math.Mat4
	DirtyViewMat bool
	DirtyProjMat bool
}

func NewBasicCamera(aspect, near, Far float32) *BasicCamera {
	var c BasicCamera
	c.Object.Reset()
	c.SetAspect(aspect)
	c.near = near
	c.Far = Far
	c.DirtyViewMat = true
	c.DirtyProjMat = true
	return &c
}

func (c *BasicCamera) Right() math.Vec3 {
	return c.UnitX
}

func (c *BasicCamera) Up() math.Vec3 {
	return c.UnitY
}

func (c *BasicCamera) Forward() math.Vec3 {
	return c.UnitZ.Scale(-1)
}

func (c *BasicCamera) SetForwardUp(forward, up math.Vec3) {
	right := forward.Cross(up).Norm()
	c.Orient(right, up) // since unitX == right and unitY == up
}

func (c *BasicCamera) SetAspect(aspect float32) {
	c.aspect = aspect
	c.DirtyProjMat = true
}

func (c *PerspectiveCamera) updateViewMatrix() {
	// TODO: elegantly make the inverse of the underlying object world matrix
	//c.viewMat.Copy(c.WorldMatrix()).MultScaling(NewVec3(1, 1, -1)).Invert()
	c.viewMat.LookAt(c.Position, c.Position.Add(c.Forward()), c.Up())
	c.DirtyViewMat = false
}

func (c *PerspectiveCamera) ViewMatrix() *math.Mat4 {
	// camera view matrix dirty iff object world matrix (its inverse) is dirty
	if c.DirtyWorldMatrix {
		c.updateViewMatrix()
	}
	c.updateViewMatrix()
	return &c.viewMat
}
