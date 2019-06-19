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
	_ "github.com/ftrvxmtrx/tga"
	"unsafe"
)

type Texture2D struct {
	id     int
	Width  int
	Height int
}

type CubeMap struct {
	id     int
	Width  int
	Height int
}

type CubeMapFace struct {
	*CubeMap
	layer CubeMapLayer
}

type FilterMode int

type WrapMode int

type CubeMapLayer int // TODO: rename?

const (
	NearestFilter FilterMode = FilterMode(gl.NEAREST)
	LinearFilter  FilterMode = FilterMode(gl.LINEAR)
)

const (
	EdgeClampWrap   WrapMode = WrapMode(gl.CLAMP_TO_EDGE)
	BorderClampWrap WrapMode = WrapMode(gl.CLAMP_TO_BORDER)
	RepeatWrap      WrapMode = WrapMode(gl.REPEAT)
)

const (
	PositiveX CubeMapLayer = CubeMapLayer(0)
	NegativeX CubeMapLayer = CubeMapLayer(1)
	PositiveY CubeMapLayer = CubeMapLayer(2)
	NegativeY CubeMapLayer = CubeMapLayer(3)
	PositiveZ CubeMapLayer = CubeMapLayer(4)
	NegativeZ CubeMapLayer = CubeMapLayer(5)
)

func NewTexture2D(filter FilterMode, wrap WrapMode, format uint32, width, height int) *Texture2D {
	var t Texture2D
	t.Width = width
	t.Height = height
	var id uint32
	gl.CreateTextures(gl.TEXTURE_2D, 1, &id)
	t.id = int(id)
	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_MIN_FILTER, int32(filter))
	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_MAG_FILTER, int32(filter))
	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_WRAP_S, int32(wrap))
	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_WRAP_T, int32(wrap))
	gl.TextureStorage2D(uint32(t.id), 1, format, int32(width), int32(height))
	return &t
}

func NewTexture2DFromImage(filter FilterMode, wrap WrapMode, format uint32, img image.Image) *Texture2D {
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

		t := NewTexture2D(filter, wrap, format, w, h)
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
		p := unsafe.Pointer(&byteSlice(img2.Pix)[0])
		gl.TextureSubImage2D(uint32(t.id), 0, 0, 0, int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, p)
		return t
	default:
		imgRGBA := image.NewRGBA(img.Bounds())
		draw.Draw(imgRGBA, imgRGBA.Bounds(), img, img.Bounds().Min, draw.Over)
		return NewTexture2DFromImage(filter, wrap, format, imgRGBA)
	}
}

func NewTexture2DUniform(rgba math.Vec4) *Texture2D {
	// TODO: floating point errors?
	r := uint8(float32(0xff) * rgba.X())
	g := uint8(float32(0xff) * rgba.Y())
	b := uint8(float32(0xff) * rgba.Z())
	a := uint8(float32(0xff) * rgba.W())
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{r, g, b, a})
	return NewTexture2DFromImage(NearestFilter, EdgeClampWrap, gl.RGBA8, img)
}

func (t *Texture2D) SetBorderColor(rgba math.Vec4) {
	gl.TextureParameterfv(uint32(t.id), gl.TEXTURE_BORDER_COLOR, &rgba[0])
}

func NewCubeMap(filter FilterMode, format uint32, width, height int) *CubeMap {
	var t CubeMap
	t.Width = width
	t.Height = height
	var id uint32
	gl.CreateTextures(gl.TEXTURE_CUBE_MAP, 1, &id)
	t.id = int(id)

	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_MIN_FILTER, int32(filter))
	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_MAG_FILTER, int32(filter))
	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(uint32(t.id), gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	gl.TextureStorage2D(uint32(t.id), 1, format, int32(width), int32(height))

	return &t
}

func NewCubeMapFromImages(filter FilterMode, img1, img2, img3, img4, img5, img6 image.Image) *CubeMap {
	w, h := img1.Bounds().Size().X, img1.Bounds().Size().Y
	t := NewCubeMap(filter, gl.RGBA8, w, h)

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

func NewCubeMapUniform(rgba math.Vec4) *CubeMap {
	// TODO: floating point errors?
	r := uint8(float32(0xff) * rgba.X())
	g := uint8(float32(0xff) * rgba.Y())
	b := uint8(float32(0xff) * rgba.Z())
	a := uint8(float32(0xff) * rgba.W())
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{r, g, b, a})
	return NewCubeMapFromImages(NearestFilter, img, img, img, img, img, img)
}

func (c *CubeMap) Face(layer CubeMapLayer) *CubeMapFace {
	var f CubeMapFace
	f.CubeMap = c
	f.layer = layer
	return &f
}
