package render

import (
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
)

type QuadShaderProgram struct {
	*graphics.ShaderProgram
	Position *graphics.Attrib
	Texture  *graphics.Uniform
}

type QuadRenderer struct {
	sp          *QuadShaderProgram
	vbo         *graphics.VertexBuffer
	tex         *graphics.Texture2D
	renderState *graphics.State
}

func NewQuadShaderProgram() *QuadShaderProgram {
	var sp QuadShaderProgram
	var err error

	vShaderFilename := "render/shaders/quadvshader.glsl" // TODO: make independent...
	fShaderFilename := "render/shaders/quadfshader.glsl" // TODO: make independent...
	sp.ShaderProgram, err = graphics.ReadShaderProgram(vShaderFilename, fShaderFilename)
	if err != nil {
		panic(err)
	}

	sp.Position = sp.Attrib("position")
	sp.Texture = sp.Uniform("tex")

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

	r.renderState = graphics.NewState()
	r.renderState.Program = r.sp.ShaderProgram
	r.renderState.BlendSourceFactor = graphics.SourceAlphaBlendFactor
	r.renderState.BlendDestinationFactor = graphics.OneMinusSourceAlphaBlendFactor
	r.renderState.PrimitiveType = graphics.Triangle

	return &r
}

func (r *QuadRenderer) Render(tex *graphics.Texture2D, fb *graphics.Framebuffer) {
	r.sp.Texture.Set(tex)
	r.renderState.Framebuffer = fb
	r.renderState.Render(6)
}
