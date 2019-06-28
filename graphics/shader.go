package graphics

import (
	"errors"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/math"
	_ "github.com/hersle/gl3d/window" // initialize graphics
	"io/ioutil"
	"strings"
	"reflect"
)

type ShaderType int

const (
	VertexShader ShaderType = iota
	FragmentShader
	GeometryShader
)

type ShaderProgram struct {
	id int
	va *vertexArray
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
type Uniform struct {
	progID   int
	location int
	glType   int
	textureUnitIndex int
}

type vertexArray struct {
	id             int
	hasIndexBuffer bool
}

func newVertexArray() *vertexArray {
	var va vertexArray
	var id uint32
	gl.CreateVertexArrays(1, &id)
	va.id = int(id)
	va.hasIndexBuffer = false
	return &va
}

// TODO: normalize should not be set for some types
func (va *vertexArray) setAttribFormat(a *Attrib, dim, typ int, normalize bool) {
	if a == nil {
		return
	}
	gl.VertexArrayAttribFormat(uint32(va.id), uint32(a.id), int32(dim), uint32(typ), normalize, 0)
}

func (va *vertexArray) setAttribSource(a *Attrib, b *Buffer, offset, stride int) {
	if a == nil {
		return
	}
	gl.VertexArrayAttribBinding(uint32(va.id), uint32(a.id), uint32(a.id))
	gl.VertexArrayVertexBuffer(uint32(va.id), uint32(a.id), uint32(b.id), offset, int32(stride))
	gl.EnableVertexArrayAttrib(uint32(va.id), uint32(a.id))
}

func (va *vertexArray) setIndexBuffer(b *Buffer) {
	gl.VertexArrayElementBuffer(uint32(va.id), uint32(b.id))
	va.hasIndexBuffer = true
}

func (va *vertexArray) bind() {
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

	err := p.link()
	if err != nil {
		return nil, err
	}

	p.va = newVertexArray()
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

func (p *ShaderProgram) linked() bool {
	var status int32
	gl.GetProgramiv(uint32(p.id), gl.LINK_STATUS, &status)
	return status == gl.TRUE
}

func (p *ShaderProgram) log() string {
	var length int32
	gl.GetProgramiv(uint32(p.id), gl.INFO_LOG_LENGTH, &length)
	log := string(make([]byte, length+1))
	gl.GetProgramInfoLog(uint32(p.id), length+1, nil, gl.Str(log))
	log = log[:len(log)-1] // remove null terminator
	return log
}

func (p *ShaderProgram) link() error {
	gl.LinkProgram(uint32(p.id))
	if p.linked() {
		return nil
	}
	return errors.New(p.log())
}

func (u *Uniform) Set(value interface{}) {
	if u == nil {
		return
	}

	switch u.glType {
	case gl.BOOL:
		value := value.(int32)
		// TODO: unnecessary?
		if value != 0{
			value = 1
		} else {
			value = 0
		}
		gl.ProgramUniform1i(uint32(u.progID), int32(u.location), int32(value))
	case gl.INT:
		value := value.(int)
		gl.ProgramUniform1i(uint32(u.progID), int32(u.location), int32(value))
	case gl.FLOAT:
		value := value.(float32)
		gl.ProgramUniform1f(uint32(u.progID), int32(u.location), value)
	case gl.FLOAT_VEC2:
		value := value.(math.Vec2)
		gl.ProgramUniform2fv(uint32(u.progID), int32(u.location), 1, &value[0])
	case gl.FLOAT_VEC3:
		value := value.(math.Vec3)
		gl.ProgramUniform3fv(uint32(u.progID), int32(u.location), 1, &value[0])
	case gl.FLOAT_VEC4:
		value := value.(math.Vec4)
		gl.ProgramUniform4fv(uint32(u.progID), int32(u.location), 1, &value[0])
	case gl.FLOAT_MAT4:
		value := value.(*math.Mat4)
		gl.ProgramUniformMatrix4fv(uint32(u.progID), int32(u.location), 1, true, &value[0])
	case gl.SAMPLER_2D:
		// TODO: other shaders can mess with this texture index
		value := value.(*Texture2D)
		gl.BindTextureUnit(uint32(u.textureUnitIndex), uint32(value.id))
		gl.ProgramUniform1i(uint32(u.progID), int32(u.location), int32(u.textureUnitIndex))
	case gl.SAMPLER_CUBE:
		// TODO: other shaders can mess with this texture index
		value := value.(*CubeMap)
		gl.BindTextureUnit(uint32(u.textureUnitIndex), uint32(value.id))
		gl.ProgramUniform1i(uint32(u.progID), int32(u.location), int32(u.textureUnitIndex))
	}
}

func (a *Attrib) SetSource(b *Buffer, el interface{}, i int) {
	if a == nil {
		return
	}

	t := reflect.TypeOf(el)

	var offset int
	var typeCand interface{}
	if t.Kind() == reflect.Struct {
		// set source to i-th field of the struct
		offset = int(t.Field(i).Offset)
		typeCand = reflect.ValueOf(el).Field(i).Interface()
	} else if t.Kind() == reflect.Array {
		// set source to i-th subelement
		offset = int(t.Elem().Size()) * i
		typeCand = reflect.ValueOf(el).Index(i).Interface()
	} else {
		panic("invalid element type")
	}

	var _type int
	switch typeCand.(type) {
	case float32, math.Vec2, math.Vec3, math.Vec4:
		_type = gl.FLOAT
	default:
		panic("unknown type " + reflect.TypeOf(typeCand).String())
	}

	stride := int(t.Size())

	a.prog.va.setAttribFormat(a, a.nComponents, _type, false)
	a.prog.va.setAttribSource(a, b, offset, stride)
}

func (p *ShaderProgram) SetAttribIndexBuffer(b *Buffer) {
	p.va.setIndexBuffer(b)
}

func (p *ShaderProgram) bind() {
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

var textureUnitsUsed int = 0

func (p *ShaderProgram) Uniform(name string) *Uniform {
	var u Uniform
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

	// TODO: allow more sampler types
	if u.glType == gl.SAMPLER_2D || u.glType == gl.SAMPLER_CUBE {
		u.textureUnitIndex = textureUnitsUsed
		textureUnitsUsed++ // TODO: make texture unit mapping more sophisticated
	}

	return &u
}
