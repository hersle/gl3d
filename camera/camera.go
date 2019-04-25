package camera

import (
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
)

type Camera interface {
	ViewMatrix() *math.Mat4
	ProjectionMatrix() *math.Mat4
	Cull(sm *object.SubMesh) bool
}
