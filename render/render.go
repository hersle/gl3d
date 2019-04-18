package render

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/scene"
)

// TODO: redesign attr/uniform access system?
type Renderer struct {
	meshRenderer *MeshRenderer
	shadowMapRenderer *ShadowMapRenderer
	skyboxRenderer    *SkyboxRenderer

	framebuffer       *graphics.Framebuffer
	RenderTarget      *graphics.Texture2D
	DepthRenderTarget *graphics.Texture2D
}

func NewRenderer() (*Renderer, error) {
	var r Renderer

	r.meshRenderer, _ = NewMeshRenderer()
	r.shadowMapRenderer = NewShadowMapRenderer()
	r.skyboxRenderer = NewSkyboxRenderer()

	w, h := 1920, 1080
	w, h = w/1, h/1
	r.RenderTarget = graphics.NewTexture2D(graphics.NearestFilter, graphics.EdgeClampWrap, gl.RGBA8, w, h)
	r.DepthRenderTarget = graphics.NewTexture2D(graphics.NearestFilter, graphics.EdgeClampWrap, gl.DEPTH_COMPONENT16, w, h)
	r.framebuffer = graphics.NewFramebuffer()
	r.framebuffer.AttachTexture2D(graphics.ColorAttachment, r.RenderTarget, 0)
	r.framebuffer.AttachTexture2D(graphics.DepthAttachment, r.DepthRenderTarget, 0)

	return &r, nil
}

func (r *Renderer) Render(s *scene.Scene, c camera.Camera) {
	r.framebuffer.ClearColor(math.NewVec4(0, 0, 0, 1))
	r.framebuffer.ClearDepth(1)

	//r.skyboxRenderer.SetFramebuffer(r.framebuffer)
	//r.skyboxRenderer.SetFramebufferSize(r.RenderTarget.Width, r.RenderTarget.Height)
	//r.skyboxRenderer.SetSkybox(s.Skybox)
	r.skyboxRenderer.Render(s.Skybox, c, r.framebuffer)

	// TODO: update shadows
	r.shadowMapRenderer.RenderShadowMaps(s)

	r.meshRenderer.Render(s, c, r.framebuffer)
}

func (r *Renderer) SetWireframe(wireframe bool) {
	r.meshRenderer.SetWireframe(wireframe)
}
