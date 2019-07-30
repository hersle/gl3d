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
	vbo         *graphics.VertexBuffer
	renderOpts  *graphics.RenderOptions
}

type ArrowShaderProgram struct {
	*graphics.ShaderProgram
	ModelMatrix      *graphics.Uniform
	ViewMatrix       *graphics.Uniform
	ProjectionMatrix *graphics.Uniform
	Color            *graphics.Uniform
	Position         *graphics.Input
	OutColor         *graphics.Output
	Depth            *graphics.Output
}

func NewArrowShaderProgram() *ArrowShaderProgram {
	var sp ArrowShaderProgram
	var err error

	vFile := "render/shaders/arrowvshader.glsl" // TODO: make independent from executable directory
	fFile := "render/shaders/arrowfshader.glsl" // TODO: make independent from executable directory
	sp.ShaderProgram, err = graphics.ReadShaderProgram(vFile, fFile, "")
	if err != nil {
		panic(err)
	}

	sp.Position = sp.InputByName("position")
	sp.ModelMatrix = sp.Uniform("modelMatrix")
	sp.ViewMatrix = sp.Uniform("viewMatrix")
	sp.ProjectionMatrix = sp.Uniform("projectionMatrix")
	sp.Color = sp.Uniform("color")
	sp.OutColor = sp.OutputColor("fragColor")
	sp.Depth = sp.OutputDepth()

	return &sp
}

func NewArrowRenderer() *ArrowRenderer {
	var r ArrowRenderer

	r.sp = NewArrowShaderProgram()

	r.renderOpts = graphics.NewRenderOptions()
	r.renderOpts.PrimitiveType = graphics.Line

	r.vbo = graphics.NewVertexBuffer()

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

func (r *ArrowRenderer) SetPosition(vbo *graphics.VertexBuffer) {
	r.sp.Position.SetSourceVertex(vbo, 0)
}

func (r *ArrowRenderer) RenderTangents(s *scene.Scene, c camera.Camera, colorTexture, depthTexture *graphics.Texture2D) {
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
	r.SetPosition(r.vbo)
	r.sp.OutColor.Set(colorTexture)
	r.sp.Depth.Set(depthTexture)
	r.sp.Render(len(r.points), r.renderOpts)
}

func (r *ArrowRenderer) RenderBitangents(s *scene.Scene, c camera.Camera, colorTexture, depthTexture *graphics.Texture2D) {
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
	r.SetPosition(r.vbo)
	r.sp.OutColor.Set(colorTexture)
	r.sp.Depth.Set(depthTexture)
	r.sp.Render(len(r.points), r.renderOpts)
}

func (r *ArrowRenderer) RenderNormals(s *scene.Scene, c camera.Camera, colorTexture, depthTexture *graphics.Texture2D) {
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
	r.SetPosition(r.vbo)
	r.sp.OutColor.Set(colorTexture)
	r.sp.Depth.Set(depthTexture)
	r.sp.Render(len(r.points), r.renderOpts)
}
