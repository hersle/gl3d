package camera

import (
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
)

type Camera interface {
	Translate(math.Vec3)
	Forward() math.Vec3
	Right() math.Vec3
	Rotate(axis math.Vec3, angle float32)
	ViewMatrix() *math.Mat4
	ProjectionMatrix() *math.Mat4
	Cull(sm *object.SubMesh) bool
}
