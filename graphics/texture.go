package graphics

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/math"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	_ "github.com/ftrvxmtrx/tga" // must be imported after jpeg?
	"unsafe"
	gomath "math"
)

type Texture2D struct {
	id     uint32
	width  int
	height int
	type_  TextureType
	levels int
}

type CubeMap struct {
	id     uint32
	width  int
	height int
	type_  TextureType
	levels int
}

type cubeMapFace struct {
	*CubeMap
	layer CubeMapLayer
}

type TextureType int

const (
	ColorTexture TextureType = iota
	DepthTexture
	StencilTexture
)

type TextureFilter int

const (
	NearestFilter TextureFilter = TextureFilter(gl.NEAREST)
	LinearFilter  TextureFilter = TextureFilter(gl.LINEAR)
)

type TextureWrap int

const (
	EdgeClampWrap   TextureWrap = TextureWrap(gl.CLAMP_TO_EDGE)
	BorderClampWrap TextureWrap = TextureWrap(gl.CLAMP_TO_BORDER)
	RepeatWrap      TextureWrap = TextureWrap(gl.REPEAT)
)

type CubeMapLayer int // TODO: rename?

const (
	PositiveX CubeMapLayer = CubeMapLayer(0)
	NegativeX CubeMapLayer = CubeMapLayer(1)
	PositiveY CubeMapLayer = CubeMapLayer(2)
	NegativeY CubeMapLayer = CubeMapLayer(3)
	PositiveZ CubeMapLayer = CubeMapLayer(4)
	NegativeZ CubeMapLayer = CubeMapLayer(5)
)

func NewTexture2D(type_ TextureType, filter TextureFilter, wrap TextureWrap, width, height int, mipmap bool) *Texture2D {
	var tex Texture2D
	tex.width = width
	tex.height = height
	tex.type_ = type_
	gl.CreateTextures(gl.TEXTURE_2D, 1, &tex.id)

	if mipmap {
		tex.levels = 1 + int(gomath.Log2(gomath.Max(float64(width), float64(height))))
	} else {
		tex.levels = 1
	}
	if mipmap && filter == LinearFilter {
		gl.TextureParameteri(tex.id, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	} else {
		gl.TextureParameteri(tex.id, gl.TEXTURE_MIN_FILTER, int32(filter))
	}
	gl.TextureParameteri(tex.id, gl.TEXTURE_MAG_FILTER, int32(filter))
	gl.TextureParameteri(tex.id, gl.TEXTURE_WRAP_S, int32(wrap))
	gl.TextureParameteri(tex.id, gl.TEXTURE_WRAP_T, int32(wrap))
	gl.TextureStorage2D(tex.id, int32(tex.levels), tex.glFormat(), int32(width), int32(height))
	return &tex
}

func LoadTexture2D(type_ TextureType, filter TextureFilter, wrap TextureWrap, img image.Image, mipmap bool) *Texture2D {
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

		tex := NewTexture2D(type_, filter, wrap, w, h, mipmap)
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
		p := unsafe.Pointer(&byteSlice(img2.Pix)[0])
		gl.TextureSubImage2D(tex.id, 0, 0, 0, int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, p)
		if mipmap {
			gl.GenerateTextureMipmap(tex.id)
		}
		return tex
	default:
		imgRGBA := image.NewRGBA(img.Bounds())
		draw.Draw(imgRGBA, imgRGBA.Bounds(), img, img.Bounds().Min, draw.Over)
		return LoadTexture2D(type_, filter, wrap, imgRGBA, mipmap)
	}
}

func NewUniformTexture2D(rgba math.Vec4) *Texture2D {
	// TODO: floating point errors?
	r := uint8(float32(0xff) * rgba.X())
	g := uint8(float32(0xff) * rgba.Y())
	b := uint8(float32(0xff) * rgba.Z())
	a := uint8(float32(0xff) * rgba.W())
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{r, g, b, a})
	return LoadTexture2D(ColorTexture, NearestFilter, EdgeClampWrap, img, false)
}

func (tex *Texture2D) Width() int {
	return tex.width
}

func (tex *Texture2D) Height() int {
	return tex.height
}

func (tex *Texture2D) Clear(rgba math.Vec4) {
	var format uint32
	switch tex.type_ {
	case ColorTexture:
		format = gl.RGBA
	case DepthTexture:
		format = gl.DEPTH_COMPONENT
	default:
		panic("invalid texture type")
	}
	type_ := uint32(gl.FLOAT)

	for level := 0; level < tex.levels; level++ {
		gl.ClearTexImage(tex.id, int32(level), format, type_, unsafe.Pointer(&rgba[0]))
	}
}

func (tex *Texture2D) SetBorderColor(rgba math.Vec4) {
	gl.TextureParameterfv(tex.id, gl.TEXTURE_BORDER_COLOR, &rgba[0])
}

func (tex *Texture2D) Display(blend Blending) {
	displayTexture(tex, blend)
}

func (tex *Texture2D) attachTo(f *framebuffer) {
	var glatt uint32
	switch tex.type_ {
	case ColorTexture:
		glatt = gl.COLOR_ATTACHMENT0
	case DepthTexture:
		glatt = gl.DEPTH_ATTACHMENT
	case StencilTexture:
		glatt = gl.STENCIL_ATTACHMENT
	default:
		panic("invalid texture format")
	}
	gl.NamedFramebufferTexture(f.id, glatt, tex.id, 0)
}

func (tex *Texture2D) glFormat() uint32 {
	switch tex.type_ {
	case ColorTexture:
		return gl.RGBA8
	case DepthTexture:
		return gl.DEPTH_COMPONENT16
	case StencilTexture:
		return gl.STENCIL_INDEX8
	default:
		panic("invalid texture type")
	}
}

func NewCubeMap(type_ TextureType, filter TextureFilter, width, height int) *CubeMap {
	var cube CubeMap
	cube.width = width
	cube.height = height
	cube.type_ = type_
	gl.CreateTextures(gl.TEXTURE_CUBE_MAP, 1, &cube.id)
	cube.levels = 1 // TODO: increase?

	gl.TextureParameteri(cube.id, gl.TEXTURE_MIN_FILTER, int32(filter))
	gl.TextureParameteri(cube.id, gl.TEXTURE_MAG_FILTER, int32(filter))
	gl.TextureParameteri(cube.id, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(cube.id, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(cube.id, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	gl.TextureStorage2D(cube.id, 1, cube.glFormat(), int32(width), int32(height))

	return &cube
}

func LoadCubeMap(filter TextureFilter, img1, img2, img3, img4, img5, img6 image.Image) *CubeMap {
	w, h := img1.Bounds().Size().X, img1.Bounds().Size().Y
	cube := NewCubeMap(ColorTexture, filter, w, h)

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
		gl.TextureSubImage3D(cube.id, 0, 0, 0, int32(i), int32(w), int32(h), 1, gl.RGBA, gl.UNSIGNED_BYTE, p)
	}

	return cube
}

func NewUniformCubeMap(rgba math.Vec4) *CubeMap {
	// TODO: floating point errors?
	r := uint8(float32(0xff) * rgba.X())
	g := uint8(float32(0xff) * rgba.Y())
	b := uint8(float32(0xff) * rgba.Z())
	a := uint8(float32(0xff) * rgba.W())
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{r, g, b, a})
	return LoadCubeMap(NearestFilter, img, img, img, img, img, img)
}

func (cube *CubeMap) Width() int {
	return cube.width
}

func (cube *CubeMap) Height() int {
	return cube.height
}

func (cube *CubeMap) Clear(rgba math.Vec4) {
	var format uint32
	switch cube.type_ {
	case ColorTexture:
		format = gl.RGBA
	case DepthTexture:
		format = gl.DEPTH_COMPONENT
	default:
		panic("invalid texture type")
	}
	type_ := uint32(gl.FLOAT)

	for level := 0; level < cube.levels; level++ {
		gl.ClearTexImage(cube.id, int32(level), format, type_, unsafe.Pointer(&rgba[0]))
	}
}

func (cube *CubeMap) Face(layer CubeMapLayer) *cubeMapFace {
	var face cubeMapFace
	face.CubeMap = cube
	face.layer = layer
	return &face
}

func (cube *CubeMap) attachTo(f *framebuffer) {
	var glatt uint32
	switch cube.type_ {
	case ColorTexture:
		glatt = gl.COLOR_ATTACHMENT0
	case DepthTexture:
		glatt = gl.DEPTH_ATTACHMENT
	case StencilTexture:
		glatt = gl.STENCIL_ATTACHMENT
	default:
		panic("invalid texture format")
	}
	gl.NamedFramebufferTexture(f.id, glatt, cube.id, 0)
}

func (cube *CubeMap) glFormat() uint32 {
	switch cube.type_ {
	case ColorTexture:
		return gl.RGBA8
	case DepthTexture:
		return gl.DEPTH_COMPONENT16
	default:
		panic("invalid texture type")
	}
}

func (face *cubeMapFace) Width() int {
	return face.CubeMap.width
}

func (face *cubeMapFace) Height() int {
	return face.CubeMap.height
}

func (face *cubeMapFace) attachTo(f *framebuffer) {
	var glatt uint32
	switch face.CubeMap.type_ {
	case ColorTexture:
		glatt = gl.COLOR_ATTACHMENT0
	case DepthTexture:
		glatt = gl.DEPTH_ATTACHMENT
	case StencilTexture:
		glatt = gl.STENCIL_ATTACHMENT
	default:
		panic("invalid texture format")
	}
	gl.NamedFramebufferTextureLayer(f.id, glatt, face.CubeMap.id, 0, int32(face.layer))
}
