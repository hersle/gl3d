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
	win *Window
	prog *ShaderProgram
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
		passID *Uniform
		shadowModelMat *Uniform
		shadowViewMat *Uniform
		shadowProjMat *Uniform
		shadowMap *Uniform
	}
	attrs struct {
		pos *Attrib
		texCoord *Attrib
		normal *Attrib
	}
	vbo, ibo *Buffer
	vao *VertexArray
	normalMat *Mat4
	ambientTexUnit, diffuseTexUnit, specularTexUnit *TextureUnit
	shadowTexUnit *TextureUnit

	shadowFb *Framebuffer
	shadowTex *Texture2D
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

	r.prog, err = ReadShaderProgram("vshader.glsl", "fshader.glsl")
	if err != nil {
		return nil, err
	}
	gls.SetShaderProgram(r.prog)

	var errs [24]error
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
	r.uniforms.passID, errs[19] = r.prog.Uniform("passID")
	r.uniforms.shadowModelMat, errs[20] = r.prog.Uniform("shadowModelMatrix")
	r.uniforms.shadowViewMat, errs[21] = r.prog.Uniform("shadowViewMatrix")
	r.uniforms.shadowProjMat, errs[22] = r.prog.Uniform("shadowProjectionMatrix")
	r.uniforms.shadowMap, errs[23] = r.prog.Uniform("shadowMap")
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
	r.shadowTexUnit = NewTextureUnit(3)

	r.uniforms.ambientMap.Set(r.ambientTexUnit)
	r.uniforms.diffuseMap.Set(r.diffuseTexUnit)
	r.uniforms.specularMap.Set(r.specularTexUnit)
	r.uniforms.shadowMap.Set(r.shadowTexUnit)

	r.normalMat = NewMat4Zero()

	r.win = win

	r.shadowTex = NewTexture2D(gl.NEAREST, gl.CLAMP_TO_BORDER)
	r.shadowTex.SetStorage(1, gl.DEPTH_COMPONENT16, 512, 512)
	r.shadowTex.SetBorderColor(NewVec4(1, 1, 1, 1))

	r.shadowFb = NewFramebuffer()
	r.shadowFb.SetTexture(gl.DEPTH_ATTACHMENT, r.shadowTex, 0)
	println(r.shadowFb.Complete())


	return &r, nil
}

func (r *Renderer) Clear() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (r *Renderer) renderMesh(s *Scene, m *Mesh, c *Camera) {
	r.normalMat.Copy(c.ViewMatrix()).Mult(m.WorldMatrix())
	r.normalMat.Invert().Transpose()

	r.uniforms.modelMat.Set(m.WorldMatrix())
	r.uniforms.viewMat.Set(c.ViewMatrix())
	r.uniforms.projMat.Set(c.ProjectionMatrix())
	r.uniforms.normalMat.Set(r.normalMat)

	r.uniforms.shadowModelMat.Set(m.WorldMatrix())
	r.uniforms.shadowViewMat.Set(s.Light.ViewMatrix())
	r.uniforms.shadowProjMat.Set(s.Light.ProjectionMatrix())

	for _, subMesh := range m.subMeshes {
		r.uniforms.ambient.Set(subMesh.mtl.ambient)
		r.uniforms.diffuse.Set(subMesh.mtl.diffuse)
		r.uniforms.specular.Set(subMesh.mtl.specular)
		r.uniforms.shine.Set(subMesh.mtl.shine)
		r.uniforms.alpha.Set(subMesh.mtl.alpha)

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
		r.shadowTexUnit.SetTexture2D(r.shadowTex)
		gls.SetVertexArray(r.vao)
		gl.DrawElements(gl.TRIANGLES, int32(subMesh.inds), gl.UNSIGNED_INT, nil)
	}
}

func (r *Renderer) shadowPass(s *Scene, c *Camera) {
	r.uniforms.passID.Set(int(1))
	r.SetViewport(0, 0, 512, 512)
	r.shadowFb.ClearDepth(1)
	gls.SetDrawFramebuffer(r.shadowFb)
	r.uniforms.viewMat.Set(s.Light.ViewMatrix())
	r.uniforms.projMat.Set(s.Light.ProjectionMatrix())

	for _, m := range s.meshes {
		r.uniforms.modelMat.Set(m.WorldMatrix())
		for _, subMesh := range m.subMeshes {
			stride := int(unsafe.Sizeof(Vertex{}))
			offset := int(unsafe.Offsetof(Vertex{}.pos))
			r.vao.SetAttribSource(r.attrs.pos, subMesh.vbo, offset, stride)
			r.vao.SetIndexBuffer(subMesh.ibo)

			gls.SetVertexArray(r.vao)
			gl.DrawElements(gl.TRIANGLES, int32(subMesh.inds), gl.UNSIGNED_INT, nil)
		}
	}
}

func (r *Renderer) Render(s *Scene, c *Camera) {
	// shadow pass
	r.shadowPass(s, c)

	// normal pass
	r.uniforms.passID.Set(int(2))
	r.SetFullViewport(r.win)
	gls.SetDrawFramebuffer(defaultFramebuffer)
	r.uniforms.lightPos.Set(s.Light.position)
	r.uniforms.ambientLight.Set(s.Light.ambient)
	r.uniforms.diffuseLight.Set(s.Light.diffuse)
	r.uniforms.specularLight.Set(s.Light.specular)
	for _, m := range s.meshes {
		r.renderMesh(s, m, c)
	}

	// draw test quad
	s.quad.subMeshes[0].mtl.ambientMapTexture = r.shadowTex
	ident := NewMat4Identity()
	r.uniforms.modelMat.Set(ident)
	r.uniforms.viewMat.Set(ident)
	r.uniforms.projMat.Set(ident)
	for _, subMesh := range s.quad.subMeshes {
		r.uniforms.ambient.Set(subMesh.mtl.ambient)
		r.uniforms.diffuse.Set(subMesh.mtl.diffuse)
		r.uniforms.specular.Set(subMesh.mtl.specular)
		r.uniforms.shine.Set(subMesh.mtl.shine)
		r.uniforms.alpha.Set(subMesh.mtl.alpha)
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

func (r *Renderer) SetViewport(l, b, w, h int) {
	gl.Viewport(int32(l), int32(b), int32(w), int32(h))
}

func (r *Renderer) SetFullViewport(win *Window) {
	w, h := win.Size()
	r.SetViewport(0, 0, w, h)
}
