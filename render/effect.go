package render

import (
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/camera"
)

type EffectRenderer struct {
	vbo *graphics.VertexBuffer
	renderState *graphics.State

	invProjectionMatrix math.Mat4

	fogSp *FogShaderProgram

	gaussianSp *GaussianShaderProgram
}

type FogShaderProgram struct {
	*graphics.ShaderProgram

	position *graphics.Input
	depthMap *graphics.Uniform
	invProjectionMatrix *graphics.Uniform
	camFar *graphics.Uniform

	color *graphics.Output
}

type GaussianShaderProgram struct {
	*graphics.ShaderProgram

	position *graphics.Input
	inTexture *graphics.Uniform
	direction *graphics.Uniform
	texDim *graphics.Uniform

	color *graphics.Output
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

	r.gaussianSp = NewGaussianShaderProgram()
	r.gaussianSp.position.SetSourceVertex(r.vbo, 0)

	r.renderState = graphics.NewState()
	r.renderState.PrimitiveType = graphics.Triangle
	//r.renderState.Framebuffer = r.framebuffer

	return &r
}

func (r *EffectRenderer) RenderFog(c camera.Camera, depthMap, fogTarget *graphics.Texture2D) {
	r.fogSp.color.Set(fogTarget)

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

func (r *EffectRenderer) RenderGaussianBlur(target, extra *graphics.Texture2D) {
	r.renderState.DisableBlending()
	r.renderState.Program = r.gaussianSp.ShaderProgram

	r.gaussianSp.color.Set(extra)
	r.gaussianSp.inTexture.Set(target)
	r.gaussianSp.texDim.Set(float32(target.Width()))
	r.gaussianSp.direction.Set(math.Vec2{1, 0})
	r.renderState.Render(6)

	r.gaussianSp.color.Set(target)
	r.gaussianSp.inTexture.Set(extra)
	r.gaussianSp.texDim.Set(float32(extra.Height()))
	r.gaussianSp.direction.Set(math.Vec2{0, 1})
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

	sp.position = sp.Input("position")
	sp.depthMap = sp.Uniform("depthTexture")
	sp.invProjectionMatrix = sp.Uniform("invProjectionMatrix")
	sp.camFar = sp.Uniform("cameraFar")
	sp.color = sp.OutputColor("fragColor")

	return &sp
}

func NewGaussianShaderProgram() *GaussianShaderProgram {
	var sp GaussianShaderProgram
	var err error

	vFile := "render/shaders/gaussianvshader.glsl" // TODO: make independent from executable directory
	fFile := "render/shaders/gaussianfshader.glsl" // TODO: make independent from executable directory
	sp.ShaderProgram, err = graphics.ReadShaderProgram(vFile, fFile, "")
	if err != nil {
		panic(err)
	}

	sp.position = sp.Input("position")
	sp.inTexture = sp.Uniform("inTexture")
	sp.direction = sp.Uniform("dir")
	sp.texDim = sp.Uniform("texDim")
	sp.color = sp.OutputColor("fragColor")

	return &sp
}
