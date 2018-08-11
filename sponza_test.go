package main

import (
	"fmt"
	"github.com/hersle/gl3d/window"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/render"
	"github.com/hersle/gl3d/scene"
	"github.com/hersle/gl3d/input"
	"time"
	"testing"
)

func TestMain(t *testing.M) {
	renderer, err := render.NewSceneRenderer()
	textRenderer := render.NewTextRenderer()
	arrowRenderer := render.NewArrowRenderer()
	quadRenderer := render.NewQuadRenderer()
	if err != nil {
		panic(err)
	}

	s := scene.NewScene()
	model, err := object.ReadMesh("assets/objects/cube/cube.obj")
	if err != nil {
		panic(err)
	}
	//model.Scale(math.NewVec3(0.02, 0.02, 0.02))
	s.AddMesh(model)

	ambient := light.NewAmbientLight(math.NewVec3(0.1, 0.1, 0.1))
	point := light.NewPointLight(math.NewVec3(1, 1, 1), math.NewVec3(1, 1, 1))
	point.AttenuationQuadratic = 0.1
	point.Place(math.NewVec3(0, 2, 0))

	s.AmbientLight = ambient
	s.AddPointLight(point)

	//s.AddSkybox(graphics.ReadCubeMapFromDir(graphics.NearestFilter, "assets/skyboxes/mountain/"))
	s.AddSkybox(graphics.NewCubeMapUniform(math.NewVec4(0.0, 0.0, 0.0, 0)))

	c := camera.NewPerspectiveCamera(60, 1, 0.1, 50)

	drawScene := true

	input.AddCameraFPSControls(c)
	input.KeySpace.Listen(func(action input.Action) {
		s.PointLights[0].Place(c.Position)
	})
	input.KeyZ.Listen(func(action input.Action) {
		drawScene = true
	})
	input.KeyX.Listen(func(action input.Action) {
		drawScene = false
	})
	input.KeyC.Listen(func(action input.Action) {
		renderer.SetWireframe(false)
	})
	input.KeyV.Listen(func(action input.Action) {
		renderer.SetWireframe(true)
	})

	time1 := time.Now()
	fps := int(0)
	frameCount := int(0)
	for !window.ShouldClose() {
		if time.Now().Sub(time1).Seconds() > 0.0 {
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
		if input.Key1.JustPressed() {
			arrowRenderer.RenderTangents(s, c)
		}
		if input.Key2.JustPressed() {
			arrowRenderer.RenderBitangents(s, c)
		}
		if input.Key3.JustPressed() {
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

		input.Update()

		frameCount++
	}
}
