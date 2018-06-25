package main

import (
	"time"
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
	s.AddMesh(&mesh1)

	c := NewCamera()

	q := NewEventQueue(win)

	var dt float64
	var time1, time2 time.Time
	time1 = time.Now()

	for !win.ShouldClose() {
		renderer.Clear()
		renderer.Render(s, c)
		renderer.Flush()
		win.Update()

		time2 = time.Now()
		dt = time2.Sub(time1).Seconds()
		time1 = time2

		for !q.empty() {
			e := q.PopEvent()
			switch e.(type) {
				case *ResizeEvent:
					e := e.(*ResizeEvent)
					renderer.SetViewport(0, 0, e.Width, e.Height)
				case *MoveEvent:
					e := e.(*MoveEvent)
					var sign float64
					if e.start {
						sign = +1
					} else {
						sign = -1
					}
					switch e.dir {
						case DirectionLeft:
							println("left")
							c.Accelerate(c.right.Scale(sign * +1))
						case DirectionRight:
							c.Accelerate(c.right.Scale(sign * -1))
							println("right")
						case DirectionForward:
							c.Accelerate(c.fwd.Scale(sign * +1))
						case DirectionBackward:
							c.Accelerate(c.fwd.Scale(sign * -1))
					}
			}
		}

		c.Update(dt)
	}
}
