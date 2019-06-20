package main

import (
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/input"
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/render"
	"github.com/hersle/gl3d/scene"
	"github.com/hersle/gl3d/window"
	"testing"
)

func TestMain(m *testing.M) {
	renderer, err := render.NewRenderer()
	if err != nil {
		panic(err)
	}

	s := scene.NewScene()
	car, err := object.ReadMesh("assets/objects/sportscar/sportscar.obj")
	if err != nil {
		panic(err)
	}
	s.AddMesh(car)

	ambient := light.NewAmbientLight(math.Vec3{0.5, 0.5, 0.5})
	point := light.NewPointLight(math.Vec3{1, 1, 1}, math.Vec3{1, 1, 1})
	point.AttenuationQuadratic = 0.00001
	point.Place(math.Vec3{0, 2, 2})
	s.AddAmbientLight(ambient)
	s.AddPointLight(point)

	//skybox := graphics.NewCubeMapUniform(math.NewVec4(0.2, 0.2, 0.2, 0))
	//s.AddSkybox(skybox)

	c := camera.NewPerspectiveCamera(60, 1, 0.1, 50)
	c.Place(point.Position)

	input.AddCameraFPSControls(c, 0.1)
	input.KeySpace.Listen(func(action input.Action) {
		s.PointLights[0].Place(c.Position)
	})

	for i := 0; i < 100; i++ {
		car.RotateY(+0.01)

		c.SetAspect(window.Aspect())
		renderer.Clear()
		renderer.RenderScene(s, c)
		renderer.RenderText(math.Vec2{-1, +1}, graphics.RenderStats.String(), 0.05)
		renderer.Render()
		graphics.RenderStats.Reset()

		window.Update()
		input.Update()
	}
}
