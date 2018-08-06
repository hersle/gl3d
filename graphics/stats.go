package graphics

import (
	_ "github.com/hersle/gl3d/window" // initialize graphics
)

type RenderStatistics struct {
	DrawCallCount int
	VertexCount   int
}

var RenderStats *RenderStatistics = &RenderStatistics{}

func (stats *RenderStatistics) Reset() {
	stats.DrawCallCount = -1 // TODO: ?
	stats.VertexCount = 0
}
