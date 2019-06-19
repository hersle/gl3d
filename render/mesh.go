package render

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/scene"
	"image"
)

var shadowCubeMap *graphics.CubeMap = nil

type MeshShaderProgram struct {
	*graphics.ShaderProgram

	Position *graphics.Attrib
	TexCoord *graphics.Attrib
	Normal   *graphics.Attrib
	Tangent  *graphics.Attrib

	ModelMatrix      *graphics.UniformMatrix4
	ViewMatrix       *graphics.UniformMatrix4
	ProjectionMatrix *graphics.UniformMatrix4

	NormalMatrix *graphics.UniformMatrix4

	Ambient     *graphics.UniformVector3
	AmbientMap  *graphics.UniformSampler
	Diffuse     *graphics.UniformVector3
	DiffuseMap  *graphics.UniformSampler
	Specular    *graphics.UniformVector3
	SpecularMap *graphics.UniformSampler
	Shine       *graphics.UniformFloat
	Alpha       *graphics.UniformFloat
	AlphaMap    *graphics.UniformSampler
	BumpMap     *graphics.UniformSampler

	LightPos               *graphics.UniformVector3
	LightDir               *graphics.UniformVector3
	AmbientLight           *graphics.UniformVector3
	DiffuseLight           *graphics.UniformVector3
	SpecularLight          *graphics.UniformVector3
	ShadowViewMatrix       *graphics.UniformMatrix4
	ShadowProjectionMatrix *graphics.UniformMatrix4
	SpotShadowMap          *graphics.UniformSampler
	CubeShadowMap          *graphics.UniformSampler
	DirShadowMap           *graphics.UniformSampler
	ShadowFar              *graphics.UniformFloat
	LightAttQuad           *graphics.UniformFloat
}

// TODO: redesign attr/uniform access system?
type MeshRenderer struct {
	sp1        *MeshShaderProgram
	sp2        *MeshShaderProgram
	sp3        *MeshShaderProgram
	sp4        *MeshShaderProgram

	renderState      *graphics.RenderState

	emptyShadowCubeMap *graphics.CubeMap

	normalMatrix math.Mat4

	vboCache map[*object.Vertex]int
	vbos []*graphics.Buffer
	ibos []*graphics.Buffer

	tex2ds map[image.Image]*graphics.Texture2D

	pointLightShadowMaps map[int]*graphics.CubeMap
	spotLightShadowMaps map[int]*graphics.Texture2D
	dirLightShadowMaps map[int]*graphics.Texture2D

	shadowSp          *ShadowMapShaderProgram
	dirshadowSp         *DirectionalLightShadowMapShaderProgram
	shadowMapFramebuffer *graphics.Framebuffer
	shadowRenderState *graphics.RenderState
}

type ShadowMapShaderProgram struct {
	*graphics.ShaderProgram
	ModelMatrix      *graphics.UniformMatrix4
	ViewMatrix       *graphics.UniformMatrix4
	ProjectionMatrix *graphics.UniformMatrix4
	LightPosition    *graphics.UniformVector3
	Far              *graphics.UniformFloat
	Position         *graphics.Attrib
}

type DirectionalLightShadowMapShaderProgram struct {
	*graphics.ShaderProgram
	ModelMatrix      *graphics.UniformMatrix4
	ViewMatrix       *graphics.UniformMatrix4
	ProjectionMatrix *graphics.UniformMatrix4
	Position         *graphics.Attrib
}

func NewShadowMapShaderProgram() *ShadowMapShaderProgram {
	var sp ShadowMapShaderProgram
	var err error

	vShaderFilename := "render/shaders/pointlightshadowmapvshader.glsl" // TODO: make independent from executable directory
	fShaderFilename := "render/shaders/pointlightshadowmapfshader.glsl" // TODO: make independent from executable directory
	sp.ShaderProgram, err = graphics.ReadShaderProgram(vShaderFilename, fShaderFilename, "")
	if err != nil {
		panic(err)
	}

	sp.ModelMatrix = sp.UniformMatrix4("modelMatrix")
	sp.ViewMatrix = sp.UniformMatrix4("viewMatrix")
	sp.ProjectionMatrix = sp.UniformMatrix4("projectionMatrix")
	sp.LightPosition = sp.UniformVector3("lightPosition")
	sp.Far = sp.UniformFloat("far")
	sp.Position = sp.Attrib("position")

	sp.Position.SetFormat(gl.FLOAT, false) // TODO: remove dependency on GL constants

	return &sp
}

func NewDirectionalLightShadowMapShaderProgram() *DirectionalLightShadowMapShaderProgram {
	var sp DirectionalLightShadowMapShaderProgram
	var err error

	vShaderFilename := "render/shaders/directionallightvshader.glsl"
	sp.ShaderProgram, err = graphics.ReadShaderProgram(vShaderFilename, "", "")
	if err != nil {
		panic(err)
	}

	sp.ModelMatrix = sp.UniformMatrix4("modelMatrix")
	sp.ViewMatrix = sp.UniformMatrix4("viewMatrix")
	sp.ProjectionMatrix = sp.UniformMatrix4("projectionMatrix")
	sp.Position = sp.Attrib("position")
	sp.Position.SetFormat(gl.FLOAT, false)

	return &sp
}

func NewMeshShaderProgram(defines []string) *MeshShaderProgram {
	var sp MeshShaderProgram
	var err error

	vShaderFilename := "render/shaders/meshvshadertemplate.glsl" // TODO: make independent from executable directory
	fShaderFilename := "render/shaders/meshfshadertemplate.glsl" // TODO: make independent from executable directory

	sp.ShaderProgram, err = graphics.ReadShaderProgramFromTemplates(vShaderFilename, fShaderFilename, "", defines)
	if err != nil {
		panic(err)
	}

	sp.Position = sp.Attrib("position")
	sp.TexCoord = sp.Attrib("texCoordV")
	sp.Normal = sp.Attrib("normalV")
	sp.Tangent = sp.Attrib("tangentV")

	sp.ModelMatrix = sp.UniformMatrix4("modelMatrix")
	sp.ViewMatrix = sp.UniformMatrix4("viewMatrix")
	sp.ProjectionMatrix = sp.UniformMatrix4("projectionMatrix")

	sp.NormalMatrix = sp.UniformMatrix4("normalMatrix")

	sp.Ambient = sp.UniformVector3("material.ambient")
	sp.AmbientMap = sp.UniformSampler("material.ambientMap")
	sp.Diffuse = sp.UniformVector3("material.diffuse")
	sp.DiffuseMap = sp.UniformSampler("material.diffuseMap")
	sp.Specular = sp.UniformVector3("material.specular")
	sp.SpecularMap = sp.UniformSampler("material.specularMap")
	sp.Shine = sp.UniformFloat("material.shine")
	sp.Alpha = sp.UniformFloat("material.alpha")
	sp.AlphaMap = sp.UniformSampler("material.alphaMap")
	sp.BumpMap = sp.UniformSampler("material.bumpMap")

	sp.LightPos = sp.UniformVector3("light.position")
	sp.LightDir = sp.UniformVector3("light.direction")
	sp.AmbientLight = sp.UniformVector3("light.ambient")
	sp.DiffuseLight = sp.UniformVector3("light.diffuse")
	sp.SpecularLight = sp.UniformVector3("light.specular")
	sp.ShadowViewMatrix = sp.UniformMatrix4("shadowViewMatrix")
	sp.ShadowProjectionMatrix = sp.UniformMatrix4("shadowProjectionMatrix")
	sp.CubeShadowMap = sp.UniformSampler("cubeShadowMap")
	sp.SpotShadowMap = sp.UniformSampler("spotShadowMap")
	sp.DirShadowMap = sp.UniformSampler("dirShadowMap")
	sp.ShadowFar = sp.UniformFloat("light.far")
	sp.LightAttQuad = sp.UniformFloat("light.attenuationQuadratic")

	sp.Position.SetFormat(gl.FLOAT, false)
	sp.Normal.SetFormat(gl.FLOAT, false)
	sp.TexCoord.SetFormat(gl.FLOAT, false)
	sp.Tangent.SetFormat(gl.FLOAT, false)

	return &sp
}

func NewMeshRenderer() (*MeshRenderer, error) {
	var r MeshRenderer

	r.sp1 = NewMeshShaderProgram([]string{"DEPTH", "AMBIENT"})
	r.sp2 = NewMeshShaderProgram([]string{"POINT"})
	r.sp3 = NewMeshShaderProgram([]string{"SPOT"})
	r.sp4 = NewMeshShaderProgram([]string{"DIR"})

	r.renderState = graphics.NewRenderState()
	r.renderState.Cull = graphics.CullBack
	r.renderState.PrimitiveType = graphics.Triangle

	r.emptyShadowCubeMap = graphics.NewCubeMapUniform(math.NewVec4(0, 0, 0, 0))

	r.vboCache = make(map[*object.Vertex]int)
	r.pointLightShadowMaps = make(map[int]*graphics.CubeMap)
	r.spotLightShadowMaps = make(map[int]*graphics.Texture2D)
	r.dirLightShadowMaps = make(map[int]*graphics.Texture2D)
	r.tex2ds = make(map[image.Image]*graphics.Texture2D)

	r.shadowSp = NewShadowMapShaderProgram()
	r.dirshadowSp = NewDirectionalLightShadowMapShaderProgram()

	r.shadowMapFramebuffer = graphics.NewFramebuffer()

	r.shadowRenderState = graphics.NewRenderState()
	r.shadowRenderState.Framebuffer = r.shadowMapFramebuffer
	r.shadowRenderState.DepthTest = graphics.LessDepthTest
	r.shadowRenderState.Cull = graphics.CullBack

	return &r, nil
}

func (r *MeshRenderer) renderMesh(sp *MeshShaderProgram, m *object.Mesh, c camera.Camera) {
	r.SetMesh(sp, m)
	r.SetCamera(sp, c)

	// TODO: cache per mesh/camera?
	if sp.NormalMatrix != nil {
		r.normalMatrix.Identity()
		r.normalMatrix.Mult(c.ViewMatrix())
		r.normalMatrix.Mult(m.WorldMatrix())
		r.normalMatrix.Invert()
		r.normalMatrix.Transpose()
		sp.NormalMatrix.Set(&r.normalMatrix)
	}

	for _, subMesh := range m.SubMeshes {
		if !c.Cull(subMesh) {
			r.SetSubMesh(sp, subMesh)
			graphics.NewRenderCommand(subMesh.Geo.Inds, 0, r.renderState).Execute()
		}
	}
}

func (r *MeshRenderer) AmbientPass(s *scene.Scene, c camera.Camera) {
	r.renderState.DisableBlending()
	r.renderState.DepthTest = graphics.LessDepthTest
	r.renderState.Program = r.sp1.ShaderProgram

	for _, m := range s.Meshes {
		r.renderMesh(r.sp1, m, c)
	}
}

func (r *MeshRenderer) LightPass(s *scene.Scene, c camera.Camera) {
	r.renderState.DepthTest = graphics.EqualDepthTest
	r.renderState.BlendSourceFactor = graphics.OneBlendFactor
	r.renderState.BlendDestinationFactor = graphics.OneBlendFactor // add to framebuffer contents

	r.PointLightPass(s, c)
	r.SpotLightPass(s, c)
	r.DirectionalLightPass(s, c)
}

func (r *MeshRenderer) PointLightPass(s *scene.Scene, c camera.Camera) {
	r.renderState.Program = r.sp2.ShaderProgram

	for _, l := range s.PointLights {
		r.SetPointLight(r.sp2, l)
		for _, m := range s.Meshes {
			r.renderMesh(r.sp2, m, c)
		}
	}
}

func (r *MeshRenderer) SpotLightPass(s *scene.Scene, c camera.Camera) {
	r.renderState.Program = r.sp3.ShaderProgram

	for _, l := range s.SpotLights {
		r.SetSpotLight(r.sp3, l)
		for _, m := range s.Meshes {
			r.renderMesh(r.sp3, m, c)
		}
	}
}

func (r *MeshRenderer) DirectionalLightPass(s *scene.Scene, c camera.Camera) {
	r.renderState.Program = r.sp4.ShaderProgram

	for _, l := range s.DirectionalLights {
		r.SetDirectionalLight(r.sp4, l)
		for _, m := range s.Meshes {
			r.renderMesh(r.sp4, m, c)
		}
	}
}

func (r *MeshRenderer) Render(s *scene.Scene, c camera.Camera, fb *graphics.Framebuffer) {
	r.RenderShadowMaps(s)

	r.renderState.Framebuffer = fb

	r.AmbientPass(s, c) // also works as depth pass
	r.LightPass(s, c)
}

func (r *MeshRenderer) SetWireframe(wireframe bool) {
	if wireframe {
		r.renderState.TriangleMode = graphics.LineTriangleMode
	} else {
		r.renderState.TriangleMode = graphics.TriangleTriangleMode
	}
}

func (r *MeshRenderer) SetCamera(sp *MeshShaderProgram, c camera.Camera) {
	sp.ViewMatrix.Set(c.ViewMatrix())
	sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *MeshRenderer) SetMesh(sp *MeshShaderProgram, m *object.Mesh) {
	sp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *MeshRenderer) SetSubMesh(sp *MeshShaderProgram, sm *object.SubMesh) {
	mtl := sm.Mtl

	tex, found := r.tex2ds[mtl.AmbientMap]
	if !found {
		tex = graphics.NewTexture2DFromImage(graphics.LinearFilter, graphics.RepeatWrap, gl.RGBA8, mtl.AmbientMap)
		r.tex2ds[mtl.AmbientMap] = tex
	}
	sp.Ambient.Set(mtl.Ambient)
	sp.AmbientMap.Set2D(tex)

	tex, found = r.tex2ds[mtl.DiffuseMap]
	if !found {
		tex = graphics.NewTexture2DFromImage(graphics.LinearFilter, graphics.RepeatWrap, gl.RGBA8, mtl.DiffuseMap)
		r.tex2ds[mtl.DiffuseMap] = tex
	}
	sp.Diffuse.Set(mtl.Diffuse)
	sp.DiffuseMap.Set2D(tex)

	tex, found = r.tex2ds[mtl.SpecularMap]
	if !found {
		tex = graphics.NewTexture2DFromImage(graphics.LinearFilter, graphics.RepeatWrap, gl.RGBA8, mtl.SpecularMap)
		r.tex2ds[mtl.SpecularMap] = tex
	}
	sp.Specular.Set(mtl.Specular)
	sp.SpecularMap.Set2D(tex)

	sp.Shine.Set(mtl.Shine)

	tex, found = r.tex2ds[mtl.AlphaMap]
	if !found {
		tex = graphics.NewTexture2DFromImage(graphics.LinearFilter, graphics.RepeatWrap, gl.RGBA8, mtl.AlphaMap)
		r.tex2ds[mtl.AlphaMap] = tex
	}
	sp.Alpha.Set(mtl.Alpha)
	sp.AlphaMap.Set2D(tex)

	tex, found = r.tex2ds[mtl.BumpMap]
	if !found {
		tex = graphics.NewTexture2DFromImage(graphics.LinearFilter, graphics.RepeatWrap, gl.RGBA8, mtl.BumpMap)
		r.tex2ds[mtl.BumpMap] = tex
	}
	sp.BumpMap.Set2D(tex)

	// upload to GPU
	var vbo *graphics.Buffer
	var ibo *graphics.Buffer
	i, found := r.vboCache[&sm.Geo.Verts[0]]
	if found {
		vbo = r.vbos[i]
		ibo = r.ibos[i]
	} else {
		vbo = graphics.NewBuffer()
		ibo = graphics.NewBuffer()
		vbo.SetData(sm.Geo.Verts, 0)
		ibo.SetData(sm.Geo.Faces, 0)

		r.vbos = append(r.vbos, vbo)
		r.ibos = append(r.ibos, ibo)
		r.vboCache[&sm.Geo.Verts[0]] = len(r.vbos)-1
	}

	var v object.Vertex
	sp.Position.SetSource(vbo, v.PositionOffset(), v.Size())
	sp.Normal.SetSource(vbo, v.NormalOffset(), v.Size())
	sp.TexCoord.SetSource(vbo, v.TexCoordOffset(), v.Size())
	sp.Tangent.SetSource(vbo, v.TangentOffset(), v.Size())
	sp.SetAttribIndexBuffer(ibo)
}

func (r *MeshRenderer) SetAmbientLight(sp *MeshShaderProgram, l *light.AmbientLight) {
	sp.AmbientLight.Set(l.Color)
	sp.LightAttQuad.Set(0)
}

func (r *MeshRenderer) SetPointLight(sp *MeshShaderProgram, l *light.PointLight) {
	sp.LightPos.Set(l.Position)
	sp.DiffuseLight.Set(l.Diffuse)
	sp.SpecularLight.Set(l.Specular)
	if l.CastShadows {
		sp.ShadowFar.Set(l.ShadowFar)
		smap, found := r.pointLightShadowMaps[l.ID]
		if !found {
			panic("set point light with no shadow map")
		}
		sp.CubeShadowMap.SetCube(smap)
	}
	sp.LightAttQuad.Set(l.AttenuationQuadratic)
}

func (r *MeshRenderer) SetSpotLight(sp *MeshShaderProgram, l *light.SpotLight) {
	sp.LightPos.Set(l.Position)
	sp.LightDir.Set(l.Forward())
	sp.DiffuseLight.Set(l.Diffuse)
	sp.SpecularLight.Set(l.Specular)
	sp.LightAttQuad.Set(l.AttenuationQuadratic)

	if l.CastShadows {
		sp.ShadowViewMatrix.Set(l.ViewMatrix())
		sp.ShadowProjectionMatrix.Set(l.ProjectionMatrix())
		sp.ShadowFar.Set(l.PerspectiveCamera.Far)
		smap, found := r.spotLightShadowMaps[l.ID]
		if !found {
			panic("set spot light with no shadow map")
		}
		sp.SpotShadowMap.Set2D(smap)
	}
}

func (r *MeshRenderer) SetDirectionalLight(sp *MeshShaderProgram, l *light.DirectionalLight) {
	sp.LightDir.Set(l.Forward())
	sp.DiffuseLight.Set(l.Diffuse)
	sp.SpecularLight.Set(l.Specular)
	sp.LightAttQuad.Set(0)

	if l.CastShadows {
		sp.ShadowViewMatrix.Set(l.ViewMatrix())
		sp.ShadowProjectionMatrix.Set(l.ProjectionMatrix())
		smap, found := r.dirLightShadowMaps[l.ID]
		if !found {
			panic("set directional light with no shadow map")
		}
		sp.DirShadowMap.Set2D(smap)
	}
}

// shadow stuff below

func (r *MeshRenderer) SetShadowCamera(c *camera.PerspectiveCamera) {
	r.shadowSp.Far.Set(c.Far)
	r.shadowSp.LightPosition.Set(c.Position)
	r.shadowSp.ViewMatrix.Set(c.ViewMatrix())
	r.shadowSp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *MeshRenderer) SetShadowMesh(m *object.Mesh) {
	r.shadowSp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *MeshRenderer) SetShadowSubMesh(sm *object.SubMesh) {
	var vbo *graphics.Buffer
	var ibo *graphics.Buffer
	i, found := r.vboCache[&sm.Geo.Verts[0]]
	if found {
		vbo = r.vbos[i]
		ibo = r.ibos[i]
	} else {
		vbo = graphics.NewBuffer()
		ibo = graphics.NewBuffer()
		vbo.SetData(sm.Geo.Verts, 0)
		ibo.SetData(sm.Geo.Faces, 0)

		r.vbos = append(r.vbos, vbo)
		r.ibos = append(r.ibos, ibo)
		r.vboCache[&sm.Geo.Verts[0]] = len(r.vbos)-1
	}

	var v object.Vertex
	r.shadowSp.Position.SetSource(vbo, v.PositionOffset(), v.Size())
	r.shadowSp.SetAttribIndexBuffer(ibo)
}

func (r *MeshRenderer) SetDirShadowCamera(sp *MeshShaderProgram, c *camera.OrthoCamera) {
	r.dirshadowSp.ViewMatrix.Set(c.ViewMatrix())
	r.dirshadowSp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *MeshRenderer) SetDirShadowMesh(sp *MeshShaderProgram, m *object.Mesh) {
	r.dirshadowSp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *MeshRenderer) SetDirShadowSubMesh(sp *MeshShaderProgram, sm *object.SubMesh) {
	var vbo *graphics.Buffer
	var ibo *graphics.Buffer
	i, found := r.vboCache[&sm.Geo.Verts[0]]
	if found {
		vbo = r.vbos[i]
		ibo = r.ibos[i]
	} else {
		vbo = graphics.NewBuffer()
		ibo = graphics.NewBuffer()
		vbo.SetData(sm.Geo.Verts, 0)
		ibo.SetData(sm.Geo.Faces, 0)

		r.vbos = append(r.vbos, vbo)
		r.ibos = append(r.ibos, ibo)
		r.vboCache[&sm.Geo.Verts[0]] = len(r.vbos)-1
	}

	var v object.Vertex
	r.dirshadowSp.Position.SetSource(vbo, v.PositionOffset(), v.Size())
	r.dirshadowSp.SetAttribIndexBuffer(ibo)
}

// render shadow map to l's shadow map
func (r *MeshRenderer) RenderPointLightShadowMap(s *scene.Scene, l *light.PointLight) {
	smap, found := r.pointLightShadowMaps[l.ID]
	if !found {
		smap = graphics.NewCubeMap(graphics.NearestFilter, gl.DEPTH_COMPONENT16, 512, 512)
		r.pointLightShadowMaps[l.ID] = smap
	}

	// TODO: re-render also when objects have moved
	/*
	if !l.DirtyShadowMap {
		return
	}
	*/

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

	c := camera.NewPerspectiveCamera(90, 1, 0.1, l.ShadowFar)
	c.Place(l.Position)

	r.shadowRenderState.Program = r.shadowSp.ShaderProgram

	// UNCOMMENT THIS LINE AND ANOTHER ONE TO DRAW SHADOW CUBE MAP AS SKYBOX
	//shadowCubeMap = l.shadowMap

	for face := 0; face < 6; face++ {
		r.shadowMapFramebuffer.AttachCubeMapFace(graphics.DepthAttachment, smap.Face(graphics.CubeMapLayer(face)), 0)
		r.shadowMapFramebuffer.ClearDepth(1)
		c.SetForwardUp(forwards[face], ups[face])

		r.SetShadowCamera(c)

		for _, m := range s.Meshes {
			r.SetShadowMesh(m)
			for _, subMesh := range m.SubMeshes {
				if !c.Cull(subMesh) {
					r.SetShadowSubMesh(subMesh)

					graphics.NewRenderCommand(subMesh.Geo.Inds, 0, r.shadowRenderState).Execute()
				}
			}
		}
	}

	//l.DirtyShadowMap = false
}

/*
func (r *MeshRenderer) RenderSpotLightShadowMap(s *scene.Scene, l *light.SpotLight) {
	smap, found := r.spotLightShadowMaps[l.ID]
	if !found {
		smap = graphics.NewTexture2D(graphics.NearestFilter, graphics.BorderClampWrap, gl.DEPTH_COMPONENT16, 512, 512)
		smap.SetBorderColor(math.NewVec4(1, 1, 1, 1))
		r.spotLightShadowMaps[l.ID] = smap
	}

	// TODO: re-render also when objects have moved
	//if !l.DirtyShadowMap {
		//return
	//}

	r.shadowMapFramebuffer.AttachTexture2D(graphics.DepthAttachment, smap, 0)
	r.shadowMapFramebuffer.ClearDepth(1)
	r.shadowRenderState.Program = r.shadowSp.ShaderProgram
	r.SetShadowCamera(&l.PerspectiveCamera)

	for _, m := range s.Meshes {
		r.SetShadowMesh(m)
		for _, subMesh := range m.SubMeshes {
			if !l.PerspectiveCamera.Cull(subMesh) {
				r.SetShadowSubMesh(subMesh)

				graphics.NewRenderCommand(graphics.Triangle, subMesh.Geo.Inds, 0, r.shadowRenderState).Execute()
			}
		}
	}

	//l.DirtyShadowMap = false
}

func (r *MeshRenderer) RenderDirectionalLightShadowMap(s *scene.Scene, l *light.DirectionalLight) {
	smap, found := r.dirLightShadowMaps[l.ID]
	if !found {
		smap = graphics.NewTexture2D(graphics.NearestFilter, graphics.BorderClampWrap, gl.DEPTH_COMPONENT16, 512, 512)
		smap.SetBorderColor(math.NewVec4(1, 1, 1, 1))
		r.dirLightShadowMaps[l.ID] = smap
	}

	// TODO: re-render also when objects have moved
	//if !l.DirtyShadowMap {
		//return
	//}

	r.shadowMapFramebuffer.AttachTexture2D(graphics.DepthAttachment, smap, 0)
	r.shadowMapFramebuffer.ClearDepth(1)
	r.shadowRenderState.Program = r.dirshadowSp.ShaderProgram
	r.SetDirShadowCamera(&l.OrthoCamera)

	for _, m := range s.Meshes {
		r.SetDirShadowMesh(m)
		for _, subMesh := range m.SubMeshes {
			if !l.OrthoCamera.Cull(subMesh) {
				r.SetDirShadowSubMesh(subMesh)

				graphics.NewRenderCommand(graphics.Triangle, subMesh.Geo.Inds, 0, r.shadowRenderState).Execute()
			}
		}
	}

	//l.DirtyShadowMap = false
}
*/

func (r *MeshRenderer) RenderShadowMaps(s *scene.Scene) {
	for _, l := range s.PointLights {
		if l.CastShadows {
			r.RenderPointLightShadowMap(s, l)
		}
	}
	/*
	for _, l := range s.SpotLights {
		if l.CastShadows {
			r.RenderSpotLightShadowMap(s, l)
		}
	}
	for _, l := range s.DirectionalLights {
		if l.CastShadows {
			r.RenderDirectionalLightShadowMap(s, l)
		}
	}
	*/
}

func PointLightInteracts(l *light.PointLight, sm *object.SubMesh) bool {
	sphere := sm.BoundingSphere()
	dist := l.Position.Sub(sphere.Center).Length()
	if dist < sphere.Radius {
		return true
	}
	dist = dist - sphere.Radius
	return dist*dist < (1 / 0.05 - 1) / l.AttenuationQuadratic
}
