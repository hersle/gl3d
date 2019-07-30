package render

import (
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
)

type QuadShaderProgram struct {
	*graphics.ShaderProgram
	Position *graphics.Input
	Texture  *graphics.Uniform
}

type QuadRenderer struct {
	sp          *QuadShaderProgram
	vbo         *graphics.VertexBuffer
	tex         *graphics.Texture2D
	renderOpts  *graphics.RenderOptions
}

func NewQuadShaderProgram() *QuadShaderProgram {
	var sp QuadShaderProgram
	var err error

	vShaderFilename := "render/shaders/quadvshader.glsl" // TODO: make independent...
	fShaderFilename := "render/shaders/quadfshader.glsl" // TODO: make independent...
	sp.ShaderProgram, err = graphics.ReadShaderProgram(vShaderFilename, fShaderFilename, "")
	if err != nil {
		panic(err)
	}

	sp.Position = sp.InputByName("position")
	sp.Texture = sp.UniformByName("tex")

	sp.Framebuffer = graphics.DefaultFramebuffer // output to default framebuffer instead

	return &sp
}

func NewQuadRenderer() *QuadRenderer {
	var r QuadRenderer

	r.sp = NewQuadShaderProgram()

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
	r.renderOpts.BlendMode = graphics.AlphaBlending
	r.renderOpts.PrimitiveType = graphics.Triangle

	return &r
}

func (r *QuadRenderer) Render(tex *graphics.Texture2D, fb *graphics.Framebuffer) {
	r.sp.Texture.Set(tex)
	r.sp.Render(6, r.renderOpts)
}
