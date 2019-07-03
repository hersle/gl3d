package render

import (
	"github.com/hersle/gl3d/camera"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/light"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"github.com/hersle/gl3d/scene"
	"image"
	"fmt"
)

type MeshShaderProgram struct {
	*graphics.ShaderProgram

	Position *graphics.Attrib
	TexCoord *graphics.Attrib
	Normal   *graphics.Attrib
	Tangent  *graphics.Attrib

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

	ShadowProjectionViewMatrix *graphics.Uniform
	ShadowMap              *graphics.Uniform
	ShadowFar              *graphics.Uniform
}

// TODO: redesign attr/uniform access system?
type MeshRenderer struct {
	sp1 *MeshShaderProgram
	sp2 *MeshShaderProgram
	sp3 *MeshShaderProgram
	sp4 *MeshShaderProgram

	renderState *graphics.State

	vboCache map[*object.Vertex]int
	vbos     []*graphics.VertexBuffer
	ibos     []*graphics.IndexBuffer

	tex2ds map[image.Image]*graphics.Texture2D

	pointLightShadowMaps map[int]*graphics.CubeMap
	spotLightShadowMaps  map[int]*graphics.Texture2D
	dirLightShadowMaps   map[int]*graphics.Texture2D

	shadowSp1             *ShadowMapShaderProgram
	shadowSp2             *ShadowMapShaderProgram
	shadowSp3             *ShadowMapShaderProgram
	shadowMapFramebuffer *graphics.Framebuffer
	shadowRenderState    *graphics.State

	shadowProjViewMat math.Mat4

	normalMatrices []math.Mat4
	renderInfos []renderInfo
}

type renderInfo struct {
	subMesh *object.SubMesh
	normalMatrix *math.Mat4
}

type ShadowMapShaderProgram struct {
	*graphics.ShaderProgram

	Position         *graphics.Attrib

	ModelMatrix      *graphics.Uniform
	ViewMatrix       *graphics.Uniform
	ProjectionMatrix *graphics.Uniform

	LightPosition    *graphics.Uniform
	LightFar              *graphics.Uniform

	ProjViewMats     []*graphics.Uniform
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
	sp.Position = sp.Attrib("position")

	sp.ProjViewMats = make([]*graphics.Uniform, 6)
	for i := 0; i < 6; i++ {
		name := fmt.Sprintf("projectionViewMatrices[%d]", i)
		sp.ProjViewMats[i] = sp.Uniform(name)
	}

	return &sp
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

	sp.Position = sp.Attrib("position")
	sp.TexCoord = sp.Attrib("texCoordV")
	sp.Normal = sp.Attrib("normalV")
	sp.Tangent = sp.Attrib("tangentV")

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

	sp.ShadowProjectionViewMatrix = sp.Uniform("shadowProjectionViewMatrix")
	sp.ShadowMap = sp.Uniform("shadowMap")
	sp.ShadowFar = sp.Uniform("lightFar")

	return &sp
}

func NewMeshRenderer() (*MeshRenderer, error) {
	var r MeshRenderer

	r.sp1 = NewMeshShaderProgram("DEPTH", "AMBIENT")
	r.sp2 = NewMeshShaderProgram("POINT", "SHADOW", "PCF")
	r.sp3 = NewMeshShaderProgram("SPOT", "SHADOW", "PCF")
	r.sp4 = NewMeshShaderProgram("DIR", "SHADOW", "PCF")

	r.renderState = graphics.NewState()
	r.renderState.Cull = graphics.CullBack
	r.renderState.PrimitiveType = graphics.Triangle

	r.vboCache = make(map[*object.Vertex]int)
	r.pointLightShadowMaps = make(map[int]*graphics.CubeMap)
	r.spotLightShadowMaps = make(map[int]*graphics.Texture2D)
	r.dirLightShadowMaps = make(map[int]*graphics.Texture2D)
	r.tex2ds = make(map[image.Image]*graphics.Texture2D)

	r.shadowSp1 = NewShadowMapShaderProgram("POINT") // point light
	r.shadowSp2 = NewShadowMapShaderProgram("SPOT") // spot light
	r.shadowSp3 = NewShadowMapShaderProgram("DIR") // directional light

	r.shadowMapFramebuffer = graphics.NewFramebuffer()

	r.shadowRenderState = graphics.NewState()
	r.shadowRenderState.Framebuffer = r.shadowMapFramebuffer
	r.shadowRenderState.DepthTest = graphics.LessDepthTest
	r.shadowRenderState.Cull = graphics.CullBack
	r.shadowRenderState.PrimitiveType = graphics.Triangle

	return &r, nil
}

func (r *MeshRenderer) ExecRenderInfo(ri *renderInfo, sp *MeshShaderProgram) {
	r.SetMesh(sp, ri.subMesh.Mesh)
	sp.NormalMatrix.Set(ri.normalMatrix)
	r.SetSubMesh(sp, ri.subMesh)
	r.renderState.Render(ri.subMesh.Geo.Inds)
}

func (r *MeshRenderer) AmbientPass(c camera.Camera) {
	r.renderState.DisableBlending()
	r.renderState.DepthTest = graphics.LessDepthTest
	r.renderState.Program = r.sp1.ShaderProgram

	r.SetCamera(r.sp1, c)

	for _, ri := range r.renderInfos {
		r.ExecRenderInfo(&ri, r.sp1)
	}
}

func (r *MeshRenderer) LightPass(s *scene.Scene, c camera.Camera) {
	r.renderState.DepthTest = graphics.EqualDepthTest
	r.renderState.BlendSourceFactor = graphics.OneBlendFactor
	r.renderState.BlendDestinationFactor = graphics.OneBlendFactor // add to framebuffer contents

	r.PointLightPass(s, c)
	r.SpotLightPass(s, c)
	r.DirectionalLightPass(s, c)
}

func (r *MeshRenderer) PointLightPass(s *scene.Scene, c camera.Camera) {
	r.renderState.Program = r.sp2.ShaderProgram

	r.SetCamera(r.sp2, c)

	for _, l := range s.PointLights {
		r.SetPointLight(r.sp2, l)
		for _, ri := range r.renderInfos {
			r.ExecRenderInfo(&ri, r.sp2)
		}
	}
}

func (r *MeshRenderer) SpotLightPass(s *scene.Scene, c camera.Camera) {
	r.renderState.Program = r.sp3.ShaderProgram

	r.SetCamera(r.sp3, c)

	for _, l := range s.SpotLights {
		r.SetSpotLight(r.sp3, l)
		for _, ri := range r.renderInfos {
			r.ExecRenderInfo(&ri, r.sp3)
		}
	}
}

func (r *MeshRenderer) DirectionalLightPass(s *scene.Scene, c camera.Camera) {
	r.renderState.Program = r.sp4.ShaderProgram

	r.SetCamera(r.sp4, c)

	for _, l := range s.DirectionalLights {
		r.SetDirectionalLight(r.sp4, l)
		for _, ri := range r.renderInfos {
			r.ExecRenderInfo(&ri, r.sp4)
		}
	}
}

func (r *MeshRenderer) Render(s *scene.Scene, c camera.Camera, fb *graphics.Framebuffer) {
	r.renderInfos = r.renderInfos[:0]
	if len(r.normalMatrices) < len(s.Meshes) {
		r.normalMatrices = make([]math.Mat4, len(s.Meshes))
	}
	for i, m := range s.Meshes {
		calcNormalMatrix := false

		for _, sm := range m.SubMeshes {
			if !c.Cull(sm) {
				if !calcNormalMatrix {
					normalMatrix := &r.normalMatrices[i]
					normalMatrix.Identity()
					normalMatrix.Mult(c.ViewMatrix())
					normalMatrix.Mult(m.WorldMatrix())
					normalMatrix.Invert()
					normalMatrix.Transpose()
					calcNormalMatrix = true
				}

				var ri renderInfo
				ri.subMesh = sm
				ri.normalMatrix = &r.normalMatrices[i]
				r.renderInfos = append(r.renderInfos, ri)
			}
		}
	}

	r.RenderShadowMaps(s)

	r.renderState.Framebuffer = fb

	r.AmbientPass(c) // also works as depth pass
	r.LightPass(s, c)
}

func (r *MeshRenderer) SetWireframe(wireframe bool) {
	if wireframe {
		r.renderState.TriangleMode = graphics.LineTriangleMode
	} else {
		r.renderState.TriangleMode = graphics.TriangleTriangleMode
	}
}

func (r *MeshRenderer) SetCamera(sp *MeshShaderProgram, c camera.Camera) {
	sp.ViewMatrix.Set(c.ViewMatrix())
	sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *MeshRenderer) SetMesh(sp *MeshShaderProgram, m *object.Mesh) {
	sp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *MeshRenderer) SetSubMesh(sp *MeshShaderProgram, sm *object.SubMesh) {
	mtl := sm.Mtl

	tex, found := r.tex2ds[mtl.AmbientMap]
	if !found {
		tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.LinearFilter, graphics.RepeatWrap, mtl.AmbientMap)
		r.tex2ds[mtl.AmbientMap] = tex
	}
	sp.MaterialAmbient.Set(mtl.Ambient)
	sp.MaterialAmbientMap.Set(tex)

	tex, found = r.tex2ds[mtl.DiffuseMap]
	if !found {
		tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.LinearFilter, graphics.RepeatWrap, mtl.DiffuseMap)
		r.tex2ds[mtl.DiffuseMap] = tex
	}
	sp.MaterialDiffuse.Set(mtl.Diffuse)
	sp.MaterialDiffuseMap.Set(tex)

	tex, found = r.tex2ds[mtl.SpecularMap]
	if !found {
		tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.LinearFilter, graphics.RepeatWrap, mtl.SpecularMap)
		r.tex2ds[mtl.SpecularMap] = tex
	}
	sp.MaterialSpecular.Set(mtl.Specular)
	sp.MaterialSpecularMap.Set(tex)

	sp.MaterialShine.Set(mtl.Shine)

	tex, found = r.tex2ds[mtl.AlphaMap]
	if !found {
		tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.LinearFilter, graphics.RepeatWrap, mtl.AlphaMap)
		r.tex2ds[mtl.AlphaMap] = tex
	}
	sp.MaterialAlpha.Set(mtl.Alpha)
	sp.MaterialAlphaMap.Set(tex)

	tex, found = r.tex2ds[mtl.BumpMap]
	if !found {
		tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.LinearFilter, graphics.RepeatWrap, mtl.BumpMap)
		r.tex2ds[mtl.BumpMap] = tex
	}
	sp.MaterialBumpMap.Set(tex)

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

	sp.SetAttribIndexBuffer(ibo)
}

func (r *MeshRenderer) SetAmbientLight(sp *MeshShaderProgram, l *light.AmbientLight) {
	sp.LightColor.Set(l.Color.Scale(l.Intensity))
	sp.LightAttenuation.Set(0)
}

func (r *MeshRenderer) SetPointLight(sp *MeshShaderProgram, l *light.PointLight) {
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

func (r *MeshRenderer) SetSpotLight(sp *MeshShaderProgram, l *light.SpotLight) {
	sp.LightPosition.Set(l.Position)
	sp.LightDirection.Set(l.Forward())
	sp.LightColor.Set(l.Color.Scale(l.Intensity))
	sp.LightAttenuation.Set(l.Attenuation)

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

func (r *MeshRenderer) SetDirectionalLight(sp *MeshShaderProgram, l *light.DirectionalLight) {
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

// shadow stuff below

func (r *MeshRenderer) SetShadowCamera(sp *ShadowMapShaderProgram, c camera.Camera) {
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

func (r *MeshRenderer) SetShadowMesh(sp *ShadowMapShaderProgram, m *object.Mesh) {
	sp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *MeshRenderer) SetShadowSubMesh(sp *ShadowMapShaderProgram, sm *object.SubMesh) {
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
	sp.SetAttribIndexBuffer(ibo)
}

// render shadow map to l's shadow map
func (r *MeshRenderer) RenderPointLightShadowMap(s *scene.Scene, l *light.PointLight) {
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

	r.shadowRenderState.Program = r.shadowSp1.ShaderProgram

	for face := 0; face < 6; face++ {
		c.SetForwardUp(forwards[face], ups[face])
		mat := math.Mat4{}
		mat.Identity()
		mat.Mult(c.ProjectionMatrix())
		mat.Mult(c.ViewMatrix())
		r.shadowSp1.ProjViewMats[face].Set(&mat)
	}

	r.shadowMapFramebuffer.Attach(smap)
	r.shadowMapFramebuffer.ClearDepth(1)

	r.SetShadowCamera(r.shadowSp1, c)

	for _, m := range s.Meshes {
		r.SetShadowMesh(r.shadowSp1, m)
		for _, subMesh := range m.SubMeshes {
			r.SetShadowSubMesh(r.shadowSp1, subMesh)

			r.shadowRenderState.Render(subMesh.Geo.Inds)
		}
	}

	//l.DirtyShadowMap = false
}

func (r *MeshRenderer) RenderSpotLightShadowMap(s *scene.Scene, l *light.SpotLight) {
	smap, found := r.spotLightShadowMaps[l.ID]
	if !found {
		smap = graphics.NewTexture2D(graphics.DepthTexture, graphics.NearestFilter, graphics.BorderClampWrap, 512, 512)
		smap.SetBorderColor(math.NewVec4(1, 1, 1, 1))
		r.spotLightShadowMaps[l.ID] = smap
	}

	// TODO: re-render also when objects have moved
	//if !l.DirtyShadowMap {
		//return
	//}

	r.shadowMapFramebuffer.Attach(smap)
	r.shadowMapFramebuffer.ClearDepth(1)
	r.shadowRenderState.Program = r.shadowSp2.ShaderProgram
	r.SetShadowCamera(r.shadowSp2, &l.PerspectiveCamera)

	for _, m := range s.Meshes {
		r.SetShadowMesh(r.shadowSp2, m)
		for _, subMesh := range m.SubMeshes {
			if !l.PerspectiveCamera.Cull(subMesh) {
				r.SetShadowSubMesh(r.shadowSp2, subMesh)

				r.shadowRenderState.Render(subMesh.Geo.Inds)
			}
		}
	}

	//l.DirtyShadowMap = false
}

func (r *MeshRenderer) RenderDirectionalLightShadowMap(s *scene.Scene, l *light.DirectionalLight) {
	smap, found := r.dirLightShadowMaps[l.ID]
	if !found {
		smap = graphics.NewTexture2D(graphics.DepthTexture, graphics.NearestFilter, graphics.BorderClampWrap, 512, 512)
		smap.SetBorderColor(math.NewVec4(1, 1, 1, 1))
		r.dirLightShadowMaps[l.ID] = smap
	}

	// TODO: re-render also when objects have moved
	//if !l.DirtyShadowMap {
		//return
	//}

	r.shadowMapFramebuffer.Attach(smap)
	r.shadowMapFramebuffer.ClearDepth(1)
	r.shadowRenderState.Program = r.shadowSp3.ShaderProgram
	r.SetShadowCamera(r.shadowSp3, &l.OrthoCamera)

	for _, m := range s.Meshes {
		r.SetShadowMesh(r.shadowSp3, m)
		for _, subMesh := range m.SubMeshes {
			if !l.OrthoCamera.Cull(subMesh) {
				r.SetShadowSubMesh(r.shadowSp3, subMesh)

				r.shadowRenderState.Render(subMesh.Geo.Inds)
			}
		}
	}

	//l.DirtyShadowMap = false
}

func (r *MeshRenderer) RenderShadowMaps(s *scene.Scene) {
	for _, l := range s.PointLights {
		if l.CastShadows {
			r.RenderPointLightShadowMap(s, l)
		}
	}
	for _, l := range s.SpotLights {
		if l.CastShadows {
			r.RenderSpotLightShadowMap(s, l)
		}
	}
	for _, l := range s.DirectionalLights {
		if l.CastShadows {
			r.RenderDirectionalLightShadowMap(s, l)
		}
	}
}

func PointLightInteracts(l *light.PointLight, sm *object.SubMesh) bool {
	sphere := sm.BoundingSphere()
	dist := l.Position.Sub(sphere.Center).Length()
	if dist < sphere.Radius {
		return true
	}
	dist = dist - sphere.Radius
	return dist*dist < (1/0.05-1)/l.Attenuation
}
