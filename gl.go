package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"errors"
	"io/ioutil"
	"reflect"
	"fmt"
	"image"
	"image/draw"
	"unsafe"
)

type RenderState struct {
	vaBound *VertexArray
	progBound *ShaderProgram
	tex2dBound *Texture2D
	drawFramebufferBound *Framebuffer
	readFramebufferBound *Framebuffer
}

type ShaderProgram struct {
	id uint32
}

type Shader struct {
	id uint32
}

type Buffer struct {
	id uint32
	size int
}

type Texture2D struct {
	id uint32
}

type Attrib struct {
	id uint32
}

type TextureUnit struct {
	id int32
}

type Uniform struct {
	pID uint32
	id uint32
	typ uint32
}

type VertexArray struct {
	id uint32
}

type Framebuffer struct {
	id uint32
}

var gls *RenderState = &RenderState{}

var defaultFramebuffer *Framebuffer = &Framebuffer{0}

func (st *RenderState) SetVertexArray(va *VertexArray) {
	if st.vaBound == nil || st.vaBound.id != va.id {
		gl.BindVertexArray(va.id)
		st.vaBound = va
	}
}

func (st *RenderState) SetShaderProgram(prog *ShaderProgram) {
	if st.progBound == nil || st.progBound.id != prog.id {
		gl.UseProgram(prog.id)
		st.progBound = prog
	}
}

func (st *RenderState) SetDrawFramebuffer(f *Framebuffer) {
	if st.drawFramebufferBound == nil || st.drawFramebufferBound.id != f.id {
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, f.id)
		st.drawFramebufferBound = f
	}
}

func (st *RenderState) SetReadFramebuffer(f *Framebuffer) {
	if st.readFramebufferBound == nil || st.readFramebufferBound.id != f.id {
		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, f.id)
		st.readFramebufferBound = f
	}
}

// TODO: draw methods

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

func NewShaderProgram(vShader, fShader *Shader) (*ShaderProgram, error) {
	var p ShaderProgram
	p.id = gl.CreateProgram()
	gl.AttachShader(p.id, vShader.id)
	gl.AttachShader(p.id, fShader.id)
	err := p.Link()
	if err != nil {
		return nil, err
	}
	gl.DetachShader(p.id, vShader.id)
	gl.DetachShader(p.id, fShader.id)
	return &p, err
}

func ReadShaderProgram(vShaderFilename, fShaderFilename string) (*ShaderProgram, error) {
	vShader, err := ReadShader(gl.VERTEX_SHADER, vShaderFilename)
	if err != nil {
		return nil, err
	}
	fShader, err := ReadShader(gl.FRAGMENT_SHADER, fShaderFilename)
	if err != nil {
		return nil, err
	}
	return NewShaderProgram(vShader, fShader)
}

func (p *ShaderProgram) Linked() bool {
	var status int32
	gl.GetProgramiv(p.id, gl.LINK_STATUS, &status)
	return status == gl.TRUE
}

func (p *ShaderProgram) Log() string {
	var length int32
	gl.GetProgramiv(p.id, gl.INFO_LOG_LENGTH, &length)
	log := string(make([]byte, length + 1))
	gl.GetProgramInfoLog(p.id, length + 1, nil, gl.Str(log))
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

func (p *ShaderProgram) Attrib(name string) (*Attrib, error) {
	var a Attrib
	loc := gl.GetAttribLocation(p.id, gl.Str(name + "\x00"))
	if loc == -1 {
		return nil, errors.New(fmt.Sprint(name, " attribute location -1"))
	}
	a.id = uint32(loc)
	return &a, nil
}

func (p *ShaderProgram) Uniform(name string) (*Uniform, error) {
	var u Uniform
	loc := gl.GetUniformLocation(p.id, gl.Str(name + "\x00"))
	if loc == -1 {
		return nil, errors.New(fmt.Sprint(name, " uniform location -1"))
	}
	u.id = uint32(loc)
	gl.GetActiveUniform(p.id, u.id, 0, nil, nil, &u.typ, nil)
	u.pID = p.id
	return &u, nil
}

func (u *Uniform) SetInteger(i int) {
	gl.ProgramUniform1i(u.pID, int32(u.id), int32(i))
}

func (u *Uniform) SetFloat(f float32) {
	gl.ProgramUniform1f(u.pID, int32(u.id), float32(f))
}

func (u *Uniform) SetVector2(v Vec2) {
	gl.ProgramUniform2fv(u.pID, int32(u.id), 1, &v[0])
}

func (u *Uniform) SetVector3(v Vec3) {
	gl.ProgramUniform3fv(u.pID, int32(u.id), 1, &v[0])
}

func (u *Uniform) SetVector4(v Vec4) {
	gl.ProgramUniform4fv(u.pID, int32(u.id), 1, &v[0])
}

func (u *Uniform) SetMatrix4(m *Mat4) {
	gl.ProgramUniformMatrix4fv(u.pID, int32(u.id), 1, true, &m[0])
}

func valueGLType(val interface{}) uint32 {
	switch val.(type) {
	case int: // TODO: int32?
		return gl.INT
	case float32:
		return gl.FLOAT
	case Vec2:
		return gl.FLOAT_VEC2
	case Vec3:
		return gl.FLOAT_VEC3
	case Vec4:
		return gl.FLOAT_VEC4
	case *Mat4:
		return gl.FLOAT_MAT4
	case *TextureUnit:
		return gl.SAMPLER_2D // TODO: INCORRECT, will malfunction with multi-D samplers
	default:
		panic("attempted to get GL type of unsupported go type")
	}
}

func (u *Uniform) Set(val interface{}) {
	// TODO: pass handler functions, compare reflect.Zero(reflect.TypeOf(val)) interfaces for types?
	// TODO: set more types
	// TODO: store uniform locations only?
	valType := valueGLType(val)

	// TODO: handle samplers correctly
	if u.typ != valType && u.typ != gl.SAMPLER_2D {
		panic("type mismatch between GL type and go value type")
	}

	switch u.typ {
	case gl.INT:
		u.SetInteger(val.(int))
	case gl.FLOAT:
		u.SetFloat(val.(float32))
	case gl.FLOAT_VEC3:
		u.SetVector3(val.(Vec3))
	case gl.FLOAT_MAT4:
		u.SetMatrix4(val.(*Mat4))
	case gl.SAMPLER_2D:
		switch val.(type) {
		case *TextureUnit:
			val := val.(*TextureUnit)
			gl.ProgramUniform1i(u.pID, int32(u.id), val.id)
		}
	default:
		panic("should never get here")
	}
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

func byteSlice(data interface{}) []byte {
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		return []byte{}
	}
	size := val.Len() * int(val.Type().Elem().Size())
	p := unsafe.Pointer(val.Index(0).UnsafeAddr())
	bytes := (*(*[1<<31]byte)(p))[:size]
	return bytes
}

func (b *Buffer) SetData(data interface{}, byteOffset int) {
	bytes := byteSlice(data)
	b.SetBytes(bytes, byteOffset)
}

func (b *Buffer) SetBytes(bytes []byte, byteOffset int) {
	size := len(bytes)
	p := unsafe.Pointer(&bytes[0])
	if size > b.size {
		b.Allocate(size)
	}
	gl.NamedBufferSubData(b.id, byteOffset, int32(size), p)
}

func NewTexture2D(pixelBehavior, edgeBehavior int32) *Texture2D {
	var t Texture2D
	gl.CreateTextures(gl.TEXTURE_2D, 1, &t.id)
	gl.TextureParameteri(t.id, gl.TEXTURE_MIN_FILTER, pixelBehavior)
	gl.TextureParameteri(t.id, gl.TEXTURE_MAG_FILTER, pixelBehavior)
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_S, edgeBehavior)
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_T, edgeBehavior)
	return &t
}

func (t *Texture2D) SetBorderColor(rgba Vec4) {
	gl.TextureParameterfv(t.id, gl.TEXTURE_BORDER_COLOR, &rgba[0])
}

func (t *Texture2D) SetStorage(levels int, format uint32, width, height int) {
	gl.TextureStorage2D(t.id, int32(levels), format, int32(width), int32(height))
}

func (t *Texture2D) SetImage(img image.Image) {
	switch img.(type) {
	case *image.RGBA:
		img := img.(*image.RGBA)
		w, h := img.Bounds().Size().X, img.Bounds().Size().Y
		t.SetStorage(1, gl.RGBA8, w, h)
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
		p := unsafe.Pointer(&byteSlice(img.Pix)[0])
		gl.TextureSubImage2D(t.id, 0, 0, 0, int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, p)
	default:
		imgRGBA := image.NewRGBA(img.Bounds())
		draw.Draw(imgRGBA, imgRGBA.Bounds(), img, img.Bounds().Min, draw.Over)
		t.SetImage(imgRGBA)
	}
}

func NewVertexArray() *VertexArray {
	var va VertexArray
	gl.CreateVertexArrays(1, &va.id)
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
}

func NewTextureUnit(id int) *TextureUnit {
	var tu TextureUnit
	tu.id = int32(id)
	return &tu
}

func (tu *TextureUnit) SetTexture2D(t *Texture2D) {
	gl.BindTextureUnit(uint32(tu.id), t.id)
}

func NewFramebuffer() *Framebuffer {
	var f Framebuffer
	gl.CreateFramebuffers(1, &f.id)
	return &f
}

func (f *Framebuffer) SetTexture(attachment uint32, t *Texture2D, level int32) {
	gl.NamedFramebufferTexture(f.id, attachment, t.id, level)
}

func (f *Framebuffer) ClearColor(rgba Vec4) {
	gl.ClearNamedFramebufferfv(f.id, gl.COLOR, 0,  &rgba[0])
}

func (f *Framebuffer) ClearDepth(clearDepth float32) {
	gl.ClearNamedFramebufferfv(f.id, gl.DEPTH, 0, &clearDepth)
}

func (f *Framebuffer) Complete() bool {
	status := gl.CheckNamedFramebufferStatus(f.id, gl.FRAMEBUFFER)
	return status == gl.FRAMEBUFFER_COMPLETE
}
