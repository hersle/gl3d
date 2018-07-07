package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"errors"
	"io/ioutil"
	"reflect"
	"fmt"
	"os"
	"image"
	"image/draw"
	_ "image/png"
)

// TODO: add opengl state object tracker

type Shader struct {
	id uint32
}

func NewShader(typ uint32, src string) (*Shader, error) {
	var s Shader
	s.id = gl.CreateShader(typ)
	s.SetSource(src)
	err := s.Compile()
	return &s, err
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
	log := string(make([]byte, length + 1))
	gl.GetShaderInfoLog(s.id, length + 1, nil, gl.Str(log))
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



type Program struct {
	id uint32
}

func NewProgram(vShader, fShader *Shader) (*Program, error) {
	var p Program
	p.id = gl.CreateProgram()
	gl.AttachShader(p.id, vShader.id)
	gl.AttachShader(p.id, fShader.id)
	err := p.Link()
	gl.DetachShader(p.id, vShader.id)
	gl.DetachShader(p.id, fShader.id)
	return &p, err
}

func ReadProgram(vShaderFilename, fShaderFilename string) (*Program, error) {
	vShader, err := ReadShader(gl.VERTEX_SHADER, vShaderFilename)
	if err != nil {
		return nil, err
	}
	fShader, err := ReadShader(gl.FRAGMENT_SHADER, fShaderFilename)
	if err != nil {
		return nil, err
	}
	return NewProgram(vShader, fShader)
}

func (p *Program) Linked() bool {
	var status int32
	gl.GetProgramiv(p.id, gl.LINK_STATUS, &status)
	return status == gl.TRUE
}

func (p *Program) Log() string {
	var length int32
	gl.GetProgramiv(p.id, gl.INFO_LOG_LENGTH, &length)
	log := string(make([]byte, length + 1))
	gl.GetProgramInfoLog(p.id, length + 1, nil, gl.Str(log))
	log = log[:len(log)-1] // remove null terminator
	return log
}

func (p *Program) Link() error {
	gl.LinkProgram(p.id)
	if p.Linked() {
		return nil
	} else {
		return errors.New(p.Log())
	}
}

// TODO: move to opengl state object
func (p *Program) Use() {
	gl.UseProgram(p.id)
}

func (p *Program) Attrib(name string) (*Attrib, error) {
	var a Attrib
	loc := gl.GetAttribLocation(p.id, gl.Str(name + "\x00"))
	if loc == -1 {
		return nil, errors.New(fmt.Sprint(name, " attribute location -1"))
	} else {
		a.id = uint32(loc)
		return &a, nil
	}
}

func (p *Program) Uniform(name string) (*Uniform, error) {
	var u Uniform
	loc := gl.GetUniformLocation(p.id, gl.Str(name + "\x00"))
	if loc == -1 {
		return nil, errors.New(fmt.Sprint(name, " uniform location -1"))
	} else {
		u.id = uint32(loc)
		gl.GetActiveUniform(p.id, u.id, 0, nil, nil, &u.typ, nil)
		return &u, nil
	}
}

func (p *Program) SetUniform(u *Uniform, val interface{}) {
	// TODO: set more types
	switch u.typ {
	case gl.FLOAT:
		switch val.(type) {
		case float32:
			val := val.(float32)
			gl.ProgramUniform1f(p.id, int32(u.id), val)
			return
		}
	case gl.FLOAT_VEC3:
		switch val.(type) {
		case Vec3:
			val := val.(Vec3)
			gl.ProgramUniform3fv(p.id, int32(u.id), 1, &val[0])
			return
		}
	case gl.FLOAT_MAT4:
		switch val.(type) {
		case *Mat4:
			val := val.(*Mat4)
			gl.ProgramUniformMatrix4fv(p.id, int32(u.id), 1, true, &val[0])
			return
		}
	default:
		panic("tried to set uniform of unknown type")
	}
	panic("tried to set uniform from unknown type")
}



type Buffer struct {
	id uint32
	size int
}

func NewBuffer() *Buffer {
	var b Buffer
	gl.CreateBuffers(1, &b.id)
	b.size = 0
	return &b
}

func (b *Buffer) Allocate(size int) {
	b.size = size
	gl.NamedBufferData(b.id, int32(b.size), nil, gl.STREAM_DRAW)
}

func (b *Buffer) SetData(data interface{}, byteOffset int) {
	// assumes all entries in data are of the same type
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Slice {
		if val.Len() > 0 {
			size := val.Len() * int(val.Type().Elem().Size())
			if size > b.size {
				b.Allocate(size)
			}
			gl.NamedBufferSubData(b.id, byteOffset, int32(size), gl.Ptr(data))
		}
	} else {
		panic("not a slice")
	}
}



type Texture2D struct {
	id uint32
}

func NewTexture2D() *Texture2D {
	var t Texture2D
	gl.CreateTextures(gl.TEXTURE_2D, 1, &t.id)
	gl.TextureParameteri(t.id, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(t.id, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_T, gl.REPEAT)
	return &t
}

// TODO: move to opengl state object
func (t *Texture2D) Bind() {
	gl.BindTexture(gl.TEXTURE_2D, t.id)
}

func (t *Texture2D) SetImage(img image.Image) {
	switch img.(type) {
	case *image.RGBA:
		img := img.(*image.RGBA)
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
		w, h := int32(img.Bounds().Size().X), int32(img.Bounds().Size().Y)
		gl.TextureStorage2D(t.id, 1, gl.RGBA8, w, h)
		gl.TextureSubImage2D(t.id, 0, 0, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
	default:
		imgRGBA := image.NewRGBA(img.Bounds())
		draw.Draw(imgRGBA, imgRGBA.Bounds(), img, img.Bounds().Min, draw.Over)
		t.SetImage(imgRGBA)
	}
}

func (t *Texture2D) ReadImage(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	t.SetImage(img)
	return nil
}



type Attrib struct {
	id uint32
}



type Uniform struct {
	id uint32
	typ uint32
}



type VertexArray struct {
	id uint32
}

func NewVertexArray() *VertexArray {
	var va VertexArray
	gl.CreateVertexArrays(1, &va.id)
	return &va
}

// TODO: move to opengl state object
func (va *VertexArray) Bind() {
	gl.BindVertexArray(va.id)
}

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
}
