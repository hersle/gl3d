package main

import (
	"os"
	"time"
	"github.com/go-gl/glfw/v3.2/glfw"
)

func main() {
	win, err := NewWindow(400, 400, "GL3D")
	if err != nil {
		panic(err)
	}

	renderer, err := NewRenderer(win)
	if err != nil {
		panic(err)
	}

	s := NewScene()
	if len(os.Args) == 2{
		filename := os.Args[1]
		model, err := ReadMesh(filename)
		if err != nil {
			panic(err)
		}
		s.AddMesh(model)
	}
	floor, err := ReadMesh("objects/floor.3d")
	if err != nil {
		panic(err)
	}
	s.AddMesh(floor)

	c := NewCamera(NewVec3(0, 0, 0), NewVec3(0, 0, 1), NewVec3(0, 1, 0), 60, 1, 0.01, 100)

	//var dt float32
	var time1, time2 time.Time
	time1 = time.Now()

	for !win.ShouldClose() {
		c.SetAspect(win.Aspect())
		renderer.SetFullViewport(win)
		renderer.Clear()
		renderer.Render(s, c)
		win.Update()

		time2 = time.Now()
		_ = time2.Sub(time1).Seconds()
		time1 = time2

		if win.glfwWin.GetKey(glfw.KeyW) == glfw.Press {
			c.MoveBy(c.fwd.Scale(+0.1))
		}
		if win.glfwWin.GetKey(glfw.KeyS) == glfw.Press {
			c.MoveBy(c.fwd.Scale(-0.1))
		}
		if win.glfwWin.GetKey(glfw.KeyD) == glfw.Press {
			c.MoveBy(c.right.Scale(+0.1))
		}
		if win.glfwWin.GetKey(glfw.KeyA) == glfw.Press {
			c.MoveBy(c.right.Scale(-0.1))
		}
		if win.glfwWin.GetKey(glfw.KeyUp) == glfw.Press {
			c.Rotate(c.right, +0.03)
		}
		if win.glfwWin.GetKey(glfw.KeyDown) == glfw.Press {
			c.Rotate(c.right, -0.03)
		}
		if win.glfwWin.GetKey(glfw.KeyLeft) == glfw.Press {
			c.Rotate(NewVec3(0, 1, 0), +0.03)
		}
		if win.glfwWin.GetKey(glfw.KeyRight) == glfw.Press {
			c.Rotate(NewVec3(0, 1, 0), -0.03)
		}
	}
}
