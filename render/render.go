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
	ArrowRenderer  *ArrowRenderer
	EffectRenderer *EffectRenderer

	sceneRenderTarget      *graphics.Texture2D
	sceneRenderTarget2     *graphics.Texture2D
	sceneDepthRenderTarget *graphics.Texture2D

	overlayRenderTarget *graphics.Texture2D

	Fog bool
	BlurRadius float32
}

func NewRenderer() (*Renderer, error) {
	var r Renderer

	r.MeshRenderer, _ = NewMeshRenderer()
	r.SkyboxRenderer = NewSkyboxRenderer()
	r.TextRenderer = NewTextRenderer()
	r.ArrowRenderer = NewArrowRenderer()
	r.EffectRenderer = NewEffectRenderer()

	w, h := 1920, 1080
	w, h = w/1, h/1

	r.sceneRenderTarget = graphics.NewTexture2D(graphics.ColorTexture, graphics.NearestFilter, graphics.EdgeClampWrap, w, h, false)
	r.sceneRenderTarget2 = graphics.NewTexture2D(graphics.ColorTexture, graphics.NearestFilter, graphics.EdgeClampWrap, w, h, false)
	r.sceneDepthRenderTarget = graphics.NewTexture2D(graphics.DepthTexture, graphics.NearestFilter, graphics.EdgeClampWrap, w, h, false)

	r.overlayRenderTarget = graphics.NewTexture2D(graphics.ColorTexture, graphics.NearestFilter, graphics.EdgeClampWrap, w, h, false)

	return &r, nil
}

func (r *Renderer) RenderScene(s *scene.Scene, c camera.Camera) {
	if s.Skybox != nil {
		r.SkyboxRenderer.Render(s.Skybox, c, r.sceneRenderTarget)
	}

	r.MeshRenderer.Render(s, c, r.sceneRenderTarget, r.sceneDepthRenderTarget)
	if r.Fog {
		r.EffectRenderer.RenderFog(c, r.sceneDepthRenderTarget, r.sceneRenderTarget)
	}
	if r.BlurRadius > 0 {
		r.EffectRenderer.RenderGaussianBlur(r.sceneRenderTarget, r.sceneRenderTarget2, r.BlurRadius)
	}
}

func (r *Renderer) RenderText(tl math.Vec2, text string, height float32, just Justification) {
	color := math.Vec3{1, 1, 0}
	r.TextRenderer.Render(tl, text, height, color, just, r.overlayRenderTarget)
}

func (r *Renderer) RenderBitangents(s *scene.Scene, c camera.Camera) {
	r.ArrowRenderer.RenderBitangents(s, c, r.sceneRenderTarget, r.sceneDepthRenderTarget)
}

func (r *Renderer) RenderNormals(s *scene.Scene, c camera.Camera) {
	r.ArrowRenderer.RenderNormals(s, c, r.sceneRenderTarget, r.sceneDepthRenderTarget)
}

func (r *Renderer) RenderTangents(s *scene.Scene, c camera.Camera) {
	r.ArrowRenderer.RenderTangents(s, c, r.sceneRenderTarget, r.sceneDepthRenderTarget)
}

func (r *Renderer) Clear() {
	graphics.Clear(math.Vec4{0, 0, 0, 0})
	r.sceneRenderTarget.Clear(math.Vec4{0, 0, 0, 0})
	r.sceneRenderTarget2.Clear(math.Vec4{0, 0, 0, 0})
	r.sceneDepthRenderTarget.Clear(math.Vec4{1, 1, 1, 1})
	r.overlayRenderTarget.Clear(math.Vec4{0, 0, 0, 0})
}

func (r *Renderer) Render() {
	r.sceneRenderTarget.Display(graphics.NoBlending)
	r.overlayRenderTarget.Display(graphics.AlphaBlending)
}

func (r *Renderer) SetWireframe(wireframe bool) {
	r.MeshRenderer.Wireframe = wireframe
}
