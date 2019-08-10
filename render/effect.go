package render

import (
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/utils"
	"fmt"
)

type EffectRenderer struct {
	vbo *graphics.VertexBuffer
	renderOpts *graphics.RenderOptions

	invProjectionMatrix math.Mat4

	fogSp *FogProgram

	gaussianSp *GaussianProgram

	randomDirectionMap *graphics.Texture2D
	ssaoProg *ssaoProgram
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

type ssaoProgram struct {
	*graphics.Program

	Position *graphics.Input
	Color               *graphics.Output
	DepthMap            *graphics.Uniform
	DepthMapWidth       *graphics.Uniform
	DepthMapHeight      *graphics.Uniform
	ProjectionMatrix *graphics.Uniform
	InvProjectionMatrix *graphics.Uniform
	Directions          []*graphics.Uniform
	DirectionMap        *graphics.Uniform
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

	r.ssaoProg = NewSSAOProgram()

	r.renderOpts = graphics.NewRenderOptions()

	verts := []math.Vec2{
		math.Vec2{-1, -1}, math.Vec2{+1, -1}, math.Vec2{+1, +1}, math.Vec2{-1, +1},
	}
	vbo := graphics.NewVertexBuffer()
	vbo.SetData(verts, 0)
	r.ssaoProg.Position.SetSourceVertex(vbo, 0)
	w := 1920 / 1
	h := 1080 / 1
	r.randomDirectionMap = graphics.NewColorTexture2D(graphics.NearestFilter, graphics.RepeatWrap, w, h, 3, 32, true, false)
	directions := make([]math.Vec3, w*h)
	for i := 0; i < w*h; i++ {
		directions[i] = utils.RandomDirection()
	}
	r.randomDirectionMap.SetData(0, 0, w, h, directions)

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

	r.renderOpts.Blending = graphics.AlphaBlending
	r.renderOpts.Primitive = graphics.Triangles

	r.fogSp.Render(6, r.renderOpts)
}

func (r *EffectRenderer) RenderGaussianBlur(target, extra *graphics.Texture2D, stddev float32) {
	r.renderOpts.Blending = graphics.NoBlending
	r.renderOpts.Primitive = graphics.Triangles

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

func (r *EffectRenderer) RenderSSAO(c camera.Camera, depthTexture, colorTexture *graphics.Texture2D) {
	r.renderOpts.Blending = graphics.NoBlending
	r.renderOpts.Primitive = graphics.TriangleFan

	r.ssaoProg.Color.Set(colorTexture)
	r.ssaoProg.DepthMap.Set(depthTexture)
	r.ssaoProg.DepthMapWidth.Set(depthTexture.Width())
	r.ssaoProg.DepthMapHeight.Set(depthTexture.Height())

	var mat math.Mat4
	mat.Identity()
	mat.Mult(c.ProjectionMatrix())
	mat.Invert()
	r.ssaoProg.InvProjectionMatrix.Set(&mat)
	r.ssaoProg.ProjectionMatrix.Set(c.ProjectionMatrix())
	r.ssaoProg.DirectionMap.Set(r.randomDirectionMap)

	r.ssaoProg.Render(6, r.renderOpts)
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

func NewSSAOProgram() *ssaoProgram {
	var sp ssaoProgram

	vFile := "render/shaders/ssaovshader.glsl" // TODO: make independent from executable directory
	fFile := "render/shaders/ssaofshader.glsl" // TODO: make independent from executable directory
	sp.Program = graphics.ReadProgram(vFile, fFile, "")

	sp.Position = sp.InputByName("position")
	sp.Color = sp.OutputColorByName("fragColor")
	sp.DepthMap = sp.UniformByName("depthMap")
	sp.DepthMapWidth = sp.UniformByName("depthMapWidth")
	sp.DepthMapHeight = sp.UniformByName("depthMapHeight")
	sp.ProjectionMatrix = sp.UniformByName("projectionMatrix")
	sp.InvProjectionMatrix = sp.UniformByName("invProjectionMatrix")

	sp.Directions = make([]*graphics.Uniform, 16)
	for i := 0; i < 16; i++ {
		name := fmt.Sprintf("directions[%d]", i)
		sp.Directions[i] = sp.UniformByName(name)
	}

	sp.DirectionMap = sp.UniformByName("directionMap")

	for i := 0; i < 16; i++ {
		sp.Directions[i].Set(utils.RandomDirection())
	}

	return &sp
}
