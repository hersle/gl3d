package graphics

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	_ "github.com/hersle/gl3d/window" // initialize graphics
)

type RenderCommand struct {
	vertexCount int
	offset      int
	state       *RenderState
}

func NewRenderCommand(vertexCount, offset int, state *RenderState) *RenderCommand {
	var cmd RenderCommand
	cmd.vertexCount = vertexCount
	cmd.offset = offset
	cmd.state = state
	return &cmd
}

func (cmd *RenderCommand) Execute() {
	cmd.state.Apply()
	if cmd.state.Program.va.hasIndexBuffer {
		gl.DrawElements(uint32(cmd.state.PrimitiveType), int32(cmd.vertexCount), gl.UNSIGNED_INT, nil)
	} else {
		gl.DrawArrays(uint32(cmd.state.PrimitiveType), int32(cmd.offset), int32(cmd.vertexCount))
	}

	RenderStats.DrawCallCount++
	RenderStats.VertexCount += cmd.vertexCount
}
