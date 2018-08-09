package main

import (
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/render"
	"github.com/hersle/gl3d/scene"
	"github.com/hersle/gl3d/window"
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

	s.AmbientLight = light.NewAmbientLight(math.NewVec3(1, 1, 1))
	s.AddSpotLight(light.NewSpotLight(math.NewVec3(1, 1, 1), math.NewVec3(1, 1, 1)))
	s.SpotLights[0].PerspectiveCamera = *camera.NewPerspectiveCamera(60, 1, 0.1, 50)
	s.SpotLights[0].Place(math.NewVec3(0, 3, 0))
	s.SpotLights[0].Orient(s.SpotLights[0].Position.Scale(-1).Norm(), math.NewVec3(0, 0, 1))
	s.AddPointLight(light.NewPointLight(math.NewVec3(1, 1, 1), math.NewVec3(1, 1, 1)))
	s.AddPointLight(light.NewPointLight(math.NewVec3(1, 1, 1), math.NewVec3(1, 1, 1)))
	s.PointLights[1].Place(math.NewVec3(5, 0, 0))
	s.AddSkybox(graphics.ReadCubeMapFromDir(graphics.NearestFilter, "assets/skyboxes/mountain/"))
	s.AddDirectionalLight(light.NewDirectionalLight(math.NewVec3(1, 1, 1), math.NewVec3(1, 1, 1)))

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
		if window.Win.GetKey(glfw.Key1) == glfw.Press {
			arrowRenderer.RenderTangents(s, c)
		}
		if window.Win.GetKey(glfw.Key2) == glfw.Press {
			arrowRenderer.RenderBitangents(s, c)
		}
		if window.Win.GetKey(glfw.Key3) == glfw.Press {
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

		if window.Win.GetKey(glfw.KeyLeftShift) == glfw.Press {
			camFactor = 0.1 // for precise camera controls
		} else {
			camFactor = 1.0
		}

		if window.Win.GetKey(glfw.KeyW) == glfw.Press {
			c.Translate(c.Forward().Scale(camFactor * +0.1))
		}
		if window.Win.GetKey(glfw.KeyS) == glfw.Press {
			c.Translate(c.Forward().Scale(camFactor * -0.1))
		}
		if window.Win.GetKey(glfw.KeyD) == glfw.Press {
			c.Translate(c.Right().Scale(camFactor * +0.1))
		}
		if window.Win.GetKey(glfw.KeyA) == glfw.Press {
			c.Translate(c.Right().Scale(camFactor * -0.1))
		}
		if window.Win.GetKey(glfw.KeyUp) == glfw.Press {
			c.Rotate(c.Right(), camFactor*+0.03)
		}
		if window.Win.GetKey(glfw.KeyDown) == glfw.Press {
			c.Rotate(c.Right(), camFactor*-0.03)
		}
		if window.Win.GetKey(glfw.KeyLeft) == glfw.Press {
			c.Rotate(math.NewVec3(0, 1, 0), camFactor*+0.03)
		}
		if window.Win.GetKey(glfw.KeyRight) == glfw.Press {
			c.Rotate(math.NewVec3(0, 1, 0), camFactor*-0.03)
		}
		if window.Win.GetKey(glfw.KeySpace) == glfw.Press {
			//s.pointLights[0].Place(c.position)
			//s.SpotLights[0].Place(c.Position)
			//s.SpotLights[0].Orient(c.UnitX, c.UnitY) // for spotlight
			s.DirectionalLights[0].Orient(c.UnitX, c.UnitY)
			s.DirectionalLights[0].Place(c.Position)
		}
		if window.Win.GetKey(glfw.KeyZ) == glfw.Press {
			drawScene = true
		}
		if window.Win.GetKey(glfw.KeyX) == glfw.Press {
			drawScene = false
		}
		if window.Win.GetKey(glfw.KeyC) == glfw.Press {
			renderer.SetWireframe(false)
		}
		if window.Win.GetKey(glfw.KeyV) == glfw.Press {
			renderer.SetWireframe(true)
		}

		frameCount++
	}
}
