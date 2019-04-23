package input

import (
	_ "github.com/hersle/gl3d/window" // initialize
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/math"
)

func AddCameraFPSControls(c *camera.PerspectiveCamera, speed float32) {
	KeyW.Listen(func(action Action) {
	switch action {
	case Hold:
		c.Translate(c.Forward().Scale(+speed))
	}
	})

	KeyS.Listen(func(action Action) {
	switch action {
	case Hold:
		c.Translate(c.Forward().Scale(-speed))
	}
	})

	KeyA.Listen(func(action Action) {
	switch action {
	case Hold:
		c.Translate(c.Right().Scale(-speed))
	}
	})

	KeyD.Listen(func(action Action) {
	switch action {
	case Hold:
		c.Translate(c.Right().Scale(+speed))
	}
	})

	KeyUp.Listen(func(action Action) {
	switch action {
	case Hold:
		c.Rotate(c.Right(), +0.03)
	}
	})

	KeyDown.Listen(func(action Action) {
	switch action {
	case Hold:
		c.Rotate(c.Right(), -0.03)
	}
	})

	KeyLeft.Listen(func(action Action) {
	switch action {
	case Hold:
		c.Rotate(math.NewVec3(0, 1, 0), +0.03)
	}
	})

	KeyRight.Listen(func(action Action) {
	switch action {
	case Hold:
		c.Rotate(math.NewVec3(0, 1, 0), -0.03)
	}
	})
}
