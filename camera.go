package main

type Camera struct {
	pos Vec3
	fwd, up, right Vec3
	viewMat, projMat, viewProjMat *Mat4
}

func NewCamera() *Camera {
	var c Camera
	c.MoveTo(NewVec3(0, 0, 0))
	c.fwd = NewVec3(0, 0, 1)
	c.up = NewVec3(0, 1, 0)
	c.right = c.fwd.Cross(c.up)
	c.viewMat = NewMat4Identity()
	c.projMat = NewMat4Identity()
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
	c.fwd = c.fwd.Rotate(axis, ang).Norm()
	c.up = c.up.Rotate(axis, ang).Norm()
	c.right = c.fwd.Cross(c.up).Norm()
}

func (c *Camera) ProjectionViewMatrix() *Mat4 {
	c.viewMat.LookAt(c.pos, c.pos.Add(c.fwd), c.up)
	c.projMat.Perspective(3.1415 / 4, 1, 0.001, 1000)
	return c.viewProjMat.Identity().Mult(c.projMat).Mult(c.viewMat)
}
