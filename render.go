package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/window"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/scene"
	"golang.org/x/image/font/basicfont"
	"path"
	"unsafe"
)

var shadowCubeMap *graphics.CubeMap = nil

// TODO: redesign attr/uniform access system?
type SceneRenderer struct {
	sp        *graphics.MeshShaderProgram
	dsp       *graphics.DepthPassShaderProgram
	vbo, ibo  *graphics.Buffer
	normalMat math.Mat4

	shadowFb *graphics.Framebuffer

	renderState      *graphics.RenderState
	depthRenderState *graphics.RenderState

	shadowMapRenderer *ShadowMapRenderer
}

func NewSceneRenderer() (*SceneRenderer, error) {
	var r SceneRenderer

	r.sp = graphics.NewMeshShaderProgram()

	r.dsp = graphics.NewDepthPassShaderProgram()

	r.shadowFb = graphics.NewFramebuffer()

	r.renderState = graphics.NewRenderState()
	r.renderState.SetShaderProgram(r.sp.ShaderProgram)
	r.renderState.SetFramebuffer(graphics.DefaultFramebuffer)
	r.renderState.SetDepthTest(true)
	r.renderState.SetDepthFunc(gl.LEQUAL) // enable drawing after depth prepass
	r.renderState.SetBlend(true)
	r.renderState.SetBlendFunction(gl.ONE, gl.ONE) // add to framebuffer contents
	r.renderState.SetCull(true)
	r.renderState.SetCullFace(gl.BACK) // CCW treated as front face by default
	r.renderState.SetPolygonMode(gl.FILL)

	r.depthRenderState = graphics.NewRenderState()
	r.depthRenderState.SetShaderProgram(r.dsp.ShaderProgram)
	r.depthRenderState.SetFramebuffer(graphics.DefaultFramebuffer)
	r.depthRenderState.SetDepthTest(true)
	r.depthRenderState.SetDepthFunc(gl.LESS) // enable drawing after depth prepass
	r.depthRenderState.SetBlend(false)
	r.depthRenderState.SetCull(true)
	r.depthRenderState.SetCullFace(gl.BACK) // CCW treated as front face by default
	r.depthRenderState.SetPolygonMode(gl.FILL)

	r.shadowMapRenderer = NewShadowMapRenderer()

	return &r, nil
}

func (r *SceneRenderer) renderMesh(m *object.Mesh, c *camera.Camera) {
	r.SetMesh(m)
	r.SetCamera(c)

	for _, subMesh := range m.SubMeshes {
		r.SetSubMesh(subMesh)
		graphics.NewRenderCommand(gl.TRIANGLES, subMesh.Inds, 0, r.renderState).Execute()
	}
}

func (r *SceneRenderer) shadowPassPointLight(s *scene.Scene, l *light.PointLight) {
	r.shadowMapRenderer.RenderPointLightShadowMap(s, l)
}

func (r *SceneRenderer) shadowPassSpotLight(s *scene.Scene, l *light.SpotLight) {
	r.shadowMapRenderer.RenderSpotLightShadowMap(s, l)
}

func (r *SceneRenderer) DepthPass(s *scene.Scene, c *camera.Camera) {
	r.SetDepthCamera(c)
	for _, m := range s.Meshes {
		r.SetDepthMesh(m)
		for _, subMesh := range m.SubMeshes {
			r.SetDepthSubMesh(subMesh)
			graphics.NewRenderCommand(gl.TRIANGLES, subMesh.Inds, 0, r.depthRenderState).Execute()
		}
	}
}

func (r *SceneRenderer) AmbientPass(s *scene.Scene, c *camera.Camera) {
	r.SetAmbientLight(s.AmbientLight)
	for _, m := range s.Meshes {
		r.renderMesh(m, c)
	}
}

func (r *SceneRenderer) PointLightPass(s *scene.Scene, c *camera.Camera) {
	for _, l := range s.PointLights {
		r.shadowPassPointLight(s, l)

		r.SetPointLight(l)

		for _, m := range s.Meshes {
			r.renderMesh(m, c)
		}
	}
}

func (r *SceneRenderer) SpotLightPass(s *scene.Scene, c *camera.Camera) {
	for _, l := range s.SpotLights {
		r.shadowPassSpotLight(s, l)

		r.SetSpotLight(l)

		for _, m := range s.Meshes {
			r.renderMesh(m, c)
		}
	}
}

func (r *SceneRenderer) Render(s *scene.Scene, c *camera.Camera) {
	r.renderState.SetViewport(window.Size())

	//r.DepthPass(s, c) // use ambient pass as depth pass too

	r.renderState.SetBlend(false) // replace framebuffer contents
	r.AmbientPass(s, c)
	r.renderState.SetBlend(true) // add to framebuffer contents
	r.PointLightPass(s, c)
	r.SpotLightPass(s, c)
}

func (r *SceneRenderer) SetWireframe(wireframe bool) {
	if wireframe {
		r.renderState.SetPolygonMode(gl.LINE)
	} else {
		r.renderState.SetPolygonMode(gl.FILL)
	}
}

func (r *SceneRenderer) SetCamera(c *camera.Camera) {
	r.sp.ViewMatrix.Set(c.ViewMatrix())
	r.sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *SceneRenderer) SetMesh(m *object.Mesh) {
	r.sp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *SceneRenderer) SetSubMesh(sm *object.SubMesh) {
	mtl := sm.Mtl

	r.sp.Ambient.Set(mtl.Ambient)
	r.sp.AmbientMap.Set2D(mtl.AmbientMap)
	r.sp.Diffuse.Set(mtl.Diffuse)
	r.sp.DiffuseMap.Set2D(mtl.DiffuseMap)
	r.sp.Specular.Set(mtl.Specular)
	r.sp.SpecularMap.Set2D(mtl.SpecularMap)
	r.sp.Shine.Set(mtl.Shine)
	r.sp.Alpha.Set(mtl.Alpha)

	if mtl.HasAlphaMap() {
		r.sp.HasAlphaMap.Set(true)
		r.sp.AlphaMap.Set2D(mtl.AlphaMap)
	} else {
		r.sp.HasAlphaMap.Set(false)
	}

	if mtl.HasBumpMap() {
		r.sp.HasBumpMap.Set(true)
		r.sp.BumpMap.Set2D(mtl.BumpMap)
	} else {
		r.sp.HasBumpMap.Set(false)
	}

	var v object.Vertex
	r.sp.Position.SetSource(sm.Vbo, v.PositionOffset(), v.Size())
	r.sp.Normal.SetSource(sm.Vbo, v.NormalOffset(), v.Size())
	r.sp.TexCoord.SetSource(sm.Vbo, v.TexCoordOffset(), v.Size())
	r.sp.Tangent.SetSource(sm.Vbo, v.TangentOffset(), v.Size())
	r.sp.SetAttribIndexBuffer(sm.Ibo)
}

func (r *SceneRenderer) SetAmbientLight(l *light.AmbientLight) {
	r.sp.LightType.Set(0)
	r.sp.AmbientLight.Set(l.Color)
}

func (r *SceneRenderer) SetPointLight(l *light.PointLight) {
	r.sp.LightType.Set(1)
	r.sp.LightPos.Set(l.Position)
	r.sp.DiffuseLight.Set(l.Diffuse)
	r.sp.SpecularLight.Set(l.Specular)
	r.sp.CubeShadowMap.SetCube(l.ShadowMap)
	r.sp.ShadowFar.Set(l.ShadowFar)
}

func (r *SceneRenderer) SetSpotLight(l *light.SpotLight) {
	r.sp.LightType.Set(2)
	r.sp.LightPos.Set(l.Position)
	r.sp.LightDir.Set(l.Forward())
	r.sp.DiffuseLight.Set(l.Diffuse)
	r.sp.SpecularLight.Set(l.Specular)
	r.sp.SpotShadowMap.Set2D(l.ShadowMap)
	r.sp.ShadowViewMatrix.Set(l.ViewMatrix())
	r.sp.ShadowProjectionMatrix.Set(l.ProjectionMatrix())
	r.sp.ShadowFar.Set(l.Camera.Far)
}

func (r *SceneRenderer) SetDepthCamera(c *camera.Camera) {
	r.dsp.ViewMatrix.Set(c.ViewMatrix())
	r.dsp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *SceneRenderer) SetDepthMesh(m *object.Mesh) {
	r.dsp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *SceneRenderer) SetDepthSubMesh(sm *object.SubMesh) {
	var v object.Vertex
	r.dsp.Position.SetSource(sm.Vbo, v.PositionOffset(), v.Size())
	r.dsp.SetAttribIndexBuffer(sm.Ibo)
}

type SkyboxRenderer struct {
	sp          *graphics.SkyboxShaderProgram
	vbo         *graphics.Buffer
	ibo         *graphics.Buffer
	tex         *graphics.CubeMap
	renderState *graphics.RenderState
}

func NewSkyboxRenderer() *SkyboxRenderer {
	var r SkyboxRenderer

	r.sp = graphics.NewSkyboxShaderProgram()

	dir := "images/skybox/mountain/"
	names := []string{"posx.jpg", "negx.jpg", "posy.jpg", "negy.jpg", "posz.jpg", "negz.jpg"}
	var filenames [6]string
	for i, name := range names {
		filenames[i] = path.Join(dir, name)
	}
	r.tex = graphics.ReadCubeMap(gl.NEAREST, filenames[0], filenames[1], filenames[2], filenames[3], filenames[4], filenames[5])

	r.vbo = graphics.NewBuffer()
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

	r.ibo = graphics.NewBuffer()
	inds := []int32{
		4, 5, 6, 4, 6, 7,
		5, 1, 2, 5, 2, 6,
		1, 0, 3, 1, 3, 2,
		0, 4, 7, 0, 7, 3,
		7, 6, 2, 7, 2, 3,
		5, 4, 0, 5, 0, 1,
	}
	r.ibo.SetData(inds, 0)

	r.SetCube(r.vbo, r.ibo)

	r.renderState = graphics.NewRenderState()
	r.renderState.SetDepthTest(false)
	r.renderState.SetFramebuffer(graphics.DefaultFramebuffer)
	r.renderState.SetShaderProgram(r.sp.ShaderProgram)
	r.renderState.SetCull(false)
	r.renderState.SetPolygonMode(gl.FILL)

	return &r
}

func (r *SkyboxRenderer) SetCamera(c *camera.Camera) {
	r.sp.ViewMatrix.Set(c.ViewMatrix())
	r.sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *SkyboxRenderer) SetSkybox(skybox *graphics.CubeMap) {
	r.sp.CubeMap.SetCube(skybox)
}

func (r *SkyboxRenderer) SetCube(vbo, ibo *graphics.Buffer) {
	r.sp.Position.SetFormat(gl.FLOAT, false)
	r.sp.Position.SetSource(vbo, 0, int(unsafe.Sizeof(math.NewVec3(0, 0, 0))))
	r.sp.SetAttribIndexBuffer(ibo)
}

func (r *SkyboxRenderer) Render(c *camera.Camera) {
	r.renderState.SetViewport(window.Size())
	r.SetCamera(c)
	r.SetSkybox(r.tex)

	graphics.NewRenderCommand(gl.TRIANGLES, 36, 0, r.renderState).Execute()
}

type TextRenderer struct {
	sp          *graphics.TextShaderProgram
	tex         *graphics.Texture2D
	vbo         *graphics.Buffer
	ibo         *graphics.Buffer
	renderState *graphics.RenderState
}

func NewTextRenderer() *TextRenderer {
	var r TextRenderer

	r.sp = graphics.NewTextShaderProgram()

	r.vbo = graphics.NewBuffer()
	r.ibo = graphics.NewBuffer()

	r.SetAttribs(r.vbo, r.ibo)

	img := basicfont.Face7x13.Mask
	r.tex = graphics.NewTexture2DFromImage(gl.NEAREST, gl.CLAMP_TO_EDGE, gl.RGBA8, img)

	r.renderState = graphics.NewRenderState()
	r.renderState.SetDepthTest(false)
	r.renderState.SetFramebuffer(graphics.DefaultFramebuffer)
	r.renderState.SetShaderProgram(r.sp.ShaderProgram)
	r.renderState.SetBlend(true)
	r.renderState.SetBlendFunction(gl.ONE_MINUS_DST_COLOR, gl.ONE_MINUS_SRC_COLOR)
	r.renderState.SetCull(false)
	r.renderState.SetPolygonMode(gl.FILL)

	return &r
}

func (r *TextRenderer) SetAtlas(tex *graphics.Texture2D) {
	r.sp.Atlas.Set2D(tex)
}

func (r *TextRenderer) SetAttribs(vbo, ibo *graphics.Buffer) {
	var v object.Vertex
	r.sp.Position.SetSource(vbo, v.PositionOffset(), v.Size())
	r.sp.TexCoord.SetSource(vbo, v.TexCoordOffset(), v.Size())
	r.sp.SetAttribIndexBuffer(ibo)
}

func (r *TextRenderer) Render(tl math.Vec2, text string, height float32) {
	r.renderState.SetViewport(window.Size())

	var verts []object.Vertex
	var inds []int32

	face := basicfont.Face7x13

	x0 := tl.X()
	imgW, imgH := face.Mask.Bounds().Dx(), face.Mask.Bounds().Dy()
	subImgW, subImgH := face.Width, face.Ascent+face.Descent
	h := height
	w := h * float32(subImgW) / float32(subImgH)

	for _, char := range text {
		for _, runeRange := range face.Ranges {
			lo, hi, offset := runeRange.Low, runeRange.High, runeRange.Offset
			if char >= lo && char < hi {
				imgX1, imgY1 := 0, imgH-(int(char-lo)+offset)*subImgH
				imgX2, imgY2 := imgX1+subImgW, imgY1-subImgH
				texX1 := float32(imgX1) / float32(imgW) // left
				texY1 := float32(imgY1) / float32(imgH) // top
				texX2 := float32(imgX2) / float32(imgW) // right
				texY2 := float32(imgY2) / float32(imgH) // bottom
				br := math.NewVec2(tl.X()+w, tl.Y()-h)
				tr := math.NewVec2(br.X(), tl.Y())
				bl := math.NewVec2(tl.X(), br.Y())

				normal := math.NewVec3(0, 0, 0)
				vert1 := object.NewVertex(bl.Vec3(0), math.NewVec2(texX1, texY2), normal, math.Vec3{})
				vert2 := object.NewVertex(br.Vec3(0), math.NewVec2(texX2, texY2), normal, math.Vec3{})
				vert3 := object.NewVertex(tr.Vec3(0), math.NewVec2(texX2, texY1), normal, math.Vec3{})
				vert4 := object.NewVertex(tl.Vec3(0), math.NewVec2(texX1, texY1), normal, math.Vec3{})
				inds = append(inds, int32(len(verts)+0))
				inds = append(inds, int32(len(verts)+1))
				inds = append(inds, int32(len(verts)+2))
				inds = append(inds, int32(len(verts)+0))
				inds = append(inds, int32(len(verts)+2))
				inds = append(inds, int32(len(verts)+3))
				verts = append(verts, vert1, vert2, vert3, vert4)
				break
			}
		}

		if char == '\n' {
			tl = math.NewVec2(x0, tl.Y()-h)
		} else if char == '\t' {
			tl = tl.Add(math.NewVec2(4*float32(face.Advance)*h/float32(subImgH), 0))
		} else {
			tl = tl.Add(math.NewVec2(float32(face.Advance)*h/float32(subImgH), 0))
		}
	}

	r.SetAtlas(r.tex)
	r.vbo.SetData(verts, 0)
	r.ibo.SetData(inds, 0)
	graphics.NewRenderCommand(gl.TRIANGLES, len(inds), 0, r.renderState).Execute()
}

type ShadowMapRenderer struct {
	sp          *graphics.ShadowMapShaderProgram
	framebuffer *graphics.Framebuffer
	renderState *graphics.RenderState
}

func NewShadowMapRenderer() *ShadowMapRenderer {
	var r ShadowMapRenderer

	r.sp = graphics.NewShadowMapShaderProgram()

	r.framebuffer = graphics.NewFramebuffer()

	r.renderState = graphics.NewRenderState()
	r.renderState.SetShaderProgram(r.sp.ShaderProgram)
	r.renderState.SetFramebuffer(r.framebuffer)
	r.renderState.SetDepthTest(true)
	r.renderState.SetBlend(false)

	return &r
}

func (r *ShadowMapRenderer) SetCamera(c *camera.Camera) {
	r.sp.Far.Set(c.Far)
	r.sp.LightPosition.Set(c.Position)
	r.sp.ViewMatrix.Set(c.ViewMatrix())
	r.sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *ShadowMapRenderer) SetMesh(m *object.Mesh) {
	r.sp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *ShadowMapRenderer) SetSubMesh(sm *object.SubMesh) {
	var v object.Vertex
	r.sp.Position.SetSource(sm.Vbo, v.PositionOffset(), v.Size())
	r.sp.SetAttribIndexBuffer(sm.Ibo)
}

// render shadow map to l's shadow map
func (r *ShadowMapRenderer) RenderPointLightShadowMap(s *scene.Scene, l *light.PointLight) {
	// TODO: re-render also when objects have moved
	if !l.DirtyShadowMap {
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

	c := camera.NewCamera(90, 1, 0.1, l.ShadowFar)
	c.Place(l.Position)

	r.renderState.SetViewport(l.ShadowMap.Width, l.ShadowMap.Height)

	// UNCOMMENT THIS LINE AND ANOTHER ONE TO DRAW SHADOW CUBE MAP AS SKYBOX
	//shadowCubeMap = l.shadowMap

	for face := 0; face < 6; face++ {
		r.framebuffer.SetTextureCubeMapFace(gl.DEPTH_ATTACHMENT, l.ShadowMap, 0, int32(face))
		r.framebuffer.ClearDepth(1)
		c.SetForwardUp(forwards[face], ups[face])

		r.SetCamera(c)

		for _, m := range s.Meshes {
			r.SetMesh(m)
			for _, subMesh := range m.SubMeshes {
				r.SetSubMesh(subMesh)

				graphics.NewRenderCommand(gl.TRIANGLES, subMesh.Inds, 0, r.renderState).Execute()
			}
		}
	}

	l.DirtyShadowMap = false
}

func (r *ShadowMapRenderer) RenderSpotLightShadowMap(s *scene.Scene, l *light.SpotLight) {
	// TODO: re-render also when objects have moved
	if !l.DirtyShadowMap {
		return
	}

	r.framebuffer.SetTexture2D(gl.DEPTH_ATTACHMENT, l.ShadowMap, 0)
	r.framebuffer.ClearDepth(1)
	r.renderState.SetViewport(l.ShadowMap.Width, l.ShadowMap.Height)
	r.SetCamera(&l.Camera)

	for _, m := range s.Meshes {
		r.SetMesh(m)
		for _, subMesh := range m.SubMeshes {
			r.SetSubMesh(subMesh)

			graphics.NewRenderCommand(gl.TRIANGLES, subMesh.Inds, 0, r.renderState).Execute()
		}
	}

	l.DirtyShadowMap = false
}

type ArrowRenderer struct {
	sp          *graphics.ArrowShaderProgram
	points      []math.Vec3
	vbo         *graphics.Buffer
	renderState *graphics.RenderState
}

func NewArrowRenderer() *ArrowRenderer {
	var r ArrowRenderer

	r.sp = graphics.NewArrowShaderProgram()

	r.renderState = graphics.NewRenderState()
	r.renderState.SetBlend(false)
	r.renderState.SetCull(false)
	r.renderState.SetDepthTest(true)
	r.renderState.SetFramebuffer(graphics.DefaultFramebuffer)
	r.renderState.SetShaderProgram(r.sp.ShaderProgram)

	r.vbo = graphics.NewBuffer()
	r.SetPosition(r.vbo)

	return &r
}

func (r *ArrowRenderer) SetCamera(c *camera.Camera) {
	r.sp.ViewMatrix.Set(c.ViewMatrix())
	r.sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *ArrowRenderer) SetMesh(m *object.Mesh) {
	r.sp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *ArrowRenderer) SetColor(color math.Vec3) {
	r.sp.Color.Set(color)
}

func (r *ArrowRenderer) SetPosition(vbo *graphics.Buffer) {
	stride := int(unsafe.Sizeof(math.NewVec3(0, 0, 0)))
	r.sp.Position.SetSource(vbo, 0, stride)
}

func (r *ArrowRenderer) RenderTangents(s *scene.Scene, c *camera.Camera) {
	r.renderState.SetViewport(window.Size())
	r.SetCamera(c)
	r.points = r.points[:0]
	r.SetColor(math.NewVec3(1, 0, 0))
	for _, m := range s.Meshes {
		r.SetMesh(m)
		for _, subMesh := range m.SubMeshes {
			for _, i := range subMesh.Faces {
				p1 := subMesh.Verts[i].Pos
				p2 := p1.Add(subMesh.Verts[i].Tangent)
				r.points = append(r.points, p1, p2)
			}
		}
	}
	r.vbo.SetData(r.points, 0)
	graphics.NewRenderCommand(gl.LINES, len(r.points), 0, r.renderState).Execute()
}

func (r *ArrowRenderer) RenderBitangents(s *scene.Scene, c *camera.Camera) {
	r.renderState.SetViewport(window.Size())
	r.SetCamera(c)
	r.points = r.points[:0]
	r.SetColor(math.NewVec3(0, 1, 0))
	for _, m := range s.Meshes {
		r.SetMesh(m)
		for _, subMesh := range m.SubMeshes {
			for _, i := range subMesh.Faces {
				p1 := subMesh.Verts[i].Pos
				p2 := p1.Add(subMesh.Verts[i].Normal.Cross(subMesh.Verts[i].Tangent))
				r.points = append(r.points, p1, p2)
			}
		}
	}
	r.vbo.SetData(r.points, 0)
	graphics.NewRenderCommand(gl.LINES, len(r.points), 0, r.renderState).Execute()
}

func (r *ArrowRenderer) RenderNormals(s *scene.Scene, c *camera.Camera) {
	r.renderState.SetViewport(window.Size())
	r.SetCamera(c)
	r.points = r.points[:0]
	r.SetColor(math.NewVec3(0, 0, 1))
	for _, m := range s.Meshes {
		r.SetMesh(m)
		for _, subMesh := range m.SubMeshes {
			for _, i := range subMesh.Faces {
				p1 := subMesh.Verts[i].Pos
				p2 := p1.Add(subMesh.Verts[i].Normal)
				r.points = append(r.points, p1, p2)
			}
		}
	}
	r.vbo.SetData(r.points, 0)
	graphics.NewRenderCommand(gl.LINES, len(r.points), 0, r.renderState).Execute()
}
