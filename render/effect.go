package render

import (
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/camera"
)

type EffectRenderer struct {
	vbo *graphics.VertexBuffer
	framebuffer *graphics.Framebuffer
	renderState *graphics.State

	invProjectionMatrix math.Mat4

	fogSp *FogShaderProgram
}

type FogShaderProgram struct {
	*graphics.ShaderProgram

	position *graphics.Attrib
	depthMap *graphics.Uniform
	invProjectionMatrix *graphics.Uniform
	camFar *graphics.Uniform
}

func NewEffectRenderer() *EffectRenderer {
	var r EffectRenderer

	r.vbo = graphics.NewVertexBuffer()
	r.vbo.SetData([]math.Vec2{
		math.Vec2{-1.0, -1.0},
		math.Vec2{+1.0, -1.0},
		math.Vec2{+1.0, +1.0},
		math.Vec2{-1.0, -1.0},
		math.Vec2{+1.0, +1.0},
		math.Vec2{-1.0, +1.0},
	}, 0)

	r.fogSp = NewFogShaderProgram()
	r.fogSp.position.SetSourceVertex(r.vbo, 0)

	r.framebuffer = graphics.NewFramebuffer()

	r.renderState = graphics.NewState()
	r.renderState.PrimitiveType = graphics.Triangle
	r.renderState.Framebuffer = r.framebuffer

	return &r
}

func (r *EffectRenderer) RenderFog(c camera.Camera, depthMap, fogTarget *graphics.Texture2D) {
	r.framebuffer.Attach(fogTarget)

	r.fogSp.depthMap.Set(depthMap)

	r.invProjectionMatrix.Identity()
	r.invProjectionMatrix.Mult(c.ProjectionMatrix())
	r.invProjectionMatrix.Invert()
	r.fogSp.invProjectionMatrix.Set(&r.invProjectionMatrix)

	switch c.(type) {
	case *camera.PerspectiveCamera:
		c := c.(*camera.PerspectiveCamera)
		r.fogSp.camFar.Set(c.Far)
	}

	r.renderState.BlendDestinationFactor = graphics.OneMinusSourceAlphaBlendFactor
	r.renderState.BlendSourceFactor = graphics.SourceAlphaBlendFactor
	r.renderState.Program = r.fogSp.ShaderProgram

	r.renderState.Render(6)
}

func NewFogShaderProgram() *FogShaderProgram {
	var sp FogShaderProgram
	var err error

	vFile := "render/shaders/fogvshader.glsl" // TODO: make independent from executable directory
	fFile := "render/shaders/fogfshader.glsl" // TODO: make independent from executable directory
	sp.ShaderProgram, err = graphics.ReadShaderProgram(vFile, fFile, "")
	if err != nil {
		panic(err)
	}

	sp.position = sp.Attrib("position")
	sp.depthMap = sp.Uniform("depthTexture")
	sp.invProjectionMatrix = sp.Uniform("invProjectionMatrix")
	sp.camFar = sp.Uniform("cameraFar")

	return &sp
}
