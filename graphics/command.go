package graphics

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	_ "github.com/hersle/gl3d/window" // initialize graphics
)

// TODO: absorb into renderstate?

type Command struct {
	vertexCount int
	state       *State
}

func NewCommand(vertexCount int, state *State) *Command {
	var cmd Command
	cmd.vertexCount = vertexCount
	cmd.state = state
	return &cmd
}

func (cmd *Command) Execute() {
	cmd.state.apply()
	if cmd.state.Program.indexBuffer == nil {
		gl.DrawArrays(uint32(cmd.state.PrimitiveType), 0, int32(cmd.vertexCount))
	} else {
		var gltype uint32
		switch cmd.state.Program.indexBuffer.index.Size() {
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

	Stats.DrawCallCount++
	Stats.VertexCount += cmd.vertexCount
}
