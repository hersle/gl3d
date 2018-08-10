package render

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/window"
	"unsafe"
)

type QuadRenderer struct {
	sp          *graphics.QuadShaderProgram
	vbo         *graphics.Buffer
	tex         *graphics.Texture2D
	renderState *graphics.RenderState
}

func NewQuadRenderer() *QuadRenderer {
	var r QuadRenderer

	r.sp = graphics.NewQuadShaderProgram()

	verts := []math.Vec2{
		math.NewVec2(-1.0, -1.0),
		math.NewVec2(+1.0, -1.0),
		math.NewVec2(+1.0, +1.0),
		math.NewVec2(-1.0, -1.0),
		math.NewVec2(+1.0, +1.0),
		math.NewVec2(-1.0, +1.0),
	}
	r.vbo = graphics.NewBuffer()
	r.vbo.SetData(verts, 0)

	r.sp.Position.SetFormat(gl.FLOAT, false)
	stride := int(unsafe.Sizeof(verts[0]))
	r.sp.Position.SetSource(r.vbo, 0, stride)

	r.renderState = graphics.NewRenderState()
	r.renderState.SetDepthTest(graphics.AlwaysDepthTest)
	r.renderState.DisableBlending()
	r.renderState.SetFramebuffer(graphics.DefaultFramebuffer)
	r.renderState.SetShaderProgram(r.sp.ShaderProgram)
	r.renderState.SetCull(false)
	r.renderState.SetPolygonMode(gl.FILL)

	return &r
}

func (r *QuadRenderer) Render(tex *graphics.Texture2D) {
	r.sp.Texture.Set2D(tex)
	r.renderState.SetViewport(window.Size())
	graphics.NewRenderCommand(graphics.Triangle, 6, 0, r.renderState).Execute()
}
