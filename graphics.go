package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"unsafe"
	"path"
	"golang.org/x/image/font/basicfont"
)

type Vertex struct {
	pos Vec3
	texCoord Vec2
	normal Vec3
	tangent Vec3
}

var shadowCubeMap *CubeMap = nil

// TODO: redesign attr/uniform access system?
type MeshRenderer struct {
	win *Window
	sp *MeshShaderProgram
	vbo, ibo *Buffer
	normalMat Mat4

	shadowFb *Framebuffer

	renderState *RenderState

	shadowMapRenderer *ShadowMapRenderer
}

func NewVertex(pos Vec3, texCoord Vec2, normal, tangent Vec3) Vertex {
	var vert Vertex
	vert.pos = pos
	vert.texCoord = texCoord
	vert.normal = normal
	vert.tangent = tangent
	return vert
}

func NewMeshRenderer(win *Window) (*MeshRenderer, error) {
	var r MeshRenderer

	win.MakeContextCurrent()

	err := gl.Init()
	if err != nil {
		return nil, err
	}

	r.sp = NewMeshShaderProgram()
	if err != nil {
		return nil, err
	}

	r.win = win

	r.shadowFb = NewFramebuffer()

	r.renderState = NewRenderState()
	r.renderState.SetShaderProgram(r.sp.ShaderProgram)
	r.renderState.SetFramebuffer(defaultFramebuffer)
	r.renderState.SetDepthTest(true)
	r.renderState.SetDepthFunc(gl.LEQUAL) // enable drawing after depth prepass
	r.renderState.SetBlend(true)
	r.renderState.SetCull(true)
	r.renderState.SetCullFace(gl.BACK) // CCW treated as front face by default
	r.renderState.SetPolygonMode(gl.FILL)

	r.shadowMapRenderer = NewShadowMapRenderer()

	return &r, nil
}

func (r *MeshRenderer) Clear() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

var enableBumpMap bool
func (r *MeshRenderer) renderMesh(m *Mesh, c *Camera) {
	r.sp.SetMesh(m)
	r.sp.SetCamera(c)

	for _, subMesh := range m.subMeshes {
		r.sp.SetSubMesh(subMesh)
		NewRenderCommand(gl.TRIANGLES, subMesh.inds, 0, r.renderState).Execute()
	}
}

func (r *MeshRenderer) shadowPassPointLight(s *Scene, l *PointLight) {
	r.shadowMapRenderer.RenderPointLightShadowMap(s, l)
}

func (r *MeshRenderer) shadowPassSpotLight(s *Scene, l *SpotLight) {
	r.shadowMapRenderer.RenderSpotLightShadowMap(s, l)
}

func (r *MeshRenderer) DepthPass(s *Scene, c *Camera) {
	// TODO: improve
	gl.Clear(gl.DEPTH_BUFFER_BIT)
	r.sp.SetCamera(c)
	for _, m := range s.meshes {
		r.sp.SetMesh(m)
		for _, subMesh := range m.subMeshes {
			r.sp.SetSubMesh(subMesh)
			NewRenderCommand(gl.TRIANGLES, subMesh.inds, 0, r.renderState).Execute()
		}
	}
}

func (r *MeshRenderer) AmbientPass(s *Scene, c *Camera) {
	r.sp.SetAmbientLight(s.ambientLight)
	for _, m := range s.meshes {
		r.renderMesh(m, c)
	}
}

func (r *MeshRenderer) PointLightPass(s *Scene, c *Camera) {
	for _, l := range s.pointLights {
		r.shadowPassPointLight(s, l)

		r.sp.SetPointLight(l)

		for _, m := range s.meshes {
			r.renderMesh(m, c)
		}
	}
}

func (r *MeshRenderer) SpotLightPass(s *Scene, c *Camera) {
	for _, l := range s.spotLights {
		r.shadowPassSpotLight(s, l)

		r.sp.SetSpotLight(l)

		for _, m := range s.meshes {
			r.renderMesh(m, c)
		}
	}
}

func (r *MeshRenderer) Render(s *Scene, c *Camera) {
	r.renderState.viewportWidth, r.renderState.viewportHeight = r.win.Size()

	r.DepthPass(s, c)

	r.renderState.SetBlendFunction(gl.ONE, gl.ZERO) // replace framebuffer contents
	r.AmbientPass(s, c)

	r.renderState.SetBlendFunction(gl.ONE, gl.ONE) // add to framebuffer contents
	r.PointLightPass(s, c)
	r.SpotLightPass(s, c)

	// UNCOMMENT THESE LINES TO DRAW SPOT LIGHT DEPTH MAP FOR DEBUGGING
	/*
	s.quad.subMeshes[0].mtl.ambientMap = s.spotLights[0].shadowMap
	ident := NewMat4Identity()
	r.uniforms.modelMat.Set(ident)
	r.uniforms.viewMat.Set(ident)
	r.uniforms.projMat.Set(ident)
	for _, subMesh := range s.quad.subMeshes {
		r.uniforms.ambient.Set(subMesh.mtl.ambient)
		r.uniforms.diffuse.Set(NewVec3(0, 0, 0))
		r.uniforms.specular.Set(NewVec3(0, 0, 0))
		r.uniforms.shine.Set(0)
		r.uniforms.alpha.Set(1)

		stride := int(unsafe.Sizeof(Vertex{}))
		offset1 := int(unsafe.Offsetof(Vertex{}.pos))
		offset3 := int(unsafe.Offsetof(Vertex{}.texCoord))
		r.attrs.pos.SetSource(subMesh.vbo, offset1, stride)
		r.attrs.texCoord.SetSource(subMesh.vbo, offset3, stride)
		r.prog.SetAttribIndexBuffer(subMesh.ibo)

		r.uniforms.ambientMap.Set2D(subMesh.mtl.ambientMap)

		NewRenderCommand(gl.TRIANGLES, subMesh.inds, 0, r.renderState).Execute()
	}
	*/
}

func (r *MeshRenderer) SetWireframe(wireframe bool) {
	if wireframe {
		r.renderState.polygonMode = gl.LINE
	} else {
		r.renderState.polygonMode = gl.FILL
	}
}

type SkyboxRenderer struct {
	win *Window
	sp *SkyboxShaderProgram
	vbo *Buffer
	ibo *Buffer
	tex *CubeMap
	renderState *RenderState
}

func NewSkyboxRenderer(win *Window) *SkyboxRenderer {
	var r SkyboxRenderer

	r.win = win

	r.sp = NewSkyboxShaderProgram()

	dir := "images/skybox/mountain/"
	names := []string{"posx.jpg", "negx.jpg", "posy.jpg", "negy.jpg", "posz.jpg", "negz.jpg"}
	var filenames [6]string
	for i, name := range names {
		filenames[i] = path.Join(dir, name)
	}
	r.tex = ReadCubeMap(gl.NEAREST, filenames[0], filenames[1], filenames[2], filenames[3], filenames[4], filenames[5])

	r.vbo = NewBuffer()
	verts := []Vec3{
		NewVec3(-1.0, -1.0, -1.0),
		NewVec3(+1.0, -1.0, -1.0),
		NewVec3(+1.0, +1.0, -1.0),
		NewVec3(-1.0, +1.0, -1.0),
		NewVec3(-1.0, -1.0, +1.0),
		NewVec3(+1.0, -1.0, +1.0),
		NewVec3(+1.0, +1.0, +1.0),
		NewVec3(-1.0, +1.0, +1.0),
	}
	r.vbo.SetData(verts, 0)

	r.ibo = NewBuffer()
	inds := []int32{
		4, 5, 6, 4, 6, 7,
		5, 1, 2, 5, 2, 6,
		1, 0, 3, 1, 3, 2,
		0, 4, 7, 0, 7, 3,
		7, 6, 2, 7, 2, 3,
		5, 4, 0, 5, 0, 1,
	}
	r.ibo.SetData(inds, 0)

	r.sp.SetCube(r.vbo, r.ibo)

	r.renderState = NewRenderState()
	r.renderState.SetDepthTest(false)
	r.renderState.SetFramebuffer(defaultFramebuffer)
	r.renderState.SetShaderProgram(r.sp.ShaderProgram)
	r.renderState.SetCull(false)
	r.renderState.SetPolygonMode(gl.FILL)

	return &r
}

func (r *SkyboxRenderer) Render(c *Camera) {
	r.renderState.viewportWidth, r.renderState.viewportHeight = r.win.Size()
	r.sp.SetCamera(c)
	r.sp.SetSkybox(r.tex)

	NewRenderCommand(gl.TRIANGLES, 36, 0, r.renderState).Execute()
}

type TextRenderer struct {
	win *Window
	prog *ShaderProgram
	uniforms struct {
		tex *UniformSampler
	}
	attrs struct {
		pos *Attrib
		texCoord *Attrib
	}
	tex *Texture2D
	vbo *Buffer
	ibo *Buffer
	renderState *RenderState
}

func NewTextRenderer(win *Window) *TextRenderer {
	var r TextRenderer

	r.win = win

	r.prog, _ = ReadShaderProgram("shaders/textvshader.glsl", "shaders/textfshader.glsl")
	r.uniforms.tex = r.prog.UniformSampler("fontAtlas")
	r.attrs.pos = r.prog.Attrib("position")
	r.attrs.texCoord = r.prog.Attrib("texCoordV")

	r.vbo = NewBuffer()
	r.ibo = NewBuffer()

	stride := int(unsafe.Sizeof(Vertex{}))
	offset1 := int(unsafe.Offsetof(Vertex{}.pos))
	offset2 := int(unsafe.Offsetof(Vertex{}.texCoord))
	r.attrs.pos.SetFormat(gl.FLOAT, false)
	r.attrs.pos.SetSource(r.vbo, offset1, stride)
	r.attrs.texCoord.SetFormat(gl.FLOAT, false)
	r.attrs.texCoord.SetSource(r.vbo, offset2, stride)
	r.prog.SetAttribIndexBuffer(r.ibo)

	img := basicfont.Face7x13.Mask
	r.tex = NewTexture2DFromImage(gl.NEAREST, gl.CLAMP_TO_EDGE, gl.RGBA8, img)

	r.renderState = NewRenderState()
	r.renderState.SetDepthTest(false)
	r.renderState.SetFramebuffer(defaultFramebuffer)
	r.renderState.SetShaderProgram(r.prog)
	r.renderState.SetBlend(true)
	r.renderState.SetBlendFunction(gl.ONE_MINUS_DST_COLOR, gl.ONE_MINUS_SRC_COLOR)
	r.renderState.SetCull(false)
	r.renderState.SetPolygonMode(gl.FILL)

	return &r
}

func (r *TextRenderer) Render(tl Vec2, text string, height float32) {
	r.renderState.viewportWidth, r.renderState.viewportHeight = r.win.Size()

	var verts []Vertex
	var inds []int32

	face := basicfont.Face7x13

	x0 := tl.X()
	imgW, imgH := face.Mask.Bounds().Dx(), face.Mask.Bounds().Dy()
	subImgW, subImgH := face.Width, face.Ascent + face.Descent
	h := height
	w := h * float32(subImgW) / float32(subImgH)

	for _, char := range text {
		for _, runeRange := range face.Ranges {
			lo, hi, offset := runeRange.Low, runeRange.High, runeRange.Offset
			if char >= lo && char < hi {
				imgX1, imgY1 := 0, imgH - (int(char-lo) + offset) * subImgH
				imgX2, imgY2 := imgX1 + subImgW, imgY1 - subImgH
				texX1 := float32(imgX1) / float32(imgW) // left
				texY1 := float32(imgY1) / float32(imgH) // top
				texX2 := float32(imgX2) / float32(imgW) // right
				texY2 := float32(imgY2) / float32(imgH) // bottom
				br := NewVec2(tl.X() + w, tl.Y() - h)
				tr := NewVec2(br.X(), tl.Y())
				bl := NewVec2(tl.X(), br.Y())

				normal := NewVec3(0, 0, 0)
				vert1 := NewVertex(bl.Vec3(0), NewVec2(texX1, texY2), normal, Vec3{})
				vert2 := NewVertex(br.Vec3(0), NewVec2(texX2, texY2), normal, Vec3{})
				vert3 := NewVertex(tr.Vec3(0), NewVec2(texX2, texY1), normal, Vec3{})
				vert4 := NewVertex(tl.Vec3(0), NewVec2(texX1, texY1), normal, Vec3{})
				inds = append(inds, int32(len(verts) + 0))
				inds = append(inds, int32(len(verts) + 1))
				inds = append(inds, int32(len(verts) + 2))
				inds = append(inds, int32(len(verts) + 0))
				inds = append(inds, int32(len(verts) + 2))
				inds = append(inds, int32(len(verts) + 3))
				verts = append(verts, vert1, vert2, vert3, vert4)
				break
			}
		}

		if char == '\n' {
			tl = NewVec2(x0, tl.Y() - h)
		} else if char == '\t' {
			tl = tl.Add(NewVec2(4 * float32(face.Advance) * h / float32(subImgH), 0))
		} else {
			tl = tl.Add(NewVec2(float32(face.Advance) * h / float32(subImgH), 0))
		}
	}

	r.uniforms.tex.Set2D(r.tex)
	r.vbo.SetData(verts, 0)
	r.ibo.SetData(inds, 0)
	NewRenderCommand(gl.TRIANGLES, len(inds), 0, r.renderState).Execute()
}

type ShadowMapRenderer struct {
	prog *ShaderProgram
	uniforms struct {
		modelMat *UniformMatrix4
		viewMat *UniformMatrix4
		projMat *UniformMatrix4
		lightPos *UniformVector3
		far *UniformFloat
	}
	attrs struct {
		pos *Attrib
	}
	framebuffer *Framebuffer
	renderState *RenderState
}

func NewShadowMapRenderer() *ShadowMapRenderer {
	var r ShadowMapRenderer
	var err error

	r.prog, err = ReadShaderProgram("shaders/pointlightshadowmapvshader.glsl", "shaders/pointlightshadowmapfshader.glsl")
	if err != nil {
		panic(err)
	}

	r.uniforms.modelMat = r.prog.UniformMatrix4("modelMatrix")
	r.uniforms.viewMat = r.prog.UniformMatrix4("viewMatrix")
	r.uniforms.projMat = r.prog.UniformMatrix4("projectionMatrix")
	r.uniforms.lightPos = r.prog.UniformVector3("lightPosition")
	r.uniforms.far = r.prog.UniformFloat("far")

	r.attrs.pos = r.prog.Attrib("position")

	r.framebuffer = NewFramebuffer()

	r.renderState = NewRenderState()
	r.renderState.SetShaderProgram(r.prog)
	r.renderState.SetFramebuffer(r.framebuffer)
	r.renderState.SetDepthTest(true)
	r.renderState.SetBlend(false)
	r.renderState.SetViewport(512, 512) // TODO: respect actual shadow map size

	return &r
}

// render shadow map to l's shadow map
func (r *ShadowMapRenderer) RenderPointLightShadowMap(s *Scene, l *PointLight) {
	forwards := []Vec3{
		NewVec3(+1, 0, 0),
		NewVec3(-1, 0, 0),
		NewVec3(0, +1, 0),
		NewVec3(0, -1, 0),
		NewVec3(0, 0, +1),
		NewVec3(0, 0, -1),
	}
	ups := []Vec3{
		NewVec3(0, -1, 0),
		NewVec3(0, -1, 0),
		NewVec3(0, 0, +1), // TODO: ?
		NewVec3(0, 0, -1), // TODO: ?
		NewVec3(0, -1, 0),
		NewVec3(0, -1, 0),
	}

	c := NewCamera(90, 1, 0.1, 50)
	c.Place(l.position)
	r.uniforms.far.Set(c.far)
	r.uniforms.projMat.Set(c.ProjectionMatrix())
	r.uniforms.lightPos.Set(l.position)

	// UNCOMMENT THIS LINE AND ANOTHER ONE TO DRAW SHADOW CUBE MAP AS SKYBOX
	//shadowCubeMap = l.shadowMap

	for face := 0; face < 6; face++ {
		r.framebuffer.SetTextureCubeMapFace(gl.DEPTH_ATTACHMENT, l.shadowMap, 0, int32(face))
		r.framebuffer.ClearDepth(1)
		c.SetForwardUp(forwards[face], ups[face])
		r.uniforms.viewMat.Set(c.ViewMatrix())

		for _, m := range s.meshes {
			r.uniforms.modelMat.Set(m.WorldMatrix())
			for _, subMesh := range m.subMeshes {
				stride := int(unsafe.Sizeof(Vertex{}))
				offset := int(unsafe.Offsetof(Vertex{}.pos))
				r.attrs.pos.SetSource(subMesh.vbo, offset, stride)
				r.prog.SetAttribIndexBuffer(subMesh.ibo)

				NewRenderCommand(gl.TRIANGLES, subMesh.inds, 0, r.renderState).Execute()
			}
		}
	}
}

func (r *ShadowMapRenderer) RenderSpotLightShadowMap(s *Scene, l *SpotLight) {
	r.framebuffer.SetTexture2D(gl.DEPTH_ATTACHMENT, l.shadowMap, 0)
	r.framebuffer.ClearDepth(1)
	r.uniforms.viewMat.Set(l.ViewMatrix())
	r.uniforms.projMat.Set(l.ProjectionMatrix())
	r.uniforms.far.Set(l.Camera.far)
	r.uniforms.lightPos.Set(l.position)

	for _, m := range s.meshes {
		r.uniforms.modelMat.Set(m.WorldMatrix())
		for _, subMesh := range m.subMeshes {
			stride := int(unsafe.Sizeof(Vertex{}))
			offset := int(unsafe.Offsetof(Vertex{}.pos))
			r.attrs.pos.SetSource(subMesh.vbo, offset, stride)
			r.prog.SetAttribIndexBuffer(subMesh.ibo)

			NewRenderCommand(gl.TRIANGLES, subMesh.inds, 0, r.renderState).Execute()
		}
	}
}

type ArrowRenderer struct {
	win *Window
	prog *ShaderProgram
	uniforms struct {
		modelMat *UniformMatrix4
		viewMat *UniformMatrix4
		projMat *UniformMatrix4
		color *UniformVector3
	}
	attrs struct {
		pos *Attrib
	}
	points []Vec3
	vbo *Buffer
	renderState *RenderState
}

func NewArrowRenderer(win *Window) *ArrowRenderer {
	var r ArrowRenderer
	var err error

	r.win = win
	r.prog, err = ReadShaderProgram("shaders/arrowvshader.glsl", "shaders/arrowfshader.glsl")
	if err != nil {
		panic(err)
	}
	r.attrs.pos = r.prog.Attrib("position")
	r.attrs.pos.SetFormat(gl.FLOAT, false)
	r.uniforms.modelMat = r.prog.UniformMatrix4("modelMatrix")
	r.uniforms.viewMat = r.prog.UniformMatrix4("viewMatrix")
	r.uniforms.projMat = r.prog.UniformMatrix4("projectionMatrix")
	r.uniforms.color = r.prog.UniformVector3("color")

	r.renderState = NewRenderState()
	r.renderState.SetBlend(false)
	r.renderState.SetCull(false)
	r.renderState.SetDepthTest(true)
	r.renderState.SetFramebuffer(defaultFramebuffer)
	r.renderState.SetShaderProgram(r.prog)

	r.vbo = NewBuffer()

	return &r
}

func (r *ArrowRenderer) RenderTangents(s *Scene, c *Camera) {
	r.renderState.viewportWidth, r.renderState.viewportHeight = r.win.Size()
	r.uniforms.viewMat.Set(c.ViewMatrix())
	r.uniforms.projMat.Set(c.ProjectionMatrix())
	stride := int(unsafe.Sizeof(NewVec3(0, 0, 0)))
	r.attrs.pos.SetSource(r.vbo, 0, stride)
	r.points = r.points[:0]
	r.uniforms.color.Set(NewVec3(1, 0, 0))
	for _, m := range s.meshes {
		r.uniforms.modelMat.Set(m.WorldMatrix())
		for _, subMesh := range m.subMeshes {
			for _, i := range subMesh.faces {
				p1 := subMesh.verts[i].pos
				p2 := p1.Add(subMesh.verts[i].tangent)
				r.points = append(r.points, p1, p2)
			}
		}
	}
	r.vbo.SetData(r.points, 0)
	NewRenderCommand(gl.LINES, len(r.points), 0, r.renderState).Execute()
}

func (r *ArrowRenderer) RenderBitangents(s *Scene, c *Camera) {
	r.renderState.viewportWidth, r.renderState.viewportHeight = r.win.Size()
	r.uniforms.viewMat.Set(c.ViewMatrix())
	r.uniforms.projMat.Set(c.ProjectionMatrix())
	stride := int(unsafe.Sizeof(NewVec3(0, 0, 0)))
	r.attrs.pos.SetSource(r.vbo, 0, stride)
	r.points = r.points[:0]
	r.uniforms.color.Set(NewVec3(0, 1, 0))
	for _, m := range s.meshes {
		r.uniforms.modelMat.Set(m.WorldMatrix())
		for _, subMesh := range m.subMeshes {
			for _, i := range subMesh.faces {
				p1 := subMesh.verts[i].pos
				p2 := p1.Add(subMesh.verts[i].normal.Cross(subMesh.verts[i].tangent))
				r.points = append(r.points, p1, p2)
			}
		}
	}
	r.vbo.SetData(r.points, 0)
	NewRenderCommand(gl.LINES, len(r.points), 0, r.renderState).Execute()
}

func (r *ArrowRenderer) RenderNormals(s *Scene, c *Camera) {
	r.renderState.viewportWidth, r.renderState.viewportHeight = r.win.Size()
	r.uniforms.viewMat.Set(c.ViewMatrix())
	r.uniforms.projMat.Set(c.ProjectionMatrix())
	stride := int(unsafe.Sizeof(NewVec3(0, 0, 0)))
	r.attrs.pos.SetSource(r.vbo, 0, stride)
	r.points = r.points[:0]
	r.uniforms.color.Set(NewVec3(0, 0, 1))
	for _, m := range s.meshes {
		r.uniforms.modelMat.Set(m.WorldMatrix())
		for _, subMesh := range m.subMeshes {
			for _, i := range subMesh.faces {
				p1 := subMesh.verts[i].pos
				p2 := p1.Add(subMesh.verts[i].normal)
				r.points = append(r.points, p1, p2)
			}
		}
	}
	r.vbo.SetData(r.points, 0)
	NewRenderCommand(gl.LINES, len(r.points), 0, r.renderState).Execute()
}
