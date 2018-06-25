package main

type Camera struct {
	pos Vec3
	fwd, up, right Vec3
	viewProjMat *Mat4
}

func NewCamera() *Camera {
	var c Camera
	c.viewProjMat = NewMat4Identity()
	return &c
}

func (c *Camera) MoveTo(pos Vec3) {
	c.pos = pos
}

func (c *Camera) MoveBy(displacement Vec3) {
	c.MoveTo(c.pos.Add(displacement))
}

func (c *Camera) ProjectionViewMatrix() *Mat4 {
	c.viewProjMat.Identity()
	c.viewProjMat.MultOrthoCentered(Vec3{4, 8, 1})
	c.viewProjMat.MultTranslation(c.pos.Scale(-1))
	return c.viewProjMat
}
