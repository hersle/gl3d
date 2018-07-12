package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"unsafe"
)

type Vertex struct {
	pos Vec3
	texCoord Vec2
	normal Vec3
}

// TODO: redesign attr/uniform access system?
type Renderer struct {
	prog *Program
	vbo, ibo *Buffer
	vao *VertexArray
	posAttr, texCoordAttr, normalAttr *Attrib
	modelMatUfm, viewMatUfm, projMatUfm *Uniform
	normalMatUfm *Uniform
	ambientUfm, ambientLightUfm, ambientMapUfm *Uniform
	diffuseUfm, diffuseLightUfm, diffuseMapUfm *Uniform
	specularUfm, specularLightUfm, shineUfm, specularMapUfm *Uniform
	lightPosUfm *Uniform
	normalMat *Mat4
	ambientTexUnit, diffuseTexUnit, specularTexUnit *TextureUnit
}

func NewVertex(pos Vec3, texCoord Vec2, normal Vec3) Vertex {
	var vert Vertex
	vert.pos = pos
	vert.texCoord = texCoord
	vert.normal = normal
	return vert
}

func NewRenderer(win *Window) (*Renderer, error) {
	var r Renderer

	win.MakeContextCurrent()

	err := gl.Init()
	if err != nil {
		return nil, err
	}

	gl.Enable(gl.DEPTH_TEST)

	r.prog, err = ReadProgram("vshader.glsl", "fshader.glsl")
	if err != nil {
		return nil, err
	}
	gls.SetProgram(r.prog)

	r.posAttr, err = r.prog.Attrib("position")
	if err != nil {
		println(err.Error())
	}
	r.texCoordAttr, err = r.prog.Attrib("texCoordV")
	if err != nil {
		println(err.Error())
	}
	r.normalAttr, err = r.prog.Attrib("normalV")
	if err != nil {
		println(err.Error())
	}
	r.modelMatUfm, err = r.prog.Uniform("modelMatrix")
	if err != nil {
		println(err.Error())
	}
	r.viewMatUfm, err = r.prog.Uniform("viewMatrix")
	if err != nil {
		println(err.Error())
	}
	r.projMatUfm, err = r.prog.Uniform("projectionMatrix")
	if err != nil {
		println(err.Error())
	}
	r.normalMatUfm, err = r.prog.Uniform("normalMatrix")
	if err != nil {
		println(err.Error())
	}
	r.ambientUfm, err = r.prog.Uniform("ambient")
	if err != nil {
		println(err.Error())
	}
	r.ambientLightUfm, err = r.prog.Uniform("ambientLight")
	if err != nil {
		println(err.Error())
	}
	r.diffuseUfm, err = r.prog.Uniform("diffuse")
	if err != nil {
		println(err.Error())
	}
	r.diffuseLightUfm, err = r.prog.Uniform("diffuseLight")
	if err != nil {
		println(err.Error())
	}
	r.specularUfm, err = r.prog.Uniform("specular")
	if err != nil {
		println(err.Error())
	}
	r.specularLightUfm, err = r.prog.Uniform("specularLight")
	if err != nil {
		println(err.Error())
	}
	r.shineUfm, err = r.prog.Uniform("shine")
	if err != nil {
		println(err.Error())
	}
	r.lightPosUfm, err = r.prog.Uniform("lightPosition")
	if err != nil {
		println(err.Error())
	}
	r.ambientMapUfm, err = r.prog.Uniform("ambientMap")
	if err != nil {
		println(err.Error())
	}
	r.diffuseMapUfm, err = r.prog.Uniform("diffuseMap")
	if err != nil {
		println(err.Error())
	}
	r.specularMapUfm, err = r.prog.Uniform("specularMap")
	if err != nil {
		println(err.Error())
	}
	r.vao = NewVertexArray()
	r.vao.SetAttribFormat(r.posAttr, 3, gl.FLOAT, false)
	r.vao.SetAttribFormat(r.normalAttr, 3, gl.FLOAT, false)
	r.vao.SetAttribFormat(r.texCoordAttr, 2, gl.FLOAT, false)

	r.ambientTexUnit = NewTextureUnit(0)
	r.prog.SetUniform(r.ambientMapUfm, r.ambientTexUnit)

	r.diffuseTexUnit = NewTextureUnit(1)
	r.prog.SetUniform(r.diffuseMapUfm, r.diffuseTexUnit)

	r.specularTexUnit = NewTextureUnit(2)
	r.prog.SetUniform(r.specularMapUfm, r.specularTexUnit)

	r.normalMat = NewMat4Zero()

	return &r, nil
}

func (r *Renderer) Clear() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (r *Renderer) renderMesh(m *Mesh, c *Camera) {
	r.prog.SetUniform(r.modelMatUfm, m.modelMat)
	r.prog.SetUniform(r.viewMatUfm, c.ViewMatrix())
	r.prog.SetUniform(r.projMatUfm, c.ProjectionMatrix())

	r.normalMat.Copy(c.ViewMatrix())
	r.normalMat.Mult(m.modelMat)
	r.normalMat.Invert()
	r.normalMat.Transpose()
	r.prog.SetUniform(r.normalMatUfm, r.normalMat)

	r.prog.SetUniform(r.ambientLightUfm, NewVec3(1, 1, 1))
	r.prog.SetUniform(r.diffuseLightUfm, NewVec3(1, 1, 1))
	r.prog.SetUniform(r.specularLightUfm, NewVec3(1, 1, 1))
	r.prog.SetUniform(r.lightPosUfm, NewVec3(0, +2.0, -5.0))

	for _, subMesh := range m.subMeshes {
		r.prog.SetUniform(r.ambientUfm, subMesh.mtl.ambient)
		r.prog.SetUniform(r.diffuseUfm, subMesh.mtl.diffuse)
		r.prog.SetUniform(r.specularUfm, subMesh.mtl.specular)
		r.prog.SetUniform(r.shineUfm, subMesh.mtl.shine)

		stride := int(unsafe.Sizeof(Vertex{}))
		offset := int(unsafe.Offsetof(Vertex{}.pos))
		r.vao.SetAttribSource(r.posAttr, subMesh.vbo, offset, stride)
		offset = int(unsafe.Offsetof(Vertex{}.normal))
		r.vao.SetAttribSource(r.normalAttr, subMesh.vbo, offset, stride)
		offset = int(unsafe.Offsetof(Vertex{}.texCoord))
		r.vao.SetAttribSource(r.texCoordAttr, subMesh.vbo, offset, stride)
		r.vao.SetIndexBuffer(subMesh.ibo)

		r.ambientTexUnit.SetTexture2D(subMesh.mtl.ambientMapTexture)
		r.diffuseTexUnit.SetTexture2D(subMesh.mtl.diffuseMapTexture)
		r.specularTexUnit.SetTexture2D(subMesh.mtl.specularMapTexture)
		gls.SetVertexArray(r.vao)
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
