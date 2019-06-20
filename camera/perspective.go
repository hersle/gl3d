package camera

import (
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	gomath "math"
)

type PerspectiveCamera struct {
	BasicCamera
	fovY float32
}

func NewPerspectiveCamera(fovYDeg, aspect, near, Far float32) *PerspectiveCamera {
	var c PerspectiveCamera
	c.BasicCamera = *NewBasicCamera(aspect, near, Far)
	c.fovY = math.Radians(fovYDeg)
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

func (c *PerspectiveCamera) Cull(sm *object.SubMesh) bool {
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

	planes := [6]*object.Plane{}
	planes[0] = object.NewPlaneFromPoints(nbl, nbr, ntr) // near
	planes[1] = object.NewPlaneFromPoints(fbr, fbl, ftl) // far
	planes[2] = object.NewPlaneFromPoints(nbl, fbl, fbr) // bottom
	planes[3] = object.NewPlaneFromPoints(ntl, ntr, ftr) // top
	planes[4] = object.NewPlaneFromPoints(fbl, nbl, ntl) // left
	planes[5] = object.NewPlaneFromPoints(nbl, fbr, ftr) // right

	bbox := sm.BoundingBox()
	bboxpts := bbox.Points()

	for _, plane := range planes {
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
