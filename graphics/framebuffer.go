package graphics

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/math"
	_ "github.com/hersle/gl3d/window" // initialize graphics
)

type Framebuffer struct {
	id uint32
}

type FramebufferAttachment int

const (
	ColorAttachment FramebufferAttachment = FramebufferAttachment(gl.COLOR_ATTACHMENT0)
	DepthAttachment FramebufferAttachment = FramebufferAttachment(gl.DEPTH_ATTACHMENT)
	StencilAttachment FramebufferAttachment = FramebufferAttachment(gl.STENCIL_ATTACHMENT)
)

var DefaultFramebuffer *Framebuffer = &Framebuffer{0}

func NewFramebuffer() *Framebuffer {
	var f Framebuffer
	gl.CreateFramebuffers(1, &f.id)
	return &f
}

func (f *Framebuffer) SetTexture2D(attachment FramebufferAttachment, t *Texture2D, level int32) {
	gl.NamedFramebufferTexture(f.id, uint32(attachment), t.id, level)
}

func (f *Framebuffer) SetTextureCubeMapFace(attachment FramebufferAttachment, cf *CubeMapFace, level int32) {
	gl.NamedFramebufferTextureLayer(f.id, uint32(attachment), cf.CubeMap.id, level, int32(cf.layer))
}

func (f *Framebuffer) ClearColor(rgba math.Vec4) {
	gl.ClearNamedFramebufferfv(f.id, gl.COLOR, 0, &rgba[0])
}

func (f *Framebuffer) ClearDepth(clearDepth float32) {
	gl.ClearNamedFramebufferfv(f.id, gl.DEPTH, 0, &clearDepth)
}

func (f *Framebuffer) Complete() bool {
	status := gl.CheckNamedFramebufferStatus(f.id, gl.FRAMEBUFFER)
	return status == gl.FRAMEBUFFER_COMPLETE
}

func (f *Framebuffer) BindDraw() {
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, f.id)
}

func (f *Framebuffer) BindRead() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, f.id)
}
