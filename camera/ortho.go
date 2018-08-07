package camera

import (
	"github.com/hersle/gl3d/math"
)

type OrthoCamera struct {
	BasicCamera
	height         float32
}

func NewOrthoCamera(height, aspect, near, Far float32) *OrthoCamera {
	var c OrthoCamera
	c.Object.Reset()
	c.height = height
	c.SetAspect(aspect)
	c.near = near
	c.Far = Far
	c.DirtyViewMat = true
	c.DirtyProjMat = true
	return &c
}

func (c *OrthoCamera) width() float32 {
	return c.aspect * c.height
}

func (c *OrthoCamera) updateProjectionMatrix() {
	c.projMat.OrthoCentered(math.NewVec3(c.width(), c.height, c.Far - c.near))
	c.DirtyProjMat = false
}

func (c *OrthoCamera) ProjectionMatrix() *math.Mat4 {
	if c.DirtyProjMat {
		c.updateProjectionMatrix()
	}
	return &c.projMat
}
