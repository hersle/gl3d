package graphics

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/window"
)

type Framebuffer struct {
	id            uint32
	width, height int
}

type FramebufferAttachment interface {
	attachTo(f *Framebuffer)
	Width() int
	Height() int
}

var DefaultFramebuffer *Framebuffer = &Framebuffer{0, 800, 800}

func NewFramebuffer() *Framebuffer {
	var fb Framebuffer
	gl.CreateFramebuffers(1, &fb.id)
	fb.width = 0
	fb.height = 0
	return &fb
}

func (fb *Framebuffer) Width() int {
	if fb == DefaultFramebuffer {
		DefaultFramebuffer.width, _ = window.Size()
	}
	return fb.width
}

func (fb *Framebuffer) Height() int {
	if fb == DefaultFramebuffer {
		_, DefaultFramebuffer.height = window.Size()
	}
	return fb.height
}

func (fb *Framebuffer) ClearColor(rgba math.Vec4) {
	gl.ClearNamedFramebufferfv(fb.id, gl.COLOR, 0, &rgba[0])
}

func (fb *Framebuffer) ClearDepth(depth float32) {
	gl.ClearNamedFramebufferfv(fb.id, gl.DEPTH, 0, &depth)
}

func (fb *Framebuffer) ClearStencil(index int) {
	value := int32(index)
	gl.ClearNamedFramebufferiv(fb.id, gl.STENCIL, 0, &value)
}

func (fb *Framebuffer) Attach(att FramebufferAttachment) {
	if fb.width == 0 && fb.height == 0 {
		fb.width = att.Width()
		fb.height = att.Height()
	} else if fb.width != att.Width() || fb.height != att.Height() {
		panic("incompatible framebuffer attachment size")
	}
	att.attachTo(fb)
}

func (fb *Framebuffer) complete() bool {
	status := gl.CheckNamedFramebufferStatus(fb.id, gl.FRAMEBUFFER)
	return status == gl.FRAMEBUFFER_COMPLETE
}

func (fb *Framebuffer) bindDraw() {
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, fb.id)
}

func (fb *Framebuffer) bindRead() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, fb.id)
}
