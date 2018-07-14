package main

import (
	"os"
	"github.com/go-gl/glfw/v3.2/glfw"
)

func main() {
	win, err := NewWindow(1280, 720, "GL3D")
	if err != nil {
		panic(err)
	}

	renderer, err := NewRenderer(win)
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
			model.Translate(NewVec3(-2, -5, +5))
		}
		if filename == "objects/sponza.obj" {
			model.Scale(NewVec3(0.02, 0.02, 0.02))
		}
		s.AddMesh(model)
	}

	c := NewCamera(60, 1, 0.01, 100)

	var camFactor float32

	for !win.ShouldClose() {
		c.SetAspect(win.Aspect())
		renderer.SetFullViewport(win)
		renderer.Clear()
		renderer.Render(s, c)
		win.Update()

		if win.glfwWin.GetKey(glfw.KeyLeftShift) == glfw.Press {
			camFactor = 0.1 // for precise camera controls
		} else {
			camFactor = 1.0
		}

		if win.glfwWin.GetKey(glfw.KeyW) == glfw.Press {
			c.Translate(c.forward.Scale(camFactor * +0.1))
		}
		if win.glfwWin.GetKey(glfw.KeyS) == glfw.Press {
			c.Translate(c.forward.Scale(camFactor * -0.1))
		}
		if win.glfwWin.GetKey(glfw.KeyD) == glfw.Press {
			c.Translate(c.right.Scale(camFactor * +0.1))
		}
		if win.glfwWin.GetKey(glfw.KeyA) == glfw.Press {
			c.Translate(c.right.Scale(camFactor * -0.1))
		}
		if win.glfwWin.GetKey(glfw.KeyUp) == glfw.Press {
			c.Rotate(c.right, camFactor * +0.03)
		}
		if win.glfwWin.GetKey(glfw.KeyDown) == glfw.Press {
			c.Rotate(c.right, camFactor * -0.03)
		}
		if win.glfwWin.GetKey(glfw.KeyLeft) == glfw.Press {
			c.Rotate(NewVec3(0, 1, 0), camFactor * +0.03)
		}
		if win.glfwWin.GetKey(glfw.KeyRight) == glfw.Press {
			c.Rotate(NewVec3(0, 1, 0), camFactor * -0.03)
		}
	}
}
