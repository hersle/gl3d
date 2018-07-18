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
	glType int32
}

// TODO: store value, have Set() function and make "Uniform" an interface?
type UniformBasic struct {
	location uint32
	glType uint32
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
		va.Bind()
		st.vaBound = va
	}
}

func (st *RenderState) SetShaderProgram(prog *ShaderProgram) {
	if st.progBound == nil || st.progBound.id != prog.id {
		prog.Bind()
		st.progBound = prog
	}
}

func (st *RenderState) SetDrawFramebuffer(f *Framebuffer) {
	if st.drawFramebufferBound == nil || st.drawFramebufferBound.id != f.id {
		f.BindDraw()
		st.drawFramebufferBound = f
	}
}

func (st *RenderState) SetReadFramebuffer(f *Framebuffer) {
	if st.readFramebufferBound == nil || st.readFramebufferBound.id != f.id {
		f.BindRead()
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

func (p *ShaderProgram) SetUniformInteger(u *UniformInteger, i int) {
	gl.ProgramUniform1i(p.id, int32(u.location), int32(i))
}

func (p *ShaderProgram) SetUniformFloat(u *UniformFloat, f float32) {
	gl.ProgramUniform1f(p.id, int32(u.location), f)
}

func (p *ShaderProgram) SetUniformVector2(u *UniformVector2, v Vec2) {
	gl.ProgramUniform2fv(p.id, int32(u.location), 1, &v[0])
}

func (p *ShaderProgram) SetUniformVector3(u *UniformVector3, v Vec3) {
	gl.ProgramUniform3fv(p.id, int32(u.location), 1, &v[0])
}

func (p *ShaderProgram) SetUniformVector4(u *UniformVector4, v Vec4) {
	gl.ProgramUniform4fv(p.id, int32(u.location), 1, &v[0])
}

func (p *ShaderProgram) SetUniformMatrix4(u *UniformMatrix4, m *Mat4) {
	gl.ProgramUniformMatrix4fv(p.id, int32(u.location), 1, true, &m[0])
}

func (p *ShaderProgram) SetUniformSampler(u *UniformSampler, tu *TextureUnit) {
	gl.ProgramUniform1i(p.id, int32(u.location), int32(tu.id))
}

func (p *ShaderProgram) Bind() {
	gl.UseProgram(p.id)
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

func (p *ShaderProgram) UniformBasic(name string) (UniformBasic, error) {
	var u UniformBasic
	loc := gl.GetUniformLocation(p.id, gl.Str(name + "\x00"))
	if loc == -1 {
		return u, errors.New(fmt.Sprint(name, " uniform location -1"))
	}
	u.location = uint32(loc)
	gl.GetActiveUniform(p.id, u.location, 0, nil, nil, &u.glType, nil)
	return u, nil
}

func (p *ShaderProgram) UniformInteger(name string) (*UniformInteger, error) {
	var u UniformInteger
	u.UniformBasic, _ = p.UniformBasic(name)
	if u.glType != gl.INT {
		panic("mismatched uniform type")
	}
	return &u, nil
}

func (p *ShaderProgram) UniformFloat(name string) (*UniformFloat, error) {
	var u UniformFloat
	u.UniformBasic, _ = p.UniformBasic(name)
	if u.glType != gl.FLOAT {
		panic("mismatched uniform type")
	}
	return &u, nil
}

func (p *ShaderProgram) UniformVector2(name string) (*UniformVector2, error) {
	var u UniformVector2
	u.UniformBasic, _ = p.UniformBasic(name)
	if u.glType != gl.FLOAT_VEC2 {
		panic("mismatched uniform type")
	}
	return &u, nil
}

func (p *ShaderProgram) UniformVector3(name string) (*UniformVector3, error) {
	var u UniformVector3
	u.UniformBasic, _ = p.UniformBasic(name)
	if u.glType != gl.FLOAT_VEC3 {
		panic("mismatched uniform type")
	}
	return &u, nil
}

func (p *ShaderProgram) UniformVector4(name string) (*UniformVector4, error) {
	var u UniformVector4
	u.UniformBasic, _ = p.UniformBasic(name)
	if u.glType != gl.FLOAT_VEC4 {
		panic("mismatched uniform type")
	}
	return &u, nil
}

func (p *ShaderProgram) UniformMatrix4(name string) (*UniformMatrix4, error) {
	var u UniformMatrix4
	u.UniformBasic, _ = p.UniformBasic(name)
	if u.glType != gl.FLOAT_MAT4 {
		panic("mismatched uniform type")
	}
	return &u, nil
}

func (p *ShaderProgram) UniformSampler(name string) (*UniformSampler, error) {
	var u UniformSampler
	u.UniformBasic, _ = p.UniformBasic(name)
	if u.glType != gl.SAMPLER_2D { // TODO: allow more sampler types
		panic("mismatched uniform type")
	}
	return &u, nil
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

func NewTexture2D(filterMode, wrapMode int32, format uint32, width, height int) *Texture2D {
	var t Texture2D
	gl.CreateTextures(gl.TEXTURE_2D, 1, &t.id)
	gl.TextureParameteri(t.id, gl.TEXTURE_MIN_FILTER, filterMode)
	gl.TextureParameteri(t.id, gl.TEXTURE_MAG_FILTER, filterMode)
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_S, wrapMode)
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_T, wrapMode)
	gl.TextureStorage2D(t.id, 1, format, int32(width), int32(height))
	return &t
}

func NewTexture2DFromImage(filterMode, wrapMode int32, format uint32, img image.Image) *Texture2D {
	switch img.(type) {
	case *image.RGBA:
		img := img.(*image.RGBA)
		w, h := img.Bounds().Size().X, img.Bounds().Size().Y
		t := NewTexture2D(filterMode, wrapMode, format, w, h)
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
		p := unsafe.Pointer(&byteSlice(img.Pix)[0])
		gl.TextureSubImage2D(t.id, 0, 0, 0, int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, p)
		return t
	default:
		imgRGBA := image.NewRGBA(img.Bounds())
		draw.Draw(imgRGBA, imgRGBA.Bounds(), img, img.Bounds().Min, draw.Over)
		return NewTexture2DFromImage(filterMode, wrapMode, format, imgRGBA)
	}
}

func (t *Texture2D) SetBorderColor(rgba Vec4) {
	gl.TextureParameterfv(t.id, gl.TEXTURE_BORDER_COLOR, &rgba[0])
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

func (va *VertexArray) Bind() {
	gl.BindVertexArray(va.id)
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

func (f *Framebuffer) BindDraw() {
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, f.id)
}

func (f *Framebuffer) BindRead() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, f.id)
}
