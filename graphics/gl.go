package graphics

import (
	"github.com/hersle/gl3d/math"
	"github.com/go-gl/gl/v4.5-core/gl"
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

type Buffer struct {
	id uint32
	size int
}

type Texture2D struct {
	id uint32
	Width int
	Height int
}

type CubeMap struct {
	id uint32
	Width int
	Height int
}

type Framebuffer struct {
	id uint32
}

type RenderStatistics struct {
	DrawCallCount int
	VertexCount int
}

func readImage(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

var DefaultFramebuffer *Framebuffer = &Framebuffer{0}

var RenderStats *RenderStatistics = &RenderStatistics{}

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
	t.Width = width
	t.Height = height
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

func ReadTexture2D(filterMode, wrapMode int32, format uint32, filename string) (*Texture2D, error) {
	img, err := readImage(filename)
	if err != nil {
		return nil, err
	}
	return NewTexture2DFromImage(filterMode, wrapMode, format, img), nil
}

func (t *Texture2D) SetBorderColor(rgba math.Vec4) {
	gl.TextureParameterfv(t.id, gl.TEXTURE_BORDER_COLOR, &rgba[0])
}

func NewCubeMap(filterMode int32, format uint32, width, height int) *CubeMap {
	var t CubeMap
	t.Width = width
	t.Height = height
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
		var imgRGBA *image.RGBA
		switch img.(type) {
		case *image.RGBA:
			imgRGBA = img.(*image.RGBA)
		default:
			imgRGBA = image.NewRGBA(img.Bounds())
			draw.Draw(imgRGBA, img.Bounds(), img, img.Bounds().Min, draw.Over)
		}
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

func (f *Framebuffer) ClearColor(rgba math.Vec4) {
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

	RenderStats.DrawCallCount++
	RenderStats.VertexCount += cmd.vertexCount
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
	stats.DrawCallCount = -1
	stats.VertexCount = 0
}
