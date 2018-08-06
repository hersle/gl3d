package main

import (
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/window"
	"github.com/go-gl/gl/v4.5-core/gl"
	"path"
	"golang.org/x/image/font/basicfont"
	"unsafe"
)

type Vertex struct {
	pos math.Vec3
	texCoord math.Vec2
	normal math.Vec3
	tangent math.Vec3
}

var shadowCubeMap *CubeMap = nil

// TODO: redesign attr/uniform access system?
type SceneRenderer struct {
	sp *MeshShaderProgram
	dsp *DepthPassShaderProgram
	vbo, ibo *Buffer
	normalMat math.Mat4

	shadowFb *Framebuffer

	renderState *RenderState
	depthRenderState *RenderState

	shadowMapRenderer *ShadowMapRenderer
}

func NewVertex(pos math.Vec3, texCoord math.Vec2, normal, tangent math.Vec3) Vertex {
	var vert Vertex
	vert.pos = pos
	vert.texCoord = texCoord
	vert.normal = normal
	vert.tangent = tangent
	return vert
}

func (_ *Vertex) Size() int {
	return int(unsafe.Sizeof(Vertex{}))
}

func (_ *Vertex) PositionOffset() int {
	return int(unsafe.Offsetof(Vertex{}.pos))
}

func (_ *Vertex) NormalOffset() int {
	return int(unsafe.Offsetof(Vertex{}.normal))
}

func (_ *Vertex) TexCoordOffset() int {
	return int(unsafe.Offsetof(Vertex{}.texCoord))
}

func (_ *Vertex) TangentOffset() int {
	return int(unsafe.Offsetof(Vertex{}.tangent))
}

func NewSceneRenderer() (*SceneRenderer, error) {
	var r SceneRenderer

	r.sp = NewMeshShaderProgram()

	r.dsp = NewDepthPassShaderProgram()


	r.shadowFb = NewFramebuffer()

	r.renderState = NewRenderState()
	r.renderState.SetShaderProgram(r.sp.ShaderProgram)
	r.renderState.SetFramebuffer(defaultFramebuffer)
	r.renderState.SetDepthTest(true)
	r.renderState.SetDepthFunc(gl.LEQUAL) // enable drawing after depth prepass
	r.renderState.SetBlend(true)
	r.renderState.SetBlendFunction(gl.ONE, gl.ONE) // add to framebuffer contents
	r.renderState.SetCull(true)
	r.renderState.SetCullFace(gl.BACK) // CCW treated as front face by default
	r.renderState.SetPolygonMode(gl.FILL)

	r.depthRenderState = NewRenderState()
	r.depthRenderState.SetShaderProgram(r.dsp.ShaderProgram)
	r.depthRenderState.SetFramebuffer(defaultFramebuffer)
	r.depthRenderState.SetDepthTest(true)
	r.depthRenderState.SetDepthFunc(gl.LESS) // enable drawing after depth prepass
	r.depthRenderState.SetBlend(false)
	r.depthRenderState.SetCull(true)
	r.depthRenderState.SetCullFace(gl.BACK) // CCW treated as front face by default
	r.depthRenderState.SetPolygonMode(gl.FILL)

	r.shadowMapRenderer = NewShadowMapRenderer()

	return &r, nil
}

func (r *SceneRenderer) renderMesh(m *Mesh, c *Camera) {
	r.sp.SetMesh(m)
	r.sp.SetCamera(c)

	for _, subMesh := range m.subMeshes {
		r.sp.SetSubMesh(subMesh)
		NewRenderCommand(gl.TRIANGLES, subMesh.inds, 0, r.renderState).Execute()
	}
}

func (r *SceneRenderer) shadowPassPointLight(s *Scene, l *PointLight) {
	r.shadowMapRenderer.RenderPointLightShadowMap(s, l)
}

func (r *SceneRenderer) shadowPassSpotLight(s *Scene, l *SpotLight) {
	r.shadowMapRenderer.RenderSpotLightShadowMap(s, l)
}

func (r *SceneRenderer) DepthPass(s *Scene, c *Camera) {
	r.dsp.SetCamera(c)
	for _, m := range s.meshes {
		r.dsp.SetMesh(m)
		for _, subMesh := range m.subMeshes {
			r.dsp.SetSubMesh(subMesh)
			NewRenderCommand(gl.TRIANGLES, subMesh.inds, 0, r.depthRenderState).Execute()
		}
	}
}

func (r *SceneRenderer) AmbientPass(s *Scene, c *Camera) {
	r.sp.SetAmbientLight(s.ambientLight)
	for _, m := range s.meshes {
		r.renderMesh(m, c)
	}
}

func (r *SceneRenderer) PointLightPass(s *Scene, c *Camera) {
	for _, l := range s.pointLights {
		r.shadowPassPointLight(s, l)

		r.sp.SetPointLight(l)

		for _, m := range s.meshes {
			r.renderMesh(m, c)
		}
	}
}

func (r *SceneRenderer) SpotLightPass(s *Scene, c *Camera) {
	for _, l := range s.spotLights {
		r.shadowPassSpotLight(s, l)

		r.sp.SetSpotLight(l)

		for _, m := range s.meshes {
			r.renderMesh(m, c)
		}
	}
}

func (r *SceneRenderer) Render(s *Scene, c *Camera) {
	r.renderState.viewportWidth, r.renderState.viewportHeight = window.Size()

	//r.DepthPass(s, c) // use ambient pass as depth pass too

	r.renderState.SetBlend(false) // replace framebuffer contents
	r.AmbientPass(s, c)
	r.renderState.SetBlend(true) // add to framebuffer contents
	r.PointLightPass(s, c)
	r.SpotLightPass(s, c)
}

func (r *SceneRenderer) SetWireframe(wireframe bool) {
	if wireframe {
		r.renderState.polygonMode = gl.LINE
	} else {
		r.renderState.polygonMode = gl.FILL
	}
}

type SkyboxRenderer struct {
	sp *SkyboxShaderProgram
	vbo *Buffer
	ibo *Buffer
	tex *CubeMap
	renderState *RenderState
}

func NewSkyboxRenderer() *SkyboxRenderer {
	var r SkyboxRenderer


	r.sp = NewSkyboxShaderProgram()

	dir := "images/skybox/mountain/"
	names := []string{"posx.jpg", "negx.jpg", "posy.jpg", "negy.jpg", "posz.jpg", "negz.jpg"}
	var filenames [6]string
	for i, name := range names {
		filenames[i] = path.Join(dir, name)
	}
	r.tex = ReadCubeMap(gl.NEAREST, filenames[0], filenames[1], filenames[2], filenames[3], filenames[4], filenames[5])

	r.vbo = NewBuffer()
	verts := []math.Vec3{
		math.NewVec3(-1.0, -1.0, -1.0),
		math.NewVec3(+1.0, -1.0, -1.0),
		math.NewVec3(+1.0, +1.0, -1.0),
		math.NewVec3(-1.0, +1.0, -1.0),
		math.NewVec3(-1.0, -1.0, +1.0),
		math.NewVec3(+1.0, -1.0, +1.0),
		math.NewVec3(+1.0, +1.0, +1.0),
		math.NewVec3(-1.0, +1.0, +1.0),
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
	r.renderState.viewportWidth, r.renderState.viewportHeight = window.Size()
	r.sp.SetCamera(c)
	r.sp.SetSkybox(r.tex)

	NewRenderCommand(gl.TRIANGLES, 36, 0, r.renderState).Execute()
}

type TextRenderer struct {
	sp *TextShaderProgram
	tex *Texture2D
	vbo *Buffer
	ibo *Buffer
	renderState *RenderState
}

func NewTextRenderer() *TextRenderer {
	var r TextRenderer

	r.sp = NewTextShaderProgram()

	r.vbo = NewBuffer()
	r.ibo = NewBuffer()

	r.sp.SetAttribs(r.vbo, r.ibo)

	img := basicfont.Face7x13.Mask
	r.tex = NewTexture2DFromImage(gl.NEAREST, gl.CLAMP_TO_EDGE, gl.RGBA8, img)

	r.renderState = NewRenderState()
	r.renderState.SetDepthTest(false)
	r.renderState.SetFramebuffer(defaultFramebuffer)
	r.renderState.SetShaderProgram(r.sp.ShaderProgram)
	r.renderState.SetBlend(true)
	r.renderState.SetBlendFunction(gl.ONE_MINUS_DST_COLOR, gl.ONE_MINUS_SRC_COLOR)
	r.renderState.SetCull(false)
	r.renderState.SetPolygonMode(gl.FILL)

	return &r
}

func (r *TextRenderer) Render(tl math.Vec2, text string, height float32) {
	r.renderState.viewportWidth, r.renderState.viewportHeight = window.Size()

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
				br := math.NewVec2(tl.X() + w, tl.Y() - h)
				tr := math.NewVec2(br.X(), tl.Y())
				bl := math.NewVec2(tl.X(), br.Y())

				normal := math.NewVec3(0, 0, 0)
				vert1 := NewVertex(bl.Vec3(0), math.NewVec2(texX1, texY2), normal, math.Vec3{})
				vert2 := NewVertex(br.Vec3(0), math.NewVec2(texX2, texY2), normal, math.Vec3{})
				vert3 := NewVertex(tr.Vec3(0), math.NewVec2(texX2, texY1), normal, math.Vec3{})
				vert4 := NewVertex(tl.Vec3(0), math.NewVec2(texX1, texY1), normal, math.Vec3{})
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
			tl = math.NewVec2(x0, tl.Y() - h)
		} else if char == '\t' {
			tl = tl.Add(math.NewVec2(4 * float32(face.Advance) * h / float32(subImgH), 0))
		} else {
			tl = tl.Add(math.NewVec2(float32(face.Advance) * h / float32(subImgH), 0))
		}
	}

	r.sp.SetAtlas(r.tex)
	r.vbo.SetData(verts, 0)
	r.ibo.SetData(inds, 0)
	NewRenderCommand(gl.TRIANGLES, len(inds), 0, r.renderState).Execute()
}

type ShadowMapRenderer struct {
	sp *ShadowMapShaderProgram
	framebuffer *Framebuffer
	renderState *RenderState
}

func NewShadowMapRenderer() *ShadowMapRenderer {
	var r ShadowMapRenderer

	r.sp = NewShadowMapShaderProgram()

	r.framebuffer = NewFramebuffer()

	r.renderState = NewRenderState()
	r.renderState.SetShaderProgram(r.sp.ShaderProgram)
	r.renderState.SetFramebuffer(r.framebuffer)
	r.renderState.SetDepthTest(true)
	r.renderState.SetBlend(false)

	return &r
}

// render shadow map to l's shadow map
func (r *ShadowMapRenderer) RenderPointLightShadowMap(s *Scene, l *PointLight) {
	// TODO: re-render also when objects have moved
	if !l.dirtyShadowMap {
		return
	}

	forwards := []math.Vec3{
		math.NewVec3(+1, 0, 0),
		math.NewVec3(-1, 0, 0),
		math.NewVec3(0, +1, 0),
		math.NewVec3(0, -1, 0),
		math.NewVec3(0, 0, +1),
		math.NewVec3(0, 0, -1),
	}
	ups := []math.Vec3{
		math.NewVec3(0, -1, 0),
		math.NewVec3(0, -1, 0),
		math.NewVec3(0, 0, +1),
		math.NewVec3(0, 0, -1),
		math.NewVec3(0, -1, 0),
		math.NewVec3(0, -1, 0),
	}

	c := NewCamera(90, 1, 0.1, l.shadowFar)
	c.Place(l.position)

	r.renderState.SetViewport(l.shadowMap.width, l.shadowMap.height)

	// UNCOMMENT THIS LINE AND ANOTHER ONE TO DRAW SHADOW CUBE MAP AS SKYBOX
	//shadowCubeMap = l.shadowMap

	for face := 0; face < 6; face++ {
		r.framebuffer.SetTextureCubeMapFace(gl.DEPTH_ATTACHMENT, l.shadowMap, 0, int32(face))
		r.framebuffer.ClearDepth(1)
		c.SetForwardUp(forwards[face], ups[face])

		r.sp.SetCamera(c)

		for _, m := range s.meshes {
			r.sp.SetMesh(m)
			for _, subMesh := range m.subMeshes {
				r.sp.SetSubMesh(subMesh)

				NewRenderCommand(gl.TRIANGLES, subMesh.inds, 0, r.renderState).Execute()
			}
		}
	}

	l.dirtyShadowMap = false
}

func (r *ShadowMapRenderer) RenderSpotLightShadowMap(s *Scene, l *SpotLight) {
	// TODO: re-render also when objects have moved
	if !l.dirtyShadowMap {
		return
	}

	r.framebuffer.SetTexture2D(gl.DEPTH_ATTACHMENT, l.shadowMap, 0)
	r.framebuffer.ClearDepth(1)
	r.renderState.SetViewport(l.shadowMap.width, l.shadowMap.height)
	r.sp.SetCamera(&l.Camera)

	for _, m := range s.meshes {
		r.sp.SetMesh(m)
		for _, subMesh := range m.subMeshes {
			r.sp.SetSubMesh(subMesh)

			NewRenderCommand(gl.TRIANGLES, subMesh.inds, 0, r.renderState).Execute()
		}
	}

	l.dirtyShadowMap = false
}

type ArrowRenderer struct {
	sp *ArrowShaderProgram
	points []math.Vec3
	vbo *Buffer
	renderState *RenderState
}

func NewArrowRenderer() *ArrowRenderer {
	var r ArrowRenderer

	r.sp = NewArrowShaderProgram()

	r.renderState = NewRenderState()
	r.renderState.SetBlend(false)
	r.renderState.SetCull(false)
	r.renderState.SetDepthTest(true)
	r.renderState.SetFramebuffer(defaultFramebuffer)
	r.renderState.SetShaderProgram(r.sp.ShaderProgram)

	r.vbo = NewBuffer()
	r.sp.SetPosition(r.vbo)

	return &r
}

func (r *ArrowRenderer) RenderTangents(s *Scene, c *Camera) {
	r.renderState.viewportWidth, r.renderState.viewportHeight = window.Size()
	r.sp.SetCamera(c)
	r.points = r.points[:0]
	r.sp.SetColor(math.NewVec3(1, 0, 0))
	for _, m := range s.meshes {
		r.sp.SetMesh(m)
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
	r.renderState.viewportWidth, r.renderState.viewportHeight = window.Size()
	r.sp.SetCamera(c)
	r.points = r.points[:0]
	r.sp.SetColor(math.NewVec3(0, 1, 0))
	for _, m := range s.meshes {
		r.sp.SetMesh(m)
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
	r.renderState.viewportWidth, r.renderState.viewportHeight = window.Size()
	r.sp.SetCamera(c)
	r.points = r.points[:0]
	r.sp.SetColor(math.NewVec3(0, 0, 1))
	for _, m := range s.meshes {
		r.sp.SetMesh(m)
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
