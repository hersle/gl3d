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
	meshRenderer   *MeshRenderer
	skyboxRenderer *SkyboxRenderer
	textRenderer   *TextRenderer
	quadRenderer   *QuadRenderer
	arrowRenderer  *ArrowRenderer

	sceneFramebuffer       *graphics.Framebuffer
	sceneRenderTarget      *graphics.Texture2D
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

	w, h := 1920, 1080
	w, h = w/1, h/1

	r.sceneRenderTarget = graphics.NewTexture2D(graphics.NearestFilter, graphics.EdgeClampWrap, w, h, gl.RGBA8)
	r.sceneDepthRenderTarget = graphics.NewTexture2D(graphics.NearestFilter, graphics.EdgeClampWrap, w, h, gl.DEPTH_COMPONENT16)
	r.sceneFramebuffer = graphics.NewFramebuffer()
	r.sceneFramebuffer.Attach(r.sceneRenderTarget)
	r.sceneFramebuffer.Attach(r.sceneDepthRenderTarget)

	r.overlayRenderTarget = graphics.NewTexture2D(graphics.NearestFilter, graphics.EdgeClampWrap, w, h, gl.RGBA8)
	r.overlayFramebuffer = graphics.NewFramebuffer()
	r.overlayFramebuffer.Attach(r.overlayRenderTarget)

	return &r, nil
}

func (r *Renderer) RenderScene(s *scene.Scene, c camera.Camera) {
	if s.Skybox != nil {
		r.skyboxRenderer.Render(s.Skybox, c, r.sceneFramebuffer)
	}

	r.meshRenderer.Render(s, c, r.sceneFramebuffer)
}

func (r *Renderer) RenderText(tl math.Vec2, text string, height float32) {
	r.textRenderer.Render(tl, text, height, r.overlayFramebuffer)
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
