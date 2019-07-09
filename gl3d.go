package main

import (
	"github.com/hersle/gl3d/engine"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/light"
	"flag"
	"runtime/pprof"
	"os"
)

var cpuprofile = flag.String("cpuprofile", "", "write CPU profile to file")

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

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
