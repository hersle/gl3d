package main

import (
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/engine"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"testing"
	gomath "math"
)

func TestMain(m *testing.M) {
	sponza, _ := object.ReadMesh("assets/objects/sponza/sponza.obj")
	sponza.Scale(math.Vec3{0.02, 0.02, 0.02})

	l1 := light.NewPointLight(math.Vec3{1, 1, 1})
	l1.Place(math.Vec3{+15, 15, 0})
	l1.Attenuation = 0.005
	l1.CastShadows = true

	l2 := light.NewSpotLight(math.Vec3{1, 0, 0})
	l2.Place(math.Vec3{0, 1, 0})
	l2.Attenuation = 0.01
	l2.CastShadows = true
	l2.FOV = 3.1415 / 4

	l3 := light.NewPointLight(math.Vec3{0, 1, 0})
	l3.Place(math.Vec3{22.59, 2.40, -9.18})
	l3.Attenuation = 0.05
	l3.CastShadows = true

	eng := engine.NewEngine()

	eng.InitializeCustom = func() {
		eng.Scene.AddMesh(sponza)
		eng.Scene.AmbientLight = light.NewAmbientLight(math.Vec3{0.8, 0.8, 0.8})
		eng.Scene.AddPointLight(l1)
		eng.Scene.AddSpotLight(l2)
		eng.Scene.AddPointLight(l3)
		eng.Camera.Place(math.Vec3{21.46, 13.37, 3.01})
		eng.Camera.Orient(math.Vec3{0.18, 0, -0.98}, math.Vec3{-0.37, 0.92, -0.07})
	}

	t := float32(0)
	eng.UpdateCustom = func(dt float32) {
		t += dt
		l1.Place(math.Vec3{15 * float32(gomath.Cos(float64(0.4*t))), 15, 0})
		l2.RotateY(float32(2*dt))
	}

	eng.Run()
}
