package camera

import (
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	gomath "math"
)

type PerspectiveCamera struct {
	BasicCamera
	fovY float32
	dirtyFrustumPlanes bool
	frustumPlanes [6]*object.Plane
}

func NewPerspectiveCamera(fovYDeg, aspect, near, Far float32) *PerspectiveCamera {
	var c PerspectiveCamera
	c.BasicCamera = *NewBasicCamera(aspect, near, Far)
	c.fovY = math.Radians(fovYDeg)
	c.updateViewMatrix()
	c.updateProjectionMatrix()
	c.dirtyFrustumPlanes = true
	return &c
}

func (c *PerspectiveCamera) Orient(unitX, unitY math.Vec3) {
	c.BasicCamera.Orient(unitX, unitY)
	c.dirtyFrustumPlanes = true
}

func (c *PerspectiveCamera) Place(position math.Vec3) {
	c.BasicCamera.Place(position)
	c.dirtyFrustumPlanes = true
}

func (c *PerspectiveCamera) Rotate(axis math.Vec3, ang float32) {
	c.BasicCamera.Rotate(axis, ang)
	c.dirtyFrustumPlanes = true
}

func (c *PerspectiveCamera) RotateX(ang float32) {
	c.BasicCamera.RotateX(ang)
	c.dirtyFrustumPlanes = true
}

func (c *PerspectiveCamera) RotateY(ang float32) {
	c.BasicCamera.RotateY(ang)
	c.dirtyFrustumPlanes = true
}

func (c *PerspectiveCamera) RotateZ(ang float32) {
	c.BasicCamera.RotateZ(ang)
	c.dirtyFrustumPlanes = true
}

func (c *PerspectiveCamera) Scale(factor math.Vec3) {
	c.BasicCamera.Scale(factor)
	c.dirtyFrustumPlanes = true
}

func (c *PerspectiveCamera) SetForwardUp(forward, up math.Vec3) {
	c.BasicCamera.SetForwardUp(forward, up)
	c.dirtyFrustumPlanes = true
}

func (c *PerspectiveCamera) SetScale(scaling math.Vec3) {
	c.BasicCamera.SetScale(scaling)
	c.dirtyFrustumPlanes = true
}

func (c *PerspectiveCamera) Translate(displacement math.Vec3) {
	c.BasicCamera.Translate(displacement)
	c.dirtyFrustumPlanes = true
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

func (c *PerspectiveCamera) updateFrustumPlanes() {
	nh := c.near * float32(gomath.Tan(float64(c.fovY/2))) * 2
	nw := nh * c.aspect

	nc := c.Position.Add(c.Forward().Scale(c.near))
	nbl := nc.Add(c.Right().Scale(-nw / 2)).Add(c.Up().Scale(-nh / 2))
	nbr := nc.Add(c.Right().Scale(+nw / 2)).Add(c.Up().Scale(-nh / 2))
	ntr := nc.Add(c.Right().Scale(+nw / 2)).Add(c.Up().Scale(+nh / 2))
	ntl := nc.Add(c.Right().Scale(-nw / 2)).Add(c.Up().Scale(+nh / 2))

	fw := (c.Far / c.near) * nw
	fh := (c.Far / c.near) * nh

	fc := c.Position.Add(c.Forward().Scale(c.Far))
	fbl := fc.Add(c.Right().Scale(-fw / 2)).Add(c.Up().Scale(-fh / 2))
	fbr := fc.Add(c.Right().Scale(+fw / 2)).Add(c.Up().Scale(-fh / 2))
	ftr := fc.Add(c.Right().Scale(+fw / 2)).Add(c.Up().Scale(+fh / 2))
	ftl := fc.Add(c.Right().Scale(-fw / 2)).Add(c.Up().Scale(+fh / 2))

	c.frustumPlanes[0] = object.NewPlaneFromPoints(nbl, nbr, ntr) // near
	c.frustumPlanes[1] = object.NewPlaneFromPoints(fbr, fbl, ftl) // far
	c.frustumPlanes[2] = object.NewPlaneFromPoints(nbl, fbl, fbr) // bottom
	c.frustumPlanes[3] = object.NewPlaneFromPoints(ntl, ntr, ftr) // top
	c.frustumPlanes[4] = object.NewPlaneFromPoints(fbl, nbl, ntl) // left
	c.frustumPlanes[5] = object.NewPlaneFromPoints(nbl, fbr, ftr) // right

	c.dirtyFrustumPlanes = false
}

func (c *PerspectiveCamera) Cull(sm *object.SubMesh) bool {
	if c.dirtyFrustumPlanes {
		c.updateFrustumPlanes()
	}

	bbox := sm.BoundingBox()
	bboxpts := bbox.Points()

	for _, plane := range c.frustumPlanes {
		nOutside := 0
		for _, pt := range bboxpts {
			if plane.SignedDistance(pt) > 0 {
				nOutside++
			}
		}
		if nOutside == len(bboxpts) {
			return true
		}
	}

	return false
}
