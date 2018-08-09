package graphics

import (
	"errors"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/math"
	_ "github.com/hersle/gl3d/window" // initialize graphics
	"io/ioutil"
)

type ShaderProgram struct {
	id uint32
	va *VertexArray
}

type Shader struct {
	id uint32
}

type Attrib struct {
	prog        *ShaderProgram
	id          uint32
	nComponents int
}

// TODO: store value, have Set() function and make "Uniform" an interface?
type UniformBasic struct {
	progID   uint32
	location uint32
	glType   uint32
}

type UniformInteger struct {
	UniformBasic
}

type UniformFloat struct {
	UniformBasic
}

type UniformVector2 struct {
	UniformBasic
}

type UniformVector3 struct {
	UniformBasic
}

type UniformVector4 struct {
	UniformBasic
}

type UniformMatrix4 struct {
	UniformBasic
}

type UniformSampler struct {
	UniformBasic
	textureUnitIndex uint32
}

type UniformBool struct {
	UniformBasic
}

type VertexArray struct {
	id             uint32
	hasIndexBuffer bool
}

func NewVertexArray() *VertexArray {
	var va VertexArray
	gl.CreateVertexArrays(1, &va.id)
	va.hasIndexBuffer = false
	return &va
}

// TODO: normalize should not be set for some types
func (va *VertexArray) SetAttribFormat(a *Attrib, dim, typ int, normalize bool) {
	gl.VertexArrayAttribFormat(va.id, a.id, int32(dim), uint32(typ), normalize, 0)
}

func (va *VertexArray) SetAttribSource(a *Attrib, b *Buffer, offset, stride int) {
	gl.VertexArrayAttribBinding(va.id, a.id, a.id)
	gl.VertexArrayVertexBuffer(va.id, a.id, b.id, offset, int32(stride))
	gl.EnableVertexArrayAttrib(va.id, a.id)
}

func (va *VertexArray) SetIndexBuffer(b *Buffer) {
	gl.VertexArrayElementBuffer(va.id, b.id)
	va.hasIndexBuffer = true
}

func (va *VertexArray) Bind() {
	gl.BindVertexArray(va.id)
}

func NewShader(typ uint32, src string) (*Shader, error) {
	var s Shader
	s.id = gl.CreateShader(typ)
	s.SetSource(src)
	err := s.Compile()
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func ReadShader(typ uint32, filename string) (*Shader, error) {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewShader(typ, string(src))
}

func (s *Shader) SetSource(src string) {
	cSrc, free := gl.Strs(src)
	defer free()
	length := int32(len(src))
	gl.ShaderSource(s.id, 1, cSrc, &length)
}

func (s *Shader) Compiled() bool {
	var status int32
	gl.GetShaderiv(s.id, gl.COMPILE_STATUS, &status)
	return status == gl.TRUE
}

func (s *Shader) Log() string {
	var length int32
	gl.GetShaderiv(s.id, gl.INFO_LOG_LENGTH, &length)
	log := string(make([]byte, length+1))
	gl.GetShaderInfoLog(s.id, length+1, nil, gl.Str(log))
	log = log[:len(log)-1] // remove null terminator
	return log
}

func (s *Shader) Compile() error {
	gl.CompileShader(s.id)
	if s.Compiled() {
		return nil
	} else {
		return errors.New(s.Log())
	}
}

func NewShaderProgram(vShader, fShader, gShader *Shader) (*ShaderProgram, error) {
	var p ShaderProgram
	p.id = gl.CreateProgram()

	if vShader != nil {
		gl.AttachShader(p.id, vShader.id)
		defer gl.DetachShader(p.id, vShader.id)
	}
	if fShader != nil {
		gl.AttachShader(p.id, fShader.id)
		defer gl.DetachShader(p.id, fShader.id)
	}
	if gShader != nil {
		gl.AttachShader(p.id, gShader.id)
		defer gl.DetachShader(p.id, gShader.id)
	}

	err := p.Link()
	if err != nil {
		return nil, err
	}

	p.va = NewVertexArray()
	return &p, err
}

func ReadShaderProgram(vShaderFilename, fShaderFilename, gShaderFilename string) (*ShaderProgram, error) {
	var vShader, fShader, gShader *Shader
	var err error

	if vShaderFilename == "" {
		vShader = nil
	} else {
		vShader, err = ReadShader(gl.VERTEX_SHADER, vShaderFilename)
		if err != nil {
			return nil, err
		}
	}

	if fShaderFilename == "" {
		fShader = nil
	} else {
		fShader, err = ReadShader(gl.FRAGMENT_SHADER, fShaderFilename)
		if err != nil {
			return nil, err
		}
	}

	if gShaderFilename == "" {
		gShader = nil
	} else {
		gShader, err = ReadShader(gl.GEOMETRY_SHADER, gShaderFilename)
		if err != nil {
			return nil, err
		}
	}

	return NewShaderProgram(vShader, fShader, gShader)
}

func (p *ShaderProgram) Linked() bool {
	var status int32
	gl.GetProgramiv(p.id, gl.LINK_STATUS, &status)
	return status == gl.TRUE
}

func (p *ShaderProgram) Log() string {
	var length int32
	gl.GetProgramiv(p.id, gl.INFO_LOG_LENGTH, &length)
	log := string(make([]byte, length+1))
	gl.GetProgramInfoLog(p.id, length+1, nil, gl.Str(log))
	log = log[:len(log)-1] // remove null terminator
	return log
}

func (p *ShaderProgram) Link() error {
	gl.LinkProgram(p.id)
	if p.Linked() {
		return nil
	}
	return errors.New(p.Log())
}

func (u *UniformInteger) Set(i int) {
	gl.ProgramUniform1i(u.progID, int32(u.location), int32(i))
}

func (u *UniformFloat) Set(f float32) {
	gl.ProgramUniform1f(u.progID, int32(u.location), f)
}

func (u *UniformVector2) Set(v math.Vec2) {
	gl.ProgramUniform2fv(u.progID, int32(u.location), 1, &v[0])
}

func (u *UniformVector3) Set(v math.Vec3) {
	gl.ProgramUniform3fv(u.progID, int32(u.location), 1, &v[0])
}

func (u *UniformVector4) Set(v math.Vec4) {
	gl.ProgramUniform4fv(u.progID, int32(u.location), 1, &v[0])
}

func (u *UniformMatrix4) Set(m *math.Mat4) {
	gl.ProgramUniformMatrix4fv(u.progID, int32(u.location), 1, true, &m[0])
}

func (u *UniformSampler) Set2D(t *Texture2D) {
	// TODO: other shaders can mess with this texture index
	gl.BindTextureUnit(u.textureUnitIndex, t.id)
	gl.ProgramUniform1i(u.progID, int32(u.location), int32(u.textureUnitIndex))
}

func (u *UniformSampler) SetCube(t *CubeMap) {
	// TODO: other shaders can mess with this texture index
	gl.BindTextureUnit(u.textureUnitIndex, t.id)
	gl.ProgramUniform1i(u.progID, int32(u.location), int32(u.textureUnitIndex))
}

func (u *UniformBool) Set(b bool) {
	var i int32
	if b {
		i = 1
	} else {
		i = 0
	}
	gl.ProgramUniform1i(u.progID, int32(u.location), i)
}

func (a *Attrib) SetFormat(typ int, normalize bool) {
	a.prog.va.SetAttribFormat(a, a.nComponents, typ, normalize)
}

func (a *Attrib) SetSource(b *Buffer, offset, stride int) {
	a.prog.va.SetAttribSource(a, b, offset, stride)
}

func (p *ShaderProgram) SetAttribIndexBuffer(b *Buffer) {
	p.va.SetIndexBuffer(b)
}

func (p *ShaderProgram) Bind() {
	gl.UseProgram(p.id)
}

func (p *ShaderProgram) Attrib(name string) *Attrib {
	var a Attrib
	loc := gl.GetAttribLocation(p.id, gl.Str(name+"\x00"))
	if loc == -1 {
		return nil
	}
	a.id = uint32(loc)
	a.prog = p

	var size int32
	var typ uint32
	gl.GetActiveAttrib(a.prog.id, a.id, 0, nil, &size, &typ, nil)

	switch typ {
	case gl.FLOAT:
		a.nComponents = 1
	case gl.FLOAT_VEC2:
		a.nComponents = 2
	case gl.FLOAT_VEC3:
		a.nComponents = 3
	case gl.FLOAT_VEC4:
		a.nComponents = 4
	default:
		panic("unrecognized attribute GL type")
	}

	return &a
}

func (p *ShaderProgram) UniformBasic(name string) *UniformBasic {
	var u UniformBasic
	loc := gl.GetUniformLocation(p.id, gl.Str(name+"\x00"))
	if loc == -1 {
		return nil
	}
	u.location = uint32(loc)
	u.progID = p.id
	index := gl.GetProgramResourceIndex(p.id, gl.UNIFORM, gl.Str(name+"\x00"))
	gl.GetActiveUniform(p.id, index, 0, nil, nil, &u.glType, nil)
	return &u
}

func (p *ShaderProgram) UniformInteger(name string) *UniformInteger {
	var u UniformInteger
	u.UniformBasic = *p.UniformBasic(name)
	if u.glType != gl.INT {
		return nil
	}
	return &u
}

func (p *ShaderProgram) UniformFloat(name string) *UniformFloat {
	var u UniformFloat
	u.UniformBasic = *p.UniformBasic(name)
	if u.glType != gl.FLOAT {
		return nil
	}
	return &u
}

func (p *ShaderProgram) UniformVector2(name string) *UniformVector2 {
	var u UniformVector2
	u.UniformBasic = *p.UniformBasic(name)
	if u.glType != gl.FLOAT_VEC2 {
		return nil
	}
	return &u
}

func (p *ShaderProgram) UniformVector3(name string) *UniformVector3 {
	var u UniformVector3
	u.UniformBasic = *p.UniformBasic(name)
	if u.glType != gl.FLOAT_VEC3 {
		return nil
	}
	return &u
}

func (p *ShaderProgram) UniformVector4(name string) *UniformVector4 {
	var u UniformVector4
	u.UniformBasic = *p.UniformBasic(name)
	if u.glType != gl.FLOAT_VEC4 {
		return nil
	}
	return &u
}

func (p *ShaderProgram) UniformMatrix4(name string) *UniformMatrix4 {
	var u UniformMatrix4
	u.UniformBasic = *p.UniformBasic(name)
	// TODO: what if things not found?
	if u.glType != gl.FLOAT_MAT4 {
		return nil
	}
	return &u
}

var textureUnitsUsed uint32 = 0

func (p *ShaderProgram) UniformSampler(name string) *UniformSampler {
	var u UniformSampler
	u.UniformBasic = *p.UniformBasic(name)
	if u.glType != gl.SAMPLER_2D && u.glType != gl.SAMPLER_CUBE { // TODO: allow more sampler types
		return nil
	}
	u.textureUnitIndex = textureUnitsUsed // TODO: make texture unit mapping more sophisticated
	textureUnitsUsed++
	return &u
}

func (p *ShaderProgram) UniformBool(name string) *UniformBool {
	var u UniformBool
	u.UniformBasic = *p.UniformBasic(name)
	// TODO: what if things not found?
	if u.glType != gl.BOOL {
		return nil
	}
	return &u
}

type MeshShaderProgram struct {
	*ShaderProgram

	Position *Attrib
	TexCoord *Attrib
	Normal   *Attrib
	Tangent  *Attrib

	ModelMatrix      *UniformMatrix4
	ViewMatrix       *UniformMatrix4
	ProjectionMatrix *UniformMatrix4

	Ambient     *UniformVector3
	AmbientMap  *UniformSampler
	Diffuse     *UniformVector3
	DiffuseMap  *UniformSampler
	Specular    *UniformVector3
	SpecularMap *UniformSampler
	Shine       *UniformFloat
	Alpha       *UniformFloat
	AlphaMap    *UniformSampler
	BumpMap     *UniformSampler
	HasBumpMap  *UniformBool

	LightType              *UniformInteger
	LightPos               *UniformVector3
	LightDir               *UniformVector3
	AmbientLight           *UniformVector3
	DiffuseLight           *UniformVector3
	SpecularLight          *UniformVector3
	ShadowViewMatrix       *UniformMatrix4
	ShadowProjectionMatrix *UniformMatrix4
	SpotShadowMap          *UniformSampler
	CubeShadowMap          *UniformSampler
	DirShadowMap           *UniformSampler
	ShadowFar              *UniformFloat
}

type SkyboxShaderProgram struct {
	*ShaderProgram
	ViewMatrix       *UniformMatrix4
	ProjectionMatrix *UniformMatrix4
	CubeMap          *UniformSampler
	Position         *Attrib
}

type TextShaderProgram struct {
	*ShaderProgram
	Atlas    *UniformSampler
	Position *Attrib
	TexCoord *Attrib
}

type ShadowMapShaderProgram struct {
	*ShaderProgram
	ModelMatrix      *UniformMatrix4
	ViewMatrix       *UniformMatrix4
	ProjectionMatrix *UniformMatrix4
	LightPosition    *UniformVector3
	Far              *UniformFloat
	Position         *Attrib
}

type ArrowShaderProgram struct {
	*ShaderProgram
	ModelMatrix      *UniformMatrix4
	ViewMatrix       *UniformMatrix4
	ProjectionMatrix *UniformMatrix4
	Color            *UniformVector3
	Position         *Attrib
}

type DepthPassShaderProgram struct {
	*ShaderProgram
	ModelMatrix      *UniformMatrix4
	ViewMatrix       *UniformMatrix4
	ProjectionMatrix *UniformMatrix4
	Position         *Attrib
}

type QuadShaderProgram struct {
	*ShaderProgram
	Position *Attrib
	Texture *UniformSampler
}

type DirectionalLightShadowMapShaderProgram struct {
	*ShaderProgram
	ModelMatrix *UniformMatrix4
	ViewMatrix *UniformMatrix4
	ProjectionMatrix *UniformMatrix4
	Position *Attrib
}

func NewMeshShaderProgram() *MeshShaderProgram {
	var sp MeshShaderProgram
	var err error

	vShaderFilename := "graphics/shaders/meshvshader.glsl" // TODO: make independent from executable directory
	fShaderFilename := "graphics/shaders/meshfshader.glsl" // TODO: make independent from executable directory

	sp.ShaderProgram, err = ReadShaderProgram(vShaderFilename, fShaderFilename, "")
	if err != nil {
		panic(err)
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
	sp.DirShadowMap = sp.UniformSampler("dirShadowMap")
	sp.ShadowFar = sp.UniformFloat("light.far")

	sp.Position.SetFormat(gl.FLOAT, false)
	sp.Normal.SetFormat(gl.FLOAT, false)
	sp.TexCoord.SetFormat(gl.FLOAT, false)
	sp.Tangent.SetFormat(gl.FLOAT, false)

	return &sp
}

func NewSkyboxShaderProgram() *SkyboxShaderProgram {
	var sp SkyboxShaderProgram
	var err error

	vShaderFilename := "graphics/shaders/skyboxvshader.glsl" // TODO: make independent from executable directory
	fShaderFilename := "graphics/shaders/skyboxfshader.glsl" // TODO: make independent from executable directory

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

func NewTextShaderProgram() *TextShaderProgram {
	var sp TextShaderProgram
	var err error

	vShaderFilename := "graphics/shaders/textvshader.glsl" // TODO: make independent from executable directory
	fShaderFilename := "graphics/shaders/textfshader.glsl" // TODO: make independent from executable directory
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

func NewShadowMapShaderProgram() *ShadowMapShaderProgram {
	var sp ShadowMapShaderProgram
	var err error

	vShaderFilename := "graphics/shaders/pointlightshadowmapvshader.glsl" // TODO: make independent from executable directory
	fShaderFilename := "graphics/shaders/pointlightshadowmapfshader.glsl" // TODO: make independent from executable directory
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

func NewArrowShaderProgram() *ArrowShaderProgram {
	var sp ArrowShaderProgram
	var err error

	vShaderFilename := "graphics/shaders/arrowvshader.glsl" // TODO: make independent from executable directory
	fShaderFilename := "graphics/shaders/arrowfshader.glsl" // TODO: make independent from executable directory
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

func NewDepthPassShaderProgram() *DepthPassShaderProgram {
	var sp DepthPassShaderProgram
	var err error

	vShaderFilename := "graphics/shaders/depthpassvshader.glsl" // TODO: make independent from executable directory
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

func NewQuadShaderProgram() *QuadShaderProgram {
	var sp QuadShaderProgram
	var err error

	vShaderFilename := "graphics/shaders/quadvshader.glsl" // TODO: make independent...
	fShaderFilename := "graphics/shaders/quadfshader.glsl" // TODO: make independent...
	sp.ShaderProgram, err = ReadShaderProgram(vShaderFilename, fShaderFilename, "")
	if err != nil {
		panic(err)
	}

	sp.Position = sp.Attrib("position")
	sp.Texture = sp.UniformSampler("tex")

	return &sp
}

func NewDirectionalLightShadowMapShaderProgram() *DirectionalLightShadowMapShaderProgram {
	var sp DirectionalLightShadowMapShaderProgram
	var err error

	vShaderFilename := "graphics/shaders/directionallightvshader.glsl"
	sp.ShaderProgram, err = ReadShaderProgram(vShaderFilename, "", "")
	if err != nil {
		panic(err)
	}

	sp.ModelMatrix = sp.UniformMatrix4("modelMatrix")
	sp.ViewMatrix = sp.UniformMatrix4("viewMatrix")
	sp.ProjectionMatrix = sp.UniformMatrix4("projectionMatrix")
	sp.Position = sp.Attrib("position")
	sp.Position.SetFormat(gl.FLOAT, false)

	return &sp
}
