package graphics

import (
	"errors"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/math"
	"io/ioutil"
	"strings"
	"fmt"
)

type shader struct {
	id uint32
}

type Program struct {
	id            uint32
	vertexArrayID uint32
	indexBuffer   *IndexBuffer
	framebuffer   *framebuffer

	inputsByLocation map[uint32]*Input
	inputLocationsByName map[string]uint32

	uniformsByLocation map[uint32]*Uniform
	uniformLocationsByName map[string]uint32

	outputColorsByLocation map[uint32]*Output
	outputColorLocationsByName map[string]uint32
}

type Input struct {
	prog           *Program
	location       uint32
	name           string
	componentCount int
	glType         uint32
	normalize      bool
	enabled        bool
}

type Output struct {
	prog     *Program
	location uint32
	name     string
}

// TODO: store value, have Set() function and make "Uniform" an interface?
type Uniform struct {
	prog             *Program
	location         uint32
	name             string
	glType           uint32
	textureUnitIndex uint32
}

var currentProg *Program

func newShader(type_ uint32, src string, defines ...string) *shader {
	var sh shader
	sh.id = gl.CreateShader(type_)

	lines := strings.Split(src, "\n")
	src = lines[0] + "\n" // #version
	for _, define := range defines {
		src = src + "#define " + define + "\n"
	}
	for _, line := range lines[1:] {
		src = src + line + "\n"
	}
	sh.setSource(src)

	err := sh.compile()
	if err != nil {
		panic(err)
	}
	return &sh
}

func (sh *shader) setSource(src string) {
	cSrc, free := gl.Strs(src)
	defer free()
	length := int32(len(src))
	gl.ShaderSource(sh.id, 1, cSrc, &length)
}

func (sh *shader) compiled() bool {
	var status int32
	gl.GetShaderiv(sh.id, gl.COMPILE_STATUS, &status)
	return status == gl.TRUE
}

func (sh *shader) log() string {
	var length int32
	gl.GetShaderiv(sh.id, gl.INFO_LOG_LENGTH, &length)
	log := string(make([]byte, length+1))
	gl.GetShaderInfoLog(sh.id, length+1, nil, gl.Str(log))
	log = log[:len(log)-1] // remove null terminator
	return log
}

func (sh *shader) compile() error {
	gl.CompileShader(sh.id)
	if sh.compiled() {
		return nil
	} else {
		return errors.New(sh.log())
	}
}

var textureUnitsUsed int = 0

func buildProgram(shaders ...*shader) *Program {
	var prog Program
	prog.id = gl.CreateProgram()

	for _, sh := range shaders {
		gl.AttachShader(prog.id, sh.id)
		defer gl.DetachShader(prog.id, sh.id)
	}

	err := prog.link()
	if err != nil {
		panic(err)
	}

	gl.CreateVertexArrays(1, &prog.vertexArrayID)
	prog.indexBuffer = nil
	prog.framebuffer = newFramebuffer()

	prog.inputsByLocation = make(map[uint32]*Input)
	prog.inputLocationsByName = make(map[string]uint32)
	for i := 0; i < prog.inputCount(); i++ {
		in := prog.inputByIndex(i)
		prog.inputsByLocation[in.location] = in
		prog.inputLocationsByName[in.name] = in.location
	}

	prog.uniformsByLocation = make(map[uint32]*Uniform)
	prog.uniformLocationsByName = make(map[string]uint32)
	for i := 0; i < prog.uniformCount(); i++ {
		for j := 0; j < prog.uniformArraySizeByIndex(i); j++ {
			ufm := prog.uniformByIndex(i, j)
			prog.uniformsByLocation[ufm.location] = ufm
			prog.uniformLocationsByName[ufm.name] = ufm.location
		}
	}

	prog.outputColorsByLocation = make(map[uint32]*Output)
	prog.outputColorLocationsByName = make(map[string]uint32)
	for i := 0; i < prog.outputColorCount(); i++ {
		out := prog.outputColorByIndex(i)
		prog.outputColorsByLocation[out.location] = out
		prog.outputColorLocationsByName[out.name] = out.location
	}

	return &prog
}

func NewProgram(vSrc, fSrc, gSrc string, defines ...string) *Program {
	shaders := make([]*shader, 0, 3)

	if vSrc != "" {
		vShader := newShader(gl.VERTEX_SHADER, vSrc, defines...)
		shaders = append(shaders, vShader)
	}
	if fSrc != "" {
		fShader := newShader(gl.FRAGMENT_SHADER, fSrc, defines...)
		shaders = append(shaders, fShader)
	}
	if gSrc != "" {
		gShader := newShader(gl.GEOMETRY_SHADER, gSrc, defines...)
		shaders = append(shaders, gShader)
	}

	return buildProgram(shaders...)
}

func ReadProgram(vFile, fFile, gFile string, defines ...string) *Program {
	var vSrc, fSrc, gSrc string

	if vFile == "" {
		vSrc = ""
	} else {
		src, err := ioutil.ReadFile(vFile)
		if err != nil {
			panic(err)
		}
		vSrc = string(src)
	}
	if fFile == "" {
		fSrc = ""
	} else {
		src, err := ioutil.ReadFile(fFile)
		if err != nil {
			panic(err)
		}
		fSrc = string(src)
	}
	if gFile == "" {
		gSrc = ""
	} else {
		src, err := ioutil.ReadFile(gFile)
		if err != nil {
			panic(err)
		}
		gSrc = string(src)
	}

	return NewProgram(vSrc, fSrc, gSrc, defines...)
}

func (prog *Program) linked() bool {
	var status int32
	gl.GetProgramiv(prog.id, gl.LINK_STATUS, &status)
	return status == gl.TRUE
}

func (prog *Program) log() string {
	var length int32
	gl.GetProgramiv(prog.id, gl.INFO_LOG_LENGTH, &length)
	log := string(make([]byte, length+1))
	gl.GetProgramInfoLog(prog.id, length+1, nil, gl.Str(log))
	log = log[:len(log)-1] // remove null terminator
	return log
}

func (prog *Program) link() error {
	gl.LinkProgram(prog.id)
	if prog.linked() {
		return nil
	}
	return errors.New(prog.log())
}

func (prog *Program) Render(vertexCount int, opts *RenderOptions) {
	if currentProg != prog {
		prog.bind()
	}
	opts.apply()

	if prog.indexBuffer == nil {
		gl.DrawArrays(opts.Primitive.glPrimitive(), 0, int32(vertexCount))
	} else {
		gltype := prog.indexBuffer.elementGlType()
		gl.DrawElements(opts.Primitive.glPrimitive(), int32(vertexCount), gltype, nil)
	}

	Stats.DrawCallCount++
	Stats.VertexCount += vertexCount
}

func (prog *Program) bind() {
	gl.UseProgram(prog.id)
	gl.BindVertexArray(prog.vertexArrayID)
	prog.framebuffer.bindDraw()
	gl.Viewport(0, 0, int32(prog.framebuffer.Width()), int32(prog.framebuffer.Height()))

	currentProg = prog
}

func (prog *Program) SetIndices(b *IndexBuffer) {
	gl.VertexArrayElementBuffer(prog.vertexArrayID, b.id)
	prog.indexBuffer = b
}

func (prog *Program) inputCount() int {
	var count int32
	gl.GetProgramiv(prog.id, gl.ACTIVE_ATTRIBUTES, &count)
	return int(count)
}

func (prog *Program) uniformCount() int {
	var count int32
	gl.GetProgramiv(prog.id, gl.ACTIVE_UNIFORMS, &count)
	return int(count)
}

func (prog *Program) outputColorCount() int {
	var count int32
	gl.GetProgramInterfaceiv(prog.id, gl.PROGRAM_OUTPUT, gl.ACTIVE_RESOURCES, &count)
	return int(count)
}

func (prog *Program) inputByIndex(i int) *Input {
	var size int32
	var type_ uint32
	var bytes [100]byte
	var namelength int32
	gl.GetActiveAttrib(prog.id, uint32(i), 95, &namelength, &size, &type_, &bytes[0])

	var in Input
	in.prog = prog
	in.name = string(bytes[:namelength])
	in.location = uint32(gl.GetAttribLocation(prog.id, gl.Str(in.name+"\x00")))

	switch type_ {
	case gl.FLOAT:
		in.componentCount = 1
	case gl.FLOAT_VEC2:
		in.componentCount = 2
	case gl.FLOAT_VEC3:
		in.componentCount = 3
	case gl.FLOAT_VEC4:
		in.componentCount = 4
	default:
		panic("unrecognized attribute GL type")
	}

	return &in
}

func (prog *Program) InputByLocation(location int) *Input {
	if location == -1 {
		return nil
	}
	return prog.inputsByLocation[uint32(location)]
}

func (prog *Program) InputByName(name string) *Input {
	location, found := prog.inputLocationsByName[name]
	if !found {
		return nil
	}
	return prog.InputByLocation(int(location))
}

func (prog *Program) uniformArraySizeByIndex(i int) int {
	var size int32
	gl.GetActiveUniform(prog.id, uint32(i), 0, nil, &size, nil, nil)
	return int(size)
}

func (prog *Program) uniformByIndex(i, j int) *Uniform {
	var size int32
	var type_ uint32
	var bytes [100]byte
	var namelength int32
	gl.GetActiveUniform(prog.id, uint32(i), 95, &namelength, &size, &type_, &bytes[0])
	basename := string(bytes[:namelength])

	var ufm Uniform
	ufm.prog = prog
	ufm.name = strings.Replace(basename, "[0]", fmt.Sprintf("[%d]", j), 1)
	ufm.location = uint32(gl.GetUniformLocation(prog.id, gl.Str(ufm.name+"\x00")))
	ufm.glType = type_

	// TODO: allow more sampler types
	if ufm.glType == gl.SAMPLER_2D || ufm.glType == gl.SAMPLER_CUBE {
		ufm.textureUnitIndex = uint32(textureUnitsUsed)
		textureUnitsUsed++ // TODO: make texture unit mapping more sophisticated
	}

	return &ufm
}

func (prog *Program) UniformByLocation(location int) *Uniform {
	ufm, found := prog.uniformsByLocation[uint32(location)]
	if !found {
		return nil
	}
	return ufm
}

func (prog *Program) UniformByName(name string) *Uniform {
	location, found := prog.uniformLocationsByName[name]
	if !found {
		return nil
	}
	return prog.UniformByLocation(int(location))
}

func (prog *Program) outputColorByIndex(i int) *Output {
	var out Output
	out.prog = prog

	var location int32
	param := uint32(gl.LOCATION)
	gl.GetProgramResourceiv(prog.id, gl.PROGRAM_OUTPUT, uint32(i), 1, &param, 1, nil, &location)
	out.location = uint32(location)

	var bytes [100]byte
	var namelength int32
	gl.GetProgramResourceName(prog.id, gl.PROGRAM_OUTPUT, uint32(i), 95, &namelength, &bytes[0])
	out.name = string(bytes[:namelength])

	return &out
}

func (prog *Program) OutputColorByLocation(location int) *Output {
	if location == -1 {
		return nil
	}
	return prog.outputColorsByLocation[uint32(location)]
}

func (prog *Program) OutputColorByName(name string) *Output {
	location, found := prog.outputColorLocationsByName[name]
	if !found {
		return nil
	}
	return prog.OutputColorByLocation(int(location))
}

func (prog *Program) OutputDepth() *Output {
	// TODO: make different from color outputs
	var out Output
	out.location = 0
	out.prog = prog
	return &out
}

func (in *Input) setSourceRaw(b *buffer, offset, stride int, type_ uint32, normalize bool) {
	if in == nil {
		return
	}

	if !in.enabled || type_ != in.glType || normalize != in.normalize {
		gl.VertexArrayAttribFormat(in.prog.vertexArrayID, in.location, int32(in.componentCount), type_, normalize, 0)
		in.glType = type_
		in.normalize = normalize
	}

	gl.VertexArrayVertexBuffer(in.prog.vertexArrayID, in.location, b.id, offset, int32(stride))

	if !in.enabled {
		gl.VertexArrayAttribBinding(in.prog.vertexArrayID, in.location, in.location)
		gl.EnableVertexArrayAttrib(in.prog.vertexArrayID, in.location)
		in.enabled = true
	}
}

func (in *Input) SetSourceVertex(b *VertexBuffer, i int) {
	offset := b.Offset(i)
	stride := b.ElementSize()
	in.setSourceRaw(&b.buffer, offset, stride, gl.FLOAT, false)
}

func (ufm *Uniform) Set(value interface{}) {
	if ufm == nil {
		return
	}

	switch ufm.glType {
	case gl.BOOL:
		value := value.(int32)
		// TODO: unnecessary?
		if value != 0{
			value = 1
		} else {
			value = 0
		}
		gl.ProgramUniform1i(ufm.prog.id, int32(ufm.location), int32(value))
	case gl.INT:
		value := value.(int)
		gl.ProgramUniform1i(ufm.prog.id, int32(ufm.location), int32(value))
	case gl.FLOAT:
		value := value.(float32)
		gl.ProgramUniform1f(ufm.prog.id, int32(ufm.location), value)
	case gl.FLOAT_VEC2:
		value := value.(math.Vec2)
		gl.ProgramUniform2fv(ufm.prog.id, int32(ufm.location), 1, &value[0])
	case gl.FLOAT_VEC3:
		value := value.(math.Vec3)
		gl.ProgramUniform3fv(ufm.prog.id, int32(ufm.location), 1, &value[0])
	case gl.FLOAT_VEC4:
		value := value.(math.Vec4)
		gl.ProgramUniform4fv(ufm.prog.id, int32(ufm.location), 1, &value[0])
	case gl.FLOAT_MAT4:
		value := value.(*math.Mat4)
		gl.ProgramUniformMatrix4fv(ufm.prog.id, int32(ufm.location), 1, true, &value[0])
	case gl.SAMPLER_2D:
		// TODO: other shaders can mess with this texture index
		value := value.(*Texture2D)
		gl.BindTextureUnit(ufm.textureUnitIndex, value.id)
		gl.ProgramUniform1i(ufm.prog.id, int32(ufm.location), int32(ufm.textureUnitIndex))
	case gl.SAMPLER_CUBE:
		// TODO: other shaders can mess with this texture index
		value := value.(*CubeMap)
		gl.BindTextureUnit(ufm.textureUnitIndex, value.id)
		gl.ProgramUniform1i(ufm.prog.id, int32(ufm.location), int32(ufm.textureUnitIndex))
	default:
		panic("invalid uniform")
	}
}

func (out *Output) Set(target renderTarget) {
	out.prog.framebuffer.attach(target)
}
