package utils

import (
	"github.com/hersle/gl3d/math"
	"math/rand"
)

func RandomDirection() math.Vec3 {
	x := float32(rand.NormFloat64())
	y := float32(rand.NormFloat64())
	z := float32(rand.NormFloat64())
	return math.Vec3{x, y, z}.Norm()
}
