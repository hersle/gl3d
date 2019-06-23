package graphics

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

type BufferTest int

const (
	AlwaysTest BufferTest = iota
	UnknownTest
	NeverTest
	LessTest
	LessEqualTest
	EqualTest
	NotEqualTest
	GreaterTest
	GreaterEqualTest
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

type StencilOperation int

const (
	KeepStencilOperation StencilOperation = iota
	ZeroStencilOperation
	ReplaceStencilOperation
	IncrementStencilOperation
	DecrementStencilOperation
)

// TODO: enable sorting of these states to reduce state changes?
type RenderState struct {
	Program                *ShaderProgram
	Framebuffer            *Framebuffer
	DepthTest              BufferTest
	StencilTest            BufferTest
	StencilTestRef         int
	StencilStencilFailOperation StencilOperation
	StencilDepthFailOperation StencilOperation
	StencilDepthPassOperation StencilOperation
	BlendSourceFactor      BlendFactor
	BlendDestinationFactor BlendFactor
	Cull                   CullMode
	TriangleMode           TriangleMode
	PrimitiveType          Primitive
}

var currentState RenderState

func NewRenderState() *RenderState {
	var rs RenderState
	rs.DisableBlending()
	rs.Framebuffer = DefaultFramebuffer
	return &rs
}

func (rs *RenderState) DisableBlending() {
	rs.BlendSourceFactor = OneBlendFactor
	rs.BlendDestinationFactor = ZeroBlendFactor
}

func (rs *RenderState) apply() {
	if currentState.Program != rs.Program {
		switch rs.Program {
		case nil:
			panic("tried to apply a render state with no shader program")
		default:
			rs.Program.va.Bind()
			rs.Program.Bind()
		}
		currentState.Program = rs.Program
	}

	if currentState.Framebuffer != rs.Framebuffer {
		switch rs.Framebuffer {
		case nil:
			panic("tried to apply a render state with no framebuffer")
		default:
			rs.Framebuffer.bindDraw()
			gl.Viewport(0, 0, int32(rs.Framebuffer.Width()), int32(rs.Framebuffer.Height()))
		}
		currentState.Framebuffer = rs.Framebuffer
	}

	if currentState.DepthTest != rs.DepthTest {
		switch rs.DepthTest {
		case NeverTest:
			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(gl.NEVER)
		case LessTest:
			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(gl.LESS)
		case LessEqualTest:
			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(gl.LEQUAL)
		case EqualTest:
			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(gl.EQUAL)
		case NotEqualTest:
			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(gl.NOTEQUAL)
		case GreaterTest:
			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(gl.GREATER)
		case GreaterEqualTest:
			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(gl.GEQUAL)
		case AlwaysTest:
			gl.Disable(gl.DEPTH_TEST)
		default:
			panic("tried to apply a render state with unknown depth test")
		}
		currentState.DepthTest = rs.DepthTest
	}

	if currentState.StencilTest != rs.StencilTest || currentState.StencilTestRef != rs.StencilTestRef {
		switch rs.StencilTest {
		case NeverTest:
			gl.Enable(gl.STENCIL_TEST)
			gl.StencilFunc(gl.NEVER, int32(rs.StencilTestRef), 0xFF)
		case LessTest:
			gl.Enable(gl.STENCIL_TEST)
			gl.StencilFunc(gl.LESS, int32(rs.StencilTestRef), 0xFF)
		case LessEqualTest:
			gl.Enable(gl.STENCIL_TEST)
			gl.StencilFunc(gl.LEQUAL, int32(rs.StencilTestRef), 0xFF)
		case EqualTest:
			gl.Enable(gl.STENCIL_TEST)
			gl.StencilFunc(gl.EQUAL, int32(rs.StencilTestRef), 0xFF)
		case NotEqualTest:
			gl.Enable(gl.STENCIL_TEST)
			gl.StencilFunc(gl.NOTEQUAL, int32(rs.StencilTestRef), 0xFF)
		case GreaterTest:
			gl.Enable(gl.STENCIL_TEST)
			gl.StencilFunc(gl.GREATER, int32(rs.StencilTestRef), 0xFF)
		case GreaterEqualTest:
			gl.Enable(gl.STENCIL_TEST)
			gl.StencilFunc(gl.GEQUAL, int32(rs.StencilTestRef), 0xFF)
		case AlwaysTest:
			gl.Disable(gl.STENCIL_TEST)
		default:
			panic("tried to apply a render state with unknown stencil test")
		}
		currentState.StencilTest = rs.StencilTest
		currentState.StencilTestRef = rs.StencilTestRef
	}

	gl.StencilOp(uint32(rs.StencilStencilFailOperation), uint32(rs.StencilDepthFailOperation), uint32(rs.StencilDepthPassOperation))
	currentState.StencilStencilFailOperation = rs.StencilStencilFailOperation
	currentState.StencilDepthFailOperation = rs.StencilDepthFailOperation
	currentState.StencilDepthPassOperation = rs.StencilDepthPassOperation

	if currentState.BlendSourceFactor != rs.BlendSourceFactor || currentState.BlendDestinationFactor != rs.BlendDestinationFactor {
		var factors [2]BlendFactor
		var funcs [2]uint32
		factors[0] = rs.BlendSourceFactor
		factors[1] = rs.BlendDestinationFactor
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
		currentState.BlendSourceFactor = rs.BlendSourceFactor
		currentState.BlendDestinationFactor = rs.BlendDestinationFactor
	}

	if currentState.Cull != rs.Cull {
		switch rs.Cull {
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
		currentState.Cull = rs.Cull
	}

	if currentState.TriangleMode != rs.TriangleMode {
		switch rs.TriangleMode {
		case PointTriangleMode:
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.POINT)
		case LineTriangleMode:
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
		case TriangleTriangleMode:
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
		default:
			panic("tried to apply a render state with an unknown polygonmode")
		}
		currentState.TriangleMode = rs.TriangleMode
	}
}

func init() {
	gl.Enable(gl.BLEND)

	// initialize cached state to default OpenGL values TODO: run apply with it?
	currentState.Program = nil
	currentState.Framebuffer = nil
	currentState.DepthTest = AlwaysTest
	currentState.StencilTest = AlwaysTest
	currentState.StencilTestRef = 0
	currentState.StencilStencilFailOperation = KeepStencilOperation
	currentState.StencilDepthFailOperation = KeepStencilOperation
	currentState.StencilDepthPassOperation = KeepStencilOperation
	currentState.DisableBlending()
	currentState.Cull = CullNothing
	currentState.TriangleMode = TriangleTriangleMode
}
