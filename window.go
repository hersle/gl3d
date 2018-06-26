package main

import (
	"github.com/go-gl/glfw/v3.2/glfw"
)

type Window struct {
	glfwWin *glfw.Window
}

func NewWindow(width, height int, title string) (*Window, error) {
	glfw.Init()

	glfw.WindowHint(glfw.ClientAPI, glfw.OpenGLAPI)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 5)

	var w Window
	var err error
	w.glfwWin, err = glfw.CreateWindow(width, height, title, nil, nil)
	return &w, err
}

func (w *Window) MakeContextCurrent() {
	w.glfwWin.MakeContextCurrent()
}

func (w *Window) ShouldClose() bool {
	return w.glfwWin.ShouldClose()
}

func (w *Window) updateGraphics() {
	w.glfwWin.SwapBuffers()
}

func (w *Window) updateEvents() {
	glfw.PollEvents()
}

func (w *Window) Update() {
	w.updateGraphics()
	w.updateEvents()
}
