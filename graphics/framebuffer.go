package graphics

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/window"
)

type framebuffer struct {
	id            uint32
	width, height int
}

type renderTarget interface {
	attachTo(f *framebuffer)
	Width() int
	Height() int
}

var defaultFramebuffer *framebuffer = &framebuffer{0, 800, 800}

func newFramebuffer() *framebuffer {
	var fb framebuffer
	gl.CreateFramebuffers(1, &fb.id)
	fb.width = 0
	fb.height = 0
	return &fb
}

func (fb *framebuffer) Width() int {
	if fb == defaultFramebuffer {
		defaultFramebuffer.width, _ = window.Size()
	}
	return fb.width
}

func (fb *framebuffer) Height() int {
	if fb == defaultFramebuffer {
		_, defaultFramebuffer.height = window.Size()
	}
	return fb.height
}

func (fb *framebuffer) clearColor(rgba math.Vec4) {
	gl.ClearNamedFramebufferfv(fb.id, gl.COLOR, 0, &rgba[0])
}

func (fb *framebuffer) clearDepth(depth float32) {
	gl.ClearNamedFramebufferfv(fb.id, gl.DEPTH, 0, &depth)
}

func (fb *framebuffer) clearStencil(index int) {
	value := int32(index)
	gl.ClearNamedFramebufferiv(fb.id, gl.STENCIL, 0, &value)
}

func (fb *framebuffer) attach(target renderTarget) {
	if fb.width == 0 && fb.height == 0 {
		fb.width = target.Width()
		fb.height = target.Height()
	} else if fb.width != target.Width() || fb.height != target.Height() {
		panic("incompatible framebuffer attachment size")
	}
	target.attachTo(fb)
}

func (fb *framebuffer) complete() bool {
	status := gl.CheckNamedFramebufferStatus(fb.id, gl.FRAMEBUFFER)
	return status == gl.FRAMEBUFFER_COMPLETE
}

func (fb *framebuffer) bindDraw() {
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, fb.id)
}

func (fb *framebuffer) bindRead() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, fb.id)
}
