package graphics

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	_ "github.com/hersle/gl3d/window" // initialize graphics
)

type RenderCommand struct {
	vertexCount int
	state       *RenderState
}

func NewRenderCommand(vertexCount int, state *RenderState) *RenderCommand {
	var cmd RenderCommand
	cmd.vertexCount = vertexCount
	cmd.state = state
	return &cmd
}

func (cmd *RenderCommand) Execute() {
	cmd.state.apply()
	if cmd.state.Program.va.indexBuffer == nil {
		gl.DrawArrays(uint32(cmd.state.PrimitiveType), 0, int32(cmd.vertexCount))
	} else {
		var gltype uint32
		switch cmd.state.Program.va.indexBuffer.index.Size() {
		case 1: // 8 bits
			gltype = gl.UNSIGNED_BYTE
		case 2: // 16 bits
			gltype = gl.UNSIGNED_SHORT
		case 4: // 32 bits
			gltype = gl.UNSIGNED_INT
		default:
			panic("invalid index buffer type")
		}
		gl.DrawElements(uint32(cmd.state.PrimitiveType), int32(cmd.vertexCount), gltype, nil)
	}

	RenderStats.DrawCallCount++
	RenderStats.VertexCount += cmd.vertexCount
}
