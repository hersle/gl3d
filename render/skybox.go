package render

import (
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/scene"
)

type SkyboxProgram struct {
	*graphics.Program
	ViewMatrix       *graphics.Uniform
	ProjectionMatrix *graphics.Uniform
	CubeMap          *graphics.Uniform
	Position         *graphics.Input
	Color            *graphics.Output
}

type SkyboxRenderer struct {
	sp          *SkyboxProgram
	vbo         *graphics.VertexBuffer
	ibo         *graphics.IndexBuffer
	tex         *graphics.CubeMap
	renderOpts  *graphics.RenderOptions
	cubemaps    map[*scene.CubeMap]*graphics.CubeMap
}

func NewSkyboxProgram() *SkyboxProgram {
	var sp SkyboxProgram

	vShaderFilename := "render/shaders/skyboxvshader.glsl" // TODO: make independent from executable directory
	fShaderFilename := "render/shaders/skyboxfshader.glsl" // TODO: make independent from executable directory

	sp.Program = graphics.ReadProgram(vShaderFilename, fShaderFilename, "")

	sp.ViewMatrix = sp.UniformByName("viewMatrix")
	sp.ProjectionMatrix = sp.UniformByName("projectionMatrix")
	sp.CubeMap = sp.UniformByName("cubeMap")
	sp.Position = sp.InputByName("positionV")
	sp.Color = sp.OutputColorByName("fragColor")

	return &sp
}

func NewSkyboxRenderer() *SkyboxRenderer {
	var r SkyboxRenderer

	r.cubemaps = make(map[*scene.CubeMap]*graphics.CubeMap)

	r.sp = NewSkyboxProgram()

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

	r.renderOpts = graphics.NewRenderOptions()
	r.renderOpts.Primitive = graphics.Triangles

	return &r
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
	r.sp.SetIndices(ibo)
}

func (r *SkyboxRenderer) Render(sb *scene.CubeMap, c camera.Camera, target *graphics.Texture2D) {
	r.setSkybox(sb)
	r.setCamera(c)
	r.sp.Color.Set(target)

	r.sp.Render(36, r.renderOpts)
}
