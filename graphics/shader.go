package graphics

import (
	"errors"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/math"
	"io/ioutil"
	"strings"
	"fmt"
)

type shaderType int

const (
	vertexShader shaderType = iota
	fragmentShader
	geometryShader
)

type Program struct {
	id int

	vaid int
	indexBuffer *IndexBuffer

	Framebuffer *Framebuffer

	inputsByLocation map[int]*Input
	inputLocationsByName map[string]int

	uniformsByLocation map[int]*Uniform
	uniformLocationsByName map[string]int
}

type shader struct {
	id int
}

type Input struct {
	prog        *Program
	location    int
	nComponents int

	glType uint32
	normalize bool
	enabled bool
}

type Output struct {
	prog *Program
	location int
}

// TODO: store value, have Set() function and make "Uniform" an interface?
type Uniform struct {
	progID   int
	location int
	glType   int
	textureUnitIndex int
	name string
}

type vertexArray struct {
	id             int
	indexBuffer    *IndexBuffer
}

var currentProg *Program

func newShader(type_ shaderType, src string, defines ...string) (*shader, error) {
	var s shader

	var gltype uint32
	switch type_ {
	case vertexShader:
		gltype = gl.VERTEX_SHADER
	case fragmentShader:
		gltype = gl.FRAGMENT_SHADER
	case geometryShader:
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

func readShader(type_ shaderType, file string, defines ...string) (*shader, error) {
	src, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return newShader(type_, string(src), defines...)
}

func (s *shader) setSource(src string) {
	cSrc, free := gl.Strs(src)
	defer free()
	length := int32(len(src))
	gl.ShaderSource(uint32(s.id), 1, cSrc, &length)
}

func (s *shader) compiled() bool {
	var status int32
	gl.GetShaderiv(uint32(s.id), gl.COMPILE_STATUS, &status)
	return status == gl.TRUE
}

func (s *shader) log() string {
	var length int32
	gl.GetShaderiv(uint32(s.id), gl.INFO_LOG_LENGTH, &length)
	log := string(make([]byte, length+1))
	gl.GetShaderInfoLog(uint32(s.id), length+1, nil, gl.Str(log))
	log = log[:len(log)-1] // remove null terminator
	return log
}

func (s *shader) compile() error {
	gl.CompileShader(uint32(s.id))
	if s.compiled() {
		return nil
	} else {
		return errors.New(s.log())
	}
}

func newProgram(shaders ...*shader) *Program {
	var p Program
	p.id = int(gl.CreateProgram())

	for _, shader := range shaders {
		gl.AttachShader(uint32(p.id), uint32(shader.id))
		defer gl.DetachShader(uint32(p.id), uint32(shader.id))
	}

	err := p.link()
	if err != nil {
		panic(err)
	}

	var id uint32
	gl.CreateVertexArrays(1, &id)
	p.vaid = int(id)
	p.indexBuffer = nil

	p.Framebuffer = NewFramebuffer()

	p.inputsByLocation = make(map[int]*Input)
	p.inputLocationsByName = make(map[string]int)

	for i := 0; i < p.inputCount(); i++ {
		var size int32
		var type_ uint32
		var bytes [100]byte
		var namelength int32
		gl.GetActiveAttrib(uint32(p.id), uint32(i), 95, &namelength, &size, &type_, &bytes[0])
		name := string(bytes[:namelength])
		location := gl.GetAttribLocation(uint32(p.id), gl.Str(name+"\x00"))

		var in Input
		in.location = int(location)
		in.prog = &p

		switch type_ {
		case gl.FLOAT:
			in.nComponents = 1
		case gl.FLOAT_VEC2:
			in.nComponents = 2
		case gl.FLOAT_VEC3:
			in.nComponents = 3
		case gl.FLOAT_VEC4:
			in.nComponents = 4
		default:
			panic("unrecognized attribute GL type")
		}

		p.inputsByLocation[in.location] = &in
		p.inputLocationsByName[name] = in.location
	}

	p.uniformsByLocation = make(map[int]*Uniform)
	p.uniformLocationsByName = make(map[string]int)

	var c int32
	gl.GetProgramiv(uint32(p.id), gl.ACTIVE_UNIFORMS, &c)
	uniformCount := int(c)

	for i := 0; i < uniformCount; i++ {
		bytes := make([]uint8, 100)
		var size int32
		var type_ uint32
		var namelength int32
		gl.GetActiveUniform(uint32(p.id), uint32(i), 95, &namelength, &size, &type_, &bytes[0])
		name := string(bytes[:namelength])

		for j := 0; j < int(size); j++ {
			fullname := strings.Replace(name, "[0]", fmt.Sprintf("[%d]", j), 1)
			loc := gl.GetUniformLocation(uint32(p.id), gl.Str(fullname+"\x00"))

			var u Uniform
			u.location = int(loc)
			u.progID = p.id
			u.glType = int(type_)
			u.name = fullname

			// TODO: allow more sampler types
			if u.glType == gl.SAMPLER_2D || u.glType == gl.SAMPLER_CUBE {
				u.textureUnitIndex = textureUnitsUsed
				textureUnitsUsed++ // TODO: make texture unit mapping more sophisticated
			}

			p.uniformsByLocation[u.location] = &u
			p.uniformLocationsByName[fullname] = u.location
		}
	}

	return &p
}

func ReadProgram(vFile, fFile, gFile string, defines ...string) *Program {
	var shaders []*shader

	if vFile != "" {
		vShader, err := readShader(vertexShader, vFile, defines...)
		if err != nil {
			panic(err)
		}
		shaders = append(shaders, vShader)
	}
	if fFile != "" {
		fShader, err := readShader(fragmentShader, fFile, defines...)
		if err != nil {
			panic(err)
		}
		shaders = append(shaders, fShader)
	}

	if gFile != "" {
		gShader, err := readShader(geometryShader, gFile, defines...)
		if err != nil {
			panic(err)
		}
		shaders = append(shaders, gShader)
	}

	return newProgram(shaders...)
}

func (p *Program) inputCount() int {
	var count int32
	gl.GetProgramiv(uint32(p.id), gl.ACTIVE_ATTRIBUTES, &count)
	return int(count)
}

func (p *Program) uniformCount() int {
	var count int32
	gl.GetProgramiv(uint32(p.id), gl.ACTIVE_UNIFORMS, &count)
	return int(count)
}

func (p *Program) linked() bool {
	var status int32
	gl.GetProgramiv(uint32(p.id), gl.LINK_STATUS, &status)
	return status == gl.TRUE
}

func (p *Program) log() string {
	var length int32
	gl.GetProgramiv(uint32(p.id), gl.INFO_LOG_LENGTH, &length)
	log := string(make([]byte, length+1))
	gl.GetProgramInfoLog(uint32(p.id), length+1, nil, gl.Str(log))
	log = log[:len(log)-1] // remove null terminator
	return log
}

func (p *Program) link() error {
	gl.LinkProgram(uint32(p.id))
	if p.linked() {
		return nil
	}
	return errors.New(p.log())
}

func (p *Program) Render(vertexCount int, opts *RenderOptions) {
	if currentProg != p {
		p.bind()
	}
	opts.apply()

	if p.indexBuffer == nil {
		gl.DrawArrays(opts.Primitive.glPrimitive(), 0, int32(vertexCount))
	} else {
		gltype := p.indexBuffer.elementGlType()
		gl.DrawElements(opts.Primitive.glPrimitive(), int32(vertexCount), gltype, nil)
	}

	Stats.DrawCallCount++
	Stats.VertexCount += vertexCount
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

func (in *Input) setSourceRaw(b *buffer, offset, stride int, type_ int, normalize bool) {
	if in == nil {
		return
	}

	if !in.enabled || uint32(type_) != in.glType || normalize != in.normalize {
		gl.VertexArrayAttribFormat(uint32(in.prog.vaid), uint32(in.location), int32(in.nComponents), uint32(type_), normalize, 0)
		in.glType = uint32(type_)
		in.normalize = normalize
	}

	gl.VertexArrayVertexBuffer(uint32(in.prog.vaid), uint32(in.location), uint32(b.id), offset, int32(stride))

	if !in.enabled {
		gl.VertexArrayAttribBinding(uint32(in.prog.vaid), uint32(in.location), uint32(in.location))
		gl.EnableVertexArrayAttrib(uint32(in.prog.vaid), uint32(in.location))
		in.enabled = true
	}
}

func (in *Input) SetSourceVertex(b *VertexBuffer, i int) {
	offset := b.Offset(i)
	stride := b.ElementSize()
	in.setSourceRaw(&b.buffer, offset, stride, gl.FLOAT, false)
}

func (p *Program) SetIndices(b *IndexBuffer) {
	gl.VertexArrayElementBuffer(uint32(p.vaid), uint32(b.id))
	p.indexBuffer = b
}

func (p *Program) bind() {
	gl.UseProgram(uint32(p.id))
	gl.BindVertexArray(uint32(p.vaid))
	p.Framebuffer.bindDraw()
	gl.Viewport(0, 0, int32(p.Framebuffer.Width()), int32(p.Framebuffer.Height()))

	currentProg = p
}

func (p *Program) InputByLocation(location int) *Input {
	if location == -1 {
		return nil
	}
	return p.inputsByLocation[location]
}

func (p *Program) InputByName(name string) *Input {
	location, found := p.inputLocationsByName[name]
	if !found {
		return nil
	}
	return p.InputByLocation(location)
}

func (p *Program) OutputColorByLocation(location int) *Output {
	if location == -1 {
		return nil
	}
	var o Output
	o.location = location
	o.prog = p
	return &o
}

func (p *Program) OutputColorByName(name string) *Output {
	location := int(gl.GetFragDataLocation(uint32(p.id), gl.Str(name+"\x00")))
	if location == -1 {
		return nil
	}
	return p.OutputColorByLocation(location)
}

func (p *Program) OutputDepth() *Output {
	var o Output
	o.location = -2 // special depth output identifier
	o.prog = p
	return &o
}

func (o *Output) Set(att FramebufferAttachment) {
	o.prog.Framebuffer.Attach(att)
}

var textureUnitsUsed int = 0

func (p *Program) UniformNames() []string {
	var c int32
	gl.GetProgramiv(uint32(p.id), gl.ACTIVE_UNIFORMS, &c)
	uniformCount := int(c)

	names := make([]string, uniformCount)

	for i := 0; i < uniformCount; i++ {
		bytes := make([]uint8, 100)
		gl.GetActiveUniform(uint32(p.id), uint32(i), 95, nil, nil, nil, &bytes[0])
		name := string(bytes)
		names = append(names, name)
	}

	return names
}

func (p *Program) UniformByLocation(location int) *Uniform {
	u, found := p.uniformsByLocation[location]
	if !found {
		return nil
	}
	return u
}

func (p *Program) UniformByName(name string) *Uniform {
	location, found := p.uniformLocationsByName[name]
	if !found {
		return nil
	}
	return p.UniformByLocation(location)
}
