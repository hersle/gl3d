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
		if filename == "assets/objects/car/car.obj" {
			model.Scale(math.Vec3{0.02, 0.02, 0.02})
			model.RotateX(-3.1415 / 2)
			model.RotateY(3.1415 - 3.1415/5)
		}
		if filename == "assets/objects/sponza/sponza.obj" || filename == "assets/objects/sponza2/sponza.obj" {
			model.Scale(math.Vec3{0.02, 0.02, 0.02})
		}
		if filename == "assets/objects/conference/conference.obj" {
			model.Scale(math.Vec3{0.02, 0.02, 0.02})
		}
		if filename == "assets/objects/racecar/racecar.obj" {
			model.Scale(math.Vec3{0.04, 0.04, 0.04})
		}
		if filename == "assets/objects/holodeck/holodeck.obj" {
			model.Scale(math.Vec3{0.04, 0.04, 0.04})
		}
		s.AddMesh(model)
	}

	s.AmbientLight = light.NewAmbientLight(math.Vec3{0.1, 0.1, 0.1})

	s.AddPointLight(light.NewPointLight(math.Vec3{1, 1, 1}, math.Vec3{1, 1, 1}))
	s.PointLights[0].AttenuationQuadratic = 0.1

	s.AddPointLight(light.NewPointLight(math.Vec3{1, 1, 1}, math.Vec3{1, 1, 1}))
	s.PointLights[1].AttenuationQuadratic = 0.1
	s.PointLights[1].Place(math.Vec3{5, 5, 0})

	filename1 := "assets/skyboxes/mountain/posx.jpg"
	filename2 := "assets/skyboxes/mountain/negx.jpg"
	filename3 := "assets/skyboxes/mountain/posy.jpg"
	filename4 := "assets/skyboxes/mountain/negy.jpg"
	filename5 := "assets/skyboxes/mountain/posz.jpg"
	filename6 := "assets/skyboxes/mountain/negz.jpg"
	skybox, err := scene.ReadCubeMap(filename1, filename2, filename3, filename4, filename5, filename6)
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
			s.PointLights[0].Place(c.Position)
		}

		text := "FPS:        " + fpsCounter.String() + "\n"
		text += "position:   " + c.Position.String() + "\n"
		text += "forward:    " + c.Forward().String() + "\n"
		text += "draw calls: " + fmt.Sprint(graphics.RenderStats.DrawCallCount) + "\n"
		text += "vertices:   " + fmt.Sprint(graphics.RenderStats.VertexCount)
		renderer.RenderText(math.Vec2{-1, +1}, text, 0.05)

		renderer.Render()
		input.Update()  // TODO: make line order not matter
		window.Update() // TODO: make line order not matter
		graphics.RenderStats.Reset()

		fpsCounter.Count()
	}
}
