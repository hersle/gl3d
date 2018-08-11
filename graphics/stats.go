package graphics

import (
	_ "github.com/hersle/gl3d/window" // initialize graphics
	"time"
)

type RenderStatistics struct {
	DrawCallCount int
	VertexCount   int
	frameStartTime time.Time
	FramesPerSecond int
}

var RenderStats *RenderStatistics = &RenderStatistics{}

func (stats *RenderStatistics) Reset() {
	stats.DrawCallCount = 0
	stats.VertexCount = 0
	frameEndTime := time.Now()
	stats.FramesPerSecond = int(1 / (frameEndTime.Sub(stats.frameStartTime).Seconds()))
	stats.frameStartTime = frameEndTime
}
