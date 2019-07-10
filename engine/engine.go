package engine

import (
	"github.com/hersle/gl3d/scene"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/window"
	"github.com/hersle/gl3d/console"
	"github.com/hersle/gl3d/render"
	"github.com/hersle/gl3d/input"
	"github.com/hersle/gl3d/utils"
	"time"
	"flag"
	"log"
	"strings"
	"reflect"
	"strconv"
	"fmt"
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

	frameCounter *utils.FrequencyCounter
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

	eng.frameCounter = utils.NewFrequencyCounter()
	eng.frameCounter.Interval = 10
}

func (eng *Engine) Update(dt float32) {
	eng.Camera.SetAspect(window.Aspect())
	if eng.UpdateCustom != nil {
		eng.UpdateCustom(dt)
	}
}

func (eng *Engine) React(dt float32) {
	input.Update()  // TODO: make line order not matter

	if eng.consoleActive {
		return
	}

	if input.KeyP.JustPressed() {
		eng.paused = !eng.paused
	}

	moveSpeed := 5.0 * dt
	lookSpeed := 2.0 * dt

	if input.KeyW.Held() {
		eng.Camera.Translate(eng.Camera.Forward().Scale(+moveSpeed))
	}

	if input.KeyS.Held() {
		eng.Camera.Translate(eng.Camera.Forward().Scale(-moveSpeed))
	}

	if input.KeyA.Held() {
		eng.Camera.Translate(eng.Camera.Right().Scale(-moveSpeed))
	}

	if input.KeyD.Held() {
		eng.Camera.Translate(eng.Camera.Right().Scale(+moveSpeed))
	}

	if input.KeyUp.Held() {
		eng.Camera.Rotate(eng.Camera.Right(), +lookSpeed)
	}

	if input.KeyDown.Held() {
		eng.Camera.Rotate(eng.Camera.Right(), -lookSpeed)
	}

	if input.KeyLeft.Held() {
		eng.Camera.Rotate(math.Vec3{0, 1, 0}, +lookSpeed)
	}

	if input.KeyRight.Held() {
		eng.Camera.Rotate(math.Vec3{0, 1, 0}, -lookSpeed)
	}
}

func (eng *Engine) Render() {
	eng.renderer.Clear()
	eng.renderer.RenderScene(eng.Scene, eng.Camera)
	if eng.consoleActive {
		eng.renderer.RenderText(math.Vec2{-1, +1}, eng.console.String(), 0.05, render.TopLeft)
	}

	period := eng.frameCounter.Period() * 1000 // ms
	framerate := eng.frameCounter.Frequency() // Hz
	text := fmt.Sprintf("%d Hz\n%.1f ms", framerate, period)
	eng.renderer.RenderText(math.Vec2{+1, +1}, text, 0.10, render.TopRight)

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
		dt := float32(t.Sub(t0).Seconds())

		eng.React(dt)
		if !eng.paused {
			eng.Update(dt)
		}
		eng.Render()

		t0 = t

		eng.frameCounter.Count()
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
	case "shadowkernelsize":
		ptr = &eng.renderer.MeshRenderer.ShadowKernelSize
	case "materialambient":
		ptr = &eng.renderer.MeshRenderer.MaterialAmbientEnabled
	case "materialdiffuse":
		ptr = &eng.renderer.MeshRenderer.MaterialDiffuseEnabled
	case "materialspecular":
		ptr = &eng.renderer.MeshRenderer.MaterialSpecularEnabled
	case "materialalpha":
		ptr = &eng.renderer.MeshRenderer.MaterialAlphaEnabled
	case "materialnormal":
		ptr = &eng.renderer.MeshRenderer.MaterialNormalEnabled
	case "shadows":
		ptr = &eng.renderer.MeshRenderer.ShadowsEnabled
	case "wireframe":
		ptr = &eng.renderer.MeshRenderer.Wireframe
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
		case *int:
			ptr := ptr.(*int)
			val, err := strconv.ParseInt(fields[1], 0, 0)
			if err == nil {
				*ptr = int(val)
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
