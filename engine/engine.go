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
	"flag"
	"log"
	"strings"
	"reflect"
	"strconv"
)
var frames = flag.Int("frames", -1, "number of frames to run")

type Engine struct {
	Scene *scene.Scene
	Camera *camera.PerspectiveCamera

	console *console.Console
	consoleActive bool

	renderer *render.Renderer

	UpdateCustom func(dt float32)
	InitializeCustom func()

	paused bool
}

func NewEngine() *Engine {
	var eng Engine
	return &eng
}

func (eng *Engine) Initialize() {
	var err error

	eng.console = console.NewConsole()
	log.SetOutput(eng.console)
	log.SetFlags(0)

	eng.renderer, err = render.NewRenderer()
	if err != nil {
		panic(err)
	}

	eng.Scene = scene.NewScene()
	eng.Camera = camera.NewPerspectiveCamera(60, 1, 0.1, 50)

	if eng.InitializeCustom != nil {
		eng.InitializeCustom()
	}

	input.ListenToText(func(char rune) {
		if eng.consoleActive {
			eng.console.AddToPrompt(char)
		}
	})
	input.KeyBackspace.Listen(func(action input.Action) {
		if action == input.Press {
			eng.console.DeleteFromPrompt()
		}
	})
	input.KeyTab.Listen(func(action input.Action) {
		if action == input.Press {
			eng.consoleActive = !eng.consoleActive
		}
	})
	input.KeyEnter.Listen(func(action input.Action) {
		if action == input.Press {
			eng.ExecuteCommand(eng.console.Prompt())
			eng.console.ClearPrompt()
		}
	})
}

func (eng *Engine) Update(dt float32) {
	eng.Camera.SetAspect(window.Aspect())
	if eng.UpdateCustom != nil {
		eng.UpdateCustom(dt)
	}
}

func (eng *Engine) React() {
	input.Update()  // TODO: make line order not matter

	if eng.consoleActive {
		return
	}

	if input.KeyP.JustPressed() {
		eng.paused = !eng.paused
	}

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
	if eng.consoleActive {
		eng.renderer.RenderText(math.Vec2{-1, +1}, eng.console.String(), 0.05)
	}
	eng.renderer.Render()
	window.Update()
}

func (eng *Engine) Run() {
	eng.Initialize()
	t0 := time.Now()
	for !window.ShouldClose() && *frames != 0 {
		if *frames > 0 {
			*frames--
		}

		t := time.Now()
		dt := t.Sub(t0).Seconds()

		eng.React()
		if !eng.paused {
			eng.Update(float32(dt))
		}
		eng.Render()

		t0 = t
	}
}

func (eng *Engine) ExecuteCommand(cmd string) {
	fields := strings.Fields(cmd)

	if len(fields) == 0 || len(fields) >= 3{
		log.Print("invalid number of fields: ", len(fields))
		return
	}

	// get pointer to field value
	var ptr interface{}
	switch fields[0] {
	case "fog":
		ptr = &eng.renderer.Fog
	case "blurradius":
		ptr = &eng.renderer.BlurRadius
	default:
		log.Print("invalid field: ", fields[0])
		return
	}

	// set field value
	if len(fields) == 2 {
		switch ptr.(type) {
		case *bool:
			ptr := ptr.(*bool)
			val, err := strconv.ParseBool(fields[1])
			if err == nil {
				*ptr = val
			} else {
				log.Print("invalid value: ", fields[1])
				return
			}
		case *float32:
			ptr := ptr.(*float32)
			val, err := strconv.ParseFloat(fields[1], 32)
			if err == nil {
				*ptr = float32(val)
			} else {
				log.Print("invalid value: ", fields[1])
				return
			}
		}
	}

	// get field value
	val := reflect.Indirect(reflect.ValueOf(ptr)).Interface()
	log.Print(fields[0], ": ", val)
}
