package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"unsafe"
)

type RGBAColor [4]uint8

type Vertex struct {
	pos Vec3
	color RGBAColor
	texCoord Vec2
}

type Renderer struct {
	prog Program
	vbo *Buffer
	ibo *Buffer
	vaoId uint32
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

	gl.GenVertexArrays(1, &r.vaoId)
	gl.BindVertexArray(r.vaoId)

	r.vbo = NewBuffer(gl.ARRAY_BUFFER)
	r.vbo.bind()

	r.ibo = NewBuffer(gl.ELEMENT_ARRAY_BUFFER)
	r.ibo.bind()

	stride := int32(unsafe.Sizeof(Vertex{}))
	offset := gl.PtrOffset(int(unsafe.Offsetof(Vertex{}.pos)))
	gl.VertexAttribPointer(r.posAttr.id, 3, gl.FLOAT, false, stride, offset)
	gl.EnableVertexAttribArray(r.posAttr.id)

	offset = gl.PtrOffset(int(unsafe.Offsetof(Vertex{}.color)))
	gl.VertexAttribPointer(r.colorAttr.id, 4, gl.UNSIGNED_BYTE, true, stride, offset)
	gl.EnableVertexAttribArray(r.colorAttr.id)

	offset = gl.PtrOffset(int(unsafe.Offsetof(Vertex{}.texCoord)))
	gl.VertexAttribPointer(r.texCoordAttr.id, 2, gl.FLOAT, false, stride, offset)
	gl.EnableVertexAttribArray(r.texCoordAttr.id)

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
	gl.UniformMatrix4fv(int32(r.projViewModelMatUfm.id), 1, true, &m[0])
}

func (r *Renderer) SetViewport(l, b, w, h int) {
	gl.Viewport(int32(l), int32(b), int32(w), int32(h))
}

func (r *Renderer) SetFullViewport(win *Window) {
	w, h := win.Size()
	r.SetViewport(0, 0, w, h)
}
