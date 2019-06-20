package graphics

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/window" // initialize graphics
)

type Framebuffer struct {
	id            int
	width, height int
}

type FramebufferAttachment int

const (
	ColorAttachment   FramebufferAttachment = FramebufferAttachment(gl.COLOR_ATTACHMENT0)
	DepthAttachment   FramebufferAttachment = FramebufferAttachment(gl.DEPTH_ATTACHMENT)
	StencilAttachment FramebufferAttachment = FramebufferAttachment(gl.STENCIL_ATTACHMENT)
)

var DefaultFramebuffer *Framebuffer = &Framebuffer{0, 800, 800}

func NewFramebuffer() *Framebuffer {
	var f Framebuffer
	var id uint32
	gl.CreateFramebuffers(1, &id)
	f.id = int(id)
	f.width = 0
	f.height = 0
	return &f
}

func (f *Framebuffer) Width() int {
	DefaultFramebuffer.width, _ = window.Size()
	return f.width
}

func (f *Framebuffer) Height() int {
	_, DefaultFramebuffer.height = window.Size()
	return f.height
}

func (f *Framebuffer) AttachTexture2D(attachment FramebufferAttachment, t *Texture2D, level int32) {
	if f.width == 0 && f.height == 0 {
		f.width = t.Width
		f.height = t.Height
	} else if f.width != t.Width || f.height != t.Height {
		panic("incompatible framebuffer attachment size")
	}
	gl.NamedFramebufferTexture(uint32(f.id), uint32(attachment), uint32(t.id), level)
}

func (f *Framebuffer) AttachCubeMapFace(attachment FramebufferAttachment, cf *CubeMapFace, level int32) {
	if f.width == 0 && f.height == 0 {
		f.width = cf.CubeMap.Width
		f.height = cf.CubeMap.Height
	} else if f.width != cf.CubeMap.Width || f.height != cf.CubeMap.Height {
		panic("incompatible framebuffer attachment size")
	}
	gl.NamedFramebufferTextureLayer(uint32(f.id), uint32(attachment), uint32(cf.CubeMap.id), level, int32(cf.layer))
}

func (f *Framebuffer) ClearColor(rgba math.Vec4) {
	gl.ClearNamedFramebufferfv(uint32(f.id), gl.COLOR, 0, &rgba[0])
}

func (f *Framebuffer) ClearDepth(clearDepth float32) {
	gl.ClearNamedFramebufferfv(uint32(f.id), gl.DEPTH, 0, &clearDepth)
}

func (f *Framebuffer) Complete() bool {
	status := gl.CheckNamedFramebufferStatus(uint32(f.id), gl.FRAMEBUFFER)
	return status == gl.FRAMEBUFFER_COMPLETE
}

func (f *Framebuffer) BindDraw() {
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, uint32(f.id))
}

func (f *Framebuffer) BindRead() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, uint32(f.id))
}
