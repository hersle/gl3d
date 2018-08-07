package render

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/scene"
	"github.com/hersle/gl3d/window"
)

var shadowCubeMap *graphics.CubeMap = nil

// TODO: redesign attr/uniform access system?
type SceneRenderer struct {
	sp        *graphics.MeshShaderProgram
	dsp       *graphics.DepthPassShaderProgram
	vbo, ibo  *graphics.Buffer
	normalMat math.Mat4

	renderState      *graphics.RenderState
	depthRenderState *graphics.RenderState

	framebuffer *graphics.Framebuffer
	RenderTarget *graphics.Texture2D
	DepthRenderTarget *graphics.Texture2D

	shadowMapRenderer *ShadowMapRenderer
}

func NewSceneRenderer() (*SceneRenderer, error) {
	var r SceneRenderer

	r.sp = graphics.NewMeshShaderProgram()

	r.dsp = graphics.NewDepthPassShaderProgram()

	w, h := window.Size()
	w, h = w / 4, h / 4
	r.RenderTarget = graphics.NewTexture2D(graphics.NearestFilter, graphics.EdgeClampWrap, gl.RGBA8, w, h)
	r.DepthRenderTarget = graphics.NewTexture2D(graphics.NearestFilter, graphics.EdgeClampWrap, gl.DEPTH_COMPONENT16, w, h)
	r.framebuffer = graphics.NewFramebuffer()
	r.framebuffer.AttachTexture2D(graphics.ColorAttachment, r.RenderTarget, 0)
	r.framebuffer.AttachTexture2D(graphics.DepthAttachment, r.DepthRenderTarget, 0)

	r.renderState = graphics.NewRenderState()
	r.renderState.SetShaderProgram(r.sp.ShaderProgram)
	r.renderState.SetFramebuffer(r.framebuffer)
	r.renderState.SetDepthTest(true)
	r.renderState.SetDepthFunc(gl.LEQUAL) // enable drawing after depth prepass
	r.renderState.SetBlend(true)
	r.renderState.SetBlendFunction(gl.ONE, gl.ONE) // add to framebuffer contents
	r.renderState.SetCull(true)
	r.renderState.SetCullFace(gl.BACK) // CCW treated as front face by default
	r.renderState.SetPolygonMode(gl.FILL)
	r.renderState.SetViewport(r.RenderTarget.Width, r.RenderTarget.Height)

	r.depthRenderState = graphics.NewRenderState()
	r.depthRenderState.SetShaderProgram(r.dsp.ShaderProgram)
	r.depthRenderState.SetFramebuffer(r.framebuffer)
	r.depthRenderState.SetDepthTest(true)
	r.depthRenderState.SetDepthFunc(gl.LESS) // enable drawing after depth prepass
	r.depthRenderState.SetBlend(false)
	r.depthRenderState.SetCull(true)
	r.depthRenderState.SetCullFace(gl.BACK) // CCW treated as front face by default
	r.depthRenderState.SetPolygonMode(gl.FILL)
	r.depthRenderState.SetViewport(r.RenderTarget.Width, r.RenderTarget.Height)

	r.shadowMapRenderer = NewShadowMapRenderer()

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

func (r *SceneRenderer) DepthPass(s *scene.Scene, c camera.Camera) {
	r.SetDepthCamera(c)
	for _, m := range s.Meshes {
		r.SetDepthMesh(m)
		for _, subMesh := range m.SubMeshes {
			r.SetDepthSubMesh(subMesh)
			graphics.NewRenderCommand(graphics.Triangle, subMesh.Inds, 0, r.depthRenderState).Execute()
		}
	}
}

func (r *SceneRenderer) AmbientPass(s *scene.Scene, c camera.Camera) {
	r.SetAmbientLight(s.AmbientLight)
	for _, m := range s.Meshes {
		r.renderMesh(m, c)
	}
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

func (r *SceneRenderer) Render(s *scene.Scene, c camera.Camera) {
	r.framebuffer.ClearColor(math.NewVec4(0, 0, 0, 1))
	r.framebuffer.ClearDepth(1)
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
	r.sp.ShadowFar.Set(l.PerspectiveCamera.Far)
}

func (r *SceneRenderer) SetDepthCamera(c camera.Camera) {
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
