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
)

type Shader uint32

func NewShaderFromString(typ uint32, src string) (Shader, error) {
	s := Shader(gl.CreateShader(typ))
	s.setSource(src)
	err := s.compile()
	return s, err
}

func NewShaderFromFile(typ uint32, filename string) (Shader, error) {
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
	vShader, err := NewShaderFromFile(gl.VERTEX_SHADER, vShaderFilename)
	if err != nil {
		return Program(0), err
	}
	fShader, err := NewShaderFromFile(gl.FRAGMENT_SHADER, fShaderFilename)
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

func (p Program) attrib(name string) (*Attrib, error) {
	var a Attrib
	loc := gl.GetAttribLocation(uint32(p), gl.Str(name + "\x00"))
	if loc == -1 {
		return nil, errors.New(fmt.Sprint(name, " attribute location -1"))
	} else {
		a.id = uint32(loc)
		return &a, nil
	}
}

func (p Program) uniform(name string) (*Uniform, error) {
	var u Uniform
	loc := gl.GetUniformLocation(uint32(p), gl.Str(name + "\x00"))
	if loc == -1 {
		return nil, errors.New(fmt.Sprint(name, " uniform location -1"))
	} else {
		u.id = uint32(loc)
		gl.GetActiveUniform(uint32(p), u.id, 0, nil, nil, &u.typ, nil)
		return &u, nil
	}
}



type Buffer struct {
	typ uint32
	id uint32
	size int
}

// TODO: type not needed
func NewBuffer(typ uint32) *Buffer {
	var b Buffer
	b.typ = typ
	gl.CreateBuffers(1, &b.id)
	b.size = 0
	return &b
}

func (b *Buffer) allocate(size int) {
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
				b.allocate(size)
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

func (t *Texture2D) bind() {
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

func (a *Attrib) SetFormat(dim int, typ int, normalize bool) {
	gl.VertexAttribFormat(a.id, int32(dim), uint32(typ), normalize, 0)
}

func (a *Attrib) SetSource(b *Buffer, offset, stride int) {
	gl.VertexAttribBinding(a.id, a.id)
	gl.BindVertexBuffer(a.id, b.id, offset, int32(stride))
	gl.EnableVertexAttribArray(a.id)
}



type Uniform struct {
	id uint32
	typ uint32
}

func (u *Uniform) Set(val interface{}) {
	switch u.typ {
	case gl.FLOAT_MAT4:
		switch val.(type) {
		case *Mat4:
			val := val.(*Mat4)
			gl.UniformMatrix4fv(int32(u.id), 1, true, &val[0])
		default:
			panic("tried to set uniform from unknown type")
		}
	default:
		panic("tried to set uniform of unknown type")
	}
}
