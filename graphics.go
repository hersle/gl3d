package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"unsafe"
)

// TODO: use image/color.RGBA; must be [4] array, not struct
type RGBAColor [4]uint8

type Vertex struct {
	pos Vec3
	color RGBAColor
	texCoord Vec2
}

type Renderer struct {
	prog *Program
	vbo *Buffer
	ibo *Buffer
	vao *VertexArray
	posAttr *Attrib
	colorAttr *Attrib
	texCoordAttr *Attrib
	projViewModelMat *Mat4
	projViewModelMatUfm *Uniform
	ambientUfm *Uniform
	ambientLightUfm *Uniform
}

func NewColor(r, g, b, a uint8) RGBAColor {
	return RGBAColor{r, g, b, a}
}

func NewRenderer(win *Window) (*Renderer, error) {
	var r Renderer

	win.MakeContextCurrent()

	err := gl.Init()
	if err != nil {
		return nil, err
	}

	r.projViewModelMat = NewMat4Zero()

	gl.Enable(gl.DEPTH_TEST)

	r.prog, err = NewProgramFromFiles("vshader.glsl", "fshader.glsl")
	if err != nil {
		return nil, err
	}
	r.prog.use()

	r.posAttr, err = r.prog.attrib("position")
	if err != nil {
		println(err.Error())
	}
	r.colorAttr, err = r.prog.attrib("colorV")
	if err != nil {
		println(err.Error())
	}
	r.texCoordAttr, err = r.prog.attrib("texCoordV")
	if err != nil {
		println(err.Error())
	}
	r.projViewModelMatUfm, err = r.prog.uniform("projectionViewModelMatrix")
	if err != nil {
		println(err.Error())
	}
	r.ambientUfm, err = r.prog.uniform("ambient")
	if err != nil {
		println(err.Error())
	}
	r.ambientLightUfm, err = r.prog.uniform("ambientLight")
	if err != nil {
		println(err.Error())
	}

	r.vao = NewVertexArray()
	r.vao.SetAttribFormat(r.posAttr, 3, gl.FLOAT, false)

	return &r, nil
}

func (r *Renderer) Clear() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (r *Renderer) renderMesh(m *Mesh, c *Camera) {
	r.projViewModelMat.Identity()
	r.projViewModelMat.Mult(c.ProjectionViewMatrix())
	r.projViewModelMat.Mult(m.modelMat)
	r.prog.SetUniform(r.projViewModelMatUfm, r.projViewModelMat)

	r.prog.SetUniform(r.ambientLightUfm, NewVec3(0.5, 0.5, 0.5))

	for _, subMesh := range m.subMeshes {
		r.prog.SetUniform(r.ambientUfm, subMesh.mtl.ambient)

		stride := int(unsafe.Sizeof(Vertex{}))
		offset := int(unsafe.Offsetof(Vertex{}.pos))
		r.vao.SetAttribSource(r.posAttr, subMesh.vbo, offset, stride)
		r.vao.SetIndexBuffer(subMesh.ibo)

		r.vao.bind()
		gl.DrawElements(gl.TRIANGLES, int32(subMesh.inds), gl.UNSIGNED_INT, nil)
	}
}

func (r *Renderer) Render(s *Scene, c *Camera) {
	for _, m := range s.meshes {
		r.renderMesh(m, c)
	}
}

func (r *Renderer) SetViewport(l, b, w, h int) {
	gl.Viewport(int32(l), int32(b), int32(w), int32(h))
}

func (r *Renderer) SetFullViewport(win *Window) {
	w, h := win.Size()
	r.SetViewport(0, 0, w, h)
}
