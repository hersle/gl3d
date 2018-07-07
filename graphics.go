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
	normal Vec3
}

type Renderer struct {
	prog *Program
	vbo *Buffer
	ibo *Buffer
	vao *VertexArray
	posAttr *Attrib
	colorAttr *Attrib
	texCoordAttr *Attrib
	normalAttr *Attrib
	modelMatUfm *Uniform
	viewMatUfm *Uniform
	projMatUfm *Uniform
	lightPosUfm *Uniform
	ambientUfm *Uniform
	ambientLightUfm *Uniform
	diffuseUfm *Uniform
	diffuseLightUfm *Uniform
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
	r.normalAttr, err = r.prog.attrib("normalV")
	if err != nil {
		println(err.Error())
	}
	r.modelMatUfm, err = r.prog.uniform("modelMatrix")
	if err != nil {
		println(err.Error())
	}
	r.viewMatUfm, err = r.prog.uniform("viewMatrix")
	if err != nil {
		println(err.Error())
	}
	r.projMatUfm, err = r.prog.uniform("projectionMatrix")
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
	r.diffuseUfm, err = r.prog.uniform("diffuse")
	if err != nil {
		println(err.Error())
	}
	r.diffuseLightUfm, err = r.prog.uniform("diffuseLight")
	if err != nil {
		println(err.Error())
	}
	r.lightPosUfm, err = r.prog.uniform("lightPosition")
	if err != nil {
		println(err.Error())
	}

	r.vao = NewVertexArray()
	r.vao.SetAttribFormat(r.posAttr, 3, gl.FLOAT, false)
	r.vao.SetAttribFormat(r.normalAttr, 3, gl.FLOAT, false)

	return &r, nil
}

func (r *Renderer) Clear() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (r *Renderer) renderMesh(m *Mesh, c *Camera) {
	c.UpdateMatrices()
	r.prog.SetUniform(r.modelMatUfm, m.modelMat)
	r.prog.SetUniform(r.viewMatUfm, c.viewMat)
	r.prog.SetUniform(r.projMatUfm, c.projMat)

	r.prog.SetUniform(r.ambientLightUfm, NewVec3(0.5, 0.5, 0.5))
	r.prog.SetUniform(r.diffuseLightUfm, NewVec3(1.0, 1.0, 1.0))
	r.prog.SetUniform(r.lightPosUfm, NewVec3(0, +2.0, -5.0))

	for _, subMesh := range m.subMeshes {
		r.prog.SetUniform(r.ambientUfm, subMesh.mtl.ambient)
		r.prog.SetUniform(r.diffuseUfm, subMesh.mtl.diffuse)

		stride := int(unsafe.Sizeof(Vertex{}))
		offset := int(unsafe.Offsetof(Vertex{}.pos))
		r.vao.SetAttribSource(r.posAttr, subMesh.vbo, offset, stride)
		offset = int(unsafe.Offsetof(Vertex{}.normal))
		r.vao.SetAttribSource(r.normalAttr, subMesh.vbo, offset, stride)
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
