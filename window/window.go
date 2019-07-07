// Package window provides access to one window to be used with OpenGL
package window

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"runtime"
)

var Win *glfw.Window

func ShouldClose() bool {
	return Win.ShouldClose()
}

func Size() (int, int) {
	return Win.GetSize()
}

func Aspect() float32 {
	width, height := Size()
	return float32(width) / float32(height)
}

func updateGraphics() {
	Win.SwapBuffers()
}

func updateEvents() {
	glfw.PollEvents()
}

func Update() {
	updateGraphics()
	updateEvents()
}

func init() {
	var err error

	runtime.LockOSThread()

	glfw.Init()
	glfw.WindowHint(glfw.ClientAPI, glfw.OpenGLAPI)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 5)

	width := 800
	height := 800
	title := "GL3D"
	Win, err = glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		panic(err)
	}

	Win.MakeContextCurrent()

	err = gl.Init()
	if err != nil {
		panic(err)
	}

	Update()
}
