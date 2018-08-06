package main

import (
	"github.com/hersle/gl3d/math"
	"github.com/go-gl/gl/v4.5-core/gl"
	"unsafe"
)

type MeshShaderProgram struct {
	*ShaderProgram

	Position *Attrib
	TexCoord *Attrib
	Normal *Attrib
	Tangent *Attrib

	ModelMatrix *UniformMatrix4
	ViewMatrix *UniformMatrix4
	ProjectionMatrix *UniformMatrix4

	Ambient *UniformVector3
	AmbientMap *UniformSampler
	Diffuse *UniformVector3
	DiffuseMap *UniformSampler
	Specular *UniformVector3
	SpecularMap *UniformSampler
	Shine *UniformFloat
	Alpha *UniformFloat
	AlphaMap *UniformSampler
	HasAlphaMap *UniformBool
	BumpMap *UniformSampler
	HasBumpMap *UniformBool

	LightType *UniformInteger
	LightPos *UniformVector3
	LightDir *UniformVector3
	AmbientLight *UniformVector3
	DiffuseLight *UniformVector3
	SpecularLight *UniformVector3
	ShadowViewMatrix *UniformMatrix4
	ShadowProjectionMatrix *UniformMatrix4
	SpotShadowMap *UniformSampler
	CubeShadowMap *UniformSampler
	ShadowFar *UniformFloat
}

type SkyboxShaderProgram struct {
	*ShaderProgram
	ViewMatrix *UniformMatrix4
	ProjectionMatrix *UniformMatrix4
	CubeMap *UniformSampler
	Position *Attrib
}

type TextShaderProgram struct {
	*ShaderProgram
	Atlas *UniformSampler
	Position *Attrib
	TexCoord *Attrib
}

type ShadowMapShaderProgram struct {
	*ShaderProgram
	ModelMatrix *UniformMatrix4
	ViewMatrix *UniformMatrix4
	ProjectionMatrix *UniformMatrix4
	LightPosition *UniformVector3
	Far *UniformFloat
	Position *Attrib
}

type ArrowShaderProgram struct {
	*ShaderProgram
	ModelMatrix *UniformMatrix4
	ViewMatrix *UniformMatrix4
	ProjectionMatrix *UniformMatrix4
	Color *UniformVector3
	Position *Attrib
}

type DepthPassShaderProgram struct {
	*ShaderProgram
	ModelMatrix *UniformMatrix4
	ViewMatrix *UniformMatrix4
	ProjectionMatrix *UniformMatrix4
	Position *Attrib
}

func NewMeshShaderProgram() *MeshShaderProgram {
	var sp MeshShaderProgram
	var err error

	vShaderFilename := "shaders/meshvshader.glsl"
	fShaderFilename := "shaders/meshfshader.glsl"

	sp.ShaderProgram, err = ReadShaderProgram(vShaderFilename, fShaderFilename, "")
	if err != nil {
		panic("error loading mesh shader")
	}

	sp.Position = sp.Attrib("position")
	sp.TexCoord = sp.Attrib("texCoordV")
	sp.Normal = sp.Attrib("normalV")
	sp.Tangent = sp.Attrib("tangentV")

	sp.ModelMatrix = sp.UniformMatrix4("modelMatrix")
	sp.ViewMatrix = sp.UniformMatrix4("viewMatrix")
	sp.ProjectionMatrix = sp.UniformMatrix4("projectionMatrix")

	sp.Ambient = sp.UniformVector3("material.ambient")
	sp.AmbientMap = sp.UniformSampler("material.ambientMap")
	sp.Diffuse = sp.UniformVector3("material.diffuse")
	sp.DiffuseMap = sp.UniformSampler("material.diffuseMap")
	sp.Specular = sp.UniformVector3("material.specular")
	sp.SpecularMap = sp.UniformSampler("material.specularMap")
	sp.Shine = sp.UniformFloat("material.shine")
	sp.Alpha = sp.UniformFloat("material.alpha")
	sp.AlphaMap = sp.UniformSampler("material.alphaMap")
	sp.HasAlphaMap = sp.UniformBool("material.hasAlphaMap")
	sp.BumpMap = sp.UniformSampler("material.bumpMap")
	sp.HasBumpMap = sp.UniformBool("material.hasBumpMap")

	sp.LightType = sp.UniformInteger("light.type")
	sp.LightPos = sp.UniformVector3("light.position")
	sp.LightDir = sp.UniformVector3("light.direction")
	sp.AmbientLight = sp.UniformVector3("light.ambient")
	sp.DiffuseLight = sp.UniformVector3("light.diffuse")
	sp.SpecularLight = sp.UniformVector3("light.specular")
	sp.ShadowViewMatrix = sp.UniformMatrix4("shadowViewMatrix")
	sp.ShadowProjectionMatrix = sp.UniformMatrix4("shadowProjectionMatrix")
	sp.CubeShadowMap = sp.UniformSampler("cubeShadowMap")
	sp.SpotShadowMap = sp.UniformSampler("spotShadowMap")
	sp.ShadowFar = sp.UniformFloat("light.far")

	sp.Position.SetFormat(gl.FLOAT, false)
	sp.Normal.SetFormat(gl.FLOAT, false)
	sp.TexCoord.SetFormat(gl.FLOAT, false)
	sp.Tangent.SetFormat(gl.FLOAT, false)

	return &sp
}

func (sp *MeshShaderProgram) SetCamera(c *Camera) {
	sp.ViewMatrix.Set(c.ViewMatrix())
	sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (sp *MeshShaderProgram) SetMesh(m *Mesh) {
	sp.ModelMatrix.Set(m.WorldMatrix())
}

func (sp *MeshShaderProgram) SetSubMesh(sm *SubMesh) {
	mtl := sm.mtl

	sp.Ambient.Set(mtl.ambient)
	sp.AmbientMap.Set2D(mtl.ambientMap)
	sp.Diffuse.Set(mtl.diffuse)
	sp.DiffuseMap.Set2D(mtl.diffuseMap)
	sp.Specular.Set(mtl.specular)
	sp.SpecularMap.Set2D(mtl.specularMap)
	sp.Shine.Set(mtl.shine)
	sp.Alpha.Set(mtl.alpha)

	if mtl.HasAlphaMap() {
		sp.HasAlphaMap.Set(true)
		sp.AlphaMap.Set2D(mtl.alphaMap)
	} else {
		sp.HasAlphaMap.Set(false)
	}

	if mtl.HasBumpMap() {
		sp.HasBumpMap.Set(true)
		sp.BumpMap.Set2D(mtl.bumpMap)
	} else {
		sp.HasBumpMap.Set(false)
	}

	var v Vertex
	sp.Position.SetSource(sm.vbo, v.PositionOffset(), v.Size())
	sp.Normal.SetSource(sm.vbo, v.NormalOffset(), v.Size())
	sp.TexCoord.SetSource(sm.vbo, v.TexCoordOffset(), v.Size())
	sp.Tangent.SetSource(sm.vbo, v.TangentOffset(), v.Size())
	sp.SetAttribIndexBuffer(sm.ibo)
}

func (sp *MeshShaderProgram) SetAmbientLight(l *AmbientLight) {
	sp.LightType.Set(0)
	sp.AmbientLight.Set(l.color)
}

func (sp *MeshShaderProgram) SetPointLight(l *PointLight) {
	sp.LightType.Set(1)
	sp.LightPos.Set(l.position)
	sp.DiffuseLight.Set(l.diffuse)
	sp.SpecularLight.Set(l.specular)
	sp.CubeShadowMap.SetCube(l.shadowMap)
	sp.ShadowFar.Set(l.shadowFar)
}

func (sp *MeshShaderProgram) SetSpotLight(l *SpotLight) {
	sp.LightType.Set(2)
	sp.LightPos.Set(l.position)
	sp.LightDir.Set(l.Forward())
	sp.DiffuseLight.Set(l.diffuse)
	sp.SpecularLight.Set(l.specular)
	sp.SpotShadowMap.Set2D(l.shadowMap)
	sp.ShadowViewMatrix.Set(l.ViewMatrix())
	sp.ShadowProjectionMatrix.Set(l.ProjectionMatrix())
	sp.ShadowFar.Set(l.Camera.far)
}

func NewSkyboxShaderProgram() *SkyboxShaderProgram {
	var sp SkyboxShaderProgram
	var err error

	vShaderFilename := "shaders/skyboxvshader.glsl"
	fShaderFilename := "shaders/skyboxfshader.glsl"

	sp.ShaderProgram, err = ReadShaderProgram(vShaderFilename, fShaderFilename, "")
	if err != nil {
		panic(err)
	}

	sp.ViewMatrix = sp.UniformMatrix4("viewMatrix")
	sp.ProjectionMatrix = sp.UniformMatrix4("projectionMatrix")
	sp.CubeMap = sp.UniformSampler("cubeMap")
	sp.Position = sp.Attrib("positionV")

	return &sp
}

func (sp *SkyboxShaderProgram) SetCamera(c *Camera) {
	sp.ViewMatrix.Set(c.ViewMatrix())
	sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (sp *SkyboxShaderProgram) SetSkybox(skybox *CubeMap) {
	sp.CubeMap.SetCube(skybox)
}

func (sp *SkyboxShaderProgram) SetCube(vbo, ibo *Buffer) {
	sp.Position.SetFormat(gl.FLOAT, false)
	sp.Position.SetSource(vbo, 0, int(unsafe.Sizeof(math.NewVec3(0, 0, 0))))
	sp.SetAttribIndexBuffer(ibo)
}

func NewTextShaderProgram() *TextShaderProgram {
	var sp TextShaderProgram
	var err error

	vShaderFilename := "shaders/textvshader.glsl"
	fShaderFilename := "shaders/textfshader.glsl"
	sp.ShaderProgram, err = ReadShaderProgram(vShaderFilename, fShaderFilename, "")
	if err != nil {
		panic(err)
	}

	sp.Atlas = sp.UniformSampler("fontAtlas")
	sp.Position = sp.Attrib("position")
	sp.TexCoord = sp.Attrib("texCoordV")

	sp.Position.SetFormat(gl.FLOAT, false)
	sp.TexCoord.SetFormat(gl.FLOAT, false)

	return &sp
}

func (sp *TextShaderProgram) SetAtlas(tex *Texture2D) {
	sp.Atlas.Set2D(tex)
}

func (sp *TextShaderProgram) SetAttribs(vbo, ibo *Buffer) {
	var v Vertex
	sp.Position.SetSource(vbo, v.PositionOffset(), v.Size())
	sp.TexCoord.SetSource(vbo, v.TexCoordOffset(), v.Size())
	sp.SetAttribIndexBuffer(ibo)
}

func NewShadowMapShaderProgram() *ShadowMapShaderProgram {
	var sp ShadowMapShaderProgram
	var err error

	vShaderFilename := "shaders/pointlightshadowmapvshader.glsl"
	fShaderFilename := "shaders/pointlightshadowmapfshader.glsl"
	sp.ShaderProgram, err = ReadShaderProgram(vShaderFilename, fShaderFilename, "")
	if err != nil {
		panic(err)
	}

	sp.ModelMatrix = sp.UniformMatrix4("modelMatrix")
	sp.ViewMatrix = sp.UniformMatrix4("viewMatrix")
	sp.ProjectionMatrix = sp.UniformMatrix4("projectionMatrix")
	sp.LightPosition = sp.UniformVector3("lightPosition")
	sp.Far = sp.UniformFloat("far")
	sp.Position = sp.Attrib("position")

	sp.Position.SetFormat(gl.FLOAT, false)

	return &sp
}

func (sp *ShadowMapShaderProgram) SetCamera(c *Camera) {
	sp.Far.Set(c.far)
	sp.LightPosition.Set(c.position)
	sp.ViewMatrix.Set(c.ViewMatrix())
	sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (sp *ShadowMapShaderProgram) SetMesh(m *Mesh) {
	sp.ModelMatrix.Set(m.WorldMatrix())
}

func (sp *ShadowMapShaderProgram) SetSubMesh(sm *SubMesh) {
	var v Vertex
	sp.Position.SetSource(sm.vbo, v.PositionOffset(), v.Size())
	sp.SetAttribIndexBuffer(sm.ibo)
}

func NewArrowShaderProgram() *ArrowShaderProgram {
	var sp ArrowShaderProgram
	var err error

	vShaderFilename := "shaders/arrowvshader.glsl"
	fShaderFilename := "shaders/arrowfshader.glsl"
	sp.ShaderProgram, err = ReadShaderProgram(vShaderFilename, fShaderFilename, "")
	if err != nil {
		panic(err)
	}

	sp.Position = sp.Attrib("position")
	sp.ModelMatrix = sp.UniformMatrix4("modelMatrix")
	sp.ViewMatrix = sp.UniformMatrix4("viewMatrix")
	sp.ProjectionMatrix = sp.UniformMatrix4("projectionMatrix")
	sp.Color = sp.UniformVector3("color")

	sp.Position.SetFormat(gl.FLOAT, false)

	return &sp
}

func (sp *ArrowShaderProgram) SetCamera(c *Camera) {
	sp.ViewMatrix.Set(c.ViewMatrix())
	sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (sp *ArrowShaderProgram) SetMesh(m *Mesh) {
	sp.ModelMatrix.Set(m.WorldMatrix())
}

func (sp *ArrowShaderProgram) SetColor(color math.Vec3) {
	sp.Color.Set(color)
}

func (sp *ArrowShaderProgram) SetPosition(vbo *Buffer) {
	stride := int(unsafe.Sizeof(math.NewVec3(0, 0, 0)))
	sp.Position.SetSource(vbo, 0, stride)
}

func NewDepthPassShaderProgram() *DepthPassShaderProgram {
	var sp DepthPassShaderProgram
	var err error

	vShaderFilename := "shaders/depthpassvshader.glsl"
	sp.ShaderProgram, err = ReadShaderProgram(vShaderFilename, "", "")
	if err != nil {
		panic(err)
	}

	sp.Position = sp.Attrib("position")
	sp.ModelMatrix = sp.UniformMatrix4("modelMatrix")
	sp.ViewMatrix = sp.UniformMatrix4("viewMatrix")
	sp.ProjectionMatrix = sp.UniformMatrix4("projectionMatrix")
	sp.Position.SetFormat(gl.FLOAT, false)

	return &sp
}

func (sp *DepthPassShaderProgram) SetCamera(c *Camera) {
	sp.ViewMatrix.Set(c.ViewMatrix())
	sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (sp *DepthPassShaderProgram) SetMesh(m *Mesh) {
	sp.ModelMatrix.Set(m.WorldMatrix())
}

func (sp *DepthPassShaderProgram) SetSubMesh(sm *SubMesh) {
	var v Vertex
	sp.Position.SetSource(sm.vbo, v.PositionOffset(), v.Size())
	sp.SetAttribIndexBuffer(sm.ibo)
}
