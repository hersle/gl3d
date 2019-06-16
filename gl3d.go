package main

import (
	"fmt"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/render"
	"github.com/hersle/gl3d/scene"
	"github.com/hersle/gl3d/window"
	"github.com/hersle/gl3d/input"
	"github.com/hersle/gl3d/material"
	"os"
	"time"
	gomath "math"
	"runtime/pprof"
	"flag"
)

var cpuprofile = flag.String("cpuprofile", "", "write CPU profile to file")

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
			model.Scale(math.NewVec3(0.02, 0.02, 0.02))
			model.RotateX(-3.1415 / 2)
			model.RotateY(3.1415 - 3.1415/5)
		}
		if filename == "assets/objects/sponza/sponza.obj" || filename == "assets/objects/sponza2/sponza.obj" {
			model.Scale(math.NewVec3(0.02, 0.02, 0.02))
		}
		if filename == "assets/objects/conference/conference.obj" {
			model.Scale(math.NewVec3(0.02, 0.02, 0.02))
		}
		if filename == "assets/objects/racecar/racecar.obj" {
			model.Scale(math.NewVec3(0.04, 0.04, 0.04))
		}
		if filename == "assets/objects/holodeck/holodeck.obj" {
			model.Scale(math.NewVec3(0.04, 0.04, 0.04))
		}
		s.AddMesh(model)
	}

	s.AmbientLight = light.NewAmbientLight(math.NewVec3(0.1, 0.1, 0.1))

	s.AddPointLight(light.NewPointLight(math.NewVec3(1, 1, 1), math.NewVec3(1, 1, 1)))
	s.PointLights[0].AttenuationQuadratic = 0.1

	s.AddPointLight(light.NewPointLight(math.NewVec3(1, 1, 1), math.NewVec3(1, 1, 1)))
	s.PointLights[1].AttenuationQuadratic = 0.1
	s.PointLights[1].Place(math.NewVec3(5, 5, 0))

	//s.AddSpotLight(light.NewSpotLight(math.NewVec3(1, 1, 1), math.NewVec3(1, 1, 1)))
	//s.SpotLights[0].AttenuationQuadratic = 0.1

	//s.AddDirectionalLight(light.NewDirectionalLight(math.NewVec3(1, 1, 1), math.NewVec3(1, 1, 1)))

	//s.AddSkybox(graphics.NewCubeMapUniform(math.NewVec4(0.3, 0.3, 0.3, 0)))
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

	ang := float32(60.0)
	near := float32(0.1)
	far := float32(50.0)
	aspect := float32(1)
	c := camera.NewPerspectiveCamera(60, aspect, near, far)
	var m2 object.Mesh
	m2.Object = *object.NewObject()
	s.AddMesh(&m2)

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

	// TODO: remove
	renderer.RenderScene(s, c)

	drawScene := true

	time1 := time.Now()
	fps := int(0)
	frameCount := int(0)
	for !window.ShouldClose() {
		if time.Now().Sub(time1).Seconds() > 0.5 {
			time2 := time.Now()
			fps = int(float64(frameCount) / (time2.Sub(time1).Seconds()))
			time1 = time2
			frameCount = 0
		}

		c.SetAspect(window.Aspect())
		renderer.Clear()
		if drawScene {
			renderer.RenderScene(s, c)
		}
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
			//s.SpotLights[0].Place(c.Position)
			//s.SpotLights[0].Orient(c.UnitX, c.UnitY)
			//s.DirectionalLights[0].Orient(c.UnitX, c.UnitY)
		}
		if input.KeyZ.JustPressed() {
			drawScene = true
		}
		if input.KeyX.JustPressed() {
			drawScene = false
		}
		if input.KeyC.JustPressed() {
			renderer.SetWireframe(false)
		}
		if input.KeyV.JustPressed() {
			renderer.SetWireframe(true)
		}
		if input.KeyF.JustPressed() {
			nearWidth := 2 * near * float32(gomath.Tan(float64(ang / 2) / 180 * gomath.Pi))
			nearHeight := nearWidth / aspect
			frustum := object.NewFrustum(c.Position, c.Forward(), c.Up(), near / 10, far / 10, nearWidth / 10, nearHeight / 10)
			m2.SubMeshes = m2.SubMeshes[:0]
			m2.AddSubMesh(object.NewSubMesh(frustum.Geometry(), material.NewDefaultMaterial(""), &m2))
		}

		text := "FPS:        " + fmt.Sprint(fps) + "\n"
		text += "position:   " + c.Position.String() + "\n"
		text += "forward:    " + c.Forward().String() + "\n"
		text += "draw calls: " + fmt.Sprint(graphics.RenderStats.DrawCallCount) + "\n"
		text += "vertices:   " + fmt.Sprint(graphics.RenderStats.VertexCount)
		renderer.RenderText(math.NewVec2(-1, +1), text, 0.05)

		renderer.Render()
		input.Update() // TODO: make line order not matter
		window.Update() // TODO: make line order not matter
		graphics.RenderStats.Reset()

		frameCount++
	}
}
