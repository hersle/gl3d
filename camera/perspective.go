package camera

import (
	"github.com/hersle/gl3d/math"
)

type PerspectiveCamera struct {
	BasicCamera
	fovY         float32
}

func NewPerspectiveCamera(fovYDeg, aspect, near, Far float32) *PerspectiveCamera {
	var c PerspectiveCamera
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

func (c *PerspectiveCamera) updateProjectionMatrix() {
	c.projMat.Perspective(c.fovY, c.aspect, c.near, c.Far)
	c.DirtyProjMat = false
}

func (c *PerspectiveCamera) ProjectionMatrix() *math.Mat4 {
	if c.DirtyProjMat {
		c.updateProjectionMatrix()
	}
	return &c.projMat
}
