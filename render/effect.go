package render

import (
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/camera"
)

type EffectRenderer struct {
	vbo *graphics.VertexBuffer
	renderOpts *graphics.RenderOptions

	invProjectionMatrix math.Mat4

	fogSp *FogProgram

	gaussianSp *GaussianProgram
}

type FogProgram struct {
	*graphics.Program

	position *graphics.Input
	depthMap *graphics.Uniform
	invProjectionMatrix *graphics.Uniform
	camFar *graphics.Uniform

	color *graphics.Output
}

type GaussianProgram struct {
	*graphics.Program

	position *graphics.Input
	inTexture *graphics.Uniform
	direction *graphics.Uniform
	texDim *graphics.Uniform
	color *graphics.Output
	stddev *graphics.Uniform
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

	r.fogSp = NewFogProgram()
	r.fogSp.position.SetSourceVertex(r.vbo, 0)

	r.gaussianSp = NewGaussianProgram()
	r.gaussianSp.position.SetSourceVertex(r.vbo, 0)

	r.renderOpts = graphics.NewRenderOptions()
	r.renderOpts.PrimitiveType = graphics.Triangle
	//r.renderOpts.Framebuffer = r.framebuffer

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

	r.renderOpts.BlendMode = graphics.AlphaBlending

	r.fogSp.Render(6, r.renderOpts)
}

func (r *EffectRenderer) RenderGaussianBlur(target, extra *graphics.Texture2D, stddev float32) {
	r.renderOpts.BlendMode = graphics.ReplaceBlending

	r.gaussianSp.stddev.Set(stddev)

	r.gaussianSp.color.Set(extra)
	r.gaussianSp.inTexture.Set(target)
	r.gaussianSp.texDim.Set(float32(target.Width()))
	r.gaussianSp.direction.Set(math.Vec2{1, 0})
	r.gaussianSp.Render(6, r.renderOpts)

	r.gaussianSp.color.Set(target)
	r.gaussianSp.inTexture.Set(extra)
	r.gaussianSp.texDim.Set(float32(extra.Height()))
	r.gaussianSp.direction.Set(math.Vec2{0, 1})
	r.gaussianSp.Render(6, r.renderOpts)
}

func NewFogProgram() *FogProgram {
	var sp FogProgram

	vFile := "render/shaders/fogvshader.glsl" // TODO: make independent from executable directory
	fFile := "render/shaders/fogfshader.glsl" // TODO: make independent from executable directory
	sp.Program = graphics.ReadProgram(vFile, fFile, "")

	sp.position = sp.InputByName("position")
	sp.depthMap = sp.UniformByName("depthTexture")
	sp.invProjectionMatrix = sp.UniformByName("invProjectionMatrix")
	sp.camFar = sp.UniformByName("cameraFar")
	sp.color = sp.OutputColorByName("fragColor")

	return &sp
}

func NewGaussianProgram() *GaussianProgram {
	var sp GaussianProgram

	vFile := "render/shaders/gaussianvshader.glsl" // TODO: make independent from executable directory
	fFile := "render/shaders/gaussianfshader.glsl" // TODO: make independent from executable directory
	sp.Program = graphics.ReadProgram(vFile, fFile, "")

	sp.position = sp.InputByName("position")
	sp.inTexture = sp.UniformByName("inTexture")
	sp.direction = sp.UniformByName("dir")
	sp.texDim = sp.UniformByName("texDim")
	sp.color = sp.OutputColorByName("fragColor")
	sp.stddev = sp.UniformByName("stddev")

	return &sp
}
