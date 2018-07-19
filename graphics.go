package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"unsafe"
	"path"
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
		modelMat *UniformMatrix4
		viewMat *UniformMatrix4
		projMat *UniformMatrix4
		normalMat *UniformMatrix4
		ambient *UniformVector3
		ambientLight *UniformVector3
		ambientMap *UniformSampler
		diffuse *UniformVector3
		diffuseLight *UniformVector3
		diffuseMap *UniformSampler
		specular *UniformVector3
		specularLight *UniformVector3
		shine *UniformFloat
		specularMap *UniformSampler
		lightPos *UniformVector3
		alpha *UniformFloat
		shadowModelMat *UniformMatrix4
		shadowViewMat *UniformMatrix4
		shadowProjMat *UniformMatrix4
		shadowMap *UniformSampler
	}
	attrs struct {
		pos *Attrib
		texCoord *Attrib
		normal *Attrib
	}
	vbo, ibo *Buffer
	vao *VertexArray
	normalMat *Mat4

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

	r.prog, err = ReadShaderProgram("vshader.glsl", "fshader.glsl")
	if err != nil {
		return nil, err
	}

	var errs [24]error
	r.attrs.pos, errs[0] = r.prog.Attrib("position")
	r.attrs.texCoord, errs[1] = r.prog.Attrib("texCoordV")
	r.attrs.normal, errs[2] = r.prog.Attrib("normalV")
	// TODO: assign uniforms only name and program, let them handle rest themselves?
	r.uniforms.modelMat, errs[3] = r.prog.UniformMatrix4("modelMatrix")
	r.uniforms.viewMat, errs[4] = r.prog.UniformMatrix4("viewMatrix")
	r.uniforms.projMat, errs[5] = r.prog.UniformMatrix4("projectionMatrix")
	r.uniforms.normalMat, errs[6] = r.prog.UniformMatrix4("normalMatrix")
	r.uniforms.ambient, errs[7] = r.prog.UniformVector3("material.ambient")
	r.uniforms.diffuse, errs[8] = r.prog.UniformVector3("material.diffuse")
	r.uniforms.specular, errs[9] = r.prog.UniformVector3("material.specular")
	r.uniforms.ambientMap, errs[10] = r.prog.UniformSampler("material.ambientMap")
	r.uniforms.diffuseMap, errs[11] = r.prog.UniformSampler("material.diffuseMap")
	r.uniforms.specularMap, errs[12] = r.prog.UniformSampler("material.specularMap")
	r.uniforms.shine, errs[13] = r.prog.UniformFloat("material.shine")
	r.uniforms.alpha, errs[14] = r.prog.UniformFloat("material.alpha")
	r.uniforms.lightPos, errs[15] = r.prog.UniformVector3("light.position")
	r.uniforms.ambientLight, errs[16] = r.prog.UniformVector3("light.ambient")
	r.uniforms.diffuseLight, errs[17] = r.prog.UniformVector3("light.diffuse")
	r.uniforms.specularLight, errs[18] = r.prog.UniformVector3("light.specular")
	// TODO: 19
	r.uniforms.shadowModelMat, errs[20] = r.prog.UniformMatrix4("shadowModelMatrix")
	r.uniforms.shadowViewMat, errs[21] = r.prog.UniformMatrix4("shadowViewMatrix")
	r.uniforms.shadowProjMat, errs[22] = r.prog.UniformMatrix4("shadowProjectionMatrix")
	r.uniforms.shadowMap, errs[23] = r.prog.UniformSampler("shadowMap")
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}

	r.vao = NewVertexArray()
	r.vao.SetAttribFormat(r.attrs.pos, 3, gl.FLOAT, false)
	r.vao.SetAttribFormat(r.attrs.normal, 3, gl.FLOAT, false)
	r.vao.SetAttribFormat(r.attrs.texCoord, 2, gl.FLOAT, false)

	r.normalMat = NewMat4Zero()

	r.win = win

	r.shadowTex = NewTexture2D(gl.NEAREST, gl.CLAMP_TO_BORDER, gl.DEPTH_COMPONENT16, 512, 512)
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

	r.prog.SetUniformMatrix4(r.uniforms.modelMat, m.WorldMatrix())
	r.prog.SetUniformMatrix4(r.uniforms.viewMat, c.ViewMatrix())
	r.prog.SetUniformMatrix4(r.uniforms.projMat, c.ProjectionMatrix())
	r.prog.SetUniformMatrix4(r.uniforms.normalMat, r.normalMat)

	r.prog.SetUniformMatrix4(r.uniforms.shadowModelMat, m.WorldMatrix())
	r.prog.SetUniformMatrix4(r.uniforms.shadowViewMat, s.Light.ViewMatrix())
	r.prog.SetUniformMatrix4(r.uniforms.shadowProjMat, s.Light.ProjectionMatrix())

	for _, subMesh := range m.subMeshes {
		r.prog.SetUniformVector3(r.uniforms.ambient, subMesh.mtl.ambient)
		r.prog.SetUniformVector3(r.uniforms.diffuse, subMesh.mtl.diffuse)
		r.prog.SetUniformVector3(r.uniforms.specular, subMesh.mtl.specular)
		r.prog.SetUniformFloat(r.uniforms.shine, subMesh.mtl.shine)
		r.prog.SetUniformFloat(r.uniforms.alpha, subMesh.mtl.alpha)

		stride := int(unsafe.Sizeof(Vertex{}))
		offset := int(unsafe.Offsetof(Vertex{}.pos))
		r.vao.SetAttribSource(r.attrs.pos, subMesh.vbo, offset, stride)
		offset = int(unsafe.Offsetof(Vertex{}.normal))
		r.vao.SetAttribSource(r.attrs.normal, subMesh.vbo, offset, stride)
		offset = int(unsafe.Offsetof(Vertex{}.texCoord))
		r.vao.SetAttribSource(r.attrs.texCoord, subMesh.vbo, offset, stride)
		r.vao.SetIndexBuffer(subMesh.ibo)

		r.prog.SetUniformSampler(r.uniforms.ambientMap, subMesh.mtl.ambientMapTexture)
		r.prog.SetUniformSampler(r.uniforms.diffuseMap, subMesh.mtl.diffuseMapTexture)
		r.prog.SetUniformSampler(r.uniforms.specularMap, subMesh.mtl.specularMapTexture)
		r.prog.SetUniformSampler(r.uniforms.shadowMap, r.shadowTex)
		gls.SetVertexArray(r.vao)
		gl.DrawElements(gl.TRIANGLES, int32(subMesh.inds), gl.UNSIGNED_INT, nil)
	}
}

func (r *Renderer) shadowPass(s *Scene, c *Camera) {
	r.SetViewport(0, 0, 512, 512)
	r.shadowFb.ClearDepth(1)
	gls.SetDrawFramebuffer(r.shadowFb)
	r.prog.SetUniformMatrix4(r.uniforms.viewMat, s.Light.ViewMatrix())
	r.prog.SetUniformMatrix4(r.uniforms.projMat, s.Light.ProjectionMatrix())

	for _, m := range s.meshes {
		r.prog.SetUniformMatrix4(r.uniforms.modelMat, m.WorldMatrix())
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
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gls.SetShaderProgram(r.prog)

	// shadow pass
	r.shadowPass(s, c)

	// normal pass
	r.SetFullViewport(r.win)
	gls.SetDrawFramebuffer(defaultFramebuffer)
	r.prog.SetUniformVector3(r.uniforms.lightPos, s.Light.position)
	r.prog.SetUniformVector3(r.uniforms.ambientLight, s.Light.ambient)
	r.prog.SetUniformVector3(r.uniforms.diffuseLight, s.Light.diffuse)
	r.prog.SetUniformVector3(r.uniforms.specularLight, s.Light.specular)
	for _, m := range s.meshes {
		r.renderMesh(s, m, c)
	}

	// draw test quad
	s.quad.subMeshes[0].mtl.ambientMapTexture = r.shadowTex
	ident := NewMat4Identity()
	r.prog.SetUniformMatrix4(r.uniforms.modelMat, ident)
	r.prog.SetUniformMatrix4(r.uniforms.viewMat, ident)
	r.prog.SetUniformMatrix4(r.uniforms.projMat, ident)
	for _, subMesh := range s.quad.subMeshes {
		r.prog.SetUniformVector3(r.uniforms.ambient, subMesh.mtl.ambient)
		stride := int(unsafe.Sizeof(Vertex{}))
		offset := int(unsafe.Offsetof(Vertex{}.pos))
		r.vao.SetAttribSource(r.attrs.pos, subMesh.vbo, offset, stride)
		offset = int(unsafe.Offsetof(Vertex{}.texCoord))
		r.vao.SetAttribSource(r.attrs.texCoord, subMesh.vbo, offset, stride)
		r.vao.SetIndexBuffer(subMesh.ibo)
		r.prog.SetUniformSampler(r.uniforms.ambientMap, subMesh.mtl.ambientMapTexture)
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

type SkyboxRenderer struct {
	win *Window
	prog *ShaderProgram
	uniforms struct {
		viewMat *UniformMatrix4
		projMat *UniformMatrix4
		cubeMap *UniformSampler
	}
	attrs struct {
		pos *Attrib
	}
	vbo *Buffer
	vao *VertexArray
	tex *CubeMap
}

func NewSkyboxRenderer() *SkyboxRenderer {
	var r SkyboxRenderer

	var err error
	r.prog, err = ReadShaderProgram("shaders/skyboxvshader.glsl", "shaders/skyboxfshader.glsl")
	if err != nil {
		panic(err)
	}
	r.uniforms.viewMat, _ = r.prog.UniformMatrix4("viewMatrix")
	r.uniforms.projMat , _= r.prog.UniformMatrix4("projectionMatrix")
	r.uniforms.cubeMap , _= r.prog.UniformSampler("cubeMap")
	r.attrs.pos, _ = r.prog.Attrib("positionV")

	dir := "images/skybox/mountain/"
	names := []string{"posx.jpg", "negx.jpg", "posy.jpg", "negy.jpg", "posz.jpg", "negz.jpg"}
	var filenames [6]string
	for i, name := range names {
		filenames[i] = path.Join(dir, name)
	}
	r.tex = ReadCubeMap(gl.NEAREST, filenames[0], filenames[1], filenames[2], filenames[3], filenames[4], filenames[5])

    verts := []Vec3{
        NewVec3(-1.0, +1.0, -1.0),
        NewVec3(-1.0, -1.0, -1.0),
        NewVec3(+1.0, -1.0, -1.0),
        NewVec3(+1.0, -1.0, -1.0),
        NewVec3(+1.0, +1.0, -1.0),
        NewVec3(-1.0, +1.0, -1.0),

        NewVec3(-1.0, -1.0, +1.0),
        NewVec3(-1.0, -1.0, -1.0),
        NewVec3(-1.0, +1.0, -1.0),
        NewVec3(-1.0, +1.0, -1.0),
        NewVec3(-1.0, +1.0, +1.0),
        NewVec3(-1.0, -1.0, +1.0),

        NewVec3(+1.0, -1.0, -1.0),
        NewVec3(+1.0, -1.0, +1.0),
        NewVec3(+1.0, +1.0, +1.0),
        NewVec3(+1.0, +1.0, +1.0),
        NewVec3(+1.0, +1.0, -1.0),
        NewVec3(+1.0, -1.0, -1.0),

        NewVec3(-1.0, -1.0, +1.0),
        NewVec3(-1.0, +1.0, +1.0),
        NewVec3(+1.0, +1.0, +1.0),
        NewVec3(+1.0, +1.0, +1.0),
        NewVec3(+1.0, -1.0, +1.0),
        NewVec3(-1.0, -1.0, +1.0),

        NewVec3(-1.0, +1.0, -1.0),
        NewVec3(+1.0, +1.0, -1.0),
        NewVec3(+1.0, +1.0, +1.0),
        NewVec3(+1.0, +1.0, +1.0),
        NewVec3(-1.0, +1.0, +1.0),
        NewVec3(-1.0, +1.0, -1.0),

        NewVec3(-1.0, -1.0, -1.0),
        NewVec3(-1.0, -1.0, +1.0),
        NewVec3(+1.0, -1.0, -1.0),
        NewVec3(+1.0, -1.0, -1.0),
        NewVec3(-1.0, -1.0, +1.0),
        NewVec3(+1.0, -1.0, +1.0),
	}

	r.vbo = NewBuffer()
	r.vbo.SetData(verts, 0)

	r.vao = NewVertexArray()
	r.vao.SetAttribFormat(r.attrs.pos, 3, gl.FLOAT, false)
	r.vao.SetAttribSource(r.attrs.pos, r.vbo, 0, int(unsafe.Sizeof(NewVec3(0, 0, 0))))

	return &r
}

func (r *SkyboxRenderer) Render(c *Camera) {
	gl.Disable(gl.DEPTH_TEST)
	r.prog.SetUniformMatrix4(r.uniforms.viewMat, c.ViewMatrix())
	r.prog.SetUniformMatrix4(r.uniforms.projMat, c.ProjectionMatrix())
	r.prog.SetUniformSamplerCube(r.uniforms.cubeMap, r.tex)
	gls.SetShaderProgram(r.prog)
	gls.SetVertexArray(r.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, 36)
}
