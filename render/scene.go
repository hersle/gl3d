package render

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/scene"
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

	LightType              *graphics.UniformInteger
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
type SceneRenderer struct {
	sp        *MeshShaderProgram

	renderState      *graphics.RenderState

	framebuffer       *graphics.Framebuffer
	RenderTarget      *graphics.Texture2D
	DepthRenderTarget *graphics.Texture2D

	shadowMapRenderer *ShadowMapRenderer
	skyboxRenderer    *SkyboxRenderer

	emptyShadowCubeMap *graphics.CubeMap
}

func NewMeshShaderProgram() *MeshShaderProgram {
	var sp MeshShaderProgram
	var err error

	vShaderFilename := "render/shaders/meshvshader.glsl" // TODO: make independent from executable directory
	fShaderFilename := "render/shaders/meshfshader.glsl" // TODO: make independent from executable directory

	sp.ShaderProgram, err = graphics.ReadShaderProgram(vShaderFilename, fShaderFilename, "")
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

	sp.LightType = sp.UniformInteger("light.type")
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

func NewSceneRenderer() (*SceneRenderer, error) {
	var r SceneRenderer

	r.sp = NewMeshShaderProgram()

	w, h := 1920, 1080
	w, h = w/1, h/1
	r.RenderTarget = graphics.NewTexture2D(graphics.NearestFilter, graphics.EdgeClampWrap, gl.RGBA8, w, h)
	r.DepthRenderTarget = graphics.NewTexture2D(graphics.NearestFilter, graphics.EdgeClampWrap, gl.DEPTH_COMPONENT16, w, h)
	r.framebuffer = graphics.NewFramebuffer()
	r.framebuffer.AttachTexture2D(graphics.ColorAttachment, r.RenderTarget, 0)
	r.framebuffer.AttachTexture2D(graphics.DepthAttachment, r.DepthRenderTarget, 0)

	r.renderState = graphics.NewRenderState()
	r.renderState.Program = r.sp.ShaderProgram
	r.renderState.Framebuffer = r.framebuffer
	r.renderState.Cull = graphics.CullBack
	r.renderState.ViewportWidth = r.RenderTarget.Width
	r.renderState.ViewportHeight = r.RenderTarget.Height

	r.shadowMapRenderer = NewShadowMapRenderer()
	r.skyboxRenderer = NewSkyboxRenderer()

	r.emptyShadowCubeMap = graphics.NewCubeMapUniform(math.NewVec4(0, 0, 0, 0))

	return &r, nil
}

func (r *SceneRenderer) renderMesh(m *object.Mesh, c camera.Camera) {
	r.SetMesh(m)
	r.SetCamera(c)

	for _, subMesh := range m.SubMeshes {
		r.SetSubMesh(subMesh)
		graphics.NewRenderCommand(graphics.Triangle, subMesh.Inds, 0, r.renderState).Execute()
	}
}

func (r *SceneRenderer) shadowPassPointLight(s *scene.Scene, l *light.PointLight) {
	r.shadowMapRenderer.RenderPointLightShadowMap(s, l)
}

func (r *SceneRenderer) shadowPassSpotLight(s *scene.Scene, l *light.SpotLight) {
	r.shadowMapRenderer.RenderSpotLightShadowMap(s, l)
}

func (r *SceneRenderer) shadowPassDirectionalLight(s *scene.Scene, l *light.DirectionalLight) {
	r.shadowMapRenderer.RenderDirectionalLightShadowMap(s, l)
}

func (r *SceneRenderer) AmbientPass(s *scene.Scene, c camera.Camera) {
	r.renderState.DisableBlending()
	r.renderState.DepthTest = graphics.LessDepthTest

	// TODO: WHY MUST THIS BE SET FOR AMBIENT LIGHT?!?! GRAPHICS DRIVER BUG?
	// TODO: FIX: AVOID BRANCHING IN SHADERS!!!!!!!!!!!!!
	r.sp.CubeShadowMap.SetCube(r.emptyShadowCubeMap)
	r.SetAmbientLight(s.AmbientLight)
	for _, m := range s.Meshes {
		r.renderMesh(m, c)
	}
}

func (r *SceneRenderer) LightPass(s *scene.Scene, c camera.Camera) {
	r.renderState.DepthTest = graphics.EqualDepthTest
	r.renderState.BlendSourceFactor = graphics.OneBlendFactor
	r.renderState.BlendDestinationFactor = graphics.OneBlendFactor // add to framebuffer contents
	r.PointLightPass(s, c)
	r.SpotLightPass(s, c)
	r.DirectionalLightPass(s, c)
}

func (r *SceneRenderer) PointLightPass(s *scene.Scene, c camera.Camera) {
	for _, l := range s.PointLights {
		r.shadowPassPointLight(s, l)
		r.SetPointLight(l)
		for _, m := range s.Meshes {
			r.renderMesh(m, c)
		}
	}
}

func (r *SceneRenderer) SpotLightPass(s *scene.Scene, c camera.Camera) {
	for _, l := range s.SpotLights {
		r.shadowPassSpotLight(s, l)
		r.SetSpotLight(l)
		for _, m := range s.Meshes {
			r.renderMesh(m, c)
		}
	}
}

func (r *SceneRenderer) DirectionalLightPass(s *scene.Scene, c camera.Camera) {
	for _, l := range s.DirectionalLights {
		r.shadowPassDirectionalLight(s, l)
		r.SetDirectionalLight(l)
		for _, m := range s.Meshes {
			r.renderMesh(m, c)
		}
	}
}

func (r *SceneRenderer) Render(s *scene.Scene, c camera.Camera) {
	r.framebuffer.ClearColor(math.NewVec4(0, 0, 0, 1))
	r.framebuffer.ClearDepth(1)

	r.skyboxRenderer.SetFramebuffer(r.framebuffer)
	r.skyboxRenderer.SetFramebufferSize(r.RenderTarget.Width, r.RenderTarget.Height)
	r.skyboxRenderer.SetSkybox(s.Skybox)
	r.skyboxRenderer.Render(c)

	r.AmbientPass(s, c) // also works as depth pass
	r.LightPass(s, c)
}

func (r *SceneRenderer) SetWireframe(wireframe bool) {
	if wireframe {
		r.renderState.TriangleMode = graphics.LineTriangleMode
	} else {
		r.renderState.TriangleMode = graphics.TriangleTriangleMode
	}
}

func (r *SceneRenderer) SetCamera(c camera.Camera) {
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
	r.sp.AlphaMap.Set2D(mtl.AlphaMap)
	r.sp.BumpMap.Set2D(mtl.BumpMap)

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
	r.sp.LightAttQuad.Set(0)
}

func (r *SceneRenderer) SetPointLight(l *light.PointLight) {
	r.sp.LightType.Set(1)
	r.sp.LightPos.Set(l.Position)
	r.sp.DiffuseLight.Set(l.Diffuse)
	r.sp.SpecularLight.Set(l.Specular)
	r.sp.CubeShadowMap.SetCube(l.ShadowMap)
	r.sp.ShadowFar.Set(l.ShadowFar)
	r.sp.LightAttQuad.Set(l.AttenuationQuadratic)
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
	r.sp.ShadowFar.Set(l.PerspectiveCamera.Far)
	r.sp.LightAttQuad.Set(l.AttenuationQuadratic)
}

func (r *SceneRenderer) SetDirectionalLight(l *light.DirectionalLight) {
	r.sp.LightType.Set(3)
	r.sp.LightDir.Set(l.Forward())
	r.sp.DiffuseLight.Set(l.Diffuse)
	r.sp.SpecularLight.Set(l.Specular)
	r.sp.DirShadowMap.Set2D(l.ShadowMap)
	r.sp.ShadowViewMatrix.Set(l.ViewMatrix())
	r.sp.ShadowProjectionMatrix.Set(l.ProjectionMatrix())
	r.sp.LightAttQuad.Set(0)
}
