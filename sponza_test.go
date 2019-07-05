package main

import (
	"fmt"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/input"
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/render"
	"github.com/hersle/gl3d/scene"
	"github.com/hersle/gl3d/utils"
	"github.com/hersle/gl3d/window"
	"testing"
	"time"
	"os"
	"runtime/pprof"
	gomath "math"
)

func TestMain(m *testing.M) {
	renderer, err := render.NewRenderer()
	if err != nil {
		panic(err)
	}

	s := scene.NewScene()
	sponza, _ := object.ReadMesh("assets/objects/sponza/sponza.obj")
	sponza.Scale(math.Vec3{0.02, 0.02, 0.02})
	s.AddMesh(sponza)

	s.AmbientLight = light.NewAmbientLight(math.Vec3{0.1, 0.1, 0.1})

	l1 := light.NewPointLight(math.Vec3{1, 1, 1})
	l1.Place(math.Vec3{+15, 15, 0})
	l1.Attenuation = 0.01
	l1.CastShadows = true
	s.AddPointLight(l1)

	l2 := light.NewSpotLight(math.Vec3{1, 0, 0})
	l2.Place(math.Vec3{0, 1, 0})
	l2.Attenuation = 0.01
	l2.CastShadows = true
	s.AddSpotLight(l2)

	l3 := light.NewPointLight(math.Vec3{0, 1, 0})
	l3.Place(math.Vec3{22.59, 2.40, -9.18})
	l3.Attenuation = 0.05
	l3.CastShadows = true
	s.AddPointLight(l3)

	c := camera.NewPerspectiveCamera(60, 1, 0.1, 70)
	c.Place(math.Vec3{21.46, 13.37, 3.01})
	c.Orient(math.Vec3{0.18, 0, -0.98}, math.Vec3{-0.37, 0.92, -0.07})

	input.AddCameraFPSControls(c, 0.1)

	fpsCounter := utils.NewFrequencyCounter()
	fpsCounter.Interval = 10
	t0 := time.Now()
	ttot := float64(0.0)
	stopped := false

	f, err := os.Create("sponza.prof")
	if err != nil {
		panic(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	for !window.ShouldClose() {
		t := time.Now()
		var dt float64
		if stopped {
			dt = 0
		} else {
			dt = t.Sub(t0).Seconds()
		}
		ttot += dt
		t0 = t

		l1.Place(math.Vec3{15 * float32(gomath.Cos(0.4*ttot)), 15, 0})
		l2.RotateY(float32(2*dt))

		c.SetAspect(window.Aspect())
		renderer.Clear()
		renderer.RenderScene(s, c)
		if input.KeySpace.JustPressed() {
			stopped = !stopped
		}

		text := "FPS:        " + fpsCounter.String() + "\n"
		text += "position:   " + c.Position.String() + "\n"
		text += "draw calls: " + fmt.Sprint(graphics.Stats.DrawCallCount) + "\n"
		text += "vertices:   " + fmt.Sprint(graphics.Stats.VertexCount)
		renderer.RenderText(math.Vec2{-1, +1}, text, 0.05)

		renderer.Render()
		input.Update()  // TODO: make line order not matter
		window.Update() // TODO: make line order not matter
		graphics.Stats.Reset()

		fpsCounter.Count()
	}
}
