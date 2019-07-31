package graphics

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	_ "github.com/hersle/gl3d/window" // initialize graphics
)

type DepthTest int

const (
	NoDepthTest DepthTest = iota
	LessDepthTest
	LessEqualDepthTest
	EqualDepthTest
	NotEqualDepthTest
	GreaterDepthTest
	GreaterEqualDepthTest
)

type Blending int

const (
	NoBlending Blending = iota
	AdditiveBlending
	AlphaBlending
)

type Culling int

const (
	NoCulling Culling = iota
	FrontCulling
	BackCulling
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
	DepthTest DepthTest
	Blending  Blending
	Culling   Culling
	Primitive Primitive
}

var currentOpts RenderOptions

func NewRenderOptions() *RenderOptions {
	var opts RenderOptions
	opts.Blending = NoBlending
	return &opts
}

func (opts *RenderOptions) apply() {
	if currentOpts.DepthTest != opts.DepthTest {
		switch opts.DepthTest {
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
		case NoDepthTest:
			gl.Disable(gl.DEPTH_TEST)
		default:
			panic("tried to apply a render state with unknown depth test")
		}
		currentOpts.DepthTest = opts.DepthTest
	}

	if currentOpts.Blending != opts.Blending {
		switch opts.Blending {
		case NoBlending:
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
		currentOpts.Blending = opts.Blending
	}

	if currentOpts.Culling != opts.Culling {
		switch opts.Culling {
		case NoCulling:
			gl.Disable(gl.CULL_FACE)
		case FrontCulling:
			gl.Enable(gl.CULL_FACE)
			gl.CullFace(gl.FRONT)
		case BackCulling:
			gl.Enable(gl.CULL_FACE)
			gl.CullFace(gl.BACK)
		default:
			panic("tried to apply a render state with an unknown culling mode")
		}
		currentOpts.Culling = opts.Culling
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
	currentOpts.DepthTest = NoDepthTest
	currentOpts.Blending = NoBlending
	currentOpts.Culling = NoCulling
}
