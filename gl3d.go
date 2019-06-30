package main

import (
	"flag"
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
	"github.com/hersle/gl3d/material"
	"os"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write CPU profile to file")
var frames = flag.Int("frames", -1, "number of frames to render")

func main() {
	flag.Parse()

	renderer, err := render.NewRenderer()
	if err != nil {
		panic(err)
	}

	s := scene.NewScene()
	for _, filename := range flag.Args() {
		model, err := object.ReadMesh(filename)
		if err != nil {
			panic(err)
		}
		s.AddMesh(model)
	}

	geo := object.NewCircle(10, math.Vec3{0, 0, 0}, math.Vec3{1, 1, 1}.Norm()).Geometry(10)
	mtl := material.NewDefaultMaterial("")
	mesh := object.NewMesh(geo, mtl)
	s.AddMesh(mesh)

	s.AmbientLight = light.NewAmbientLight(math.Vec3{0.1, 0.1, 0.1})

	l := light.NewSpotLight(math.Vec3{1, 1, 1}, math.Vec3{1, 1, 1})
	l.Attenuation = 0.1
	l.CastShadows = true
	s.AddSpotLight(l)

	f1 := "assets/skyboxes/mountain/posx.jpg"
	f2 := "assets/skyboxes/mountain/negx.jpg"
	f3 := "assets/skyboxes/mountain/posy.jpg"
	f4 := "assets/skyboxes/mountain/negy.jpg"
	f5 := "assets/skyboxes/mountain/posz.jpg"
	f6 := "assets/skyboxes/mountain/negz.jpg"
	skybox, err := scene.ReadCubeMap(f1, f2, f3, f4, f5, f6)
	if err != nil {
		panic(err)
	}
	s.AddSkybox(skybox)

	c := camera.NewPerspectiveCamera(60, 1, 0.1, 50)

	input.AddCameraFPSControls(c, 0.1)

	if *cpuprofile != "" {
		println("profiling")
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	fpsCounter := utils.NewFrequencyCounter()
	fpsCounter.Interval = 10
	for !window.ShouldClose() {
		if *frames == 0 {
			break
		} else if *frames > 0 {
			*frames--
		}

		c.SetAspect(window.Aspect())
		renderer.Clear()
		renderer.RenderScene(s, c)
		if input.Key1.Held() {
			renderer.RenderTangents(s, c)
		}
		if input.Key2.Held() {
			renderer.RenderBitangents(s, c)
		}
		if input.Key3.Held() {
			renderer.RenderNormals(s, c)
		}
		if input.KeySpace.JustPressed() {
			l.Place(c.Position)
			l.Orient(c.UnitX, c.UnitY)
		}
		if input.KeyMinus.JustPressed() {
			for _, mesh := range s.Meshes {
				mesh.Scale(math.Vec3{2.0, 2.0, 2.0})
			}
		}
		if input.KeySlash.JustPressed() {
			for _, mesh := range s.Meshes {
				mesh.Scale(math.Vec3{0.5, 0.5, 0.5})
			}
		}

		text := "FPS:        " + fpsCounter.String() + "\n"
		text += "position:   " + c.Position.String() + "\n"
		text += "forward:    " + c.Forward().String() + "\n"
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
