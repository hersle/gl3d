package main

type Camera struct {
	pos, vel Vec3
	fwd, up, right Vec3
	viewProjMat *Mat4
}

func NewCamera() *Camera {
	var c Camera
	c.MoveTo(NewVec3(0, 0, 0))
	c.vel = NewVec3(0, 0, 0)
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

func (c *Camera) ProjectionViewMatrix() *Mat4 {
	c.viewProjMat.Identity()
	c.viewProjMat.MultOrthoCentered(Vec3{4, 8, 1})
	c.viewProjMat.MultTranslation(c.pos.Scale(-1))
	return c.viewProjMat
}

func (c *Camera) Accelerate(dvel Vec3) {
	c.vel = c.vel.Add(dvel)
}

func (c *Camera) Update(dt float64) {
	c.MoveBy(c.vel.Scale(dt))
}
