package render

import (
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/scene"
)

type SkyboxShaderProgram struct {
	*graphics.ShaderProgram
	ViewMatrix       *graphics.Uniform
	ProjectionMatrix *graphics.Uniform
	CubeMap          *graphics.Uniform
	Position         *graphics.Attrib
}

type SkyboxRenderer struct {
	sp          *SkyboxShaderProgram
	vbo         *graphics.VertexBuffer
	ibo         *graphics.IndexBuffer
	tex         *graphics.CubeMap
	renderState *graphics.RenderState
	cubemaps    map[*scene.CubeMap]*graphics.CubeMap
}

func NewSkyboxShaderProgram() *SkyboxShaderProgram {
	var sp SkyboxShaderProgram
	var err error

	vShaderFilename := "render/shaders/skyboxvshader.glsl" // TODO: make independent from executable directory
	fShaderFilename := "render/shaders/skyboxfshader.glsl" // TODO: make independent from executable directory

	sp.ShaderProgram, err = graphics.ReadShaderProgram(vShaderFilename, fShaderFilename)
	if err != nil {
		panic(err)
	}

	sp.ViewMatrix = sp.Uniform("viewMatrix")
	sp.ProjectionMatrix = sp.Uniform("projectionMatrix")
	sp.CubeMap = sp.Uniform("cubeMap")
	sp.Position = sp.Attrib("positionV")

	return &sp
}

func NewSkyboxRenderer() *SkyboxRenderer {
	var r SkyboxRenderer

	r.cubemaps = make(map[*scene.CubeMap]*graphics.CubeMap)

	r.sp = NewSkyboxShaderProgram()

	r.vbo = graphics.NewVertexBuffer()
	verts := []math.Vec3{
		math.Vec3{-1.0, -1.0, -1.0},
		math.Vec3{+1.0, -1.0, -1.0},
		math.Vec3{+1.0, +1.0, -1.0},
		math.Vec3{-1.0, +1.0, -1.0},
		math.Vec3{-1.0, -1.0, +1.0},
		math.Vec3{+1.0, -1.0, +1.0},
		math.Vec3{+1.0, +1.0, +1.0},
		math.Vec3{-1.0, +1.0, +1.0},
	}
	r.vbo.SetData(verts, 0)

	r.ibo = graphics.NewIndexBuffer()
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
	r.renderState.PrimitiveType = graphics.Triangle

	return &r
}

func (r *SkyboxRenderer) setFramebuffer(framebuffer *graphics.Framebuffer) {
	r.renderState.Framebuffer = framebuffer
}

func (r *SkyboxRenderer) setCamera(c camera.Camera) {
	r.sp.ViewMatrix.Set(c.ViewMatrix())
	r.sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *SkyboxRenderer) setSkybox(skybox *scene.CubeMap) {
	cm, found := r.cubemaps[skybox]
	if !found {
		img1 := skybox.Posx
		img2 := skybox.Negx
		img3 := skybox.Posy
		img4 := skybox.Negy
		img5 := skybox.Posz
		img6 := skybox.Negz
		cm = graphics.LoadCubeMap(graphics.NearestFilter, img1, img2, img3, img4, img5, img6)
		r.cubemaps[skybox] = cm
	}

	r.sp.CubeMap.Set(cm)
}

func (r *SkyboxRenderer) setCube(vbo *graphics.VertexBuffer, ibo *graphics.IndexBuffer) {
	r.sp.Position.SetSourceVertex(vbo, 0)
	r.sp.SetAttribIndexBuffer(ibo)
}

func (r *SkyboxRenderer) Render(sb *scene.CubeMap, c camera.Camera, fb *graphics.Framebuffer) {
	r.setSkybox(sb)
	r.setCamera(c)
	r.setFramebuffer(fb)

	graphics.NewRenderCommand(36, r.renderState).Execute()
}
