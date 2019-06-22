package scene

import (
	"testing"
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
)

func TestString(t *testing.T) {
	n := NewScene()
	n.AddAmbientLight(light.NewAmbientLight(math.Vec3{0, 0, 0}))
	n.AddAmbientLight(light.NewAmbientLight(math.Vec3{0, 0, 0}))
	n.Children[0].AddAmbientLight(light.NewAmbientLight(math.Vec3{0, 0, 0}))
	println(n.String())
}
