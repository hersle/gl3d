package render

import (
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/material"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/scene"
	"image"
	"fmt"
	gomath "math"
)

var blueTexture *graphics.Texture2D
var whiteTexture *graphics.Texture2D
var blackTexture *graphics.Texture2D

var whiteCubeMap *graphics.CubeMap

type MeshRenderer struct {
	scene *scene.Scene
	camera camera.Camera

	depthProg *MeshProgram
	ambientProg *MeshProgram
	pointLitProg *MeshProgram
	spotLitProg *MeshProgram
	dirLitProg *MeshProgram

	renderOpts *graphics.RenderOptions

	vboCache map[*object.Vertex]int
	vbos     []*graphics.VertexBuffer
	ibos     []*graphics.IndexBuffer

	tex2ds map[image.Image]*graphics.Texture2D

	pointLightShadowMaps map[int]*graphics.CubeMap
	spotLightShadowMaps  map[int]*graphics.Texture2D
	dirLightShadowMaps   map[int]*graphics.Texture2D

	shadowSp1             *ShadowMapProgram
	shadowSp2             *ShadowMapProgram
	shadowSp3             *ShadowMapProgram
	shadowRenderOpts      *graphics.RenderOptions

	shadowProjViewMat math.Mat4

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

	AmbientOcclusionMap *graphics.Uniform
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

func NewMeshRenderer() (*MeshRenderer, error) {
	var r MeshRenderer

	r.depthProg = NewMeshProgram("DEPTH")
	r.ambientProg = NewMeshProgram("AMBIENT")
	r.pointLitProg = NewMeshProgram("POINT", "SHADOW", "PCF")
	r.spotLitProg = NewMeshProgram("SPOT", "SHADOW", "PCF")
	r.dirLitProg = NewMeshProgram("DIR", "SHADOW", "PCF")

	r.renderOpts = graphics.NewRenderOptions()

	r.vboCache = make(map[*object.Vertex]int)
	r.pointLightShadowMaps = make(map[int]*graphics.CubeMap)
	r.spotLightShadowMaps = make(map[int]*graphics.Texture2D)
	r.dirLightShadowMaps = make(map[int]*graphics.Texture2D)
	r.tex2ds = make(map[image.Image]*graphics.Texture2D)

	r.shadowSp1 = NewShadowMapProgram("POINT") // point light
	r.shadowSp2 = NewShadowMapProgram("SPOT") // spot light
	r.shadowSp3 = NewShadowMapProgram("DIR") // directional light

	r.shadowRenderOpts = graphics.NewRenderOptions()

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

	return &r, nil
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

	sp.AmbientOcclusionMap = sp.UniformByName("aoMap")

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

func (r *MeshRenderer) Prepare(s *scene.Scene, c camera.Camera, colorTexture, depthTexture *graphics.Texture2D) {
	r.depthProg.Depth.Set(depthTexture)
	r.ambientProg.Color.Set(colorTexture)
	r.pointLitProg.Color.Set(colorTexture)
	r.spotLitProg.Color.Set(colorTexture)
	r.dirLitProg.Color.Set(colorTexture)
	r.ambientProg.Depth.Set(depthTexture)
	r.pointLitProg.Depth.Set(depthTexture)
	r.spotLitProg.Depth.Set(depthTexture)
	r.dirLitProg.Depth.Set(depthTexture)

	r.scene = s
	r.camera = c

	r.renderOpts.Culling = graphics.BackCulling
	r.renderOpts.Primitive = graphics.Triangles

	if r.Wireframe {
		r.renderOpts.Primitive = graphics.TriangleOutlines
	} else {
		r.renderOpts.Primitive = graphics.Triangles
	}

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

func (r *MeshRenderer) RenderDepth() {
	r.renderOpts.Blending = graphics.NoBlending
	r.renderOpts.DepthTest = graphics.LessDepthTest

	r.setCamera(r.depthProg, r.camera)
	r.renderMeshes(r.depthProg)

	for _, l := range r.scene.PointLights {
		r.ambientProg.LightColor.Set(l.Color)
		r.pointLightMesh.Place(l.Position)
		r.setMesh(r.ambientProg, r.pointLightMesh)
		for _, subMesh := range r.pointLightMesh.SubMeshes {
			r.setSubMesh(r.ambientProg, subMesh)
			r.ambientProg.Render(subMesh.Geo.Inds, r.renderOpts)
		}
	}

	for _, l := range r.scene.SpotLights {
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

func (r *MeshRenderer) RenderAmbient(aoMap *graphics.Texture2D) {
	r.renderOpts.Blending = graphics.NoBlending
	r.renderOpts.DepthTest = graphics.EqualDepthTest

	r.ambientProg.LightColor.Set(r.scene.AmbientLight.Color)
	r.ambientProg.AmbientOcclusionMap.Set(aoMap)
	r.setCamera(r.ambientProg, r.camera)
	r.renderMeshes(r.ambientProg)

	// render light source
	// TODO: do with shaders instead for fancier effects?
	for _, l := range r.scene.PointLights {
		r.ambientProg.LightColor.Set(l.Color)
		r.pointLightMesh.Place(l.Position)
		r.setMesh(r.ambientProg, r.pointLightMesh)
		for _, subMesh := range r.pointLightMesh.SubMeshes {
			r.setSubMesh(r.ambientProg, subMesh)
			r.ambientProg.Render(subMesh.Geo.Inds, r.renderOpts)
		}
	}

	for _, l := range r.scene.SpotLights {
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

func (r *MeshRenderer) RenderLights() {
	r.renderOpts.DepthTest = graphics.EqualDepthTest
	r.renderOpts.Blending = graphics.AdditiveBlending // add to framebuffer contents

	r.setCamera(r.pointLitProg, r.camera)
	for _, l := range r.scene.PointLights {
		r.setPointLight(r.pointLitProg, l)
		r.renderMeshes(r.pointLitProg)
	}

	r.setCamera(r.spotLitProg, r.camera)
	for _, l := range r.scene.SpotLights {
		r.setSpotLight(r.spotLitProg, l)
		r.renderMeshes(r.spotLitProg)
	}

	r.setCamera(r.dirLitProg, r.camera)
	for _, l := range r.scene.DirectionalLights {
		r.setDirectionalLight(r.dirLitProg, l)
		r.renderMeshes(r.dirLitProg)
	}
}

func (r *MeshRenderer) renderMeshes(sp *MeshProgram) {
	j := 0
	for i, m := range r.scene.Meshes {
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
		tex, found := r.tex2ds[mtl.AmbientMap]
		if !found {
			tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.LinearFilter, graphics.RepeatWrap, mtl.AmbientMap, true)
			r.tex2ds[mtl.AmbientMap] = tex
		}
		sp.MaterialAmbient.Set(mtl.Ambient)
		sp.MaterialAmbientMap.Set(tex)
	} else {
		sp.MaterialAmbient.Set(math.Vec3{0, 0, 0})
		sp.MaterialAmbientMap.Set(blackTexture)
	}

	if r.MaterialDiffuseEnabled {
		tex, found := r.tex2ds[mtl.DiffuseMap]
		if !found {
			tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.LinearFilter, graphics.RepeatWrap, mtl.DiffuseMap, true)
			r.tex2ds[mtl.DiffuseMap] = tex
		}
		sp.MaterialDiffuse.Set(mtl.Diffuse)
		sp.MaterialDiffuseMap.Set(tex)
	} else {
		sp.MaterialDiffuse.Set(math.Vec3{0, 0, 0})
		sp.MaterialDiffuseMap.Set(blackTexture)
	}

	if r.MaterialSpecularEnabled {
		tex, found := r.tex2ds[mtl.SpecularMap]
		if !found {
			tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.LinearFilter, graphics.RepeatWrap, mtl.SpecularMap, true)
			r.tex2ds[mtl.SpecularMap] = tex
		}
		sp.MaterialSpecular.Set(mtl.Specular)
		sp.MaterialSpecularMap.Set(tex)
		sp.MaterialShine.Set(mtl.Shine)
	} else {
		sp.MaterialSpecular.Set(math.Vec3{0, 0, 0})
		sp.MaterialSpecularMap.Set(blackTexture)
		sp.MaterialShine.Set(float32(0))
	}

	if r.MaterialAlphaEnabled {
		tex, found := r.tex2ds[mtl.AlphaMap]
		if !found {
			tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.LinearFilter, graphics.RepeatWrap, mtl.AlphaMap, true)
			r.tex2ds[mtl.AlphaMap] = tex
		}
		sp.MaterialAlpha.Set(mtl.Alpha)
		sp.MaterialAlphaMap.Set(tex)
	} else {
		sp.MaterialAlpha.Set(float32(1.0))
		sp.MaterialAlphaMap.Set(whiteTexture)
	}

	if r.MaterialNormalEnabled {
		tex, found := r.tex2ds[mtl.BumpMap]
		if !found {
			tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.LinearFilter, graphics.RepeatWrap, mtl.BumpMap, true)
			r.tex2ds[mtl.BumpMap] = tex
		}
		sp.MaterialBumpMap.Set(tex)
		sp.MaterialBumpMapWidth.Set(tex.Width())
		sp.MaterialBumpMapHeight.Set(tex.Height())
	} else {
		sp.MaterialBumpMap.Set(whiteTexture)
		sp.MaterialBumpMapWidth.Set(whiteTexture.Width())
		sp.MaterialBumpMapHeight.Set(whiteTexture.Height())
	}

	// upload to GPU
	var vbo *graphics.VertexBuffer
	var ibo *graphics.IndexBuffer
	i, found := r.vboCache[&sm.Geo.Verts[0]]
	if found {
		vbo = r.vbos[i]
		ibo = r.ibos[i]
	} else {
		vbo = graphics.NewVertexBuffer()
		ibo = graphics.NewIndexBuffer()
		vbo.SetData(sm.Geo.Verts, 0)
		ibo.SetData(sm.Geo.Faces, 0)

		r.vbos = append(r.vbos, vbo)
		r.ibos = append(r.ibos, ibo)
		r.vboCache[&sm.Geo.Verts[0]] = len(r.vbos) - 1
	}

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
		smap, found := r.pointLightShadowMaps[l.ID]
		if !found {
			panic("set point light with no shadow map")
		}
		sp.ShadowMap.Set(smap)
	} else {
		sp.ShadowFar.Set(float32(100))
		sp.ShadowMap.Set(whiteCubeMap)
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
		r.shadowProjViewMat.Identity()
		r.shadowProjViewMat.Mult(l.ProjectionMatrix())
		r.shadowProjViewMat.Mult(l.ViewMatrix())
		sp.ShadowProjectionViewMatrix.Set(&r.shadowProjViewMat)
		sp.ShadowFar.Set(l.PerspectiveCamera.Far)
		smap, found := r.spotLightShadowMaps[l.ID]
		if !found {
			panic("set spot light with no shadow map")
		}
		sp.ShadowMap.Set(smap)
	} else {
		sp.ShadowFar.Set(float32(100))
		sp.ShadowMap.Set(whiteTexture)
	}
}

func (r *MeshRenderer) setDirectionalLight(sp *MeshProgram, l *light.DirectionalLight) {
	sp.LightDirection.Set(l.Forward())
	sp.LightColor.Set(l.Color.Scale(l.Intensity))
	sp.LightAttenuation.Set(float32(0))

	if r.ShadowsEnabled && l.CastShadows {
		r.shadowProjViewMat.Identity()
		r.shadowProjViewMat.Mult(l.ProjectionMatrix())
		r.shadowProjViewMat.Mult(l.ViewMatrix())
		sp.ShadowProjectionViewMatrix.Set(&r.shadowProjViewMat)
		smap, found := r.dirLightShadowMaps[l.ID]
		if !found {
			panic("set directional light with no shadow map")
		}
		sp.ShadowMap.Set(smap)
	} else {
		sp.ShadowFar.Set(float32(100))
		sp.ShadowMap.Set(whiteTexture)
	}
}

// shadow stuff below

func (r *MeshRenderer) setShadowCamera(sp *ShadowMapProgram, c camera.Camera) {
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

func (r *MeshRenderer) setShadowMesh(sp *ShadowMapProgram, m *object.Mesh) {
	sp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *MeshRenderer) setShadowSubMesh(sp *ShadowMapProgram, sm *object.SubMesh) {
	var vbo *graphics.VertexBuffer
	var ibo *graphics.IndexBuffer
	i, found := r.vboCache[&sm.Geo.Verts[0]]
	if found {
		vbo = r.vbos[i]
		ibo = r.ibos[i]
	} else {
		vbo = graphics.NewVertexBuffer()
		ibo = graphics.NewIndexBuffer()
		vbo.SetData(sm.Geo.Verts, 0)
		ibo.SetData(sm.Geo.Faces, 0)

		r.vbos = append(r.vbos, vbo)
		r.ibos = append(r.ibos, ibo)
		r.vboCache[&sm.Geo.Verts[0]] = len(r.vbos) - 1
	}

	sp.Position.SetSourceVertex(vbo, 0)
	sp.SetIndices(ibo)
}

func (r *MeshRenderer) RenderShadows() {
	r.shadowRenderOpts.DepthTest = graphics.LessDepthTest
	r.shadowRenderOpts.Culling = graphics.BackCulling
	r.shadowRenderOpts.Primitive = graphics.Triangles

	for _, l := range r.scene.PointLights {
		if l.CastShadows {
			r.renderPointLightShadowMap(l)
		}
	}
	for _, l := range r.scene.SpotLights {
		if l.CastShadows {
			r.renderSpotLightShadowMap(l)
		}
	}
	for _, l := range r.scene.DirectionalLights {
		if l.CastShadows {
			r.renderDirectionalLightShadowMap(l)
		}
	}
}

// render shadow map to l's shadow map
func (r *MeshRenderer) renderPointLightShadowMap(l *light.PointLight) {
	smap, found := r.pointLightShadowMaps[l.ID]
	if !found {
		smap = graphics.NewCubeMap(graphics.DepthTexture, graphics.LinearFilter, 512, 512)
		r.pointLightShadowMaps[l.ID] = smap
	}

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

	r.setShadowCamera(r.shadowSp1, c)

	for _, m := range r.scene.Meshes {
		r.setShadowMesh(r.shadowSp1, m)
		for _, subMesh := range m.SubMeshes {
			r.setShadowSubMesh(r.shadowSp1, subMesh)

			r.shadowSp1.Render(subMesh.Geo.Inds, r.shadowRenderOpts)
		}
	}

	//l.DirtyShadowMap = false
}

func (r *MeshRenderer) renderSpotLightShadowMap(l *light.SpotLight) {
	smap, found := r.spotLightShadowMaps[l.ID]
	if !found {
		smap = graphics.NewTexture2D(graphics.DepthTexture, graphics.LinearFilter, graphics.BorderClampWrap, 512, 512, false)
		smap.SetBorderColor(math.NewVec4(1, 1, 1, 1))
		r.spotLightShadowMaps[l.ID] = smap
	}

	// TODO: re-render also when objects have moved
	//if !l.DirtyShadowMap {
		//return
	//}

	r.shadowSp2.Depth.Set(smap)
	smap.Clear(math.Vec4{1, 1, 1, 1})
	r.setShadowCamera(r.shadowSp2, &l.PerspectiveCamera)

	for _, m := range r.scene.Meshes {
		r.setShadowMesh(r.shadowSp2, m)
		for _, subMesh := range m.SubMeshes {
			if !l.PerspectiveCamera.Cull(subMesh) {
				r.setShadowSubMesh(r.shadowSp2, subMesh)

				r.shadowSp2.Render(subMesh.Geo.Inds, r.shadowRenderOpts)
			}
		}
	}

	//l.DirtyShadowMap = false
}

func (r *MeshRenderer) renderDirectionalLightShadowMap(l *light.DirectionalLight) {
	smap, found := r.dirLightShadowMaps[l.ID]
	if !found {
		smap = graphics.NewTexture2D(graphics.DepthTexture, graphics.LinearFilter, graphics.BorderClampWrap, 512, 512, false)
		smap.SetBorderColor(math.NewVec4(1, 1, 1, 1))
		r.dirLightShadowMaps[l.ID] = smap
	}

	// TODO: re-render also when objects have moved
	//if !l.DirtyShadowMap {
		//return
	//}

	r.shadowSp3.Depth.Set(smap)
	smap.Clear(math.Vec4{1, 1, 1, 1})
	r.setShadowCamera(r.shadowSp3, &l.OrthoCamera)

	for _, m := range r.scene.Meshes {
		r.setShadowMesh(r.shadowSp3, m)
		for _, subMesh := range m.SubMeshes {
			if !l.OrthoCamera.Cull(subMesh) {
				r.setShadowSubMesh(r.shadowSp3, subMesh)

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

func init() {
	blueTexture = graphics.NewUniformTexture2D(math.Vec4{0.5, 0.5, 1, 0})
	whiteTexture = graphics.NewUniformTexture2D(math.Vec4{1, 1, 1, 1})
	blackTexture = graphics.NewUniformTexture2D(math.Vec4{0, 0, 0, 1})

	whiteCubeMap = graphics.NewUniformCubeMap(math.Vec4{1, 1, 1, 1})
}
