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
	"os"
	"time"
)

func main() {
	renderer, err := render.NewSceneRenderer()
	if err != nil {
		panic(err)
	}

	s := scene.NewScene()
	for _, filename := range os.Args[1:] {
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

	//s.AddSkybox(graphics.ReadCubeMapFromDir(graphics.NearestFilter, "assets/skyboxes/mountain/"))
	s.AddSkybox(graphics.NewCubeMapUniform(math.NewVec4(0.0, 0.0, 0.0, 0)))

	c := camera.NewPerspectiveCamera(60, 1, 0.1, 50)

	var camFactor float32

	textRenderer := render.NewTextRenderer()
	arrowRenderer := render.NewArrowRenderer()
	quadRenderer := render.NewQuadRenderer()

	// TODO: remove
	renderer.Render(s, c)

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
		graphics.DefaultFramebuffer.ClearColor(math.NewVec4(0, 0, 0, 0))
		graphics.DefaultFramebuffer.ClearDepth(1)
		if drawScene {
			renderer.Render(s, c)
		}
		quadRenderer.Render(renderer.RenderTarget)
		//quadRenderer.Render(s.DirectionalLights[0].ShadowMap)
		if input.Key1.Pressed() {
			arrowRenderer.RenderTangents(s, c)
		}
		if input.Key2.Pressed() {
			arrowRenderer.RenderBitangents(s, c)
		}
		if input.Key3.Pressed() {
			arrowRenderer.RenderNormals(s, c)
		}
		text := "FPS:        " + fmt.Sprint(fps) + "\n"
		text += "position:   " + c.Position.String() + "\n"
		text += "forward:    " + c.Forward().String() + "\n"
		text += "draw calls: " + fmt.Sprint(graphics.RenderStats.DrawCallCount) + "\n"
		text += "vertices:   " + fmt.Sprint(graphics.RenderStats.VertexCount)
		textRenderer.Render(math.NewVec2(-1, +1), text, 0.05)
		window.Update()

		graphics.RenderStats.Reset()

		if input.KeyLeftShift.Pressed() {
			camFactor = 0.1 // for precise camera controls
		} else {
			camFactor = 1.0
		}

		if input.KeyW.Pressed() {
			c.Translate(c.Forward().Scale(camFactor * +0.1))
		}
		if input.KeyS.Pressed() {
			c.Translate(c.Forward().Scale(camFactor * -0.1))
		}
		if input.KeyD.Pressed() {
			c.Translate(c.Right().Scale(camFactor * +0.1))
		}
		if input.KeyA.Pressed() {
			c.Translate(c.Right().Scale(camFactor * -0.1))
		}
		if input.KeyUp.Pressed() {
			c.Rotate(c.Right(), camFactor*+0.03)
		}
		if input.KeyDown.Pressed() {
			c.Rotate(c.Right(), camFactor*-0.03)
		}
		if input.KeyLeft.Pressed() {
			c.Rotate(math.NewVec3(0, 1, 0), camFactor*+0.03)
		}
		if input.KeyRight.Pressed() {
			c.Rotate(math.NewVec3(0, 1, 0), camFactor*-0.03)
		}
		if input.KeySpace.Pressed() {
			s.PointLights[0].Place(c.Position)
			//s.SpotLights[0].Place(c.Position)
			//s.SpotLights[0].Orient(c.UnitX, c.UnitY)
			//s.DirectionalLights[0].Orient(c.UnitX, c.UnitY)
		}
		if input.KeyZ.Pressed() {
			drawScene = true
		}
		if input.KeyX.Pressed() {
			drawScene = false
		}
		if input.KeyC.Pressed() {
			renderer.SetWireframe(false)
		}
		if input.KeyV.Pressed() {
			renderer.SetWireframe(true)
		}

		input.Update()

		frameCount++
	}
}
