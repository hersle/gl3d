package main

import (
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/material"
	"github.com/hersle/gl3d/engine"
	"testing"
	gomath "math"
)

func TestMain(m *testing.M) {
	geo := object.NewPlane(math.Vec3{0, 0, 0}, math.Vec3{0, 1, 0}).Geometry(20)
	mtl := material.NewDefaultMaterial("")
	floor := object.NewMesh(geo, mtl)

	geo = object.NewSphere(math.Vec3{1, 1, 0}, 1).Geometry(10)
	ball := object.NewMesh(geo, mtl)

	geo = object.NewBox(math.Vec3{0, 4, 0}, math.Vec3{1, 0, 0}, math.Vec3{0, 1, 0}, 1, 2, 3).Geometry()
	box := object.NewMesh(geo, mtl)

	l1 := light.NewPointLight(math.Vec3{0, 0, 1})
	l1.Place(math.Vec3{0, 8, 0})
	l1.CastShadows = true
	l1.Attenuation = 0.01

	l2 := light.NewSpotLight(math.Vec3{1, 0, 0})
	l2.CastShadows = true
	l2.Place(math.Vec3{8, 6, 8})
	l2.Orient(math.Vec3{1, 0, 0}, math.Vec3{0, 0, -1})
	l2.RotateX(3.1415 / 4)
	l2.RotateY(3.1415 / 4)
	l2.FOV = 3.1415 / 8

	l3 := light.NewPointLight(math.Vec3{0, 1, 0})
	l3.CastShadows = true
	l3.Attenuation = 0.01

	c := camera.NewPerspectiveCamera(60, 1, 0.1, 50)
	c.Place(math.Vec3{0, 1, +10})

	eng := engine.NewEngine()

	eng.InitializeCustom = func() {
		eng.Scene.AddMesh(floor)
		eng.Scene.AddMesh(ball)
		eng.Scene.AddMesh(box)
		eng.Scene.AddPointLight(l1)
		eng.Scene.AddSpotLight(l2)
		eng.Scene.AddPointLight(l3)
	}

	t := float32(0)
	eng.UpdateCustom = func(dt float32) {
		t += dt
		ball.Place(math.Vec3{float32(3*gomath.Cos(float64(t))), 2, float32(3*gomath.Sin(float64(t)))})
		box.RotateX(0.01)
		box.RotateY(0.02)
		box.RotateZ(0.03)
		box.SetScale(math.Vec3{1, 1, 1}.Scale(1 + 0.5 * float32(gomath.Sin(float64(t)))))
		l3.Place(math.Vec3{0, float32(5*gomath.Sin(float64(t))), 0})
	}

	eng.Run()
}
