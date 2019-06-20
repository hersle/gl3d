package graphics

import (
	"errors"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/math"
	_ "github.com/hersle/gl3d/window" // initialize graphics
	"io/ioutil"
	"strings"
)

type ShaderType int

const (
	VertexShader ShaderType = iota
	FragmentShader
	GeometryShader
)

type ShaderProgram struct {
	id int
	va *VertexArray
}

type Shader struct {
	id int
}

type Attrib struct {
	prog        *ShaderProgram
	id          int
	nComponents int
}

// TODO: store value, have Set() function and make "Uniform" an interface?
type uniformBasic struct {
	progID   int
	location int
	glType   int
}

type UniformInteger struct {
	uniformBasic
}

type UniformFloat struct {
	uniformBasic
}

type UniformVector2 struct {
	uniformBasic
}

type UniformVector3 struct {
	uniformBasic
}

type UniformVector4 struct {
	uniformBasic
}

type UniformMatrix4 struct {
	uniformBasic
}

type UniformSampler struct {
	uniformBasic
	textureUnitIndex int
}

type UniformBool struct {
	uniformBasic
}

type VertexArray struct {
	id             int
	hasIndexBuffer bool
}

func NewVertexArray() *VertexArray {
	var va VertexArray
	var id uint32
	gl.CreateVertexArrays(1, &id)
	va.id = int(id)
	va.hasIndexBuffer = false
	return &va
}

// TODO: normalize should not be set for some types
func (va *VertexArray) SetAttribFormat(a *Attrib, dim, typ int, normalize bool) {
	if a == nil {
		return
	}
	gl.VertexArrayAttribFormat(uint32(va.id), uint32(a.id), int32(dim), uint32(typ), normalize, 0)
}

func (va *VertexArray) SetAttribSource(a *Attrib, b *Buffer, offset, stride int) {
	if a == nil {
		return
	}
	gl.VertexArrayAttribBinding(uint32(va.id), uint32(a.id), uint32(a.id))
	gl.VertexArrayVertexBuffer(uint32(va.id), uint32(a.id), uint32(b.id), offset, int32(stride))
	gl.EnableVertexArrayAttrib(uint32(va.id), uint32(a.id))
}

func (va *VertexArray) SetIndexBuffer(b *Buffer) {
	gl.VertexArrayElementBuffer(uint32(va.id), uint32(b.id))
	va.hasIndexBuffer = true
}

func (va *VertexArray) Bind() {
	gl.BindVertexArray(uint32(va.id))
}

func NewShader(_type ShaderType, src string) (*Shader, error) {
	var s Shader

	var gltype uint32
	switch _type {
	case VertexShader:
		gltype = gl.VERTEX_SHADER
	case FragmentShader:
		gltype = gl.FRAGMENT_SHADER
	case GeometryShader:
		gltype = gl.GEOMETRY_SHADER
	default:
		panic("unknown shader type")
	}

	s.id = int(gl.CreateShader(gltype))
	s.SetSource(src)
	err := s.Compile()
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func ReadShader(_type ShaderType, filename string) (*Shader, error) {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewShader(_type, string(src))
}

func NewShaderFromTemplate(_type ShaderType, src string, defines []string) (*Shader, error) {
	lines := strings.Split(src, "\n")
	src = lines[0] + "\n" // #version
	for _, define := range defines {
		src = src + "#define " + define + "\n"
	}
	for _, line := range lines[1:] {
		src = src + line + "\n"
	}
	println("shader template source:\n", src)

	return NewShader(_type, src)
}

func ReadShaderFromTemplate(_type ShaderType, filename string, defines []string) (*Shader, error) {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewShaderFromTemplate(_type, string(src), defines)
}

func (s *Shader) SetSource(src string) {
	cSrc, free := gl.Strs(src)
	defer free()
	length := int32(len(src))
	gl.ShaderSource(uint32(s.id), 1, cSrc, &length)
}

func (s *Shader) Compiled() bool {
	var status int32
	gl.GetShaderiv(uint32(s.id), gl.COMPILE_STATUS, &status)
	return status == gl.TRUE
}

func (s *Shader) Log() string {
	var length int32
	gl.GetShaderiv(uint32(s.id), gl.INFO_LOG_LENGTH, &length)
	log := string(make([]byte, length+1))
	gl.GetShaderInfoLog(uint32(s.id), length+1, nil, gl.Str(log))
	log = log[:len(log)-1] // remove null terminator
	return log
}

func (s *Shader) Compile() error {
	gl.CompileShader(uint32(s.id))
	if s.Compiled() {
		return nil
	} else {
		return errors.New(s.Log())
	}
}

func NewShaderProgram(vShader, fShader, gShader *Shader) (*ShaderProgram, error) {
	var p ShaderProgram
	p.id = int(gl.CreateProgram())

	if vShader != nil {
		gl.AttachShader(uint32(p.id), uint32(vShader.id))
		defer gl.DetachShader(uint32(p.id), uint32(vShader.id))
	}
	if fShader != nil {
		gl.AttachShader(uint32(p.id), uint32(fShader.id))
		defer gl.DetachShader(uint32(p.id), uint32(fShader.id))
	}
	if gShader != nil {
		gl.AttachShader(uint32(p.id), uint32(gShader.id))
		defer gl.DetachShader(uint32(p.id), uint32(gShader.id))
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
		vShader, err = ReadShader(VertexShader, vShaderFilename)
		if err != nil {
			return nil, err
		}
	}

	if fShaderFilename == "" {
		fShader = nil
	} else {
		fShader, err = ReadShader(FragmentShader, fShaderFilename)
		if err != nil {
			return nil, err
		}
	}

	if gShaderFilename == "" {
		gShader = nil
	} else {
		gShader, err = ReadShader(GeometryShader, gShaderFilename)
		if err != nil {
			return nil, err
		}
	}

	return NewShaderProgram(vShader, fShader, gShader)
}

func ReadShaderProgramFromTemplates(vShaderFilename, fShaderFilename, gShaderFilename string, defines []string) (*ShaderProgram, error) {
	var vShader, fShader, gShader *Shader
	var err error

	if vShaderFilename == "" {
		vShader = nil
	} else {
		vShader, err = ReadShaderFromTemplate(VertexShader, vShaderFilename, defines)
		if err != nil {
			return nil, err
		}
	}

	if fShaderFilename == "" {
		fShader = nil
	} else {
		fShader, err = ReadShaderFromTemplate(FragmentShader, fShaderFilename, defines)
		if err != nil {
			return nil, err
		}
	}

	if gShaderFilename == "" {
		gShader = nil
	} else {
		gShader, err = ReadShaderFromTemplate(GeometryShader, gShaderFilename, defines)
		if err != nil {
			return nil, err
		}
	}

	return NewShaderProgram(vShader, fShader, gShader)
}

func (p *ShaderProgram) Linked() bool {
	var status int32
	gl.GetProgramiv(uint32(p.id), gl.LINK_STATUS, &status)
	return status == gl.TRUE
}

func (p *ShaderProgram) Log() string {
	var length int32
	gl.GetProgramiv(uint32(p.id), gl.INFO_LOG_LENGTH, &length)
	log := string(make([]byte, length+1))
	gl.GetProgramInfoLog(uint32(p.id), length+1, nil, gl.Str(log))
	log = log[:len(log)-1] // remove null terminator
	return log
}

func (p *ShaderProgram) Link() error {
	gl.LinkProgram(uint32(p.id))
	if p.Linked() {
		return nil
	}
	return errors.New(p.Log())
}

func (u *UniformInteger) Set(i int) {
	if u == nil {
		return
	}
	gl.ProgramUniform1i(uint32(u.progID), int32(u.location), int32(i))
}

func (u *UniformFloat) Set(f float32) {
	if u == nil {
		return
	}
	gl.ProgramUniform1f(uint32(u.progID), int32(u.location), f)
}

func (u *UniformVector2) Set(v math.Vec2) {
	if u == nil {
		return
	}
	gl.ProgramUniform2fv(uint32(u.progID), int32(u.location), 1, &v[0])
}

func (u *UniformVector3) Set(v math.Vec3) {
	if u == nil {
		return
	}
	gl.ProgramUniform3fv(uint32(u.progID), int32(u.location), 1, &v[0])
}

func (u *UniformVector4) Set(v math.Vec4) {
	if u == nil {
		return
	}
	gl.ProgramUniform4fv(uint32(u.progID), int32(u.location), 1, &v[0])
}

func (u *UniformMatrix4) Set(m *math.Mat4) {
	if u == nil {
		return
	}
	gl.ProgramUniformMatrix4fv(uint32(u.progID), int32(u.location), 1, true, &m[0])
}

func (u *UniformSampler) Set2D(t *Texture2D) {
	if u == nil {
		return
	}
	// TODO: other shaders can mess with this texture index
	gl.BindTextureUnit(uint32(u.textureUnitIndex), uint32(t.id))
	gl.ProgramUniform1i(uint32(u.progID), int32(u.location), int32(u.textureUnitIndex))
}

func (u *UniformSampler) SetCube(t *CubeMap) {
	if u == nil {
		return
	}
	// TODO: other shaders can mess with this texture index
	gl.BindTextureUnit(uint32(u.textureUnitIndex), uint32(t.id))
	gl.ProgramUniform1i(uint32(u.progID), int32(u.location), int32(u.textureUnitIndex))
}

func (u *UniformBool) Set(b bool) {
	if u == nil {
		return
	}
	var i int32
	if b {
		i = 1
	} else {
		i = 0
	}
	gl.ProgramUniform1i(uint32(u.progID), int32(u.location), i)
}

func (a *Attrib) SetFormat(_type int, normalize bool) {
	if a == nil {
		return
	}
	a.prog.va.SetAttribFormat(a, a.nComponents, _type, normalize)
}

func (a *Attrib) SetSource(b *Buffer, offset, stride int) {
	if a == nil {
		return
	}
	a.prog.va.SetAttribSource(a, b, offset, stride)
}

func (p *ShaderProgram) SetAttribIndexBuffer(b *Buffer) {
	p.va.SetIndexBuffer(b)
}

func (p *ShaderProgram) Bind() {
	gl.UseProgram(uint32(p.id))
}

func (p *ShaderProgram) Attrib(name string) *Attrib {
	var a Attrib
	loc := gl.GetAttribLocation(uint32(p.id), gl.Str(name+"\x00"))
	if loc == -1 {
		return nil
	}
	a.id = int(loc)
	a.prog = p

	var size int32
	var typ uint32
	gl.GetActiveAttrib(uint32(a.prog.id), uint32(a.id), 0, nil, &size, &typ, nil)

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

func (p *ShaderProgram) uniformBasic(name string) *uniformBasic {
	var u uniformBasic
	loc := gl.GetUniformLocation(uint32(p.id), gl.Str(name+"\x00"))
	if loc == -1 {
		println("error getting uniform " + name)
		return nil
	}
	u.location = int(loc)
	u.progID = p.id
	index := gl.GetProgramResourceIndex(uint32(p.id), gl.UNIFORM, gl.Str(name+"\x00"))
	var gltype uint32
	gl.GetActiveUniform(uint32(p.id), index, 0, nil, nil, &gltype, nil)
	u.glType = int(gltype)
	return &u
}

func (p *ShaderProgram) UniformInteger(name string) *UniformInteger {
	var u UniformInteger
	ptr := p.uniformBasic(name)
	if ptr == nil {
		return nil
	}
	u.uniformBasic = *ptr
	if u.glType != gl.INT {
		return nil
	}
	return &u
}

func (p *ShaderProgram) UniformFloat(name string) *UniformFloat {
	var u UniformFloat
	ptr := p.uniformBasic(name)
	if ptr == nil {
		return nil
	}
	u.uniformBasic = *ptr
	if u.glType != gl.FLOAT {
		return nil
	}
	return &u
}

func (p *ShaderProgram) UniformVector2(name string) *UniformVector2 {
	var u UniformVector2
	ptr := p.uniformBasic(name)
	if ptr == nil {
		return nil
	}
	u.uniformBasic = *ptr
	if u.glType != gl.FLOAT_VEC2 {
		return nil
	}
	return &u
}

func (p *ShaderProgram) UniformVector3(name string) *UniformVector3 {
	var u UniformVector3
	ptr := p.uniformBasic(name)
	if ptr == nil {
		return nil
	}
	u.uniformBasic = *ptr
	if u.glType != gl.FLOAT_VEC3 {
		return nil
	}
	return &u
}

func (p *ShaderProgram) UniformVector4(name string) *UniformVector4 {
	var u UniformVector4
	ptr := p.uniformBasic(name)
	if ptr == nil {
		return nil
	}
	u.uniformBasic = *ptr
	if u.glType != gl.FLOAT_VEC4 {
		return nil
	}
	return &u
}

func (p *ShaderProgram) UniformMatrix4(name string) *UniformMatrix4 {
	var u UniformMatrix4
	ptr := p.uniformBasic(name)
	if ptr == nil {
		return nil
	}
	u.uniformBasic = *ptr
	// TODO: what if things not found?
	if u.glType != gl.FLOAT_MAT4 {
		return nil
	}
	return &u
}

var textureUnitsUsed int = 0

func (p *ShaderProgram) UniformSampler(name string) *UniformSampler {
	var u UniformSampler
	ptr := p.uniformBasic(name)
	if ptr == nil {
		return nil
	}
	u.uniformBasic = *ptr
	if u.glType != gl.SAMPLER_2D && u.glType != gl.SAMPLER_CUBE { // TODO: allow more sampler types
		return nil
	}
	u.textureUnitIndex = textureUnitsUsed // TODO: make texture unit mapping more sophisticated
	textureUnitsUsed++
	return &u
}

func (p *ShaderProgram) UniformBool(name string) *UniformBool {
	var u UniformBool
	ptr := p.uniformBasic(name)
	if ptr == nil {
		return nil
	}
	u.uniformBasic = *ptr
	// TODO: what if things not found?
	if u.glType != gl.BOOL {
		return nil
	}
	return &u
}
