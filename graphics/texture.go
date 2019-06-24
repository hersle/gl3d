package graphics

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/math"
	_ "github.com/hersle/gl3d/window" // initialize graphics
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	_ "github.com/ftrvxmtrx/tga" // must be imported after jpeg?
	"unsafe"
)

type baseTexture struct {
	id     int
	width  int
	height int
}

type ColorTexture struct {
	baseTexture
}

type DepthTexture struct {
	baseTexture
}

//type Texture interface {
//}

type baseCubeMap struct {
	id     int
	width  int
	height int
}

type ColorCubeMap struct {
	baseCubeMap
}

type DepthCubeMap struct {
	baseCubeMap
}

type ColorCubeMapFace struct {
	*ColorCubeMap
	layer CubeMapLayer
}

type DepthCubeMapFace struct {
	*DepthCubeMap
	layer CubeMapLayer
}

type CubeMapLayer int // TODO: rename?

type TextureFilter int

type TextureWrap int

const (
	NearestFilter TextureFilter = TextureFilter(gl.NEAREST)
	LinearFilter  TextureFilter = TextureFilter(gl.LINEAR)
)

const (
	EdgeClampWrap   TextureWrap = TextureWrap(gl.CLAMP_TO_EDGE)
	BorderClampWrap TextureWrap = TextureWrap(gl.CLAMP_TO_BORDER)
	RepeatWrap      TextureWrap = TextureWrap(gl.REPEAT)
)

const (
	PositiveX CubeMapLayer = CubeMapLayer(0)
	NegativeX CubeMapLayer = CubeMapLayer(1)
	PositiveY CubeMapLayer = CubeMapLayer(2)
	NegativeY CubeMapLayer = CubeMapLayer(3)
	PositiveZ CubeMapLayer = CubeMapLayer(4)
	NegativeZ CubeMapLayer = CubeMapLayer(5)
)

func newBaseTexture(filter TextureFilter, wrap TextureWrap, format uint32, w, h int) baseTexture {
	var t baseTexture
	t.width = w
	t.height = h
	var id uint32
	gl.CreateTextures(gl.TEXTURE_2D, 1, &id)
	t.id = int(id)
	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_MIN_FILTER, int32(filter))
	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_MAG_FILTER, int32(filter))
	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_WRAP_S, int32(wrap))
	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_WRAP_T, int32(wrap))
	gl.TextureStorage2D(uint32(t.id), 1, format, int32(w), int32(h))
	return t
}

func NewColorTexture(filter TextureFilter, wrap TextureWrap, w, h int) *ColorTexture {
	var t ColorTexture
	t.baseTexture = newBaseTexture(filter, wrap, gl.RGBA8, w, h)
	return &t
}

func NewDepthTexture(filter TextureFilter, wrap TextureWrap, w, h int) *DepthTexture {
	var t DepthTexture
	t.baseTexture = newBaseTexture(filter, wrap, gl.DEPTH_COMPONENT16, w, h)
	return &t
}

func LoadColorTexture(filter TextureFilter, wrap TextureWrap, img image.Image) *ColorTexture {
	switch img.(type) {
	case *image.RGBA:
		img := img.(*image.RGBA)
		w, h := img.Bounds().Size().X, img.Bounds().Size().Y

		img2 := image.NewRGBA(img.Bounds())
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				img2.SetRGBA(x, img.Bounds().Max.Y-y-1, img.RGBAAt(x, y))
			}
		}

		t := NewColorTexture(filter, wrap, w, h)
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
		p := unsafe.Pointer(&byteSlice(img2.Pix)[0])
		gl.TextureSubImage2D(uint32(t.id), 0, 0, 0, int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, p)
		return t
	default:
		imgRGBA := image.NewRGBA(img.Bounds())
		draw.Draw(imgRGBA, imgRGBA.Bounds(), img, img.Bounds().Min, draw.Over)
		return LoadColorTexture(filter, wrap, imgRGBA)
	}
}

func NewUniformColorTexture(rgba math.Vec4) *ColorTexture {
	// TODO: floating point errors?
	r := uint8(float32(0xff) * rgba.X())
	g := uint8(float32(0xff) * rgba.Y())
	b := uint8(float32(0xff) * rgba.Z())
	a := uint8(float32(0xff) * rgba.W())
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{r, g, b, a})
	return LoadColorTexture(NearestFilter, EdgeClampWrap, img)
}

func (t *baseTexture) Width() int {
	return t.width
}

func (t *baseTexture) Height() int {
	return t.height
}

func (t *ColorTexture) SetBorder(rgba math.Vec4) {
	gl.TextureParameterfv(uint32(t.id), gl.TEXTURE_BORDER_COLOR, &rgba[0])
}

func (t *DepthTexture) SetBorder(depth float32) {
	// TODO: use single parameter GL function?
	rgba := math.Vec4{depth, depth, depth, depth}
	gl.TextureParameterfv(uint32(t.id), gl.TEXTURE_BORDER_COLOR, &rgba[0])
}

func (t *baseTexture) attachToFramebuffer(f *Framebuffer, att uint32) {
	gl.NamedFramebufferTexture(uint32(f.id), att, uint32(t.id), 0)
}

func (t *ColorTexture) attachToFramebuffer(f *Framebuffer) {
	t.baseTexture.attachToFramebuffer(f, gl.COLOR_ATTACHMENT0)
}

func (t *DepthTexture) attachToFramebuffer(f *Framebuffer) {
	t.baseTexture.attachToFramebuffer(f, gl.DEPTH_ATTACHMENT)
}

func (cf *ColorCubeMapFace) attachToFramebuffer(f *Framebuffer) {
	gl.NamedFramebufferTextureLayer(uint32(f.id), gl.COLOR_ATTACHMENT0, uint32(cf.ColorCubeMap.id), 0, int32(cf.layer))
}

func (cf *DepthCubeMapFace) attachToFramebuffer(f *Framebuffer) {
	gl.NamedFramebufferTextureLayer(uint32(f.id), gl.DEPTH_ATTACHMENT, uint32(cf.DepthCubeMap.id), 0, int32(cf.layer))
}

func newBaseCubeMap(filter TextureFilter, format uint32, width, height int) baseCubeMap {
	var t baseCubeMap
	t.width = width
	t.height = height
	var id uint32
	gl.CreateTextures(gl.TEXTURE_CUBE_MAP, 1, &id)
	t.id = int(id)

	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_MIN_FILTER, int32(filter))
	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_MAG_FILTER, int32(filter))
	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	gl.TextureStorage2D(uint32(t.id), 1, format, int32(width), int32(height))

	return t
}

func NewColorCubeMap(filter TextureFilter, width, height int) *ColorCubeMap {
	var c ColorCubeMap
	c.baseCubeMap = newBaseCubeMap(filter, gl.RGBA8, width, height)
	return &c
}

func NewDepthCubeMap(filter TextureFilter, width, height int) *DepthCubeMap {
	var c DepthCubeMap
	c.baseCubeMap = newBaseCubeMap(filter, gl.DEPTH_COMPONENT16, width, height)
	return &c
}

func LoadColorCubeMap(filter TextureFilter, img1, img2, img3, img4, img5, img6 image.Image) *ColorCubeMap {
	w, h := img1.Bounds().Size().X, img1.Bounds().Size().Y
	t := NewColorCubeMap(filter, w, h)

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
		gl.TextureSubImage3D(uint32(t.id), 0, 0, 0, int32(i), int32(w), int32(h), 1, gl.RGBA, gl.UNSIGNED_BYTE, p)
	}

	return t
}

func NewUniformColorCubeMap(rgba math.Vec4) *ColorCubeMap {
	// TODO: floating point errors?
	r := uint8(float32(0xff) * rgba.X())
	g := uint8(float32(0xff) * rgba.Y())
	b := uint8(float32(0xff) * rgba.Z())
	a := uint8(float32(0xff) * rgba.W())
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{r, g, b, a})
	return LoadColorCubeMap(NearestFilter, img, img, img, img, img, img)
}

func (c *ColorCubeMap) Face(layer CubeMapLayer) *ColorCubeMapFace {
	var f ColorCubeMapFace
	f.ColorCubeMap = c
	f.layer = layer
	return &f
}

func (c *DepthCubeMap) Face(layer CubeMapLayer) *DepthCubeMapFace {
	var f DepthCubeMapFace
	f.DepthCubeMap = c
	f.layer = layer
	return &f
}

func (f *ColorCubeMapFace) Width() int {
	return f.ColorCubeMap.width
}

func (f *ColorCubeMapFace) Height() int {
	return f.ColorCubeMap.height
}

func (f *DepthCubeMapFace) Width() int {
	return f.DepthCubeMap.width
}

func (f *DepthCubeMapFace) Height() int {
	return f.DepthCubeMap.height
}
