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

type DebugRenderer struct {
	sp          *DebugShaderProgram
	points      []math.Vec3
	vbo         *graphics.Buffer
	ibo         *graphics.Buffer
	renderState *graphics.RenderState
}

type DebugShaderProgram struct {
	*graphics.ShaderProgram
	ModelMatrix      *graphics.UniformMatrix4
	ViewMatrix       *graphics.UniformMatrix4
	ProjectionMatrix *graphics.UniformMatrix4
	Color            *graphics.UniformVector3
	Position         *graphics.Attrib
}

func NewDebugShaderProgram() *DebugShaderProgram {
	var sp DebugShaderProgram
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

func NewDebugRenderer() *DebugRenderer {
	var r DebugRenderer

	r.sp = NewDebugShaderProgram()

	r.renderState = graphics.NewRenderState()
	r.renderState.Program = r.sp.ShaderProgram
	r.renderState.PrimitiveType = graphics.Line
	r.renderState.DepthTest = graphics.LessEqualTest

	r.vbo = graphics.NewBuffer()
	r.ibo = graphics.NewBuffer()
	r.SetPosition(r.vbo)

	return &r
}

func (r *DebugRenderer) SetCamera(c camera.Camera) {
	r.sp.ViewMatrix.Set(c.ViewMatrix())
	r.sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *DebugRenderer) SetMesh(m *object.Mesh) {
	r.sp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *DebugRenderer) SetColor(color math.Vec3) {
	r.sp.Color.Set(color)
}

func (r *DebugRenderer) SetPosition(vbo *graphics.Buffer) {
	stride := int(unsafe.Sizeof(math.Vec3{0, 0, 0}))
	r.sp.Position.SetSource(vbo, 0, stride)
}

func (r *DebugRenderer) RenderTangents(s *scene.Scene, c camera.Camera, fb *graphics.Framebuffer) {
	r.renderState.DepthTest = graphics.LessEqualTest
	r.renderState.Framebuffer = fb
	r.renderState.PrimitiveType = graphics.Line
	r.sp.SetAttribIndexBuffer(nil)
	stride := int(unsafe.Sizeof(math.Vec3{0, 0, 0}))
	r.sp.Position.SetSource(r.vbo, 0, stride)

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

func (r *DebugRenderer) RenderBitangents(s *scene.Scene, c camera.Camera, fb *graphics.Framebuffer) {
	r.renderState.DepthTest = graphics.LessEqualTest
	r.renderState.Framebuffer = fb
	r.renderState.PrimitiveType = graphics.Line
	r.sp.SetAttribIndexBuffer(nil)
	stride := int(unsafe.Sizeof(math.Vec3{0, 0, 0}))
	r.sp.Position.SetSource(r.vbo, 0, stride)

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

func (r *DebugRenderer) RenderNormals(s *scene.Scene, c camera.Camera, fb *graphics.Framebuffer) {
	r.renderState.DepthTest = graphics.LessEqualTest
	r.renderState.Framebuffer = fb
	r.renderState.PrimitiveType = graphics.Line
	r.sp.SetAttribIndexBuffer(nil)
	stride := int(unsafe.Sizeof(math.Vec3{0, 0, 0}))
	r.sp.Position.SetSource(r.vbo, 0, stride)

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

func (r *DebugRenderer) AddLine(p1, p2 math.Vec3) {
	println(p1.String())
	println(p2.String())
	r.points = append(r.points, p1, p2)
}

func (r *DebugRenderer) RenderCameraFrustum(c camera.Camera, vc camera.Camera, fb *graphics.Framebuffer) {
	r.SetCamera(vc)
	r.points = r.points[:0]
	r.SetColor(math.Vec3{1, 1, 0})
	var m math.Mat4
	m.Identity()
	r.sp.ModelMatrix.Set(&m)

	r.renderState.DepthTest = graphics.LessEqualTest
	r.renderState.Framebuffer = fb
	r.renderState.PrimitiveType = graphics.Line
	r.sp.SetAttribIndexBuffer(nil)
	stride := int(unsafe.Sizeof(math.Vec3{0, 0, 0}))
	r.sp.Position.SetSource(r.vbo, 0, stride)

	switch c.(type) {
	case *camera.PerspectiveCamera:
		c := c.(*camera.PerspectiveCamera)
		f := c.Frustum()

		nw := f.NearWidth
		nh := f.NearHeight

		nc := f.Org.Add(f.Dir.Scale(f.NearDist))
		nbl := nc.Add(f.Right.Scale(-nw / 2)).Add(f.Up.Scale(-nh / 2))
		nbr := nc.Add(f.Right.Scale(+nw / 2)).Add(f.Up.Scale(-nh / 2))
		ntr := nc.Add(f.Right.Scale(+nw / 2)).Add(f.Up.Scale(+nh / 2))
		ntl := nc.Add(f.Right.Scale(-nw / 2)).Add(f.Up.Scale(+nh / 2))

		fw := (f.FarDist / f.NearDist) * f.NearWidth
		fh := (f.FarDist / f.NearDist) * f.NearHeight

		fc := f.Org.Add(f.Dir.Scale(f.FarDist))
		fbl := fc.Add(f.Right.Scale(-fw / 2)).Add(f.Up.Scale(-fh / 2))
		fbr := fc.Add(f.Right.Scale(+fw / 2)).Add(f.Up.Scale(-fh / 2))
		ftr := fc.Add(f.Right.Scale(+fw / 2)).Add(f.Up.Scale(+fh / 2))
		ftl := fc.Add(f.Right.Scale(-fw / 2)).Add(f.Up.Scale(+fh / 2))

		// near rectangle
		r.AddLine(nbl, nbr)
		r.AddLine(nbr, ntr)
		r.AddLine(ntr, ntl)
		r.AddLine(ntl, nbl)

		// near-far joining lines
		r.AddLine(nbl, fbl)
		r.AddLine(nbr, fbr)
		r.AddLine(ntr, ftr)
		r.AddLine(ntl, ftl)

		// far rectangle
		r.AddLine(fbl, fbr)
		r.AddLine(fbr, ftr)
		r.AddLine(ftr, ftl)
		r.AddLine(ftl, fbl)

		r.vbo.SetData(r.points, 0)
		graphics.NewRenderCommand(len(r.points), r.renderState).Execute()
	default:
		panic("can only render perspective camera frustum")
	}
}

func (r *DebugRenderer) RenderMeshWireframe(mesh *object.Mesh, c camera.Camera, fb *graphics.Framebuffer) {
	r.SetCamera(c)
	r.points = r.points[:0]
	r.SetColor(math.Vec3{1, 0, 0})

	r.renderState.DepthTest = graphics.LessEqualTest
	r.renderState.Framebuffer = fb
	r.renderState.PrimitiveType = graphics.Triangle
	r.renderState.TriangleMode = graphics.LineTriangleMode

	var v object.Vertex
	r.SetMesh(mesh)
	for _, subMesh := range mesh.SubMeshes {
		r.vbo.SetData(subMesh.Geo.Verts, 0)
		r.ibo.SetData(subMesh.Geo.Faces, 0)
		r.sp.Position.SetSource(r.vbo, v.PositionOffset(), v.Size())
		r.sp.SetAttribIndexBuffer(r.ibo)
		graphics.NewRenderCommand(subMesh.Geo.Inds, r.renderState).Execute()
	}
}
