package render

import (
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/scene"
)

// TODO: redesign attr/uniform access system?
type Renderer struct {
	meshRenderer   *MeshRenderer
	skyboxRenderer *SkyboxRenderer
	textRenderer   *TextRenderer
	quadRenderer   *QuadRenderer
	arrowRenderer  *ArrowRenderer
	effectRenderer *EffectRenderer

	sceneFramebuffer       *graphics.Framebuffer
	sceneRenderTarget      *graphics.Texture2D
	sceneRenderTarget2     *graphics.Texture2D
	sceneDepthRenderTarget *graphics.Texture2D

	overlayFramebuffer  *graphics.Framebuffer
	overlayRenderTarget *graphics.Texture2D
}

func NewRenderer() (*Renderer, error) {
	var r Renderer

	r.meshRenderer, _ = NewMeshRenderer()
	r.skyboxRenderer = NewSkyboxRenderer()
	r.textRenderer = NewTextRenderer()
	r.quadRenderer = NewQuadRenderer()
	r.arrowRenderer = NewArrowRenderer()
	r.effectRenderer = NewEffectRenderer()

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
		r.skyboxRenderer.Render(s.Skybox, c, r.sceneFramebuffer)
	}

	r.meshRenderer.Render(s, c, r.sceneFramebuffer)
	r.effectRenderer.RenderFog(c, r.sceneDepthRenderTarget, r.sceneRenderTarget)
	r.effectRenderer.RenderGaussianBlur(r.sceneRenderTarget, r.sceneRenderTarget2)
}

func (r *Renderer) RenderText(tl math.Vec2, text string, height float32) {
	color := math.Vec3{1, 1, 1}
	r.textRenderer.Render(tl, text, height, color, r.overlayFramebuffer)
}

func (r *Renderer) RenderQuad(tex *graphics.Texture2D) {
	r.quadRenderer.Render(tex, graphics.DefaultFramebuffer)
}

func (r *Renderer) RenderBitangents(s *scene.Scene, c camera.Camera) {
	r.arrowRenderer.RenderBitangents(s, c, r.sceneFramebuffer)
}

func (r *Renderer) RenderNormals(s *scene.Scene, c camera.Camera) {
	r.arrowRenderer.RenderNormals(s, c, r.sceneFramebuffer)
}

func (r *Renderer) RenderTangents(s *scene.Scene, c camera.Camera) {
	r.arrowRenderer.RenderTangents(s, c, r.sceneFramebuffer)
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
	r.meshRenderer.SetWireframe(wireframe)
}
