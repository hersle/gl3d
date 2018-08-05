package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"errors"
	"io/ioutil"
	"reflect"
	"image"
	"image/draw"
	"unsafe"
	"os"
)

// TODO: enable sorting of these states to reduce state changes?
type RenderState struct {
	prog *ShaderProgram
	framebuffer *Framebuffer
	depthTest bool
	depthFunc uint32
	blend bool
	blendSrcFactor uint32
	blendDstFactor uint32
	viewportWidth int
	viewportHeight int
	cull bool
	cullFace uint32
	polygonMode uint32
}

type RenderCommand struct {
	primitiveType uint32
	vertexCount int
	offset int
	state *RenderState
}

type ShaderProgram struct {
	id uint32
	va *VertexArray
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
	width int
	height int
}

type CubeMap struct {
	id uint32
	width int
	height int
}

type Attrib struct {
	prog *ShaderProgram
	id uint32
	nComponents int
}

// TODO: store value, have Set() function and make "Uniform" an interface?
type UniformBasic struct {
	progID uint32
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
	textureUnitIndex uint32
}

type UniformBool struct {
	UniformBasic
}

type VertexArray struct {
	id uint32
	hasIndexBuffer bool
}

type Framebuffer struct {
	id uint32
}

type RenderStatistics struct {
	drawCallCount int
	vertexCount int
}

var defaultFramebuffer *Framebuffer = &Framebuffer{0}

var RenderStats *RenderStatistics = &RenderStatistics{}

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
	p.va = NewVertexArray()
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

func (u *UniformInteger) Set(i int) {
	gl.ProgramUniform1i(u.progID, int32(u.location), int32(i))
}

func (u *UniformFloat) Set(f float32) {
	gl.ProgramUniform1f(u.progID, int32(u.location), f)
}

func (u *UniformVector2) Set(v Vec2) {
	gl.ProgramUniform2fv(u.progID, int32(u.location), 1, &v[0])
}

func (u *UniformVector3) Set(v Vec3) {
	gl.ProgramUniform3fv(u.progID, int32(u.location), 1, &v[0])
}

func (u *UniformVector4) Set(v Vec4) {
	gl.ProgramUniform4fv(u.progID, int32(u.location), 1, &v[0])
}

func (u *UniformMatrix4) Set(m *Mat4) {
	gl.ProgramUniformMatrix4fv(u.progID, int32(u.location), 1, true, &m[0])
}

func (u *UniformSampler) Set2D(t *Texture2D) {
	// TODO: other shaders can mess with this texture index
	gl.BindTextureUnit(u.textureUnitIndex, t.id)
	gl.ProgramUniform1i(u.progID, int32(u.location), int32(u.textureUnitIndex))
}

func (u *UniformSampler) SetCube(t *CubeMap) {
	// TODO: other shaders can mess with this texture index
	gl.BindTextureUnit(u.textureUnitIndex, t.id)
	gl.ProgramUniform1i(u.progID, int32(u.location), int32(u.textureUnitIndex))
}

func (u *UniformBool) Set(b bool) {
	var i int32
	if b {
		i = 1
	} else {
		i = 0
	}
	gl.ProgramUniform1i(u.progID, int32(u.location), i)
}

func (a *Attrib) SetFormat(typ int, normalize bool) {
	a.prog.va.SetAttribFormat(a, a.nComponents, typ, normalize)
}

func (a *Attrib) SetSource(b *Buffer, offset, stride int) {
	a.prog.va.SetAttribSource(a, b, offset, stride)
}

func (p *ShaderProgram) SetAttribIndexBuffer(b *Buffer) {
	p.va.SetIndexBuffer(b)
}

func (p *ShaderProgram) Bind() {
	gl.UseProgram(p.id)
}

func (p *ShaderProgram) Attrib(name string) *Attrib {
	var a Attrib
	loc := gl.GetAttribLocation(p.id, gl.Str(name + "\x00"))
	if loc == -1 {
		return nil
	}
	a.id = uint32(loc)
	a.prog = p

	var size int32
	var typ uint32
	gl.GetActiveAttrib(a.prog.id, a.id, 0, nil, &size, &typ, nil)

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

func (p *ShaderProgram) UniformBasic(name string) *UniformBasic {
	var u UniformBasic
	loc := gl.GetUniformLocation(p.id, gl.Str(name + "\x00"))
	if loc == -1 {
		return nil
	}
	u.location = uint32(loc)
	u.progID = p.id
	index := gl.GetProgramResourceIndex(p.id, gl.UNIFORM, gl.Str(name + "\x00"))
	gl.GetActiveUniform(p.id, index, 0, nil, nil, &u.glType, nil)
	return &u
}

func (p *ShaderProgram) UniformInteger(name string) *UniformInteger {
	var u UniformInteger
	u.UniformBasic = *p.UniformBasic(name)
	if u.glType != gl.INT {
		return nil
	}
	return &u
}

func (p *ShaderProgram) UniformFloat(name string) *UniformFloat {
	var u UniformFloat
	u.UniformBasic = *p.UniformBasic(name)
	if u.glType != gl.FLOAT {
		return nil
	}
	return &u
}

func (p *ShaderProgram) UniformVector2(name string) *UniformVector2 {
	var u UniformVector2
	u.UniformBasic = *p.UniformBasic(name)
	if u.glType != gl.FLOAT_VEC2 {
		return nil
	}
	return &u
}

func (p *ShaderProgram) UniformVector3(name string) *UniformVector3 {
	var u UniformVector3
	u.UniformBasic = *p.UniformBasic(name)
	if u.glType != gl.FLOAT_VEC3 {
		return nil
	}
	return &u
}

func (p *ShaderProgram) UniformVector4(name string) *UniformVector4 {
	var u UniformVector4
	u.UniformBasic = *p.UniformBasic(name)
	if u.glType != gl.FLOAT_VEC4 {
		return nil
	}
	return &u
}

func (p *ShaderProgram) UniformMatrix4(name string) *UniformMatrix4 {
	var u UniformMatrix4
	u.UniformBasic = *p.UniformBasic(name)
	// TODO: what if things not found?
	if u.glType != gl.FLOAT_MAT4 {
		return nil
	}
	return &u
}

func (p *ShaderProgram) UniformSampler(name string) *UniformSampler {
	var u UniformSampler
	u.UniformBasic = *p.UniformBasic(name)
	if u.glType != gl.SAMPLER_2D && u.glType != gl.SAMPLER_CUBE { // TODO: allow more sampler types
		return nil
	}
	u.textureUnitIndex = u.location // TODO: make texture unit mapping more sophisticated
	return &u
}

func (p *ShaderProgram) UniformBool(name string) *UniformBool {
	var u UniformBool
	u.UniformBasic = *p.UniformBasic(name)
	// TODO: what if things not found?
	if u.glType != gl.BOOL {
		return nil
	}
	return &u
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
	t.width = width
	t.height = height
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

		img2 := image.NewRGBA(img.Bounds())
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				img2.SetRGBA(x, img.Bounds().Max.Y - y, img.RGBAAt(x, y))
			}
		}

		t := NewTexture2D(filterMode, wrapMode, format, w, h)
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
		p := unsafe.Pointer(&byteSlice(img2.Pix)[0])
		gl.TextureSubImage2D(t.id, 0, 0, 0, int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, p)
		return t
	default:
		imgRGBA := image.NewRGBA(img.Bounds())
		draw.Draw(imgRGBA, imgRGBA.Bounds(), img, img.Bounds().Min, draw.Over)
		return NewTexture2DFromImage(filterMode, wrapMode, format, imgRGBA)
	}
}

func ReadTexture2D(filterMode, wrapMode int32, format uint32, filename string) *Texture2D {
	file, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil
	}
	return NewTexture2DFromImage(filterMode, wrapMode, format, img)
}

func (t *Texture2D) SetBorderColor(rgba Vec4) {
	gl.TextureParameterfv(t.id, gl.TEXTURE_BORDER_COLOR, &rgba[0])
}

func NewCubeMap(filterMode int32, format uint32, width, height int) *CubeMap {
	var t CubeMap
	t.width = width
	t.height = height
	gl.CreateTextures(gl.TEXTURE_CUBE_MAP, 1, &t.id)

	gl.TextureParameteri(t.id, gl.TEXTURE_MIN_FILTER, filterMode)
	gl.TextureParameteri(t.id, gl.TEXTURE_MAG_FILTER, filterMode)
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	gl.TextureStorage2D(t.id, 1, format, int32(width), int32(height))

	return &t
}

func NewCubeMapFromImages(filterMode int32, img1, img2, img3, img4, img5, img6 image.Image) *CubeMap {
	w, h := img1.Bounds().Size().X, img1.Bounds().Size().Y
	t := NewCubeMap(filterMode, gl.RGBA8, w, h)

	imgs := []image.Image{img1, img2, img3, img4, img5, img6}
	for i, img := range imgs {
		imgRGBA := image.NewRGBA(img.Bounds())
		draw.Draw(imgRGBA, img.Bounds(), img, img.Bounds().Min, draw.Over)
		p := gl.Ptr(imgRGBA.Pix)
		gl.TextureSubImage3D(t.id, 0, 0, 0, int32(i), int32(w), int32(h), 1, gl.RGBA, gl.UNSIGNED_BYTE, p)
	}

	return t
}

func ReadCubeMap(filterMode int32, filename1, filename2, filename3, filename4, filename5, filename6 string) *CubeMap {
	var imgs [6]image.Image
	var errs [6]error
	imgs[0], errs[0] = readImage(filename1)
	imgs[1], errs[1] = readImage(filename2)
	imgs[2], errs[2] = readImage(filename3)
	imgs[3], errs[3] = readImage(filename4)
	imgs[4], errs[4] = readImage(filename5)
	imgs[5], errs[5] = readImage(filename6)
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}
	return NewCubeMapFromImages(filterMode, imgs[0], imgs[1], imgs[2], imgs[3], imgs[4], imgs[5])
}

func NewVertexArray() *VertexArray {
	var va VertexArray
	gl.CreateVertexArrays(1, &va.id)
	va.hasIndexBuffer = false
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
	va.hasIndexBuffer = true
}

func (va *VertexArray) Bind() {
	gl.BindVertexArray(va.id)
}

func NewFramebuffer() *Framebuffer {
	var f Framebuffer
	gl.CreateFramebuffers(1, &f.id)
	return &f
}

func (f *Framebuffer) SetTexture2D(attachment uint32, t *Texture2D, level int32) {
	gl.NamedFramebufferTexture(f.id, attachment, t.id, level)
}

func (f *Framebuffer) SetTextureCubeMapFace(attachment uint32, t *CubeMap, level int32, layer int32) {
	gl.NamedFramebufferTextureLayer(f.id, attachment, t.id, level, layer)
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

func NewRenderCommand(primitiveType uint32, vertexCount, offset int, state *RenderState) *RenderCommand {
	var cmd RenderCommand
	cmd.primitiveType = primitiveType
	cmd.vertexCount = vertexCount
	cmd.offset = offset
	cmd.state = state
	return &cmd
}

func (cmd *RenderCommand) Execute() {
	cmd.state.Apply()
	if cmd.state.prog.va.hasIndexBuffer {
		gl.DrawElements(cmd.primitiveType, int32(cmd.vertexCount), gl.UNSIGNED_INT, nil)
	} else {
		gl.DrawArrays(cmd.primitiveType, int32(cmd.offset), int32(cmd.vertexCount))
	}

	RenderStats.drawCallCount++
	RenderStats.vertexCount += cmd.vertexCount
}

func NewRenderState() *RenderState {
	var rs RenderState
	return &rs
}

func (rs *RenderState) SetShaderProgram(prog *ShaderProgram) {
	rs.prog = prog
}

func (rs *RenderState) SetFramebuffer(fb *Framebuffer) {
	rs.framebuffer = fb
}

func (rs *RenderState) SetDepthTest(depthTest bool) {
	rs.depthTest = depthTest
}

func (rs *RenderState) SetDepthFunc(depthFunc uint32) {
	rs.depthFunc = depthFunc
}

func (rs *RenderState) SetBlend(blend bool) {
	rs.blend = blend
}

func (rs *RenderState) SetBlendFunction(blendSrcFactor, blendDstFactor uint32) {
	rs.blendSrcFactor = blendSrcFactor
	rs.blendDstFactor = blendDstFactor
}

func (rs *RenderState) SetViewport(width, height int) {
	rs.viewportWidth = width
	rs.viewportHeight = height
}

func (rs *RenderState) SetCull(cull bool) {
	rs.cull = cull
}

func (rs *RenderState) SetCullFace(cullFace uint32) {
	rs.cullFace = cullFace
}

func (rs *RenderState) SetPolygonMode(mode uint32) {
	rs.polygonMode = mode
}

func (rs *RenderState) Apply() {
	rs.prog.va.Bind()
	rs.prog.Bind()

	rs.framebuffer.BindDraw()

	if rs.depthTest {
		gl.Enable(gl.DEPTH_TEST)
		gl.DepthFunc(rs.depthFunc)
	} else {
		gl.Disable(gl.DEPTH_TEST)
	}

	if rs.blend {
		gl.Enable(gl.BLEND)
		gl.BlendFunc(rs.blendSrcFactor, rs.blendDstFactor)
	} else {
		gl.Disable(gl.BLEND)
	}

	if rs.cull {
		gl.Enable(gl.CULL_FACE)
		gl.CullFace(rs.cullFace)
	} else {
		gl.Disable(gl.CULL_FACE)
	}

	gl.PolygonMode(gl.FRONT_AND_BACK, rs.polygonMode)

	gl.Viewport(0, 0, int32(rs.viewportWidth), int32(rs.viewportHeight))
}

func (stats *RenderStatistics) Reset() {
	stats.drawCallCount = 0
	stats.vertexCount = 0
}
