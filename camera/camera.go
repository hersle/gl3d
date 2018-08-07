package camera

import (
	"github.com/hersle/gl3d/math"
)

type Camera interface {
	ViewMatrix() *math.Mat4
	ProjectionMatrix() *math.Mat4
}
