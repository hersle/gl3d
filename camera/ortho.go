package camera

import (
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
)

type OrthoCamera struct {
	BasicCamera
	height float32
}

func NewOrthoCamera(height, aspect, near, Far float32) *OrthoCamera {
	var c OrthoCamera
	c.BasicCamera = *NewBasicCamera(aspect, near, Far)
	c.height = height
	return &c
}

func (c *OrthoCamera) width() float32 {
	return c.aspect * c.height
}

func (c *OrthoCamera) updateProjectionMatrix() {
	mat := math.Mat4Stack.New()
	defer math.Mat4Stack.Pop()

	c.projMat.OrthoCentered(math.Vec3{c.width(), c.height, c.Far - c.near})
	c.projMat.Mult(mat.Translation(math.Vec3{0, 0, +((c.Far - c.near) / 2) + c.near}))
	c.DirtyProjMat = false
}

func (c *OrthoCamera) ProjectionMatrix() *math.Mat4 {
	if c.DirtyProjMat {
		c.updateProjectionMatrix()
	}
	return &c.projMat
}

func (c *OrthoCamera) Cull(sm *object.SubMesh) bool {
	return false
}
