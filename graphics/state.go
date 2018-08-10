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

// TODO: enable sorting of these states to reduce state changes?
type RenderState struct {
	Program        *ShaderProgram
	Framebuffer    *Framebuffer
	DepthTest      DepthTest
	BlendSourceFactor BlendFactor
	BlendDestinationFactor BlendFactor
	ViewportWidth  int
	ViewportHeight int
	Cull           CullMode
	TriangleMode   TriangleMode
}

func NewRenderState() *RenderState {
	var rs RenderState
	rs.DisableBlending()
	rs.SetFramebuffer(DefaultFramebuffer)
	return &rs
}

func (rs *RenderState) SetShaderProgram(prog *ShaderProgram) {
	rs.Program = prog
}

func (rs *RenderState) SetFramebuffer(fb *Framebuffer) {
	rs.Framebuffer = fb
}

func (rs *RenderState) SetDepthTest(depthTest DepthTest) {
	rs.DepthTest = depthTest
}

func (rs *RenderState) SetBlendFactors(blendSrcFactor, blendDstFactor BlendFactor) {
	rs.BlendSourceFactor = blendSrcFactor
	rs.BlendDestinationFactor = blendDstFactor
}

func (rs *RenderState) DisableBlending() {
	rs.SetBlendFactors(OneBlendFactor, ZeroBlendFactor)
}

func (rs *RenderState) SetViewport(width, height int) {
	rs.ViewportWidth = width
	rs.ViewportHeight = height
}

func (rs *RenderState) SetCull(cull CullMode) {
	rs.Cull = cull
}

func (rs *RenderState) SetTriangleMode(mode TriangleMode) {
	rs.TriangleMode = mode
}

func (rs *RenderState) Apply() {
	rs.Program.va.Bind()
	rs.Program.Bind()

	rs.Framebuffer.BindDraw()

	switch (rs.DepthTest) {
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

	gl.Viewport(0, 0, int32(rs.ViewportWidth), int32(rs.ViewportHeight))
}

func init() {
	gl.Enable(gl.BLEND)
}
