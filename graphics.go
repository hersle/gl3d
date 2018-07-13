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
	uniforms struct {
		modelMat *Uniform
		viewMat *Uniform
		projMat *Uniform
		normalMat *Uniform
		ambient *Uniform
		ambientLight *Uniform
		ambientMap *Uniform
		diffuse *Uniform
		diffuseLight *Uniform
		diffuseMap *Uniform
		specular *Uniform
		specularLight *Uniform
		shine *Uniform
		specularMap *Uniform
		lightPos *Uniform
		alpha *Uniform
	}
	//posAttr, texCoordAttr, normalAttr *Attrib
	attrs struct {
		pos *Attrib
		texCoord *Attrib
		normal *Attrib
	}
	vbo, ibo *Buffer
	vao *VertexArray
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
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	r.prog, err = ReadProgram("vshader.glsl", "fshader.glsl")
	if err != nil {
		return nil, err
	}
	gls.SetProgram(r.prog)

	var errs [19]error
	r.attrs.pos, errs[0] = r.prog.Attrib("position")
	r.attrs.texCoord, errs[1] = r.prog.Attrib("texCoordV")
	r.attrs.normal, errs[2] = r.prog.Attrib("normalV")
	r.uniforms.modelMat, errs[3] = r.prog.Uniform("modelMatrix")
	r.uniforms.viewMat, errs[4] = r.prog.Uniform("viewMatrix")
	r.uniforms.projMat, errs[5] = r.prog.Uniform("projectionMatrix")
	r.uniforms.normalMat, errs[6] = r.prog.Uniform("normalMatrix")
	r.uniforms.ambient, errs[7] = r.prog.Uniform("material.ambient")
	r.uniforms.diffuse, errs[8] = r.prog.Uniform("material.diffuse")
	r.uniforms.specular, errs[9] = r.prog.Uniform("material.specular")
	r.uniforms.ambientMap, errs[10] = r.prog.Uniform("material.ambientMap")
	r.uniforms.diffuseMap, errs[11] = r.prog.Uniform("material.diffuseMap")
	r.uniforms.specularMap, errs[12] = r.prog.Uniform("material.specularMap")
	r.uniforms.shine, errs[13] = r.prog.Uniform("material.shine")
	r.uniforms.alpha, errs[14] = r.prog.Uniform("material.alpha")
	r.uniforms.lightPos, errs[15] = r.prog.Uniform("light.position")
	r.uniforms.ambientLight, errs[16] = r.prog.Uniform("light.ambient")
	r.uniforms.diffuseLight, errs[17] = r.prog.Uniform("light.diffuse")
	r.uniforms.specularLight, errs[18] = r.prog.Uniform("light.specular")
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}

	r.vao = NewVertexArray()
	r.vao.SetAttribFormat(r.attrs.pos, 3, gl.FLOAT, false)
	r.vao.SetAttribFormat(r.attrs.normal, 3, gl.FLOAT, false)
	r.vao.SetAttribFormat(r.attrs.texCoord, 2, gl.FLOAT, false)

	r.ambientTexUnit = NewTextureUnit(0)
	r.diffuseTexUnit = NewTextureUnit(1)
	r.specularTexUnit = NewTextureUnit(2)

	r.prog.SetUniform(r.uniforms.ambientMap, r.ambientTexUnit)
	r.prog.SetUniform(r.uniforms.diffuseMap, r.diffuseTexUnit)
	r.prog.SetUniform(r.uniforms.specularMap, r.specularTexUnit)

	r.normalMat = NewMat4Zero()

	return &r, nil
}

func (r *Renderer) Clear() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (r *Renderer) renderMesh(m *Mesh, c *Camera) {
	r.normalMat.Copy(c.ViewMatrix())
	r.normalMat.Mult(m.modelMat)
	r.normalMat.Invert()
	r.normalMat.Transpose()

	r.prog.SetUniform(r.uniforms.modelMat, m.modelMat)
	r.prog.SetUniform(r.uniforms.viewMat, c.ViewMatrix())
	r.prog.SetUniform(r.uniforms.projMat, c.ProjectionMatrix())
	r.prog.SetUniform(r.uniforms.normalMat, r.normalMat)

	r.prog.SetUniform(r.uniforms.ambientLight, NewVec3(1, 1, 1))
	r.prog.SetUniform(r.uniforms.diffuseLight, NewVec3(1, 1, 1))
	r.prog.SetUniform(r.uniforms.specularLight, NewVec3(1, 1, 1))
	r.prog.SetUniform(r.uniforms.lightPos, NewVec3(0, +2.0, -5.0))

	for _, subMesh := range m.subMeshes {
		r.prog.SetUniform(r.uniforms.ambient, subMesh.mtl.ambient)
		r.prog.SetUniform(r.uniforms.diffuse, subMesh.mtl.diffuse)
		r.prog.SetUniform(r.uniforms.specular, subMesh.mtl.specular)
		r.prog.SetUniform(r.uniforms.shine, subMesh.mtl.shine)
		r.prog.SetUniform(r.uniforms.alpha, subMesh.mtl.alpha)

		stride := int(unsafe.Sizeof(Vertex{}))
		offset := int(unsafe.Offsetof(Vertex{}.pos))
		r.vao.SetAttribSource(r.attrs.pos, subMesh.vbo, offset, stride)
		offset = int(unsafe.Offsetof(Vertex{}.normal))
		r.vao.SetAttribSource(r.attrs.normal, subMesh.vbo, offset, stride)
		offset = int(unsafe.Offsetof(Vertex{}.texCoord))
		r.vao.SetAttribSource(r.attrs.texCoord, subMesh.vbo, offset, stride)
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
