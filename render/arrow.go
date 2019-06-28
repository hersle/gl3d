package render

import (
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/scene"
)

type ArrowRenderer struct {
	sp          *ArrowShaderProgram
	points      []math.Vec3
	vbo         *graphics.Buffer
	renderState *graphics.RenderState
}

type ArrowShaderProgram struct {
	*graphics.ShaderProgram
	ModelMatrix      *graphics.Uniform
	ViewMatrix       *graphics.Uniform
	ProjectionMatrix *graphics.Uniform
	Color            *graphics.Uniform
	Position         *graphics.Attrib
}

func NewArrowShaderProgram() *ArrowShaderProgram {
	var sp ArrowShaderProgram
	var err error

	vShaderFilename := "render/shaders/arrowvshader.glsl" // TODO: make independent from executable directory
	fShaderFilename := "render/shaders/arrowfshader.glsl" // TODO: make independent from executable directory
	sp.ShaderProgram, err = graphics.ReadShaderProgram(vShaderFilename, fShaderFilename, "")
	if err != nil {
		panic(err)
	}

	sp.Position = sp.Attrib("position")
	sp.ModelMatrix = sp.Uniform("modelMatrix")
	sp.ViewMatrix = sp.Uniform("viewMatrix")
	sp.ProjectionMatrix = sp.Uniform("projectionMatrix")
	sp.Color = sp.Uniform("color")

	return &sp
}

func NewArrowRenderer() *ArrowRenderer {
	var r ArrowRenderer

	r.sp = NewArrowShaderProgram()

	r.renderState = graphics.NewRenderState()
	r.renderState.Program = r.sp.ShaderProgram
	r.renderState.PrimitiveType = graphics.Line

	r.vbo = graphics.NewBuffer()
	r.SetPosition(r.vbo)

	return &r
}

func (r *ArrowRenderer) SetCamera(c camera.Camera) {
	r.sp.ViewMatrix.Set(c.ViewMatrix())
	r.sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *ArrowRenderer) SetMesh(m *object.Mesh) {
	r.sp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *ArrowRenderer) SetColor(color math.Vec3) {
	r.sp.Color.Set(color)
}

func (r *ArrowRenderer) SetPosition(vbo *graphics.Buffer) {
	r.sp.Position.SetSource(vbo, math.Vec3{}, 0)
}

func (r *ArrowRenderer) RenderTangents(s *scene.Scene, c camera.Camera, fb *graphics.Framebuffer) {
	r.SetCamera(c)
	r.points = r.points[:0]
	r.SetColor(math.Vec3{1, 0, 0})
	for _, m := range s.Meshes {
		r.SetMesh(m)
		for _, subMesh := range m.SubMeshes {
			for _, i := range subMesh.Geo.Faces {
				p1 := subMesh.Geo.Verts[i].Position
				p2 := p1.Add(subMesh.Geo.Verts[i].Tangent)
				r.points = append(r.points, p1, p2)
			}
		}
	}
	r.vbo.SetData(r.points, 0)
	r.renderState.Framebuffer = fb
	graphics.NewRenderCommand(len(r.points), r.renderState).Execute()
}

func (r *ArrowRenderer) RenderBitangents(s *scene.Scene, c camera.Camera, fb *graphics.Framebuffer) {
	r.SetCamera(c)
	r.points = r.points[:0]
	r.SetColor(math.Vec3{0, 1, 0})
	for _, m := range s.Meshes {
		r.SetMesh(m)
		for _, subMesh := range m.SubMeshes {
			for _, i := range subMesh.Geo.Faces {
				p1 := subMesh.Geo.Verts[i].Position
				p2 := p1.Add(subMesh.Geo.Verts[i].Bitangent())
				r.points = append(r.points, p1, p2)
			}
		}
	}
	r.vbo.SetData(r.points, 0)
	r.renderState.Framebuffer = fb
	graphics.NewRenderCommand(len(r.points), r.renderState).Execute()
}

func (r *ArrowRenderer) RenderNormals(s *scene.Scene, c camera.Camera, fb *graphics.Framebuffer) {
	r.SetCamera(c)
	r.points = r.points[:0]
	r.SetColor(math.Vec3{0, 0, 1})
	for _, m := range s.Meshes {
		r.SetMesh(m)
		for _, subMesh := range m.SubMeshes {
			for _, i := range subMesh.Geo.Faces {
				p1 := subMesh.Geo.Verts[i].Position
				p2 := p1.Add(subMesh.Geo.Verts[i].Normal)
				r.points = append(r.points, p1, p2)
			}
		}
	}
	r.vbo.SetData(r.points, 0)
	r.renderState.Framebuffer = fb
	graphics.NewRenderCommand(len(r.points), r.renderState).Execute()
}
