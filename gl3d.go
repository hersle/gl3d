package main

import (
	"github.com/hersle/gl3d/engine"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/light"
	"flag"
)

func main() {
	flag.Parse()

	eng := engine.NewEngine()

	eng.InitializeCustom = func() {
		eng.Scene.AddAmbientLight(light.NewAmbientLight(math.Vec3{0.5, 0.5, 0.5}))
		for _, filename := range flag.Args() {
			model, err := object.ReadMesh(filename)
			if err != nil {
				panic(err)
			}
			eng.Scene.AddMesh(model)
		}
	}

	eng.Run()
}
