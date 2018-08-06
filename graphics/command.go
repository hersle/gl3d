package graphics

import (
	_ "github.com/hersle/gl3d/window" // initialize graphics
	"github.com/go-gl/gl/v4.5-core/gl"
)

type RenderCommand struct {
	primitiveType uint32
	vertexCount   int
	offset        int
	state         *RenderState
}

func NewRenderCommand(primitiveType uint32, vertexCount, offset int, state *RenderState) *RenderCommand {
	var cmd RenderCommand
	cmd.primitiveType = primitiveType
	cmd.vertexCount = vertexCount
	cmd.offset = offset
	cmd.state = state
	return &cmd
}

func (cmd *RenderCommand) Execute() {
	cmd.state.Apply()
	if cmd.state.prog.va.hasIndexBuffer {
		gl.DrawElements(cmd.primitiveType, int32(cmd.vertexCount), gl.UNSIGNED_INT, nil)
	} else {
		gl.DrawArrays(cmd.primitiveType, int32(cmd.offset), int32(cmd.vertexCount))
	}

	RenderStats.DrawCallCount++
	RenderStats.VertexCount += cmd.vertexCount
}
