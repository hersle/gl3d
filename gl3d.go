package main

import (
	//"fmt"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	//"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/render"
	"github.com/hersle/gl3d/scene"
	"github.com/hersle/gl3d/window"
	"github.com/hersle/gl3d/input"
	"os"
	//"time"
)

func main() {
	renderer, err := render.NewRenderer(800, 800)
	if err != nil {
		panic(err)
	}

	var filename string
	if len(os.Args[1:]) > 0 {
		filename = os.Args[1]
	} else {
		filename = "assets/objects/cube/cube.obj"
	}
	model1, err := object.ReadMesh(filename)
	model2, err := object.ReadMesh(filename)
	model2.Translate(math.NewVec3(3, 0, 0))

	scene := scene.NewScene()
	scene.AddMesh(model1)
	scene.AddMesh(model2)

	c := camera.NewPerspectiveCamera(60, 1, 0.1, 50)

	input.AddCameraFPSControls(c)

	for !window.ShouldClose() {
		c.SetAspect(window.Aspect())
		graphics.DefaultFramebuffer.ClearColor(math.NewVec4(0, 0, 0, 1))
		graphics.DefaultFramebuffer.ClearDepth(1)
		renderer.SetViewportSize(window.Size())
		renderer.RenderScene(scene, c)
		window.Update()
		input.Update()
	}
}
