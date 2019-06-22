package render

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/scene"
	"unsafe"
)

type ArrowRenderer struct {
	sp          *ArrowShaderProgram
	points      []math.Vec3
	vbo         *graphics.Buffer
	renderState *graphics.RenderState
}

type ArrowShaderProgram struct {
	*graphics.ShaderProgram
	ModelMatrix      *graphics.UniformMatrix4
	ViewMatrix       *graphics.UniformMatrix4
	ProjectionMatrix *graphics.UniformMatrix4
	Color            *graphics.UniformVector3
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
	sp.ModelMatrix = sp.UniformMatrix4("modelMatrix")
	sp.ViewMatrix = sp.UniformMatrix4("viewMatrix")
	sp.ProjectionMatrix = sp.UniformMatrix4("projectionMatrix")
	sp.Color = sp.UniformVector3("color")

	sp.Position.SetFormat(gl.FLOAT, false) // TODO: remove dependency on GL constants

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
	stride := int(unsafe.Sizeof(math.Vec3{0, 0, 0}))
	r.sp.Position.SetSource(vbo, 0, stride)
}

func (r *ArrowRenderer) RenderTangents(s *scene.Node, c camera.Camera, fb *graphics.Framebuffer) {
	r.SetCamera(c)
	r.points = r.points[:0]
	r.SetColor(math.Vec3{1, 0, 0})

	f := func(n *scene.Node, _ int) {
		if n.Mesh == nil {
			return
		}

		m := n.Mesh

		r.SetMesh(m)
		for _, subMesh := range m.SubMeshes {
			for _, i := range subMesh.Geo.Faces {
				p1 := subMesh.Geo.Verts[i].Position
				p2 := p1.Add(subMesh.Geo.Verts[i].Tangent)
				r.points = append(r.points, p1, p2)
			}
		}
	}

	s.Traverse(f)

	r.vbo.SetData(r.points, 0)
	r.renderState.Framebuffer = fb
	graphics.NewRenderCommand(len(r.points), r.renderState).Execute()
}

func (r *ArrowRenderer) RenderBitangents(s *scene.Node, c camera.Camera, fb *graphics.Framebuffer) {
	r.SetCamera(c)
	r.points = r.points[:0]
	r.SetColor(math.Vec3{0, 1, 0})

	f := func(n *scene.Node, _ int) {
		if n.Mesh == nil {
			return
		}

		m := n.Mesh

		r.SetMesh(m)
		for _, subMesh := range m.SubMeshes {
			for _, i := range subMesh.Geo.Faces {
				p1 := subMesh.Geo.Verts[i].Position
				p2 := p1.Add(subMesh.Geo.Verts[i].Bitangent())
				r.points = append(r.points, p1, p2)
			}
		}
	}

	s.Traverse(f)

	r.vbo.SetData(r.points, 0)
	r.renderState.Framebuffer = fb
	graphics.NewRenderCommand(len(r.points), r.renderState).Execute()
}

func (r *ArrowRenderer) RenderNormals(s *scene.Node, c camera.Camera, fb *graphics.Framebuffer) {
	r.SetCamera(c)
	r.points = r.points[:0]
	r.SetColor(math.Vec3{0, 0, 1})

	f := func(n *scene.Node, _ int) {
		if n.Mesh == nil {
			return
		}

		m := n.Mesh

		r.SetMesh(m)
		for _, subMesh := range m.SubMeshes {
			for _, i := range subMesh.Geo.Faces {
				p1 := subMesh.Geo.Verts[i].Position
				p2 := p1.Add(subMesh.Geo.Verts[i].Normal)
				r.points = append(r.points, p1, p2)
			}
		}
	}

	s.Traverse(f)
	r.vbo.SetData(r.points, 0)
	r.renderState.Framebuffer = fb
	graphics.NewRenderCommand(len(r.points), r.renderState).Execute()
}
