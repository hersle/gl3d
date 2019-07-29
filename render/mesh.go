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

type MeshRenderer struct {
	ambientProg *MeshShaderProgram
	pointLitProg *MeshShaderProgram
	spotLitProg *MeshShaderProgram
	dirLitProg *MeshShaderProgram

	renderOpts *graphics.RenderOptions

	vboCache map[*object.Vertex]int
	iboCache map[*object.Vertex]int
	vbos     []*graphics.VertexBuffer
	ibos     []*graphics.IndexBuffer

	tex2ds map[image.Image]*graphics.Texture2D

	pointLightShadowMaps map[int]*graphics.CubeMap
	spotLightShadowMaps  map[int]*graphics.Texture2D
	dirLightShadowMaps   map[int]*graphics.Texture2D

	shadowSp1             *ShadowMapShaderProgram
	shadowSp2             *ShadowMapShaderProgram
	shadowSp3             *ShadowMapShaderProgram
	shadowRenderOpts      *graphics.RenderOptions

	shadowProjViewMat math.Mat4

	pointLightMesh *object.Mesh
	spotLightMesh *object.Mesh

	normalMatrices []math.Mat4

	cullCache []bool
}

type MeshShaderProgram struct {
	*graphics.ShaderProgram

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

	LightPosition          *graphics.Uniform
	LightDirection         *graphics.Uniform
	LightColor             *graphics.Uniform
	LightAttenuation       *graphics.Uniform
	LightCosAngle          *graphics.Uniform

	ShadowProjectionViewMatrix *graphics.Uniform
	ShadowMap              *graphics.Uniform
	ShadowFar              *graphics.Uniform
}

type ShadowMapShaderProgram struct {
	*graphics.ShaderProgram

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

	r.ambientProg = NewMeshShaderProgram("DEPTH", "AMBIENT")
	r.pointLitProg = NewMeshShaderProgram("POINT", "SHADOW", "PCF")
	r.spotLitProg = NewMeshShaderProgram("SPOT", "SHADOW", "PCF")
	r.dirLitProg = NewMeshShaderProgram("DIR", "SHADOW", "PCF")

	r.renderOpts = graphics.NewRenderOptions()

	r.vboCache = make(map[*object.Vertex]int)
	r.iboCache = make(map[*object.Vertex]int)
	r.pointLightShadowMaps = make(map[int]*graphics.CubeMap)
	r.spotLightShadowMaps = make(map[int]*graphics.Texture2D)
	r.dirLightShadowMaps = make(map[int]*graphics.Texture2D)
	r.tex2ds = make(map[image.Image]*graphics.Texture2D)

	r.shadowSp1 = NewShadowMapShaderProgram("POINT") // point light
	r.shadowSp2 = NewShadowMapShaderProgram("SPOT") // spot light
	r.shadowSp3 = NewShadowMapShaderProgram("DIR") // directional light

	r.shadowRenderOpts = graphics.NewRenderOptions()

	geo := object.NewSphere(math.Vec3{0, 0, 0}, 0.1).Geometry(6)
	mtl := material.NewDefaultMaterial("")
	mtl.Ambient = math.Vec3{1, 1, 1}
	r.pointLightMesh = object.NewMesh(geo, mtl)

	geo = object.NewCone(math.Vec3{0, 0, -1}, math.Vec3{0, 0, 0}, 0.5).Geometry(6)
	r.spotLightMesh = object.NewMesh(geo, mtl)

	return &r, nil
}

func NewMeshShaderProgram(defines ...string) *MeshShaderProgram {
	var sp MeshShaderProgram
	var err error

	vFile := "render/shaders/meshvshadertemplate.glsl" // TODO: make independent from executable directory
	fFile := "render/shaders/meshfshadertemplate.glsl" // TODO: make independent from executable directory
	sp.ShaderProgram, err = graphics.ReadShaderProgram(vFile, fFile, "", defines...)
	if err != nil {
		panic(err)
	}

	sp.Position = sp.Input("position")
	sp.TexCoord = sp.Input("texCoordV")
	sp.Normal = sp.Input("normalV")
	sp.Tangent = sp.Input("tangentV")

	sp.Color = sp.OutputColor("fragColor")
	sp.Depth = sp.OutputDepth()

	sp.ModelMatrix = sp.Uniform("modelMatrix")
	sp.ViewMatrix = sp.Uniform("viewMatrix")
	sp.ProjectionMatrix = sp.Uniform("projectionMatrix")
	sp.NormalMatrix = sp.Uniform("normalMatrix")

	sp.MaterialAmbient = sp.Uniform("materialAmbient")
	sp.MaterialAmbientMap = sp.Uniform("materialAmbientMap")
	sp.MaterialDiffuse = sp.Uniform("materialDiffuse")
	sp.MaterialDiffuseMap = sp.Uniform("materialDiffuseMap")
	sp.MaterialSpecular = sp.Uniform("materialSpecular")
	sp.MaterialSpecularMap = sp.Uniform("materialSpecularMap")
	sp.MaterialShine = sp.Uniform("materialShine")
	sp.MaterialAlpha = sp.Uniform("materialAlpha")
	sp.MaterialAlphaMap = sp.Uniform("materialAlphaMap")
	sp.MaterialBumpMap = sp.Uniform("materialBumpMap")

	sp.LightPosition = sp.Uniform("lightPosition")
	sp.LightDirection = sp.Uniform("lightDirection")
	sp.LightColor = sp.Uniform("lightColor")
	sp.LightAttenuation = sp.Uniform("lightAttenuation")
	sp.LightCosAngle = sp.Uniform("lightCosAng")

	sp.ShadowProjectionViewMatrix = sp.Uniform("shadowProjectionViewMatrix")
	sp.ShadowMap = sp.Uniform("shadowMap")
	sp.ShadowFar = sp.Uniform("lightFar")

	return &sp
}

func NewShadowMapShaderProgram(defines ...string) *ShadowMapShaderProgram {
	var sp ShadowMapShaderProgram
	var err error

	vFile := "render/shaders/shadowmapvshadertemplate.glsl" // TODO: make independent from executable directory
	fFile := "render/shaders/shadowmapfshadertemplate.glsl" // TODO: make independent from executable directory
	gFile := "render/shaders/shadowmapgshadertemplate.glsl" // TODO: make independent from executable directory
	sp.ShaderProgram, err = graphics.ReadShaderProgram(vFile, fFile, gFile, defines...)
	if err != nil {
		panic(err)
	}

	sp.ModelMatrix = sp.Uniform("modelMatrix")
	sp.ViewMatrix = sp.Uniform("viewMatrix")
	sp.ProjectionMatrix = sp.Uniform("projectionMatrix")
	sp.LightPosition = sp.Uniform("lightPosition")
	sp.LightFar = sp.Uniform("far")
	sp.Position = sp.Input("position")

	sp.ProjViewMats = make([]*graphics.Uniform, 6)
	for i := 0; i < 6; i++ {
		name := fmt.Sprintf("projectionViewMatrices[%d]", i)
		sp.ProjViewMats[i] = sp.Uniform(name)
	}

	sp.Depth = sp.OutputDepth()

	return &sp
}

func (r *MeshRenderer) Render(s *scene.Scene, c camera.Camera, colorTexture, depthTexture *graphics.Texture2D) {
	r.renderOpts.Cull = graphics.CullBack
	r.renderOpts.PrimitiveType = graphics.Triangle

	r.ambientProg.Color.Set(colorTexture)
	r.pointLitProg.Color.Set(colorTexture)
	r.spotLitProg.Color.Set(colorTexture)
	r.dirLitProg.Color.Set(colorTexture)
	r.ambientProg.Depth.Set(depthTexture)
	r.pointLitProg.Depth.Set(depthTexture)
	r.spotLitProg.Depth.Set(depthTexture)
	r.dirLitProg.Depth.Set(depthTexture)

	r.preparationPass(s, c)
	r.shadowPass(s)
	r.depthAmbientPass(s, c)
	r.lightPass(s, c)
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
}

func (r *MeshRenderer) depthAmbientPass(s *scene.Scene, c camera.Camera) {
	r.renderOpts.DisableBlending()
	r.renderOpts.DepthTest = graphics.LessDepthTest

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
			ibo := r.subMeshIndexBuffer(subMesh)
			r.ambientProg.RenderIndexed(ibo, subMesh.Geo.Inds, r.renderOpts)
		}
	}

	for _, l := range s.SpotLights {
		r.ambientProg.LightColor.Set(l.Color)
		r.spotLightMesh.Place(l.Position)
		r.spotLightMesh.Orient(l.UnitX, l.UnitY)
		r.setMesh(r.ambientProg, r.spotLightMesh)
		for _, subMesh := range r.spotLightMesh.SubMeshes {
			r.setSubMesh(r.ambientProg, subMesh)
			ibo := r.subMeshIndexBuffer(subMesh)
			r.ambientProg.RenderIndexed(ibo, subMesh.Geo.Inds, r.renderOpts)
		}
	}
}

func (r *MeshRenderer) lightPass(s *scene.Scene, c camera.Camera) {
	r.renderOpts.DepthTest = graphics.EqualDepthTest
	r.renderOpts.BlendSourceFactor = graphics.OneBlendFactor
	r.renderOpts.BlendDestinationFactor = graphics.OneBlendFactor // add to framebuffer contents

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

func (r *MeshRenderer) renderMeshes(s *scene.Scene, c camera.Camera, sp *MeshShaderProgram) {
	j := 0
	for i, m := range s.Meshes {
		r.setMesh(sp, m)
		sp.NormalMatrix.Set(&r.normalMatrices[i])
		for _, sm := range m.SubMeshes {
			if !r.cullCache[j] {
				r.setSubMesh(sp, sm)
				ibo := r.subMeshIndexBuffer(sm)
				sp.RenderIndexed(ibo, sm.Geo.Inds, r.renderOpts)
			}
			j++
		}
	}
}

func (r *MeshRenderer) setCamera(sp *MeshShaderProgram, c camera.Camera) {
	sp.ViewMatrix.Set(c.ViewMatrix())
	sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *MeshRenderer) setMesh(sp *MeshShaderProgram, m *object.Mesh) {
	sp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *MeshRenderer) subMeshVertexBuffer(sm *object.SubMesh) *graphics.VertexBuffer {
	var vbo *graphics.VertexBuffer
	i, found := r.vboCache[&sm.Geo.Verts[0]]
	if found {
		vbo = r.vbos[i]
	} else {
		vbo = graphics.NewVertexBuffer()
		vbo.SetData(sm.Geo.Verts, 0)
		r.vbos = append(r.vbos, vbo)
		r.vboCache[&sm.Geo.Verts[0]] = len(r.vbos) - 1
	}
	return vbo
}

func (r *MeshRenderer) subMeshIndexBuffer(sm *object.SubMesh) *graphics.IndexBuffer {
	var ibo *graphics.IndexBuffer
	i, found := r.iboCache[&sm.Geo.Verts[0]]
	if found {
		ibo = r.ibos[i]
	} else {
		ibo = graphics.NewIndexBuffer()
		ibo.SetData(sm.Geo.Faces, 0)
		r.ibos = append(r.ibos, ibo)
		r.iboCache[&sm.Geo.Verts[0]] = len(r.ibos) - 1
	}
	return ibo
}

func (r *MeshRenderer) setSubMesh(sp *MeshShaderProgram, sm *object.SubMesh) {
	mtl := sm.Mtl

	tex, found := r.tex2ds[mtl.AmbientMap]
	if !found {
		tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.LinearFilter, graphics.RepeatWrap, mtl.AmbientMap, true)
		r.tex2ds[mtl.AmbientMap] = tex
	}
	sp.MaterialAmbient.Set(mtl.Ambient)
	sp.MaterialAmbientMap.Set(tex)

	tex, found = r.tex2ds[mtl.DiffuseMap]
	if !found {
		tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.LinearFilter, graphics.RepeatWrap, mtl.DiffuseMap, true)
		r.tex2ds[mtl.DiffuseMap] = tex
	}
	sp.MaterialDiffuse.Set(mtl.Diffuse)
	sp.MaterialDiffuseMap.Set(tex)

	tex, found = r.tex2ds[mtl.SpecularMap]
	if !found {
		tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.LinearFilter, graphics.RepeatWrap, mtl.SpecularMap, true)
		r.tex2ds[mtl.SpecularMap] = tex
	}
	sp.MaterialSpecular.Set(mtl.Specular)
	sp.MaterialSpecularMap.Set(tex)

	sp.MaterialShine.Set(mtl.Shine)

	tex, found = r.tex2ds[mtl.AlphaMap]
	if !found {
		tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.LinearFilter, graphics.RepeatWrap, mtl.AlphaMap, true)
		r.tex2ds[mtl.AlphaMap] = tex
	}
	sp.MaterialAlpha.Set(mtl.Alpha)
	sp.MaterialAlphaMap.Set(tex)

	tex, found = r.tex2ds[mtl.BumpMap]
	if !found {
		tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.LinearFilter, graphics.RepeatWrap, mtl.BumpMap, true)
		r.tex2ds[mtl.BumpMap] = tex
	}
	sp.MaterialBumpMap.Set(tex)

	vbo := r.subMeshVertexBuffer(sm)

	sp.Position.SetSourceVertex(vbo, 0)
	sp.Normal.SetSourceVertex(vbo, 2)
	sp.TexCoord.SetSourceVertex(vbo, 1)
	sp.Tangent.SetSourceVertex(vbo, 3)
}

func (r *MeshRenderer) setAmbientLight(sp *MeshShaderProgram, l *light.AmbientLight) {
	sp.LightColor.Set(l.Color.Scale(l.Intensity))
}

func (r *MeshRenderer) setPointLight(sp *MeshShaderProgram, l *light.PointLight) {
	sp.LightPosition.Set(l.Position)
	sp.LightColor.Set(l.Color.Scale(l.Intensity))
	if l.CastShadows {
		sp.ShadowFar.Set(l.ShadowFar)
		smap, found := r.pointLightShadowMaps[l.ID]
		if !found {
			panic("set point light with no shadow map")
		}
		sp.ShadowMap.Set(smap)
	}
	sp.LightAttenuation.Set(l.Attenuation)
}

func (r *MeshRenderer) setSpotLight(sp *MeshShaderProgram, l *light.SpotLight) {
	sp.LightPosition.Set(l.Position)
	sp.LightDirection.Set(l.Forward())
	sp.LightColor.Set(l.Color.Scale(l.Intensity))
	sp.LightAttenuation.Set(l.Attenuation)
	sp.LightCosAngle.Set(float32(gomath.Cos(float64(l.FOV/2))))

	if l.CastShadows {
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
	}
}

func (r *MeshRenderer) setDirectionalLight(sp *MeshShaderProgram, l *light.DirectionalLight) {
	sp.LightDirection.Set(l.Forward())
	sp.LightColor.Set(l.Color.Scale(l.Intensity))
	sp.LightAttenuation.Set(float32(0))

	if l.CastShadows {
		r.shadowProjViewMat.Identity()
		r.shadowProjViewMat.Mult(l.ProjectionMatrix())
		r.shadowProjViewMat.Mult(l.ViewMatrix())
		sp.ShadowProjectionViewMatrix.Set(&r.shadowProjViewMat)
		smap, found := r.dirLightShadowMaps[l.ID]
		if !found {
			panic("set directional light with no shadow map")
		}
		sp.ShadowMap.Set(smap)
	}
}

func (r *MeshRenderer) SetWireframe(wireframe bool) {
	if wireframe {
		r.renderOpts.TriangleMode = graphics.LineTriangleMode
	} else {
		r.renderOpts.TriangleMode = graphics.TriangleTriangleMode
	}
}

// shadow stuff below

func (r *MeshRenderer) setShadowCamera(sp *ShadowMapShaderProgram, c camera.Camera) {
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

func (r *MeshRenderer) setShadowMesh(sp *ShadowMapShaderProgram, m *object.Mesh) {
	sp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *MeshRenderer) setShadowSubMesh(sp *ShadowMapShaderProgram, sm *object.SubMesh) {
	vbo := r.subMeshVertexBuffer(sm)

	sp.Position.SetSourceVertex(vbo, 0)
}

func (r *MeshRenderer) shadowPass(s *scene.Scene) {
	r.shadowRenderOpts.DepthTest = graphics.LessDepthTest
	r.shadowRenderOpts.Cull = graphics.CullBack
	r.shadowRenderOpts.PrimitiveType = graphics.Triangle

	for _, l := range s.PointLights {
		if l.CastShadows {
			r.renderPointLightShadowMap(s, l)
		}
	}
	for _, l := range s.SpotLights {
		if l.CastShadows {
			r.renderSpotLightShadowMap(s, l)
		}
	}
	for _, l := range s.DirectionalLights {
		if l.CastShadows {
			r.renderDirectionalLightShadowMap(s, l)
		}
	}
}

// render shadow map to l's shadow map
func (r *MeshRenderer) renderPointLightShadowMap(s *scene.Scene, l *light.PointLight) {
	smap, found := r.pointLightShadowMaps[l.ID]
	if !found {
		smap = graphics.NewCubeMap(graphics.DepthTexture, graphics.NearestFilter, 512, 512)
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

	for _, m := range s.Meshes {
		r.setShadowMesh(r.shadowSp1, m)
		for _, subMesh := range m.SubMeshes {
			r.setShadowSubMesh(r.shadowSp1, subMesh)

			ibo := r.subMeshIndexBuffer(subMesh)
			r.shadowSp1.RenderIndexed(ibo, subMesh.Geo.Inds, r.shadowRenderOpts)
		}
	}

	//l.DirtyShadowMap = false
}

func (r *MeshRenderer) renderSpotLightShadowMap(s *scene.Scene, l *light.SpotLight) {
	smap, found := r.spotLightShadowMaps[l.ID]
	if !found {
		smap = graphics.NewTexture2D(graphics.DepthTexture, graphics.NearestFilter, graphics.BorderClampWrap, 512, 512, false)
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

	for _, m := range s.Meshes {
		r.setShadowMesh(r.shadowSp2, m)
		for _, subMesh := range m.SubMeshes {
			if !l.PerspectiveCamera.Cull(subMesh) {
				r.setShadowSubMesh(r.shadowSp2, subMesh)

				ibo := r.subMeshIndexBuffer(subMesh)
				r.shadowSp2.RenderIndexed(ibo, subMesh.Geo.Inds, r.shadowRenderOpts)
			}
		}
	}

	//l.DirtyShadowMap = false
}

func (r *MeshRenderer) renderDirectionalLightShadowMap(s *scene.Scene, l *light.DirectionalLight) {
	smap, found := r.dirLightShadowMaps[l.ID]
	if !found {
		smap = graphics.NewTexture2D(graphics.DepthTexture, graphics.NearestFilter, graphics.BorderClampWrap, 512, 512, false)
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

	for _, m := range s.Meshes {
		r.setShadowMesh(r.shadowSp3, m)
		for _, subMesh := range m.SubMeshes {
			if !l.OrthoCamera.Cull(subMesh) {
				r.setShadowSubMesh(r.shadowSp3, subMesh)

				ibo := r.subMeshIndexBuffer(subMesh)
				r.shadowSp3.RenderIndexed(ibo, subMesh.Geo.Inds, r.shadowRenderOpts)
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
