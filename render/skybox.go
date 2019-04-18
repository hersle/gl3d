package render

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"unsafe"
)

type SkyboxShaderProgram struct {
	*graphics.ShaderProgram
	ViewMatrix       *graphics.UniformMatrix4
	ProjectionMatrix *graphics.UniformMatrix4
	CubeMap          *graphics.UniformSampler
	Position         *graphics.Attrib
}

type SkyboxRenderer struct {
	sp          *SkyboxShaderProgram
	vbo         *graphics.Buffer
	ibo         *graphics.Buffer
	tex         *graphics.CubeMap
	renderState *graphics.RenderState
}

func NewSkyboxShaderProgram() *SkyboxShaderProgram {
	var sp SkyboxShaderProgram
	var err error

	vShaderFilename := "render/shaders/skyboxvshader.glsl" // TODO: make independent from executable directory
	fShaderFilename := "render/shaders/skyboxfshader.glsl" // TODO: make independent from executable directory

	sp.ShaderProgram, err = graphics.ReadShaderProgram(vShaderFilename, fShaderFilename, "")
	if err != nil {
		panic(err)
	}

	sp.ViewMatrix = sp.UniformMatrix4("viewMatrix")
	sp.ProjectionMatrix = sp.UniformMatrix4("projectionMatrix")
	sp.CubeMap = sp.UniformSampler("cubeMap")
	sp.Position = sp.Attrib("positionV")

	return &sp
}

func NewSkyboxRenderer() *SkyboxRenderer {
	var r SkyboxRenderer

	r.sp = NewSkyboxShaderProgram()

	r.vbo = graphics.NewBuffer()
	verts := []math.Vec3{
		math.NewVec3(-1.0, -1.0, -1.0),
		math.NewVec3(+1.0, -1.0, -1.0),
		math.NewVec3(+1.0, +1.0, -1.0),
		math.NewVec3(-1.0, +1.0, -1.0),
		math.NewVec3(-1.0, -1.0, +1.0),
		math.NewVec3(+1.0, -1.0, +1.0),
		math.NewVec3(+1.0, +1.0, +1.0),
		math.NewVec3(-1.0, +1.0, +1.0),
	}
	r.vbo.SetData(verts, 0)

	r.ibo = graphics.NewBuffer()
	inds := []int32{
		4, 5, 6, 4, 6, 7,
		5, 1, 2, 5, 2, 6,
		1, 0, 3, 1, 3, 2,
		0, 4, 7, 0, 7, 3,
		7, 6, 2, 7, 2, 3,
		5, 4, 0, 5, 0, 1,
	}
	r.ibo.SetData(inds, 0)

	r.setCube(r.vbo, r.ibo)

	r.renderState = graphics.NewRenderState()
	r.renderState.Program = r.sp.ShaderProgram

	return &r
}

func (r *SkyboxRenderer) setFramebuffer(framebuffer *graphics.Framebuffer) {
	r.renderState.Framebuffer = framebuffer
	r.renderState.ViewportWidth = framebuffer.Width
	r.renderState.ViewportHeight = framebuffer.Height
}

func (r *SkyboxRenderer) setCamera(c camera.Camera) {
	r.sp.ViewMatrix.Set(c.ViewMatrix())
	r.sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *SkyboxRenderer) setSkybox(skybox *graphics.CubeMap) {
	r.sp.CubeMap.SetCube(skybox)
}

func (r *SkyboxRenderer) setCube(vbo, ibo *graphics.Buffer) {
	r.sp.Position.SetFormat(gl.FLOAT, false)
	r.sp.Position.SetSource(vbo, 0, int(unsafe.Sizeof(math.NewVec3(0, 0, 0))))
	r.sp.SetAttribIndexBuffer(ibo)
}

func (r *SkyboxRenderer) Render(sb *graphics.CubeMap, c camera.Camera, fb *graphics.Framebuffer) {
	r.setSkybox(sb)
	r.setCamera(c)
	r.setFramebuffer(fb)

	graphics.NewRenderCommand(graphics.Triangle, 36, 0, r.renderState).Execute()
}
