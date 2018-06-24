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
		color := NewColor(0xff, 0xff, 0xff, 0x80)

		renderer.Clear()
		renderer.RenderTriangle(v1, v2, v3, color)
		renderer.Flush()
		win.Update()
	}
}
