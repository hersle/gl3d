package graphics

import (
	_ "github.com/hersle/gl3d/window" // initialize graphics
	"github.com/go-gl/gl/v4.5-core/gl"
)

// TODO: enable sorting of these states to reduce state changes?
type RenderState struct {
	prog           *ShaderProgram
	framebuffer    *Framebuffer
	depthTest      bool
	depthFunc      uint32
	blend          bool
	blendSrcFactor uint32
	blendDstFactor uint32
	viewportWidth  int
	viewportHeight int
	cull           bool
	cullFace       uint32
	polygonMode    uint32
}

func NewRenderState() *RenderState {
	var rs RenderState
	return &rs
}

func (rs *RenderState) SetShaderProgram(prog *ShaderProgram) {
	rs.prog = prog
}

func (rs *RenderState) SetFramebuffer(fb *Framebuffer) {
	rs.framebuffer = fb
}

func (rs *RenderState) SetDepthTest(depthTest bool) {
	rs.depthTest = depthTest
}

func (rs *RenderState) SetDepthFunc(depthFunc uint32) {
	rs.depthFunc = depthFunc
}

func (rs *RenderState) SetBlend(blend bool) {
	rs.blend = blend
}

func (rs *RenderState) SetBlendFunction(blendSrcFactor, blendDstFactor uint32) {
	rs.blendSrcFactor = blendSrcFactor
	rs.blendDstFactor = blendDstFactor
}

func (rs *RenderState) SetViewport(width, height int) {
	rs.viewportWidth = width
	rs.viewportHeight = height
}

func (rs *RenderState) SetCull(cull bool) {
	rs.cull = cull
}

func (rs *RenderState) SetCullFace(cullFace uint32) {
	rs.cullFace = cullFace
}

func (rs *RenderState) SetPolygonMode(mode uint32) {
	rs.polygonMode = mode
}

func (rs *RenderState) Apply() {
	rs.prog.va.Bind()
	rs.prog.Bind()

	rs.framebuffer.BindDraw()

	if rs.depthTest {
		gl.Enable(gl.DEPTH_TEST)
		gl.DepthFunc(rs.depthFunc)
	} else {
		gl.Disable(gl.DEPTH_TEST)
	}

	if rs.blend {
		gl.Enable(gl.BLEND)
		gl.BlendFunc(rs.blendSrcFactor, rs.blendDstFactor)
	} else {
		gl.Disable(gl.BLEND)
	}

	if rs.cull {
		gl.Enable(gl.CULL_FACE)
		gl.CullFace(rs.cullFace)
	} else {
		gl.Disable(gl.CULL_FACE)
	}

	gl.PolygonMode(gl.FRONT_AND_BACK, rs.polygonMode)

	gl.Viewport(0, 0, int32(rs.viewportWidth), int32(rs.viewportHeight))
}
