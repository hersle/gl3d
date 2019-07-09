package engine

import (
	"github.com/hersle/gl3d/scene"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/window"
	"github.com/hersle/gl3d/console"
	"github.com/hersle/gl3d/render"
	"github.com/hersle/gl3d/input"
	"time"
)

type Engine struct {
	Scene *scene.Scene
	Camera *camera.PerspectiveCamera

	console *console.Console

	renderer *render.Renderer

	UpdateCustom func(dt float32)
	InitializeCustom func()
}

func NewEngine() *Engine {
	var eng Engine
	return &eng
}

func (eng *Engine) Initialize() {
	var err error

	eng.renderer, err = render.NewRenderer()
	if err != nil {
		panic(err)
	}

	eng.Scene = scene.NewScene()
	eng.Camera = camera.NewPerspectiveCamera(60, 1, 0.1, 50)

	if eng.InitializeCustom != nil {
		eng.InitializeCustom()
	}
}

func (eng *Engine) Update(dt float32) {
	eng.Camera.SetAspect(window.Aspect())
	if eng.UpdateCustom != nil {
		eng.UpdateCustom(dt)
	}
}

func (eng *Engine) React() {
	input.Update()  // TODO: make line order not matter

	speed := float32(0.1)

	if input.KeyW.Held() {
		eng.Camera.Translate(eng.Camera.Forward().Scale(+speed))
	}

	if input.KeyS.Held() {
		eng.Camera.Translate(eng.Camera.Forward().Scale(-speed))
	}

	if input.KeyA.Held() {
		eng.Camera.Translate(eng.Camera.Right().Scale(-speed))
	}

	if input.KeyD.Held() {
		eng.Camera.Translate(eng.Camera.Right().Scale(+speed))
	}

	if input.KeyUp.Held() {
		eng.Camera.Rotate(eng.Camera.Right(), +0.03)
	}

	if input.KeyDown.Held() {
		eng.Camera.Rotate(eng.Camera.Right(), -0.03)
	}

	if input.KeyLeft.Held() {
		eng.Camera.Rotate(math.Vec3{0, 1, 0}, +0.03)
	}

	if input.KeyRight.Held() {
		eng.Camera.Rotate(math.Vec3{0, 1, 0}, -0.03)
	}
}

func (eng *Engine) Render() {
	eng.renderer.Clear()
	eng.renderer.RenderScene(eng.Scene, eng.Camera)
	eng.renderer.Render()
	window.Update()
}

func (eng *Engine) Run() {
	eng.Initialize()
	t0 := time.Now()
	for !window.ShouldClose() {
		t := time.Now()
		dt := t.Sub(t0).Seconds()

		eng.React()
		eng.Update(float32(dt))
		eng.Render()

		t0 = t
	}
}
