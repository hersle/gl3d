package main

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
	s.AddMesh(&mesh1)

	c := NewCamera()

	for !win.ShouldClose() {
		renderer.Clear()
		renderer.Render(s, c)
		renderer.Flush()
		win.Update()
		c.Move(Vec3{0.001, 0.0, 0.0})
	}
}
