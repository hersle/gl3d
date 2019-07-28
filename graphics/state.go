package graphics

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	_ "github.com/hersle/gl3d/window" // initialize graphics
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
type RenderOptions struct {
	DepthTest              DepthTest
	BlendSourceFactor      BlendFactor
	BlendDestinationFactor BlendFactor
	Cull                   CullMode
	TriangleMode           TriangleMode
	PrimitiveType          Primitive
}

var currentOpts RenderOptions

func NewRenderOptions() *RenderOptions {
	var opts RenderOptions
	opts.DisableBlending()
	return &opts
}

func (opts *RenderOptions) DisableBlending() {
	opts.BlendSourceFactor = OneBlendFactor
	opts.BlendDestinationFactor = ZeroBlendFactor
}

func (opts *RenderOptions) apply() {
	if currentOpts.DepthTest != opts.DepthTest {
		switch opts.DepthTest {
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
		currentOpts.DepthTest = opts.DepthTest
	}

	if currentOpts.BlendSourceFactor != opts.BlendSourceFactor || currentOpts.BlendDestinationFactor != opts.BlendDestinationFactor {
		var factors [2]BlendFactor
		var funcs [2]uint32
		factors[0] = opts.BlendSourceFactor
		factors[1] = opts.BlendDestinationFactor
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
		currentOpts.BlendSourceFactor = opts.BlendSourceFactor
		currentOpts.BlendDestinationFactor = opts.BlendDestinationFactor
	}

	if currentOpts.Cull != opts.Cull {
		switch opts.Cull {
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
		currentOpts.Cull = opts.Cull
	}

	if currentOpts.TriangleMode != opts.TriangleMode {
		switch opts.TriangleMode {
		case PointTriangleMode:
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.POINT)
		case LineTriangleMode:
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
		case TriangleTriangleMode:
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
		default:
			panic("tried to apply a render state with an unknown polygonmode")
		}
		currentOpts.TriangleMode = opts.TriangleMode
	}
}

func init() {
	gl.Enable(gl.BLEND)

	// initialize cached state to default OpenGL values TODO: run apply with it?
	currentOpts.DepthTest = AlwaysDepthTest
	currentOpts.DisableBlending()
	currentOpts.Cull = CullNothing
	currentOpts.TriangleMode = TriangleTriangleMode
}
