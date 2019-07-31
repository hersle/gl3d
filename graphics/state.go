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

type BlendMode int

const (
	ReplaceBlending BlendMode = iota
	AdditiveBlending
	AlphaBlending
)

type CullMode int

const (
	CullNothing CullMode = iota
	CullFront
	CullBack
)

type Primitive int

const (
	Points Primitive = iota
	Lines
	LineStrip
	LineLoop
	Triangles
	TriangleStrip
	TriangleFan
	TriangleOutlines
	TriangleOutlineStrip
	TriangleOutlineFan
)

// TODO: enable sorting of these states to reduce state changes?
type RenderOptions struct {
	DepthTest              DepthTest
	BlendMode              BlendMode
	Cull                   CullMode
	Primitive              Primitive
}

var currentOpts RenderOptions

func NewRenderOptions() *RenderOptions {
	var opts RenderOptions
	opts.BlendMode = ReplaceBlending
	return &opts
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

	if currentOpts.BlendMode != opts.BlendMode {
		switch opts.BlendMode {
		case ReplaceBlending:
			gl.Disable(gl.BLEND)
		case AdditiveBlending:
			gl.Enable(gl.BLEND)
			gl.BlendFunc(gl.ONE, gl.ONE)
		case AlphaBlending:
			gl.Enable(gl.BLEND)
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		default:
			panic("invalid blend mode")
		}
		currentOpts.BlendMode = opts.BlendMode
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

	if currentOpts.Primitive != opts.Primitive {
		switch opts.Primitive {
		case Triangles, TriangleStrip, TriangleFan:
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
		case TriangleOutlines, TriangleOutlineStrip, TriangleOutlineFan:
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
		}
		currentOpts.Primitive = opts.Primitive
	}
}

func (p Primitive) glPrimitive() uint32 {
	switch p {
	case Points:
		return gl.POINTS
	case Lines:
		return gl.LINES
	case LineStrip:
		return gl.LINE_STRIP
	case LineLoop:
		return gl.LINE_LOOP
	case Triangles, TriangleOutlines:
		return gl.TRIANGLES
	case TriangleStrip, TriangleOutlineStrip:
		return gl.TRIANGLE_STRIP
	case TriangleFan, TriangleOutlineFan:
		return gl.TRIANGLE_FAN
	default:
		panic("invalid primitive")
	}
}

func init() {
	gl.Enable(gl.BLEND)

	// initialize cached state to default OpenGL values TODO: run apply with it?
	currentOpts.DepthTest = AlwaysDepthTest
	currentOpts.BlendMode = ReplaceBlending
	currentOpts.Cull = CullNothing
}
