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
type RenderState struct {
	Program                *ShaderProgram
	Framebuffer            *Framebuffer
	DepthTest              DepthTest
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
			rs.Program.bind()
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
		currentState.DepthTest = rs.DepthTest
	}

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
	currentState.DepthTest = AlwaysDepthTest
	currentState.DisableBlending()
	currentState.Cull = CullNothing
	currentState.TriangleMode = TriangleTriangleMode
}
