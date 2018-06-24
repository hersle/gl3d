package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"errors"
	"io/ioutil"
	"reflect"
)

type ShaderType uint32
type Shader uint32

const (
	VertexShader ShaderType = gl.VERTEX_SHADER
	FragmentShader ShaderType = gl.FRAGMENT_SHADER
)

func NewShaderFromString(typ ShaderType, src string) (Shader, error) {
	s := Shader(gl.CreateShader(uint32(typ)))
	s.setSource(src)
	err := s.compile()
	return s, err
}

func NewShaderFromFile(typ ShaderType, filename string) (Shader, error) {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return Shader(0), err
	}
	return NewShaderFromString(typ, string(src))
}

func (s Shader) setSource(src string) {
	cSrc, free := gl.Strs(src)
	defer free()
	length := int32(len(src))
	gl.ShaderSource(uint32(s), 1, cSrc, &length)
}

func (s Shader) compiled() bool {
	var status int32
	gl.GetShaderiv(uint32(s), gl.COMPILE_STATUS, &status)
	return status == gl.TRUE
}

func (s Shader) log() string {
	var length int32
	gl.GetShaderiv(uint32(s), gl.INFO_LOG_LENGTH, &length)
	var log string = string(make([]byte, length + 1))
	gl.GetShaderInfoLog(uint32(s), length + 1, nil, gl.Str(log))
	log = log[:len(log)-1] // remove null terminator
	return log
}

func (s Shader) compile() error {
	gl.CompileShader(uint32(s))
	if s.compiled() {
		return nil
	} else {
		return errors.New(s.log())
	}
}



type Program uint32

func NewProgramFromShaders(vShader, fShader Shader) (Program, error) {
	p := Program(gl.CreateProgram())
	gl.AttachShader(uint32(p), uint32(vShader))
	gl.AttachShader(uint32(p), uint32(fShader))
	err := p.link()
	return p, err
}

func NewProgramFromFiles(vShaderFilename, fShaderFilename string) (Program, error) {
	vShader, err := NewShaderFromFile(VertexShader, vShaderFilename)
	if err != nil {
		return Program(0), err
	}
	fShader, err := NewShaderFromFile(FragmentShader, fShaderFilename)
	if err != nil {
		return Program(0), err
	}
	return NewProgramFromShaders(vShader, fShader)
}

func (p Program) linked() bool {
	var status int32
	gl.GetProgramiv(uint32(p), gl.LINK_STATUS, &status)
	return status == gl.TRUE
}

func (p Program) log() string {
	var length int32
	gl.GetProgramiv(uint32(p), gl.INFO_LOG_LENGTH, &length)
	var log string = string(make([]byte, length + 1))
	gl.GetProgramInfoLog(uint32(p), length + 1, nil, gl.Str(log))
	log = log[:len(log)-1] // remove null terminator
	return log
}

func (p Program) link() error {
	gl.LinkProgram(uint32(p))
	if p.linked() {
		return nil
	} else {
		return errors.New(p.log())
	}
}

func (p Program) use() {
	gl.UseProgram(uint32(p))
}

func (p Program) attribLocation(name string) (uint32, error) {
	loc := gl.GetAttribLocation(uint32(p), gl.Str(name + "\x00"))
	err := error(nil)
	if loc == -1 {
		err = errors.New("attribute location -1")
	}
	return uint32(loc), err

}

func (p Program) uniformLocation(name string) (uint32, error) {
	loc := gl.GetUniformLocation(uint32(p), gl.Str(name + "\x00"))
	err := error(nil)
	if loc == -1 {
		err = errors.New("uniform location -1")
	}
	return uint32(loc), err
}



type Buffer struct {
	typ uint32
	id uint32
	size int
}

func NewBuffer(typ uint32) *Buffer {
	var b Buffer
	b.typ = typ
	gl.GenBuffers(1, &b.id)
	b.size = 0
	return &b
}

func (b *Buffer) bind() {
	// TODO: track currently bound buffer to avoid unnecessarily binding buffers
	gl.BindBuffer(b.typ, b.id)
}

func (b *Buffer) allocate(size int) {
	b.bind()
	b.size = size
	gl.BufferData(b.typ, b.size, nil, gl.STREAM_DRAW)
}

func (b *Buffer) SetData(data interface{}, byteOffset int) {
	// assumes all entries in data are of the same type

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Slice {
		if val.Len() > 0 {
			b.bind()
			size := val.Len() * int(val.Type().Elem().Size())
			if size > b.size {
				b.allocate(size)
			}
			gl.BufferSubData(b.typ, byteOffset, size, gl.Ptr(data))
		}
	} else {
		panic("not a slice")
	}
}

