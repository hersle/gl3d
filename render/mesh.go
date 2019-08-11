package render

import (
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/material"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/scene"
	"github.com/hersle/gl3d/utils"
	"image"
	"fmt"
	gomath "math"
)

type MeshRenderer struct {
	depthProg *MeshProgram
	ambientProg *MeshProgram
	ssaoProg *ssaoProgram
	ssaoBlurProg *ssaoBlurProgram
	pointLitProg *MeshProgram
	spotLitProg *MeshProgram
	dirLitProg *MeshProgram

	shadowMapRenderer *ShadowMapRenderer

	resources *meshResourceManager

	renderOpts *graphics.RenderOptions

	pointLightMesh *object.Mesh
	spotLightMesh *object.Mesh

	normalMatrices []math.Mat4

	cullCache []bool

	ShadowKernelSize int

	MaterialAmbientEnabled bool
	MaterialDiffuseEnabled bool
	MaterialSpecularEnabled bool
	MaterialAlphaEnabled bool
	MaterialNormalEnabled bool
	ShadowsEnabled bool
	Wireframe bool

	AmbientOcclusion bool
	randomDirectionMap *graphics.Texture2D
	aoMap *graphics.Texture2D
	blurredAoMap *graphics.Texture2D
}

type meshResourceManager struct {
	vbos map[*object.Vertex]*graphics.VertexBuffer
	ibos map[*int32]*graphics.IndexBuffer

	textures map[image.Image]*graphics.Texture2D

	pointLightShadowMaps map[int]*graphics.CubeMap
	spotLightShadowMaps  map[int]*graphics.Texture2D
	dirLightShadowMaps   map[int]*graphics.Texture2D

	// default textures
	blueTexture *graphics.Texture2D
	whiteTexture *graphics.Texture2D
	blackTexture *graphics.Texture2D
	whiteCubeMap *graphics.CubeMap
}

type ShadowMapRenderer struct {
	resources *meshResourceManager

	shadowSp1             *ShadowMapProgram
	shadowSp2             *ShadowMapProgram
	shadowSp3             *ShadowMapProgram
	shadowRenderOpts      *graphics.RenderOptions

	shadowProjViewMat math.Mat4
}

type MeshProgram struct {
	*graphics.Program

	Position *graphics.Input
	TexCoord *graphics.Input
	Normal   *graphics.Input
	Tangent  *graphics.Input

	Color *graphics.Output
	Depth *graphics.Output

	ModelMatrix      *graphics.Uniform
	ViewMatrix       *graphics.Uniform
	ProjectionMatrix *graphics.Uniform
	NormalMatrix     *graphics.Uniform

	MaterialAmbient     *graphics.Uniform
	MaterialAmbientMap  *graphics.Uniform
	MaterialDiffuse     *graphics.Uniform
	MaterialDiffuseMap  *graphics.Uniform
	MaterialSpecular    *graphics.Uniform
	MaterialSpecularMap *graphics.Uniform
	MaterialShine       *graphics.Uniform
	MaterialAlpha       *graphics.Uniform
	MaterialAlphaMap    *graphics.Uniform
	MaterialBumpMap     *graphics.Uniform
	MaterialBumpMapWidth  *graphics.Uniform
	MaterialBumpMapHeight *graphics.Uniform

	LightPosition          *graphics.Uniform
	LightDirection         *graphics.Uniform
	LightColor             *graphics.Uniform
	LightAttenuation       *graphics.Uniform
	LightCosAngle          *graphics.Uniform

	ShadowProjectionViewMatrix *graphics.Uniform
	ShadowMap              *graphics.Uniform
	ShadowFar              *graphics.Uniform
	ShadowKernelSize       *graphics.Uniform

	AoMap *graphics.Uniform
}

type ShadowMapProgram struct {
	*graphics.Program

	Position         *graphics.Input

	ModelMatrix      *graphics.Uniform
	ViewMatrix       *graphics.Uniform
	ProjectionMatrix *graphics.Uniform

	LightPosition    *graphics.Uniform
	LightFar         *graphics.Uniform

	ProjViewMats     []*graphics.Uniform

	Depth *graphics.Output
}

type ssaoProgram struct {
	*graphics.Program

	Position *graphics.Input
	Color               *graphics.Output
	DepthMap            *graphics.Uniform
	DepthMapWidth       *graphics.Uniform
	DepthMapHeight      *graphics.Uniform
	ProjectionMatrix *graphics.Uniform
	InvProjectionMatrix *graphics.Uniform
	Directions          []*graphics.Uniform
	DirectionMap        *graphics.Uniform
}

type ssaoBlurProgram struct {
	*graphics.Program

	Color *graphics.Output
	aoMap *graphics.Uniform
	aoMapWidth *graphics.Uniform
	aoMapHeight *graphics.Uniform
}

func NewMeshRenderer() (*MeshRenderer, error) {
	var r MeshRenderer

	r.depthProg = NewMeshProgram("DEPTH")
	r.ssaoProg = NewSSAOProgram()
	r.ssaoBlurProg = NewSSAOBlurProgram()
	r.ambientProg = NewMeshProgram("AMBIENT")
	r.pointLitProg = NewMeshProgram("POINT", "SHADOW", "PCF")
	r.spotLitProg = NewMeshProgram("SPOT", "SHADOW", "PCF")
	r.dirLitProg = NewMeshProgram("DIR", "SHADOW", "PCF")

	r.resources = newMeshResourceManager()

	r.shadowMapRenderer = NewShadowMapRenderer(r.resources) // share resources

	r.renderOpts = graphics.NewRenderOptions()

	geo := object.NewSphere(math.Vec3{0, 0, 0}, 0.1).Geometry(6)
	mtl := material.NewDefaultMaterial("")
	mtl.Ambient = math.Vec3{1, 1, 1}
	r.pointLightMesh = object.NewMesh(geo, mtl)

	geo = object.NewCone(math.Vec3{0, 0, -1}, math.Vec3{0, 0, 0}, 0.5).Geometry(6)
	r.spotLightMesh = object.NewMesh(geo, mtl)

	r.MaterialAmbientEnabled = true
	r.MaterialDiffuseEnabled = true
	r.MaterialSpecularEnabled = true
	r.MaterialAlphaEnabled = true
	r.MaterialNormalEnabled = true
	r.ShadowsEnabled = true
	r.AmbientOcclusion = true

	w := 1920 / 1
	h := 1080 / 1
	r.randomDirectionMap = graphics.NewColorTexture2D(graphics.NearestFilter, graphics.RepeatWrap, w, h, 3, 32, true, false)
	directions := make([]math.Vec3, w*h)
	for i := 0; i < w*h; i++ {
		directions[i] = utils.RandomDirection()
	}
	r.randomDirectionMap.SetData(0, 0, w, h, directions)

	r.aoMap = graphics.NewColorTexture2D(graphics.LinearFilter, graphics.RepeatWrap, 1920 / 1, 1080 / 1, 1, 8, false, false)
	r.blurredAoMap = graphics.NewColorTexture2D(graphics.LinearFilter, graphics.RepeatWrap, 1920 / 1, 1080 / 1, 1, 8, false, false)

	return &r, nil
}

func NewShadowMapRenderer(resources *meshResourceManager) *ShadowMapRenderer {
	var r ShadowMapRenderer

	if resources == nil {
		r.resources = newMeshResourceManager()
	} else {
		r.resources = resources
	}

	r.shadowSp1 = NewShadowMapProgram("POINT") // point light
	r.shadowSp2 = NewShadowMapProgram("SPOT") // spot light
	r.shadowSp3 = NewShadowMapProgram("DIR") // directional light

	r.shadowRenderOpts = graphics.NewRenderOptions()
	r.shadowRenderOpts.DepthTest = graphics.LessDepthTest
	r.shadowRenderOpts.Culling = graphics.BackCulling
	r.shadowRenderOpts.Primitive = graphics.Triangles

	return &r
}

func NewMeshProgram(defines ...string) *MeshProgram {
	var sp MeshProgram

	vFile := "render/shaders/meshvshadertemplate.glsl" // TODO: make independent from executable directory
	fFile := "render/shaders/meshfshadertemplate.glsl" // TODO: make independent from executable directory
	sp.Program = graphics.ReadProgram(vFile, fFile, "", defines...)

	sp.Position = sp.InputByName("position")
	sp.TexCoord = sp.InputByName("texCoordV")
	sp.Normal = sp.InputByName("normalV")
	sp.Tangent = sp.InputByName("tangentV")

	sp.Color = sp.OutputColorByName("fragColor")
	sp.Depth = sp.OutputDepth()

	sp.ModelMatrix = sp.UniformByName("modelMatrix")
	sp.ViewMatrix = sp.UniformByName("viewMatrix")
	sp.ProjectionMatrix = sp.UniformByName("projectionMatrix")
	sp.NormalMatrix = sp.UniformByName("normalMatrix")

	sp.MaterialAmbient = sp.UniformByName("materialAmbient")
	sp.MaterialAmbientMap = sp.UniformByName("materialAmbientMap")
	sp.MaterialDiffuse = sp.UniformByName("materialDiffuse")
	sp.MaterialDiffuseMap = sp.UniformByName("materialDiffuseMap")
	sp.MaterialSpecular = sp.UniformByName("materialSpecular")
	sp.MaterialSpecularMap = sp.UniformByName("materialSpecularMap")
	sp.MaterialShine = sp.UniformByName("materialShine")
	sp.MaterialAlpha = sp.UniformByName("materialAlpha")
	sp.MaterialAlphaMap = sp.UniformByName("materialAlphaMap")
	sp.MaterialBumpMap = sp.UniformByName("materialBumpMap")
	sp.MaterialBumpMapWidth = sp.UniformByName("materialBumpMapWidth")
	sp.MaterialBumpMapHeight = sp.UniformByName("materialBumpMapHeight")

	sp.LightPosition = sp.UniformByName("lightPosition")
	sp.LightDirection = sp.UniformByName("lightDirection")
	sp.LightColor = sp.UniformByName("lightColor")
	sp.LightAttenuation = sp.UniformByName("lightAttenuation")
	sp.LightCosAngle = sp.UniformByName("lightCosAng")

	sp.ShadowProjectionViewMatrix = sp.UniformByName("shadowProjectionViewMatrix")
	sp.ShadowMap = sp.UniformByName("shadowMap")
	sp.ShadowFar = sp.UniformByName("lightFar")
	sp.ShadowKernelSize = sp.UniformByName("kernelSize")

	sp.AoMap = sp.UniformByName("aoMap")

	return &sp
}

func NewShadowMapProgram(defines ...string) *ShadowMapProgram {
	var sp ShadowMapProgram

	vFile := "render/shaders/shadowmapvshadertemplate.glsl" // TODO: make independent from executable directory
	fFile := "render/shaders/shadowmapfshadertemplate.glsl" // TODO: make independent from executable directory
	gFile := "render/shaders/shadowmapgshadertemplate.glsl" // TODO: make independent from executable directory
	sp.Program = graphics.ReadProgram(vFile, fFile, gFile, defines...)

	sp.ModelMatrix = sp.UniformByName("modelMatrix")
	sp.ViewMatrix = sp.UniformByName("viewMatrix")
	sp.ProjectionMatrix = sp.UniformByName("projectionMatrix")
	sp.LightPosition = sp.UniformByName("lightPosition")
	sp.LightFar = sp.UniformByName("far")
	sp.Position = sp.InputByName("position")

	sp.ProjViewMats = make([]*graphics.Uniform, 6)
	for i := 0; i < 6; i++ {
		name := fmt.Sprintf("projectionViewMatrices[%d]", i)
		sp.ProjViewMats[i] = sp.UniformByName(name)
	}

	sp.Depth = sp.OutputDepth()

	return &sp
}

func NewSSAOProgram() *ssaoProgram {
	var sp ssaoProgram

	vFile := "render/shaders/ssaovshader.glsl" // TODO: make independent from executable directory
	fFile := "render/shaders/ssaofshader.glsl" // TODO: make independent from executable directory
	sp.Program = graphics.ReadProgram(vFile, fFile, "")

	sp.Position = sp.InputByName("position")
	sp.Color = sp.OutputColorByName("fragColor")
	sp.DepthMap = sp.UniformByName("depthMap")
	sp.DepthMapWidth = sp.UniformByName("depthMapWidth")
	sp.DepthMapHeight = sp.UniformByName("depthMapHeight")
	sp.ProjectionMatrix = sp.UniformByName("projectionMatrix")
	sp.InvProjectionMatrix = sp.UniformByName("invProjectionMatrix")
	sp.DirectionMap = sp.UniformByName("directionMap")

	return &sp
}

func NewSSAOBlurProgram() *ssaoBlurProgram {
	var sp ssaoBlurProgram

	vFile := "render/shaders/ssaoblurvshader.glsl" // TODO: make independent from executable directory
	fFile := "render/shaders/ssaoblurfshader.glsl" // TODO: make independent from executable directory
	sp.Program = graphics.ReadProgram(vFile, fFile, "")

	sp.Color = sp.OutputColorByName("fragColor")
	sp.aoMap = sp.UniformByName("aoMap")
	sp.aoMapWidth = sp.UniformByName("aoMapWidth")
	sp.aoMapHeight = sp.UniformByName("aoMapHeight")

	return &sp
}

func (r *MeshRenderer) Render(s *scene.Scene, c camera.Camera, colorTexture, depthTexture *graphics.Texture2D) {
	r.renderOpts.Culling = graphics.BackCulling
	r.renderOpts.Primitive = graphics.Triangles

	r.depthProg.Depth.Set(depthTexture)
	r.ambientProg.Color.Set(colorTexture)
	r.pointLitProg.Color.Set(colorTexture)
	r.spotLitProg.Color.Set(colorTexture)
	r.dirLitProg.Color.Set(colorTexture)
	r.ambientProg.Depth.Set(depthTexture)
	r.pointLitProg.Depth.Set(depthTexture)
	r.spotLitProg.Depth.Set(depthTexture)
	r.dirLitProg.Depth.Set(depthTexture)

	if r.Wireframe {
		r.renderOpts.Primitive = graphics.TriangleOutlines
	} else {
		r.renderOpts.Primitive = graphics.Triangles
	}

	r.preparationPass(s, c)
	r.shadowPass(s)
	r.depthPass(s, c)
	r.ssaoPass(depthTexture, c)
	r.ambientPass(s, c)
	r.lightPass(s, c)
}

func (r *MeshRenderer) ssaoPass(depthMap *graphics.Texture2D, c camera.Camera) {
	if !r.AmbientOcclusion {
		return
	}

	r.renderOpts.Primitive = graphics.TriangleFan
	r.renderOpts.Blending = graphics.NoBlending
	r.ssaoProg.Color.Set(r.aoMap)
	r.ssaoProg.DepthMap.Set(depthMap)
	r.ssaoProg.DepthMapWidth.Set(depthMap.Width())
	r.ssaoProg.DepthMapHeight.Set(depthMap.Height())

	var mat math.Mat4
	mat.Identity()
	mat.Mult(c.ProjectionMatrix())
	mat.Invert()
	r.ssaoProg.InvProjectionMatrix.Set(&mat)
	r.ssaoProg.ProjectionMatrix.Set(c.ProjectionMatrix())
	r.ssaoProg.DirectionMap.Set(r.randomDirectionMap)

	r.ssaoProg.Render(4, r.renderOpts)

	r.blurAoMap()
}

func (r *MeshRenderer) blurAoMap() {
	if !r.AmbientOcclusion {
		return
	}

	r.ssaoBlurProg.aoMap.Set(r.aoMap)
	r.ssaoBlurProg.aoMapWidth.Set(r.aoMap.Width())
	r.ssaoBlurProg.aoMapHeight.Set(r.aoMap.Height())
	r.ssaoBlurProg.Color.Set(r.blurredAoMap)

	var opts graphics.RenderOptions
	opts.Primitive = graphics.TriangleFan
	r.ssaoBlurProg.Render(4, &opts)
}

func (r *MeshRenderer) preparationPass(s *scene.Scene, c camera.Camera) {
	// precalculate normal matrices for use in multiple rendering passes
	if len(r.normalMatrices) > len(s.Meshes) {
		 r.normalMatrices = r.normalMatrices[:len(s.Meshes)]
	} else {
		r.normalMatrices = make([]math.Mat4, len(s.Meshes))
	}
	subMeshCount := 0
	for i, m := range s.Meshes {
		normalMatrix := &r.normalMatrices[i]
		normalMatrix.Identity()
		normalMatrix.Mult(c.ViewMatrix())
		normalMatrix.Mult(m.WorldMatrix())
		normalMatrix.Invert()
		normalMatrix.Transpose()
		subMeshCount += len(m.SubMeshes)
	}

	// precalculate culling for use in multiple rendering passes
	if len(r.cullCache) > subMeshCount {
		r.cullCache = r.cullCache[:subMeshCount]
	} else {
		r.cullCache = make([]bool, subMeshCount)
	}
	i := 0
	for _, m := range s.Meshes {
		for _, sm := range m.SubMeshes {
			r.cullCache[i] = c.Cull(sm)
			i++
		}
	}

	r.ambientProg.ShadowKernelSize.Set(r.ShadowKernelSize)
	r.pointLitProg.ShadowKernelSize.Set(r.ShadowKernelSize)
	r.spotLitProg.ShadowKernelSize.Set(r.ShadowKernelSize)
	r.dirLitProg.ShadowKernelSize.Set(r.ShadowKernelSize)
}

func (r *MeshRenderer) depthPass(s *scene.Scene, c camera.Camera) {
	r.renderOpts.Blending = graphics.NoBlending
	r.renderOpts.DepthTest = graphics.LessDepthTest

	r.setCamera(r.depthProg, c)
	r.renderMeshes(s, c, r.depthProg)

	for _, l := range s.PointLights {
		r.ambientProg.LightColor.Set(l.Color)
		r.pointLightMesh.Place(l.Position)
		r.setMesh(r.ambientProg, r.pointLightMesh)
		for _, subMesh := range r.pointLightMesh.SubMeshes {
			r.setSubMesh(r.ambientProg, subMesh)
			r.ambientProg.Render(subMesh.Geo.Inds, r.renderOpts)
		}
	}

	for _, l := range s.SpotLights {
		r.ambientProg.LightColor.Set(l.Color)
		r.spotLightMesh.Place(l.Position)
		r.spotLightMesh.Orient(l.UnitX, l.UnitY)
		r.setMesh(r.ambientProg, r.spotLightMesh)
		for _, subMesh := range r.spotLightMesh.SubMeshes {
			r.setSubMesh(r.ambientProg, subMesh)
			r.ambientProg.Render(subMesh.Geo.Inds, r.renderOpts)
		}
	}
}

func (r *MeshRenderer) ambientPass(s *scene.Scene, c camera.Camera) {
	r.renderOpts.Primitive = graphics.Triangles
	r.renderOpts.Blending = graphics.NoBlending
	r.renderOpts.DepthTest = graphics.EqualDepthTest

	if r.AmbientOcclusion {
		r.ambientProg.AoMap.Set(r.blurredAoMap)
	} else {
		r.ambientProg.AoMap.Set(r.resources.whiteTexture)
	}
	r.ambientProg.LightColor.Set(s.AmbientLight.Color)
	r.setCamera(r.ambientProg, c)
	r.renderMeshes(s, c, r.ambientProg)

	// render light source
	// TODO: do with shaders instead for fancier effects?
	for _, l := range s.PointLights {
		r.ambientProg.LightColor.Set(l.Color)
		r.pointLightMesh.Place(l.Position)
		r.setMesh(r.ambientProg, r.pointLightMesh)
		for _, subMesh := range r.pointLightMesh.SubMeshes {
			r.setSubMesh(r.ambientProg, subMesh)
			r.ambientProg.Render(subMesh.Geo.Inds, r.renderOpts)
		}
	}

	for _, l := range s.SpotLights {
		r.ambientProg.LightColor.Set(l.Color)
		r.spotLightMesh.Place(l.Position)
		r.spotLightMesh.Orient(l.UnitX, l.UnitY)
		r.setMesh(r.ambientProg, r.spotLightMesh)
		for _, subMesh := range r.spotLightMesh.SubMeshes {
			r.setSubMesh(r.ambientProg, subMesh)
			r.ambientProg.Render(subMesh.Geo.Inds, r.renderOpts)
		}
	}
}

func (r *MeshRenderer) lightPass(s *scene.Scene, c camera.Camera) {
	r.renderOpts.DepthTest = graphics.EqualDepthTest
	r.renderOpts.Blending = graphics.AdditiveBlending // add to framebuffer contents

	r.setCamera(r.pointLitProg, c)
	for _, l := range s.PointLights {
		r.setPointLight(r.pointLitProg, l)
		r.renderMeshes(s, c, r.pointLitProg)
	}

	r.setCamera(r.spotLitProg, c)
	for _, l := range s.SpotLights {
		r.setSpotLight(r.spotLitProg, l)
		r.renderMeshes(s, c, r.spotLitProg)
	}

	r.setCamera(r.dirLitProg, c)
	for _, l := range s.DirectionalLights {
		r.setDirectionalLight(r.dirLitProg, l)
		r.renderMeshes(s, c, r.dirLitProg)
	}
}

func (r *MeshRenderer) renderMeshes(s *scene.Scene, c camera.Camera, sp *MeshProgram) {
	j := 0
	for i, m := range s.Meshes {
		r.setMesh(sp, m)
		sp.NormalMatrix.Set(&r.normalMatrices[i])
		for _, sm := range m.SubMeshes {
			if !r.cullCache[j] {
				r.setSubMesh(sp, sm)
				sp.Render(sm.Geo.Inds, r.renderOpts)
			}
			j++
		}
	}
}

func (r *MeshRenderer) setCamera(sp *MeshProgram, c camera.Camera) {
	sp.ViewMatrix.Set(c.ViewMatrix())
	sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *MeshRenderer) setMesh(sp *MeshProgram, m *object.Mesh) {
	sp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *MeshRenderer) setSubMesh(sp *MeshProgram, sm *object.SubMesh) {
	mtl := sm.Mtl

	if r.MaterialAmbientEnabled {
		tex := r.resources.texture(mtl.AmbientMap)
		sp.MaterialAmbient.Set(mtl.Ambient)
		sp.MaterialAmbientMap.Set(tex)
	} else {
		sp.MaterialAmbient.Set(math.Vec3{0, 0, 0})
		sp.MaterialAmbientMap.Set(r.resources.blackTexture)
	}

	if r.MaterialDiffuseEnabled {
		tex := r.resources.texture(mtl.DiffuseMap)
		sp.MaterialDiffuse.Set(mtl.Diffuse)
		sp.MaterialDiffuseMap.Set(tex)
	} else {
		sp.MaterialDiffuse.Set(math.Vec3{0, 0, 0})
		sp.MaterialDiffuseMap.Set(r.resources.blackTexture)
	}

	if r.MaterialSpecularEnabled {
		tex := r.resources.texture(mtl.SpecularMap)
		sp.MaterialSpecular.Set(mtl.Specular)
		sp.MaterialSpecularMap.Set(tex)
		sp.MaterialShine.Set(mtl.Shine)
	} else {
		sp.MaterialSpecular.Set(math.Vec3{0, 0, 0})
		sp.MaterialSpecularMap.Set(r.resources.blackTexture)
		sp.MaterialShine.Set(float32(0))
	}

	if r.MaterialAlphaEnabled {
		tex := r.resources.texture(mtl.AlphaMap)
		sp.MaterialAlpha.Set(mtl.Alpha)
		sp.MaterialAlphaMap.Set(tex)
	} else {
		sp.MaterialAlpha.Set(float32(1.0))
		sp.MaterialAlphaMap.Set(r.resources.whiteTexture)
	}

	if r.MaterialNormalEnabled {
		tex := r.resources.texture(mtl.BumpMap)
		sp.MaterialBumpMap.Set(tex)
		sp.MaterialBumpMapWidth.Set(tex.Width())
		sp.MaterialBumpMapHeight.Set(tex.Height())
	} else {
		sp.MaterialBumpMap.Set(r.resources.whiteTexture)
		sp.MaterialBumpMapWidth.Set(r.resources.whiteTexture.Width())
		sp.MaterialBumpMapHeight.Set(r.resources.whiteTexture.Height())
	}

	vbo := r.resources.vertexBuffer(sm)
	ibo := r.resources.indexBuffer(sm)

	sp.Position.SetSourceVertex(vbo, 0)
	sp.Normal.SetSourceVertex(vbo, 2)
	sp.TexCoord.SetSourceVertex(vbo, 1)
	sp.Tangent.SetSourceVertex(vbo, 3)
	sp.SetIndices(ibo)
}

func (r *MeshRenderer) setAmbientLight(sp *MeshProgram, l *light.AmbientLight) {
	sp.LightColor.Set(l.Color.Scale(l.Intensity))
}

func (r *MeshRenderer) setPointLight(sp *MeshProgram, l *light.PointLight) {
	sp.LightPosition.Set(l.Position)
	sp.LightColor.Set(l.Color.Scale(l.Intensity))
	if r.ShadowsEnabled && l.CastShadows {
		sp.ShadowFar.Set(l.ShadowFar)
		smap := r.resources.pointShadowMap(l)
		sp.ShadowMap.Set(smap)
	} else {
		sp.ShadowFar.Set(float32(100))
		sp.ShadowMap.Set(r.resources.whiteCubeMap)
	}
	sp.LightAttenuation.Set(l.Attenuation)
}

func (r *MeshRenderer) setSpotLight(sp *MeshProgram, l *light.SpotLight) {
	sp.LightPosition.Set(l.Position)
	sp.LightDirection.Set(l.Forward())
	sp.LightColor.Set(l.Color.Scale(l.Intensity))
	sp.LightAttenuation.Set(l.Attenuation)
	sp.LightCosAngle.Set(float32(gomath.Cos(float64(l.FOV/2))))

	if r.ShadowsEnabled && l.CastShadows {
		var m math.Mat4
		m.Identity()
		m.Mult(l.ProjectionMatrix())
		m.Mult(l.ViewMatrix())
		sp.ShadowProjectionViewMatrix.Set(&m)
		sp.ShadowFar.Set(l.PerspectiveCamera.Far)
		smap := r.resources.spotShadowMap(l)
		sp.ShadowMap.Set(smap)
	} else {
		sp.ShadowFar.Set(float32(100))
		sp.ShadowMap.Set(r.resources.whiteTexture)
	}
}

func (r *MeshRenderer) setDirectionalLight(sp *MeshProgram, l *light.DirectionalLight) {
	sp.LightDirection.Set(l.Forward())
	sp.LightColor.Set(l.Color.Scale(l.Intensity))
	sp.LightAttenuation.Set(float32(0))

	if r.ShadowsEnabled && l.CastShadows {
		var m math.Mat4
		m.Identity()
		m.Mult(l.ProjectionMatrix())
		m.Mult(l.ViewMatrix())
		sp.ShadowProjectionViewMatrix.Set(&m)
		smap := r.resources.dirShadowMap(l)
		sp.ShadowMap.Set(smap)
	} else {
		sp.ShadowFar.Set(float32(100))
		sp.ShadowMap.Set(r.resources.whiteTexture)
	}
}

// shadow stuff below

func (r *ShadowMapRenderer) setCamera(sp *ShadowMapProgram, c camera.Camera) {
	sp.ViewMatrix.Set(c.ViewMatrix())
	sp.ProjectionMatrix.Set(c.ProjectionMatrix())

	switch c.(type) {
	case *camera.PerspectiveCamera:
		c := c.(*camera.PerspectiveCamera)
		sp.LightFar.Set(c.Far)
		sp.LightPosition.Set(c.Position)
	case *camera.OrthoCamera:
		c := c.(*camera.OrthoCamera)
		sp.LightFar.Set(c.Far)
		sp.LightPosition.Set(c.Position)
	}
}

func (r *ShadowMapRenderer) setMesh(sp *ShadowMapProgram, m *object.Mesh) {
	sp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *ShadowMapRenderer) setSubMesh(sp *ShadowMapProgram, sm *object.SubMesh) {
	vbo := r.resources.vertexBuffer(sm)
	ibo := r.resources.indexBuffer(sm)

	sp.Position.SetSourceVertex(vbo, 0)
	sp.SetIndices(ibo)
}

func (r *MeshRenderer) shadowPass(s *scene.Scene) {
	for _, l := range s.PointLights {
		if l.CastShadows {
			smap := r.resources.pointShadowMap(l)
			r.shadowMapRenderer.renderPointLightShadowMap(s, l, smap)
		}
	}
	for _, l := range s.SpotLights {
		if l.CastShadows {
			smap := r.resources.spotShadowMap(l)
			r.shadowMapRenderer.renderSpotLightShadowMap(s, l, smap)
		}
	}
	for _, l := range s.DirectionalLights {
		if l.CastShadows {
			smap := r.resources.dirShadowMap(l)
			r.shadowMapRenderer.renderDirectionalLightShadowMap(s, l, smap)
		}
	}
}

// render shadow map to l's shadow map
func (r *ShadowMapRenderer) renderPointLightShadowMap(s *scene.Scene, l *light.PointLight, smap *graphics.CubeMap) {
	// TODO: re-render also when objects have moved
	/*
		if !l.DirtyShadowMap {
			return
		}
	*/

	forwards := []math.Vec3{
		math.Vec3{+1, 0, 0},
		math.Vec3{-1, 0, 0},
		math.Vec3{0, +1, 0},
		math.Vec3{0, -1, 0},
		math.Vec3{0, 0, +1},
		math.Vec3{0, 0, -1},
	}
	ups := []math.Vec3{
		math.Vec3{0, -1, 0},
		math.Vec3{0, -1, 0},
		math.Vec3{0, 0, +1},
		math.Vec3{0, 0, -1},
		math.Vec3{0, -1, 0},
		math.Vec3{0, -1, 0},
	}

	c := camera.NewPerspectiveCamera(90, 1, 0.1, l.ShadowFar)
	c.Place(l.Position)

	for face := 0; face < 6; face++ {
		c.SetForwardUp(forwards[face], ups[face])
		r.shadowProjViewMat.Identity()
		r.shadowProjViewMat.Mult(c.ProjectionMatrix())
		r.shadowProjViewMat.Mult(c.ViewMatrix())
		r.shadowSp1.ProjViewMats[face].Set(&r.shadowProjViewMat)
	}

	smap.Clear(math.Vec4{1, 1, 1, 1})
	r.shadowSp1.Depth.Set(smap)

	r.setCamera(r.shadowSp1, c)

	for _, m := range s.Meshes {
		r.setMesh(r.shadowSp1, m)
		for _, subMesh := range m.SubMeshes {
			r.setSubMesh(r.shadowSp1, subMesh)

			r.shadowSp1.Render(subMesh.Geo.Inds, r.shadowRenderOpts)
		}
	}

	//l.DirtyShadowMap = false
}

func (r *ShadowMapRenderer) renderSpotLightShadowMap(s *scene.Scene, l *light.SpotLight, smap *graphics.Texture2D) {
	// TODO: re-render also when objects have moved
	//if !l.DirtyShadowMap {
		//return
	//}

	r.shadowSp2.Depth.Set(smap)
	smap.Clear(math.Vec4{1, 1, 1, 1})
	r.setCamera(r.shadowSp2, &l.PerspectiveCamera)

	for _, m := range s.Meshes {
		r.setMesh(r.shadowSp2, m)
		for _, subMesh := range m.SubMeshes {
			if !l.PerspectiveCamera.Cull(subMesh) {
				r.setSubMesh(r.shadowSp2, subMesh)

				r.shadowSp2.Render(subMesh.Geo.Inds, r.shadowRenderOpts)
			}
		}
	}

	//l.DirtyShadowMap = false
}

func (r *ShadowMapRenderer) renderDirectionalLightShadowMap(s *scene.Scene, l *light.DirectionalLight, smap *graphics.Texture2D) {
	// TODO: re-render also when objects have moved
	//if !l.DirtyShadowMap {
		//return
	//}

	r.shadowSp3.Depth.Set(smap)
	smap.Clear(math.Vec4{1, 1, 1, 1})
	r.setCamera(r.shadowSp3, &l.OrthoCamera)

	for _, m := range s.Meshes {
		r.setMesh(r.shadowSp3, m)
		for _, subMesh := range m.SubMeshes {
			if !l.OrthoCamera.Cull(subMesh) {
				r.setSubMesh(r.shadowSp3, subMesh)

				r.shadowSp3.Render(subMesh.Geo.Inds, r.shadowRenderOpts)
			}
		}
	}

	//l.DirtyShadowMap = false
}

func pointLightInteracts(l *light.PointLight, sm *object.SubMesh) bool {
	sphere := sm.BoundingSphere()
	dist := l.Position.Sub(sphere.Center).Length()
	if dist < sphere.Radius {
		return true
	}
	dist = dist - sphere.Radius
	return dist*dist < (1/0.05-1)/l.Attenuation
}

func newMeshResourceManager() *meshResourceManager {
	var rman meshResourceManager

	rman.vbos = make(map[*object.Vertex]*graphics.VertexBuffer)
	rman.ibos = make(map[*int32]*graphics.IndexBuffer)
	rman.pointLightShadowMaps = make(map[int]*graphics.CubeMap)
	rman.spotLightShadowMaps = make(map[int]*graphics.Texture2D)
	rman.dirLightShadowMaps = make(map[int]*graphics.Texture2D)
	rman.textures = make(map[image.Image]*graphics.Texture2D)

	rman.blueTexture = graphics.NewUniformTexture2D(math.Vec4{0.5, 0.5, 1, 0})
	rman.whiteTexture = graphics.NewUniformTexture2D(math.Vec4{1, 1, 1, 1})
	rman.blackTexture = graphics.NewUniformTexture2D(math.Vec4{0, 0, 0, 1})
	rman.whiteCubeMap = graphics.NewUniformCubeMap(math.Vec4{1, 1, 1, 1})

	return &rman
}

func (rman *meshResourceManager) vertexBuffer(sm *object.SubMesh) *graphics.VertexBuffer {
	vbo, found := rman.vbos[&sm.Geo.Verts[0]]
	if !found {
		vbo = graphics.NewVertexBuffer()
		vbo.SetData(sm.Geo.Verts, 0)
		rman.vbos[&sm.Geo.Verts[0]] = vbo
	}
	return vbo
}

func (rman *meshResourceManager) indexBuffer(sm *object.SubMesh) *graphics.IndexBuffer {
	ibo, found := rman.ibos[&sm.Geo.Faces[0]]
	if !found {
		ibo = graphics.NewIndexBuffer()
		ibo.SetData(sm.Geo.Faces, 0)
		rman.ibos[&sm.Geo.Faces[0]] = ibo
	}
	return ibo
}

func (rman *meshResourceManager) texture(img image.Image) *graphics.Texture2D {
	tex, found := rman.textures[img]
	if !found {
		tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.LinearFilter, graphics.RepeatWrap, img, true)
		rman.textures[img] = tex
	}
	return tex
}

func (rman *meshResourceManager) pointShadowMap(l *light.PointLight) *graphics.CubeMap {
	smap, found := rman.pointLightShadowMaps[l.ID]
	if !found {
		smap = graphics.NewCubeMap(graphics.DepthTexture, graphics.LinearFilter, 512, 512)
		rman.pointLightShadowMaps[l.ID] = smap
	}
	return smap
}

func (rman *meshResourceManager) spotShadowMap(l *light.SpotLight) *graphics.Texture2D {
	smap, found := rman.spotLightShadowMaps[l.ID]
	if !found {
		smap = graphics.NewTexture2D(graphics.DepthTexture, graphics.LinearFilter, graphics.BorderClampWrap, 512, 512, false)
		smap.SetBorderColor(math.NewVec4(1, 1, 1, 1))
		rman.spotLightShadowMaps[l.ID] = smap
	}
	return smap
}

func (rman *meshResourceManager) dirShadowMap(l *light.DirectionalLight) *graphics.Texture2D {
	smap, found := rman.dirLightShadowMaps[l.ID]
	if !found {
		smap = graphics.NewTexture2D(graphics.DepthTexture, graphics.LinearFilter, graphics.BorderClampWrap, 512, 512, false)
		smap.SetBorderColor(math.NewVec4(1, 1, 1, 1))
		rman.dirLightShadowMaps[l.ID] = smap
	}
	return smap
}
