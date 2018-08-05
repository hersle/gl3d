package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"unsafe"
)

type MeshShaderProgram struct {
	*ShaderProgram

	position *Attrib
	texCoord *Attrib
	normal *Attrib
	tangent *Attrib

	modelMatrix *UniformMatrix4
	viewMatrix *UniformMatrix4
	projectionMatrix *UniformMatrix4

	ambient *UniformVector3
	ambientMap *UniformSampler
	diffuse *UniformVector3
	diffuseMap *UniformSampler
	specular *UniformVector3
	specularMap *UniformSampler
	shine *UniformFloat
	alpha *UniformFloat
	alphaMap *UniformSampler
	hasAlphaMap *UniformBool
	bumpMap *UniformSampler
	hasBumpMap *UniformBool

	lightType *UniformInteger
	lightPos *UniformVector3
	lightDir *UniformVector3
	ambientLight *UniformVector3
	diffuseLight *UniformVector3
	specularLight *UniformVector3
	shadowViewMatrix *UniformMatrix4
	shadowProjectionMatrix *UniformMatrix4
	spotShadowMap *UniformSampler
	cubeShadowMap *UniformSampler
}

type SkyboxShaderProgram struct {
	*ShaderProgram
	viewMatrix *UniformMatrix4
	projectionMatrix *UniformMatrix4
	cubeMap *UniformSampler
	position *Attrib
}

type TextShaderProgram struct {
	*ShaderProgram
	atlas *UniformSampler
	position *Attrib
	texCoord *Attrib
}

type ShadowMapShaderProgram struct {
	*ShaderProgram
	modelMatrix *UniformMatrix4
	viewMatrix *UniformMatrix4
	projectionMatrix *UniformMatrix4
	lightPosition *UniformVector3
	far *UniformFloat
	position *Attrib
}

type ArrowShaderProgram struct {
	*ShaderProgram
	modelMatrix *UniformMatrix4
	viewMatrix *UniformMatrix4
	projectionMatrix *UniformMatrix4
	color *UniformVector3
	position *Attrib
}

type DepthPassShaderProgram struct {
	*ShaderProgram
	modelMatrix *UniformMatrix4
	viewMatrix *UniformMatrix4
	projectionMatrix *UniformMatrix4
	position *Attrib
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

	sp.position = sp.Attrib("position")
	sp.texCoord = sp.Attrib("texCoordV")
	sp.normal = sp.Attrib("normalV")
	sp.tangent = sp.Attrib("tangentV")

	sp.modelMatrix = sp.UniformMatrix4("modelMatrix")
	sp.viewMatrix = sp.UniformMatrix4("viewMatrix")
	sp.projectionMatrix = sp.UniformMatrix4("projectionMatrix")

	sp.ambient = sp.UniformVector3("material.ambient")
	sp.ambientMap = sp.UniformSampler("material.ambientMap")
	sp.diffuse = sp.UniformVector3("material.diffuse")
	sp.diffuseMap = sp.UniformSampler("material.diffuseMap")
	sp.specular = sp.UniformVector3("material.specular")
	sp.specularMap = sp.UniformSampler("material.specularMap")
	sp.shine = sp.UniformFloat("material.shine")
	sp.alpha = sp.UniformFloat("material.alpha")
	sp.alphaMap = sp.UniformSampler("material.alphaMap")
	sp.hasAlphaMap = sp.UniformBool("material.hasAlphaMap")
	sp.bumpMap = sp.UniformSampler("material.bumpMap")
	sp.hasBumpMap = sp.UniformBool("material.hasBumpMap")

	sp.lightType = sp.UniformInteger("light.type")
	sp.lightPos = sp.UniformVector3("light.position")
	sp.lightDir = sp.UniformVector3("light.direction")
	sp.ambientLight = sp.UniformVector3("light.ambient")
	sp.diffuseLight = sp.UniformVector3("light.diffuse")
	sp.specularLight = sp.UniformVector3("light.specular")
	sp.shadowViewMatrix = sp.UniformMatrix4("shadowViewMatrix")
	sp.shadowProjectionMatrix = sp.UniformMatrix4("shadowProjectionMatrix")
	sp.cubeShadowMap = sp.UniformSampler("cubeShadowMap")
	sp.spotShadowMap = sp.UniformSampler("spotShadowMap")

	sp.position.SetFormat(gl.FLOAT, false)
	sp.normal.SetFormat(gl.FLOAT, false)
	sp.texCoord.SetFormat(gl.FLOAT, false)
	sp.tangent.SetFormat(gl.FLOAT, false)

	return &sp
}

func (sp *MeshShaderProgram) SetCamera(c *Camera) {
	sp.viewMatrix.Set(c.ViewMatrix())
	sp.projectionMatrix.Set(c.ProjectionMatrix())
}

func (sp *MeshShaderProgram) SetMesh(m *Mesh) {
	sp.modelMatrix.Set(m.WorldMatrix())
}

func (sp *MeshShaderProgram) SetSubMesh(sm *SubMesh) {
	mtl := sm.mtl

	sp.ambient.Set(mtl.ambient)
	sp.ambientMap.Set2D(mtl.ambientMap)
	sp.diffuse.Set(mtl.diffuse)
	sp.diffuseMap.Set2D(mtl.diffuseMap)
	sp.specular.Set(mtl.specular)
	sp.specularMap.Set2D(mtl.specularMap)
	sp.shine.Set(mtl.shine)
	sp.alpha.Set(mtl.alpha)

	if mtl.HasAlphaMap() {
		sp.hasAlphaMap.Set(true)
		sp.alphaMap.Set2D(mtl.alphaMap)
	} else {
		sp.hasAlphaMap.Set(false)
	}

	if mtl.HasBumpMap() {
		sp.hasBumpMap.Set(true)
		sp.bumpMap.Set2D(mtl.bumpMap)
	} else {
		sp.hasBumpMap.Set(false)
	}

	stride := int(unsafe.Sizeof(Vertex{}))
	offset1 := int(unsafe.Offsetof(Vertex{}.pos))
	offset2 := int(unsafe.Offsetof(Vertex{}.normal))
	offset3 := int(unsafe.Offsetof(Vertex{}.texCoord))
	offset4 := int(unsafe.Offsetof(Vertex{}.tangent))
	sp.position.SetSource(sm.vbo, offset1, stride)
	sp.normal.SetSource(sm.vbo, offset2, stride)
	sp.texCoord.SetSource(sm.vbo, offset3, stride)
	sp.tangent.SetSource(sm.vbo, offset4, stride)
	sp.SetAttribIndexBuffer(sm.ibo)
}

func (sp *MeshShaderProgram) SetAmbientLight(l *AmbientLight) {
	sp.lightType.Set(0)
	sp.ambientLight.Set(l.color)
}

func (sp *MeshShaderProgram) SetPointLight(l *PointLight) {
	sp.lightType.Set(1)
	sp.lightPos.Set(l.position)
	sp.diffuseLight.Set(l.diffuse)
	sp.specularLight.Set(l.specular)
	sp.cubeShadowMap.SetCube(l.shadowMap)
}

func (sp *MeshShaderProgram) SetSpotLight(l *SpotLight) {
	sp.lightType.Set(2)
	sp.lightPos.Set(l.position)
	sp.lightDir.Set(l.Forward())
	sp.diffuseLight.Set(l.diffuse)
	sp.specularLight.Set(l.specular)
	sp.spotShadowMap.Set2D(l.shadowMap)
	sp.shadowViewMatrix.Set(l.ViewMatrix())
	sp.shadowProjectionMatrix.Set(l.ProjectionMatrix())
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

	sp.viewMatrix = sp.UniformMatrix4("viewMatrix")
	sp.projectionMatrix = sp.UniformMatrix4("projectionMatrix")
	sp.cubeMap = sp.UniformSampler("cubeMap")
	sp.position = sp.Attrib("positionV")

	return &sp
}

func (sp *SkyboxShaderProgram) SetCamera(c *Camera) {
	sp.viewMatrix.Set(c.ViewMatrix())
	sp.projectionMatrix.Set(c.ProjectionMatrix())
}

func (sp *SkyboxShaderProgram) SetSkybox(skybox *CubeMap) {
	sp.cubeMap.SetCube(skybox)
}

func (sp *SkyboxShaderProgram) SetCube(vbo, ibo *Buffer) {
	sp.position.SetFormat(gl.FLOAT, false)
	sp.position.SetSource(vbo, 0, int(unsafe.Sizeof(NewVec3(0, 0, 0))))
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

	sp.atlas = sp.UniformSampler("fontAtlas")
	sp.position = sp.Attrib("position")
	sp.texCoord = sp.Attrib("texCoordV")

	sp.position.SetFormat(gl.FLOAT, false)
	sp.texCoord.SetFormat(gl.FLOAT, false)

	return &sp
}

func (sp *TextShaderProgram) SetAtlas(tex *Texture2D) {
	sp.atlas.Set2D(tex)
}

func (sp *TextShaderProgram) SetAttribs(vbo, ibo *Buffer) {
	stride := int(unsafe.Sizeof(Vertex{}))
	offset1 := int(unsafe.Offsetof(Vertex{}.pos))
	offset2 := int(unsafe.Offsetof(Vertex{}.texCoord))
	sp.position.SetSource(vbo, offset1, stride)
	sp.texCoord.SetSource(vbo, offset2, stride)
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

	sp.modelMatrix = sp.UniformMatrix4("modelMatrix")
	sp.viewMatrix = sp.UniformMatrix4("viewMatrix")
	sp.projectionMatrix = sp.UniformMatrix4("projectionMatrix")
	sp.lightPosition = sp.UniformVector3("lightPosition")
	sp.far = sp.UniformFloat("far")
	sp.position = sp.Attrib("position")

	sp.position.SetFormat(gl.FLOAT, false)

	return &sp
}

func (sp *ShadowMapShaderProgram) SetCamera(c *Camera) {
	sp.far.Set(c.far)
	sp.lightPosition.Set(c.position)
	sp.viewMatrix.Set(c.ViewMatrix())
	sp.projectionMatrix.Set(c.ProjectionMatrix())
}

func (sp *ShadowMapShaderProgram) SetMesh(m *Mesh) {
	sp.modelMatrix.Set(m.WorldMatrix())
}

func (sp *ShadowMapShaderProgram) SetSubMesh(sm *SubMesh) {
	stride := int(unsafe.Sizeof(Vertex{}))
	offset := int(unsafe.Offsetof(Vertex{}.pos))
	sp.position.SetSource(sm.vbo, offset, stride)
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

	sp.position = sp.Attrib("position")
	sp.modelMatrix = sp.UniformMatrix4("modelMatrix")
	sp.viewMatrix = sp.UniformMatrix4("viewMatrix")
	sp.projectionMatrix = sp.UniformMatrix4("projectionMatrix")
	sp.color = sp.UniformVector3("color")

	sp.position.SetFormat(gl.FLOAT, false)

	return &sp
}

func (sp *ArrowShaderProgram) SetCamera(c *Camera) {
	sp.viewMatrix.Set(c.ViewMatrix())
	sp.projectionMatrix.Set(c.ProjectionMatrix())
}

func (sp *ArrowShaderProgram) SetMesh(m *Mesh) {
	sp.modelMatrix.Set(m.WorldMatrix())
}

func (sp *ArrowShaderProgram) SetColor(color Vec3) {
	sp.color.Set(color)
}

func (sp *ArrowShaderProgram) SetPosition(vbo *Buffer) {
	stride := int(unsafe.Sizeof(NewVec3(0, 0, 0)))
	sp.position.SetSource(vbo, 0, stride)
}

func NewDepthPassShaderProgram() *DepthPassShaderProgram {
	var sp DepthPassShaderProgram
	var err error

	vShaderFilename := "shaders/depthpassvshader.glsl"
	sp.ShaderProgram, err = ReadShaderProgram(vShaderFilename, "", "")
	if err != nil {
		panic(err)
	}

	sp.position = sp.Attrib("position")
	sp.modelMatrix = sp.UniformMatrix4("modelMatrix")
	sp.viewMatrix = sp.UniformMatrix4("viewMatrix")
	sp.projectionMatrix = sp.UniformMatrix4("projectionMatrix")
	sp.position.SetFormat(gl.FLOAT, false)

	return &sp
}

func (sp *DepthPassShaderProgram) SetCamera(c *Camera) {
	sp.viewMatrix.Set(c.ViewMatrix())
	sp.projectionMatrix.Set(c.ProjectionMatrix())
}

func (sp *DepthPassShaderProgram) SetMesh(m *Mesh) {
	sp.modelMatrix.Set(m.WorldMatrix())
}

func (sp *DepthPassShaderProgram) SetSubMesh(sm *SubMesh) {
	stride := int(unsafe.Sizeof(Vertex{}))
	offset1 := int(unsafe.Offsetof(Vertex{}.pos))
	sp.position.SetSource(sm.vbo, offset1, stride)
	sp.SetAttribIndexBuffer(sm.ibo)
}
