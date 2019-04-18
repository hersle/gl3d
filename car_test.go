package main

import (
	"github.com/hersle/gl3d/window"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	//"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/render"
	//"github.com/hersle/gl3d/scene"
	"github.com/hersle/gl3d/input"
	"testing"
)

func TestMain(t *testing.M) {
	println("hey\n\n")
	renderer, err := render.NewRenderer()
	if err != nil {
		panic(err)
	}
	println("created renderer")

	car, err := object.ReadMesh("assets/objects/sportscar/sportscar.obj")
	if err != nil {
		panic(err)
	}
	println("loaded car")

	//ambient := light.NewAmbientLight(math.NewVec3(0.5, 0.5, 0.5))

	c := camera.NewPerspectiveCamera(60, 1, 0.1, 50)
	c.Place(math.NewVec3(1, 1, 1))

	input.AddCameraFPSControls(c)

	for !window.ShouldClose() {
		println("loop")

		car.RotateY(+0.01)

		c.SetAspect(window.Aspect())
		renderer.SubmitMesh(car, c)
		graphics.RenderStats.Reset()

		window.Update()
		input.Update()
	}
}
