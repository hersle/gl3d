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

	q := NewEventQueue(win)

	for !win.ShouldClose() {
		c.Move(Vec3{0.001, 0.0, 0.0})

		renderer.Clear()
		renderer.Render(s, c)
		renderer.Flush()
		win.Update()

		for !q.empty() {
			e := q.PopEvent()
			switch e.(type) {
				case *ResizeEvent:
					e := e.(*ResizeEvent)
					renderer.SetViewport(0, 0, e.Width, e.Height)
					println("resize event")
			}
		}
	}
}
