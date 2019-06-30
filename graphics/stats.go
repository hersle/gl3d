package graphics

import (
	"fmt"
)

type Statistics struct {
	DrawCallCount   int
	VertexCount     int
}

var Stats *Statistics = &Statistics{}

func (stats *Statistics) Reset() {
	stats.DrawCallCount = 0
	stats.VertexCount = 0
}

func (stats *Statistics) String() string {
	text := fmt.Sprint(stats.DrawCallCount) + " draw calls, "
	text += fmt.Sprint(stats.VertexCount) + " vertices"
	return text
}
