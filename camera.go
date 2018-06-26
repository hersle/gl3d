package main

// TODO: connect fwd, up, right with viewProjection matrix

type Camera struct {
	pos Vec3
	fwd, up, right Vec3
	viewProjMat *Mat4
}

func NewCamera() *Camera {
	var c Camera
	c.MoveTo(NewVec3(0, 0, 0))
	c.fwd = NewVec3(0, 0, 1)
	c.right = NewVec3(1, 0, 0)
	c.up = NewVec3(0, 1, 0)
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
	c.fwd = c.fwd.Rotate(axis, ang)
	c.up = c.up.Rotate(axis, ang)
	c.right = c.fwd.Cross(c.up)
}

func (c *Camera) ProjectionViewMatrix() *Mat4 {
	c.viewProjMat.Identity()
	c.viewProjMat.MultPerspective(3.1415 / 2, 1, 0.001, 1000)
	c.viewProjMat.MultLookAt(c.pos, c.pos.Add(c.fwd), c.up)
	return c.viewProjMat
}
