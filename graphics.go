package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"unsafe"
)

// TODO: use image/color.RGBA
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
	verts []Vertex
	inds []int32
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

	r.vao = NewVertexArray()

	r.vbo = NewBuffer()
	r.ibo = NewBuffer()

	r.vao.SetIndexBuffer(r.ibo)

	stride := int(unsafe.Sizeof(Vertex{}))

	offset := int(unsafe.Offsetof(Vertex{}.pos))
	r.vao.SetAttribFormat(r.posAttr, 3, gl.FLOAT, false)
	r.vao.SetAttribSource(r.posAttr, r.vbo, offset, stride)

	offset = int(unsafe.Offsetof(Vertex{}.color))
	r.vao.SetAttribFormat(r.colorAttr, 4, gl.UNSIGNED_BYTE, true)
	r.vao.SetAttribSource(r.colorAttr, r.vbo, offset, stride)

	offset = int(unsafe.Offsetof(Vertex{}.texCoord))
	r.vao.SetAttribFormat(r.texCoordAttr, 2, gl.FLOAT, false)
	r.vao.SetAttribSource(r.texCoordAttr, r.vbo, offset, stride)

	return &r, nil
}

func (r *Renderer) Clear() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (r *Renderer) renderMesh(m *Mesh, c *Camera) {
	r.projViewModelMat.Identity()
	r.projViewModelMat.Mult(c.ProjectionViewMatrix())
	r.projViewModelMat.Mult(m.modelMat)
	r.SetProjectionViewModelMatrix(r.projViewModelMat)
	m.tex.bind()
	for _, i := range m.faces {
		r.inds = append(r.inds, int32(len(r.verts) + i))
	}
	for _, vert := range m.verts {
		r.verts = append(r.verts, vert)
	}

	r.vbo.SetData(r.verts, 0)
	r.ibo.SetData(r.inds, 0)

	r.vao.bind()
	gl.DrawElements(gl.TRIANGLES, int32(len(r.inds)), gl.UNSIGNED_INT, nil)

	r.verts = r.verts[:0]
	r.inds = r.inds[:0]
}

func (r *Renderer) Render(s *Scene, c *Camera) {
	for _, m := range s.meshes {
		r.renderMesh(m, c)
	}
}

func (r *Renderer) SetProjectionViewModelMatrix(m *Mat4) {
	r.projViewModelMatUfm.Set(m)
}

func (r *Renderer) SetViewport(l, b, w, h int) {
	gl.Viewport(int32(l), int32(b), int32(w), int32(h))
}

func (r *Renderer) SetFullViewport(win *Window) {
	w, h := win.Size()
	r.SetViewport(0, 0, w, h)
}
