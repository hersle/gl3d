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

func (c *Camera) Place(pos Vec3) {
	c.pos = pos
}

func (c *Camera) Move(displacement Vec3) {
	c.pos = c.pos.Add(displacement)
}

func (c *Camera) ViewProjectionMatrix() *Mat4 {
	c.viewProjMat.Identity()
	c.viewProjMat.MultOrthoCentered(Vec3{4, 8, 1})
	c.viewProjMat.MultTranslation(c.pos.Scale(-1))
	return c.viewProjMat
}
