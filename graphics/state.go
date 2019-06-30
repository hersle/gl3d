package graphics

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

type DepthTest int

const (
	AlwaysDepthTest DepthTest = iota
	UnknownDepthTest
	NeverDepthTest
	LessDepthTest
	LessEqualDepthTest
	EqualDepthTest
	NotEqualDepthTest
	GreaterDepthTest
	GreaterEqualDepthTest
)

type BlendFactor int

const (
	ZeroBlendFactor BlendFactor = iota
	OneBlendFactor
	SourceColorBlendFactor
	OneMinusSourceColorBlendFactor
	DestinationColorBlendFactor
	OneMinusDestinationColorBlendFactor
	SourceAlphaBlendFactor
	OneMinusSourceAlphaBlendFactor
	DestinationAlphaBlendFactor
	OneMinusDestinationAlphaBlendFactor
)

type CullMode int

const (
	CullNothing CullMode = iota
	CullFront
	CullBack
)

type TriangleMode int

const (
	TriangleTriangleMode TriangleMode = iota
	PointTriangleMode
	LineTriangleMode
)

type Primitive int

const (
	Point    Primitive = Primitive(gl.POINTS)
	Line     Primitive = Primitive(gl.LINES)
	Triangle Primitive = Primitive(gl.TRIANGLES)
)

// TODO: enable sorting of these states to reduce state changes?
type State struct {
	Program                *ShaderProgram
	Framebuffer            *Framebuffer
	DepthTest              DepthTest
	BlendSourceFactor      BlendFactor
	BlendDestinationFactor BlendFactor
	Cull                   CullMode
	TriangleMode           TriangleMode
	PrimitiveType          Primitive
}

var currentState State

func NewState() *State {
	var state State
	state.DisableBlending()
	state.Framebuffer = DefaultFramebuffer
	return &state
}

func (state *State) DisableBlending() {
	state.BlendSourceFactor = OneBlendFactor
	state.BlendDestinationFactor = ZeroBlendFactor
}

func (state *State) apply() {
	if currentState.Program != state.Program {
		switch state.Program {
		case nil:
			panic("tried to apply a render state with no shader program")
		default:
			state.Program.bind()
		}
		currentState.Program = state.Program
	}

	if currentState.Framebuffer != state.Framebuffer {
		switch state.Framebuffer {
		case nil:
			panic("tried to apply a render state with no framebuffer")
		default:
			state.Framebuffer.bindDraw()
			gl.Viewport(0, 0, int32(state.Framebuffer.Width()), int32(state.Framebuffer.Height()))
		}
		currentState.Framebuffer = state.Framebuffer
	}

	if currentState.DepthTest != state.DepthTest {
		switch state.DepthTest {
		case NeverDepthTest:
			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(gl.NEVER)
		case LessDepthTest:
			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(gl.LESS)
		case LessEqualDepthTest:
			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(gl.LEQUAL)
		case EqualDepthTest:
			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(gl.EQUAL)
		case NotEqualDepthTest:
			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(gl.NOTEQUAL)
		case GreaterDepthTest:
			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(gl.GREATER)
		case GreaterEqualDepthTest:
			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(gl.GEQUAL)
		case AlwaysDepthTest:
			gl.Disable(gl.DEPTH_TEST)
		default:
			panic("tried to apply a render state with unknown depth test")
		}
		currentState.DepthTest = state.DepthTest
	}

	if currentState.BlendSourceFactor != state.BlendSourceFactor || currentState.BlendDestinationFactor != state.BlendDestinationFactor {
		var factors [2]BlendFactor
		var funcs [2]uint32
		factors[0] = state.BlendSourceFactor
		factors[1] = state.BlendDestinationFactor
		for i := 0; i < 2; i++ {
			switch factors[i] {
			case ZeroBlendFactor:
				funcs[i] = gl.ZERO
			case OneBlendFactor:
				funcs[i] = gl.ONE
			case SourceColorBlendFactor:
				funcs[i] = gl.SRC_COLOR
			case OneMinusSourceColorBlendFactor:
				funcs[i] = gl.ONE_MINUS_SRC_COLOR
			case DestinationColorBlendFactor:
				funcs[i] = gl.DST_COLOR
			case OneMinusDestinationColorBlendFactor:
				funcs[i] = gl.ONE_MINUS_DST_COLOR
			case SourceAlphaBlendFactor:
				funcs[i] = gl.SRC_ALPHA
			case OneMinusSourceAlphaBlendFactor:
				funcs[i] = gl.ONE_MINUS_SRC_ALPHA
			case DestinationAlphaBlendFactor:
				funcs[i] = gl.DST_ALPHA
			case OneMinusDestinationAlphaBlendFactor:
				funcs[i] = gl.ONE_MINUS_DST_ALPHA
			default:
				panic("tried to apply a render state with an unknown blending factor")
			}
		}
		gl.BlendFunc(funcs[0], funcs[1])
		currentState.BlendSourceFactor = state.BlendSourceFactor
		currentState.BlendDestinationFactor = state.BlendDestinationFactor
	}

	if currentState.Cull != state.Cull {
		switch state.Cull {
		case CullNothing:
			gl.Disable(gl.CULL_FACE)
		case CullFront:
			gl.Enable(gl.CULL_FACE)
			gl.CullFace(gl.FRONT)
		case CullBack:
			gl.Enable(gl.CULL_FACE)
			gl.CullFace(gl.BACK)
		default:
			panic("tried to apply a render state with an unknown culling mode")
		}
		currentState.Cull = state.Cull
	}

	if currentState.TriangleMode != state.TriangleMode {
		switch state.TriangleMode {
		case PointTriangleMode:
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.POINT)
		case LineTriangleMode:
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
		case TriangleTriangleMode:
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
		default:
			panic("tried to apply a render state with an unknown polygonmode")
		}
		currentState.TriangleMode = state.TriangleMode
	}
}

func (state *State) Render(vertexCount int) {
	state.apply()

	if state.Program.indexBuffer == nil {
		gl.DrawArrays(uint32(state.PrimitiveType), 0, int32(vertexCount))
	} else {
		gltype := state.Program.indexBuffer.elementGlType()
		gl.DrawElements(uint32(state.PrimitiveType), int32(vertexCount), gltype, nil)
	}

	Stats.DrawCallCount++
	Stats.VertexCount += vertexCount
}

func init() {
	gl.Enable(gl.BLEND)

	// initialize cached state to default OpenGL values TODO: run apply with it?
	currentState.Program = nil
	currentState.Framebuffer = nil
	currentState.DepthTest = AlwaysDepthTest
	currentState.DisableBlending()
	currentState.Cull = CullNothing
	currentState.TriangleMode = TriangleTriangleMode
}
