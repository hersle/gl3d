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

	for !win.ShouldClose() {
		v1 := Vec3{0.5, 0.5, 0}
		v2 := Vec3{0.0, 0.0, 0}
		v3 := Vec3{0.5, 0.0, 0}

		renderer.Clear()
		renderer.DrawTriangle(v1, v2, v3)
		renderer.Flush()
		win.Update()
	}
}
