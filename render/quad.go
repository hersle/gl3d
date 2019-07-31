package render

import (
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
)

type QuadProgram struct {
	*graphics.Program
	Position *graphics.Input
	Texture  *graphics.Uniform
}

type QuadRenderer struct {
	sp          *QuadProgram
	vbo         *graphics.VertexBuffer
	tex         *graphics.Texture2D
	renderOpts  *graphics.RenderOptions
}

func NewQuadProgram() *QuadProgram {
	var sp QuadProgram

	vShaderFilename := "render/shaders/quadvshader.glsl" // TODO: make independent...
	fShaderFilename := "render/shaders/quadfshader.glsl" // TODO: make independent...
	sp.Program = graphics.ReadProgram(vShaderFilename, fShaderFilename, "")

	sp.Position = sp.InputByName("position")
	sp.Texture = sp.UniformByName("tex")

	sp.Framebuffer = graphics.DefaultFramebuffer // output to default framebuffer instead

	return &sp
}

func NewQuadRenderer() *QuadRenderer {
	var r QuadRenderer

	r.sp = NewQuadProgram()

	verts := []math.Vec2{
		math.Vec2{-1.0, -1.0},
		math.Vec2{+1.0, -1.0},
		math.Vec2{+1.0, +1.0},
		math.Vec2{-1.0, -1.0},
		math.Vec2{+1.0, +1.0},
		math.Vec2{-1.0, +1.0},
	}
	r.vbo = graphics.NewVertexBuffer()
	r.vbo.SetData(verts, 0)

	r.sp.Position.SetSourceVertex(r.vbo, 0)

	r.renderOpts = graphics.NewRenderOptions()
	r.renderOpts.Blending = graphics.AlphaBlending
	r.renderOpts.Primitive = graphics.Triangles

	return &r
}

func (r *QuadRenderer) Render(tex *graphics.Texture2D, fb *graphics.Framebuffer) {
	r.sp.Texture.Set(tex)
	r.sp.Render(6, r.renderOpts)
}
