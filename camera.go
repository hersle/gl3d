package main

type Camera struct {
	pos, vel Vec3
	yawAng, yawAngVel float64
	pitchAng, pitchAngVel float64
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
	c.yawAng = 0
	c.yawAngVel = 0
	c.pitchAng = 0
	c.pitchAngVel = 0
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
	c.viewProjMat.MultOrthoCentered(Vec3{10, 10, 10})
	println(c.viewProjMat.String())
	c.viewProjMat.MultTranslation(c.pos.Scale(-1))
	c.viewProjMat.MultRotationY(c.yawAng)
	c.viewProjMat.MultRotationX(c.pitchAng)
	return c.viewProjMat
}

func (c *Camera) Accelerate(dvel Vec3) {
	c.vel = c.vel.Add(dvel)
}

func (c *Camera) Update(dt float64) {
	c.yawAng = c.yawAng + c.yawAngVel * dt
	c.pitchAng = c.pitchAng + c.pitchAngVel * dt
	c.MoveBy(c.vel.Scale(dt))
}
