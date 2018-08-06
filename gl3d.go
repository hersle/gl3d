package main

import (
	"github.com/hersle/gl3d/window"
	"os"
	"github.com/go-gl/glfw/v3.2/glfw"
	"time"
	"fmt"
)

func main() {
	renderer, err := NewSceneRenderer()
	if err != nil {
		panic(err)
	}

	s := NewScene()
	for _, filename := range os.Args[1:] {
		model, err := ReadMesh(filename)
		if err != nil {
			panic(err)
		}
		if filename == "objects/car.obj" {
			model.Scale(NewVec3(0.02, 0.02, 0.02))
			model.RotateX(-3.1415/2)
			model.RotateY(3.1415 - 3.1415/5)
		}
		if filename == "objects/sponza.obj" || filename == "objects/sponza2/sponza.obj" {
			model.Scale(NewVec3(0.02, 0.02, 0.02))
		}
		if filename == "objects/conference.obj" {
			model.Scale(NewVec3(0.02, 0.02, 0.02))
		}
		if filename == "objects/scrubPine.obj" {
			model.Scale(NewVec3(0.02, 0.02, 0.02))
		}
		if filename == "objects/racecar.obj" {
			model.Scale(NewVec3(0.04, 0.04, 0.04))
		}
		if filename == "objects/holodeck/holodeck.obj" {
			model.Scale(NewVec3(0.04, 0.04, 0.04))
		}
		if filename == "objects/oak/white_oak.obj" {
			model.Scale(NewVec3(0.04, 0.04, 0.04))
		}
		s.AddMesh(model)
	}

	s.ambientLight = NewAmbientLight(NewVec3(1, 1, 1))
	s.AddSpotLight(NewSpotLight(NewVec3(1, 1, 1), NewVec3(1, 1, 1)))
	s.spotLights[0].Camera = *NewCamera(60, 1, 0.1, 50)
	s.spotLights[0].Place(NewVec3(0, 3, 0))
	s.spotLights[0].Orient(s.spotLights[0].position.Scale(-1).Norm(), NewVec3(0, 0, 1))
	s.AddPointLight(NewPointLight(NewVec3(1, 1, 1), NewVec3(1, 1, 1)))
	s.AddPointLight(NewPointLight(NewVec3(1, 1, 1), NewVec3(1, 1, 1)))
	s.pointLights[1].Place(NewVec3(5, 0, 0))

	c := NewCamera(60, 1, 0.1, 50)

	var camFactor float32

	skyboxRenderer := NewSkyboxRenderer() // disable while working with multiple lights
	textRenderer := NewTextRenderer()
	arrowRenderer := NewArrowRenderer()

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
		defaultFramebuffer.ClearColor(NewVec4(0, 0, 0, 0))
		defaultFramebuffer.ClearDepth(1)
		skyboxRenderer.Render(c)
		if drawScene {
			renderer.Render(s, c)
		}
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
		text += "position:   " + c.position.String() + "\n"
		text += "forward:    " + c.Forward().String() + "\n"
		text += "draw calls: " + fmt.Sprint(RenderStats.drawCallCount) + "\n"
		text += "vertices:   " + fmt.Sprint(RenderStats.vertexCount)
		textRenderer.Render(NewVec2(-1, +1), text, 0.05)
		window.Update()

		RenderStats.Reset()

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
			c.Rotate(c.Right(), camFactor * +0.03)
		}
		if window.Win.GetKey(glfw.KeyDown) == glfw.Press {
			c.Rotate(c.Right(), camFactor * -0.03)
		}
		if window.Win.GetKey(glfw.KeyLeft) == glfw.Press {
			c.Rotate(NewVec3(0, 1, 0), camFactor * +0.03)
		}
		if window.Win.GetKey(glfw.KeyRight) == glfw.Press {
			c.Rotate(NewVec3(0, 1, 0), camFactor * -0.03)
		}
		if window.Win.GetKey(glfw.KeySpace) == glfw.Press {
			//s.pointLights[0].Place(c.position)
			s.spotLights[0].Place(c.position)
			s.spotLights[0].Orient(c.unitX, c.unitY) // for spotlight
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
