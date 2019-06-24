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

type FramebufferAttachment interface {
	attachToFramebuffer(f *Framebuffer)
	Width() int
	Height() int
}

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
	if f == DefaultFramebuffer {
		DefaultFramebuffer.width, _ = window.Size()
	}
	return f.width
}

func (f *Framebuffer) Height() int {
	if f == DefaultFramebuffer {
		_, DefaultFramebuffer.height = window.Size()
	}
	return f.height
}

func (f *Framebuffer) Attach(att FramebufferAttachment) {
	if f.width == 0 && f.height == 0 {
		f.width = att.Width()
		f.height = att.Height()
	} else if f.width != att.Width() || f.height != att.Height() {
		panic("incompatible framebuffer attachment size")
	}
	att.attachToFramebuffer(f)
}

func (f *Framebuffer) ClearColor(rgba math.Vec4) {
	gl.ClearNamedFramebufferfv(uint32(f.id), gl.COLOR, 0, &rgba[0])
}

func (f *Framebuffer) ClearDepth(depth float32) {
	gl.ClearNamedFramebufferfv(uint32(f.id), gl.DEPTH, 0, &depth)
}

func (f *Framebuffer) complete() bool {
	status := gl.CheckNamedFramebufferStatus(uint32(f.id), gl.FRAMEBUFFER)
	return status == gl.FRAMEBUFFER_COMPLETE
}

func (f *Framebuffer) bindDraw() {
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, uint32(f.id))
}

func (f *Framebuffer) bindRead() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, uint32(f.id))
}
