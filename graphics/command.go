package graphics

import (
	_ "github.com/hersle/gl3d/window" // initialize graphics
	"github.com/go-gl/gl/v4.5-core/gl"
)

type RenderCommand struct {
	primitive Primitive
	vertexCount   int
	offset        int
	state         *RenderState
}

type Primitive int

const (
	Point Primitive = Primitive(gl.POINTS)
	Line Primitive = Primitive(gl.LINES)
	Triangle Primitive = Primitive(gl.TRIANGLES)
)

func NewRenderCommand(primitive Primitive, vertexCount, offset int, state *RenderState) *RenderCommand {
	var cmd RenderCommand
	cmd.primitive = primitive
	cmd.vertexCount = vertexCount
	cmd.offset = offset
	cmd.state = state
	return &cmd
}

func (cmd *RenderCommand) Execute() {
	cmd.state.Apply()
	if cmd.state.prog.va.hasIndexBuffer {
		gl.DrawElements(uint32(cmd.primitive), int32(cmd.vertexCount), gl.UNSIGNED_INT, nil)
	} else {
		gl.DrawArrays(uint32(cmd.primitive), int32(cmd.offset), int32(cmd.vertexCount))
	}

	RenderStats.DrawCallCount++
	RenderStats.VertexCount += cmd.vertexCount
}
