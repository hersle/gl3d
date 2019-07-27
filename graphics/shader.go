package graphics

import (
	"errors"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/math"
	"io/ioutil"
	"strings"
	"strconv"
)

type ShaderType int

const (
	VertexShader ShaderType = iota
	FragmentShader
	GeometryShader
)

type ShaderProgram struct {
	id int

	vaid int
	indexBuffer *IndexBuffer

	Framebuffer *Framebuffer
}

type Shader struct {
	id int
}

type Attrib struct {
	prog        *ShaderProgram
	id          int
	nComponents int

	glType uint32
	normalize bool
	enabled bool
}

type Output struct {
	prog *ShaderProgram
	id int
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
	indexBuffer    *IndexBuffer
}

func NewShader(type_ ShaderType, src string, defines ...string) (*Shader, error) {
	var s Shader

	var gltype uint32
	switch type_ {
	case VertexShader:
		gltype = gl.VERTEX_SHADER
	case FragmentShader:
		gltype = gl.FRAGMENT_SHADER
	case GeometryShader:
		gltype = gl.GEOMETRY_SHADER
	default:
		panic("unknown shader type")
	}

	lines := strings.Split(src, "\n")
	src = lines[0] + "\n" // #version
	for _, define := range defines {
		src = src + "#define " + define + "\n"
	}
	for _, line := range lines[1:] {
		src = src + line + "\n"
	}

	s.id = int(gl.CreateShader(gltype))
	s.setSource(src)
	err := s.compile()
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func ReadShader(type_ ShaderType, file string, defines ...string) (*Shader, error) {
	src, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return NewShader(type_, string(src), defines...)
}

func (s *Shader) setSource(src string) {
	cSrc, free := gl.Strs(src)
	defer free()
	length := int32(len(src))
	gl.ShaderSource(uint32(s.id), 1, cSrc, &length)
}

func (s *Shader) compiled() bool {
	var status int32
	gl.GetShaderiv(uint32(s.id), gl.COMPILE_STATUS, &status)
	return status == gl.TRUE
}

func (s *Shader) log() string {
	var length int32
	gl.GetShaderiv(uint32(s.id), gl.INFO_LOG_LENGTH, &length)
	log := string(make([]byte, length+1))
	gl.GetShaderInfoLog(uint32(s.id), length+1, nil, gl.Str(log))
	log = log[:len(log)-1] // remove null terminator
	return log
}

func (s *Shader) compile() error {
	gl.CompileShader(uint32(s.id))
	if s.compiled() {
		return nil
	} else {
		return errors.New(s.log())
	}
}

func NewShaderProgram(shaders ...*Shader) (*ShaderProgram, error) {
	var p ShaderProgram
	p.id = int(gl.CreateProgram())

	for _, shader := range shaders {
		gl.AttachShader(uint32(p.id), uint32(shader.id))
		defer gl.DetachShader(uint32(p.id), uint32(shader.id))
	}

	err := p.link()
	if err != nil {
		return nil, err
	}

	var id uint32
	gl.CreateVertexArrays(1, &id)
	p.vaid = int(id)
	p.indexBuffer = nil

	p.Framebuffer = NewFramebuffer()

	return &p, err
}

func ReadShaderProgram(vFile, fFile, gFile string, defines ...string) (*ShaderProgram, error) {
	var shaders []*Shader

	if vFile != "" {
		vShader, err := ReadShader(VertexShader, vFile, defines...)
		if err != nil {
			return nil, err
		}
		shaders = append(shaders, vShader)
	}
	if fFile != "" {
		fShader, err := ReadShader(FragmentShader, fFile, defines...)
		if err != nil {
			return nil, err
		}
		shaders = append(shaders, fShader)
	}

	if gFile != "" {
		gShader, err := ReadShader(GeometryShader, gFile, defines...)
		if err != nil {
			return nil, err
		}
		shaders = append(shaders, gShader)
	}

	return NewShaderProgram(shaders...)
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
	default:
		panic("invalid uniform")
	}
}

func (a *Attrib) SetSourceRaw(b *Buffer, offset, stride int, type_ int, normalize bool) {
	if a == nil {
		return
	}

	if !a.enabled || uint32(type_) != a.glType || normalize != a.normalize {
		gl.VertexArrayAttribFormat(uint32(a.prog.vaid), uint32(a.id), int32(a.nComponents), uint32(type_), normalize, 0)
		a.glType = uint32(type_)
		a.normalize = normalize
	}

	gl.VertexArrayVertexBuffer(uint32(a.prog.vaid), uint32(a.id), uint32(b.id), offset, int32(stride))

	if !a.enabled {
		gl.VertexArrayAttribBinding(uint32(a.prog.vaid), uint32(a.id), uint32(a.id))
		gl.EnableVertexArrayAttrib(uint32(a.prog.vaid), uint32(a.id))
		a.enabled = true
	}
}

func (a *Attrib) SetSourceVertex(b *VertexBuffer, i int) {
	offset := b.Offset(i)
	stride := b.ElementSize()
	a.SetSourceRaw(&b.Buffer, offset, stride, gl.FLOAT, false)
}

func (p *ShaderProgram) SetAttribIndexBuffer(b *IndexBuffer) {
	gl.VertexArrayElementBuffer(uint32(p.vaid), uint32(b.id))
	p.indexBuffer = b
}

func (p *ShaderProgram) bind() {
	gl.UseProgram(uint32(p.id))
	gl.BindVertexArray(uint32(p.vaid))
	p.Framebuffer.bindDraw()
	gl.Viewport(0, 0, int32(p.Framebuffer.Width()), int32(p.Framebuffer.Height()))
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

func (p *ShaderProgram) OutputColor(name string) *Output {
	var o Output
	loc := gl.GetFragDataLocation(uint32(p.id), gl.Str(name+"\x00"))
	if loc == -1 {
		return nil
	}
	o.id = int(loc)
	o.prog = p
	return &o
}

func (p *ShaderProgram) OutputDepth() *Output {
	var o Output
	o.id = -2 // special depth output identifier
	o.prog = p
	return &o
}

func (o *Output) Set(att FramebufferAttachment) {
	o.prog.Framebuffer.Attach(att)
}

var textureUnitsUsed int = 0

func (p *ShaderProgram) UniformNames() []string {
	var c int32
	gl.GetProgramiv(uint32(p.id), gl.ACTIVE_UNIFORMS, &c)
	uniformCount := int(c)

	names := make([]string, uniformCount)

	for i := 0; i < uniformCount; i++ {
		bytes := make([]uint8, 100)
		gl.GetActiveUniform(uint32(p.id), uint32(i), 95, nil, nil, nil, &bytes[0])
		name := string(bytes)
		names = append(names, name)
		println(name)
	}

	return names
}

func (p *ShaderProgram) Uniform(name string) *Uniform {
	// handle variable names referencing array elements
	arrayindex := 0
	i1 := strings.LastIndex(name, "[")
	i2 := strings.LastIndex(name, "]")
	if i1 != -1 && i2 != -1 {
		// [0]-name must be used for querying correct type, etc.
		arrayindex, _ = strconv.Atoi(name[i1+1:i2])
		name = name[:i1] + "[0]" + name[i2+1:]
	}

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

	u.location += arrayindex // location increments by one in uniform arrays

	// TODO: allow more sampler types
	if u.glType == gl.SAMPLER_2D || u.glType == gl.SAMPLER_CUBE {
		u.textureUnitIndex = textureUnitsUsed
		textureUnitsUsed++ // TODO: make texture unit mapping more sophisticated
	}

	return &u
}
