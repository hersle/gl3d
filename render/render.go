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

	r.sceneRenderTarget = graphics.NewTexture2D(graphics.NearestFilter, graphics.EdgeClampWrap, gl.RGBA8, w, h)
	r.sceneDepthRenderTarget = graphics.NewTexture2D(graphics.NearestFilter, graphics.EdgeClampWrap, gl.DEPTH_COMPONENT16, w, h)
	r.sceneFramebuffer = graphics.NewFramebuffer()
	r.sceneFramebuffer.AttachTexture2D(graphics.ColorAttachment, r.sceneRenderTarget, 0)
	r.sceneFramebuffer.AttachTexture2D(graphics.DepthAttachment, r.sceneDepthRenderTarget, 0)

	r.overlayRenderTarget = graphics.NewTexture2D(graphics.NearestFilter, graphics.EdgeClampWrap, gl.RGBA8, w, h)
	r.overlayFramebuffer = graphics.NewFramebuffer()
	r.overlayFramebuffer.AttachTexture2D(graphics.ColorAttachment, r.overlayRenderTarget, 0)

	return &r, nil
}

func (r *Renderer) RenderScene(s *scene.Node, c camera.Camera) {
	/*
	if s.Skybox != nil {
		r.skyboxRenderer.Render(s.Skybox, c, r.sceneFramebuffer)
	}
	*/

	r.meshRenderer.Render(s, c, r.sceneFramebuffer)
}

func (r *Renderer) RenderText(tl math.Vec2, text string, height float32) {
	r.textRenderer.Render(tl, text, height, r.overlayFramebuffer)
}

func (r *Renderer) RenderQuad(tex *graphics.Texture2D) {
	r.quadRenderer.Render(tex, graphics.DefaultFramebuffer)
}

func (r *Renderer) RenderBitangents(s *scene.Node, c camera.Camera) {
	r.arrowRenderer.RenderBitangents(s, c, r.sceneFramebuffer)
}

func (r *Renderer) RenderNormals(s *scene.Node, c camera.Camera) {
	r.arrowRenderer.RenderNormals(s, c, r.sceneFramebuffer)
}

func (r *Renderer) RenderTangents(s *scene.Node, c camera.Camera) {
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
