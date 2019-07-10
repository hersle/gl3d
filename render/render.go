package render

import (
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/scene"
)

// TODO: redesign attr/uniform access system?
type Renderer struct {
	MeshRenderer   *MeshRenderer
	SkyboxRenderer *SkyboxRenderer
	TextRenderer   *TextRenderer
	QuadRenderer   *QuadRenderer
	ArrowRenderer  *ArrowRenderer
	EffectRenderer *EffectRenderer

	sceneFramebuffer       *graphics.Framebuffer
	sceneRenderTarget      *graphics.Texture2D
	sceneRenderTarget2     *graphics.Texture2D
	sceneDepthRenderTarget *graphics.Texture2D

	overlayFramebuffer  *graphics.Framebuffer
	overlayRenderTarget *graphics.Texture2D

	Fog bool
	BlurRadius float32
}

func NewRenderer() (*Renderer, error) {
	var r Renderer

	r.MeshRenderer, _ = NewMeshRenderer()
	r.SkyboxRenderer = NewSkyboxRenderer()
	r.TextRenderer = NewTextRenderer()
	r.QuadRenderer = NewQuadRenderer()
	r.ArrowRenderer = NewArrowRenderer()
	r.EffectRenderer = NewEffectRenderer()

	w, h := 1920, 1080
	w, h = w/1, h/1

	r.sceneRenderTarget = graphics.NewTexture2D(graphics.ColorTexture, graphics.NearestFilter, graphics.EdgeClampWrap, w, h, false)
	r.sceneRenderTarget2 = graphics.NewTexture2D(graphics.ColorTexture, graphics.NearestFilter, graphics.EdgeClampWrap, w, h, false)
	r.sceneDepthRenderTarget = graphics.NewTexture2D(graphics.DepthTexture, graphics.NearestFilter, graphics.EdgeClampWrap, w, h, false)
	r.sceneFramebuffer = graphics.NewFramebuffer()
	r.sceneFramebuffer.Attach(r.sceneRenderTarget)
	r.sceneFramebuffer.Attach(r.sceneDepthRenderTarget)

	r.overlayRenderTarget = graphics.NewTexture2D(graphics.ColorTexture, graphics.NearestFilter, graphics.EdgeClampWrap, w, h, false)
	r.overlayFramebuffer = graphics.NewFramebuffer()
	r.overlayFramebuffer.Attach(r.overlayRenderTarget)

	return &r, nil
}

func (r *Renderer) RenderScene(s *scene.Scene, c camera.Camera) {
	if s.Skybox != nil {
		r.SkyboxRenderer.Render(s.Skybox, c, r.sceneFramebuffer)
	}

	r.MeshRenderer.Render(s, c, r.sceneFramebuffer)
	if r.Fog {
		r.EffectRenderer.RenderFog(c, r.sceneDepthRenderTarget, r.sceneRenderTarget)
	}
	if r.BlurRadius > 0 {
		r.EffectRenderer.RenderGaussianBlur(r.sceneRenderTarget, r.sceneRenderTarget2, r.BlurRadius)
	}
}

func (r *Renderer) RenderText(tl math.Vec2, text string, height float32, just Justification) {
	color := math.Vec3{1, 1, 0}
	r.TextRenderer.Render(tl, text, height, color, just, r.overlayFramebuffer)
}

func (r *Renderer) RenderQuad(tex *graphics.Texture2D) {
	r.QuadRenderer.Render(tex, graphics.DefaultFramebuffer)
}

func (r *Renderer) RenderBitangents(s *scene.Scene, c camera.Camera) {
	r.ArrowRenderer.RenderBitangents(s, c, r.sceneFramebuffer)
}

func (r *Renderer) RenderNormals(s *scene.Scene, c camera.Camera) {
	r.ArrowRenderer.RenderNormals(s, c, r.sceneFramebuffer)
}

func (r *Renderer) RenderTangents(s *scene.Scene, c camera.Camera) {
	r.ArrowRenderer.RenderTangents(s, c, r.sceneFramebuffer)
}

func (r *Renderer) Clear() {
	graphics.DefaultFramebuffer.ClearColor(math.Vec4{0, 0, 0, 0})
	r.sceneFramebuffer.ClearColor(math.Vec4{0, 0, 0, 0})
	r.sceneFramebuffer.ClearDepth(1)
	r.overlayFramebuffer.ClearColor(math.Vec4{0, 0, 0, 0})
}

func (r *Renderer) Render() {
	r.RenderQuad(r.sceneRenderTarget)
	r.RenderQuad(r.overlayRenderTarget)
}

func (r *Renderer) SetWireframe(wireframe bool) {
	r.MeshRenderer.SetWireframe(wireframe)
}
