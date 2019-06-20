package graphics

import (
	"fmt"
	_ "github.com/hersle/gl3d/window" // initialize graphics
)

type RenderStatistics struct {
	DrawCallCount   int
	VertexCount     int
}

var RenderStats *RenderStatistics = &RenderStatistics{}

func (stats *RenderStatistics) Reset() {
	stats.DrawCallCount = 0
	stats.VertexCount = 0
}

func (stats *RenderStatistics) String() string {
	text := fmt.Sprint(stats.DrawCallCount) + " draw calls, "
	text += fmt.Sprint(stats.VertexCount) + " vertices"
	return text
}
