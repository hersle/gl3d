package main

import (
	"github.com/hersle/gl3d/window"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/render"
	"github.com/hersle/gl3d/scene"
	"github.com/hersle/gl3d/input"
	"testing"
)

func TestMain(t *testing.M) {
	renderer, err := render.NewSceneRenderer()
	textRenderer := render.NewTextRenderer()
	quadRenderer := render.NewQuadRenderer()
	if err != nil {
		panic(err)
	}

	s := scene.NewScene()
	car, err := object.ReadMesh("assets/objects/sportscar/sportscar.obj")
	if err != nil {
		panic(err)
	}
	s.AddMesh(car)

	ambient := light.NewAmbientLight(math.NewVec3(0.5, 0.5, 0.5))
	point := light.NewPointLight(math.NewVec3(1, 1, 1), math.NewVec3(1, 1, 1))
	point.AttenuationQuadratic = 0.1
	point.Place(math.NewVec3(0, 2, 2))
	s.AddAmbientLight(ambient)
	s.AddPointLight(point)

	skybox := graphics.NewCubeMapUniform(math.NewVec4(0.2, 0.2, 0.2, 0))
	s.AddSkybox(skybox)

	c := camera.NewPerspectiveCamera(60, 1, 0.1, 50)
	c.Place(point.Position)

	input.AddCameraFPSControls(c)
	input.KeySpace.Listen(func(action input.Action) {
		s.PointLights[0].Place(c.Position)
	})

	for !window.ShouldClose() {
		car.RotateY(+0.01)

		c.SetAspect(window.Aspect())
		renderer.Render(s, c)
		quadRenderer.Render(renderer.RenderTarget)
		textRenderer.Render(math.NewVec2(-1, +1), graphics.RenderStats.String(), 0.05)
		graphics.RenderStats.Reset()

		window.Update()
		input.Update()
	}
}
