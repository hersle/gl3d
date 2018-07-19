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
	normalMat *Mat4

	shadowFb *Framebuffer
	shadowTex *Texture2D

	renderState1 *RenderState
	renderState2 *RenderState
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

	r.attrs.pos.SetFormat(3, gl.FLOAT, false)
	r.attrs.normal.SetFormat(3, gl.FLOAT, false)
	r.attrs.texCoord.SetFormat(2, gl.FLOAT, false)

	r.normalMat = NewMat4Zero()

	r.win = win

	r.shadowTex = NewTexture2D(gl.NEAREST, gl.CLAMP_TO_BORDER, gl.DEPTH_COMPONENT16, 512, 512)
	r.shadowTex.SetBorderColor(NewVec4(1, 1, 1, 1))

	r.shadowFb = NewFramebuffer()
	r.shadowFb.SetTexture(gl.DEPTH_ATTACHMENT, r.shadowTex, 0)
	println(r.shadowFb.Complete())

	r.renderState1 = NewRenderState()
	r.renderState1.SetVertexArray(r.prog.va) // TODO: remove 
	r.renderState1.SetShaderProgram(r.prog)
	r.renderState1.SetFramebuffer(defaultFramebuffer)
	r.renderState1.SetDepthTest(true)
	r.renderState1.SetBlend(true)
	r.renderState1.SetBlendFunction(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	r.renderState2 = NewRenderState()
	r.renderState2.SetVertexArray(r.prog.va) // TODO: remove 
	r.renderState2.SetShaderProgram(r.prog)
	r.renderState2.SetFramebuffer(r.shadowFb)
	r.renderState2.SetDepthTest(true)
	r.renderState2.SetBlend(false)
	r.renderState2.SetViewport(512, 512)

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
		r.attrs.pos.SetSource(subMesh.vbo, offset, stride)
		offset = int(unsafe.Offsetof(Vertex{}.normal))
		r.attrs.normal.SetSource(subMesh.vbo, offset, stride)
		offset = int(unsafe.Offsetof(Vertex{}.texCoord))
		r.attrs.texCoord.SetSource(subMesh.vbo, offset, stride)
		r.prog.SetAttribIndexBuffer(subMesh.ibo)

		r.uniforms.ambientMap.Set2D(subMesh.mtl.ambientMapTexture)
		r.uniforms.diffuseMap.Set2D(subMesh.mtl.diffuseMapTexture)
		r.uniforms.specularMap.Set2D(subMesh.mtl.specularMapTexture)
		r.uniforms.shadowMap.Set2D(r.shadowTex)
		NewRenderCommand(gl.TRIANGLES, subMesh.inds, 0, r.renderState1).Execute()
	}
}

func (r *Renderer) shadowPass(s *Scene, c *Camera) {
	r.shadowFb.ClearDepth(1)
	r.uniforms.viewMat.Set(s.Light.ViewMatrix())
	r.uniforms.projMat.Set(s.Light.ProjectionMatrix())

	for _, m := range s.meshes {
		r.uniforms.modelMat.Set(m.WorldMatrix())
		for _, subMesh := range m.subMeshes {
			stride := int(unsafe.Sizeof(Vertex{}))
			offset := int(unsafe.Offsetof(Vertex{}.pos))
			r.attrs.pos.SetSource(subMesh.vbo, offset, stride)
			r.prog.SetAttribIndexBuffer(subMesh.ibo)

			NewRenderCommand(gl.TRIANGLES, subMesh.inds, 0, r.renderState2).Execute()
		}
	}
}

func (r *Renderer) Render(s *Scene, c *Camera) {
	// shadow pass
	r.shadowPass(s, c)

	// normal pass
	r.renderState1.viewportWidth, r.renderState1.viewportHeight = r.win.Size()
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
		stride := int(unsafe.Sizeof(Vertex{}))
		offset := int(unsafe.Offsetof(Vertex{}.pos))
		r.attrs.pos.SetSource(subMesh.vbo, offset, stride)
		offset = int(unsafe.Offsetof(Vertex{}.texCoord))
		r.attrs.texCoord.SetSource(subMesh.vbo, offset, stride)
		r.prog.SetAttribIndexBuffer(subMesh.ibo)

		r.uniforms.ambientMap.Set2D(subMesh.mtl.ambientMapTexture)
		NewRenderCommand(gl.TRIANGLES, subMesh.inds, 0, r.renderState1).Execute()
	}
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
	tex *CubeMap
	renderState *RenderState
}

func NewSkyboxRenderer(win *Window) *SkyboxRenderer {
	var r SkyboxRenderer

	r.win = win

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

	r.attrs.pos.SetFormat(3, gl.FLOAT, false)
	r.attrs.pos.SetSource(r.vbo, 0, int(unsafe.Sizeof(NewVec3(0, 0, 0))))

	r.renderState = NewRenderState()
	r.renderState.SetDepthTest(false)
	r.renderState.SetFramebuffer(defaultFramebuffer)
	r.renderState.SetShaderProgram(r.prog)
	r.renderState.SetVertexArray(r.prog.va) // TODO: remove

	return &r
}

func (r *SkyboxRenderer) Render(c *Camera) {
	r.renderState.viewportWidth, r.renderState.viewportHeight = r.win.Size()
	r.uniforms.viewMat.Set(c.ViewMatrix())
	r.uniforms.projMat.Set(c.ProjectionMatrix())
	r.uniforms.cubeMap.SetCube(r.tex)
	NewRenderCommand(gl.TRIANGLES, 36, 0, r.renderState).Execute()
}
