package graphics

import (
	_ "github.com/ftrvxmtrx/tga"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/math"
	_ "github.com/hersle/gl3d/window" // initialize graphics
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path"
	"unsafe"
)

type Texture2D struct {
	id     uint32
	Width  int
	Height int
}

type CubeMap struct {
	id     uint32
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

func NewTexture2D(filter FilterMode, wrap WrapMode, format uint32, width, height int) *Texture2D {
	var t Texture2D
	t.Width = width
	t.Height = height
	gl.CreateTextures(gl.TEXTURE_2D, 1, &t.id)
	gl.TextureParameteri(t.id, gl.TEXTURE_MIN_FILTER, int32(filter))
	gl.TextureParameteri(t.id, gl.TEXTURE_MAG_FILTER, int32(filter))
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_S, int32(wrap))
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_T, int32(wrap))
	gl.TextureStorage2D(t.id, 1, format, int32(width), int32(height))
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
		gl.TextureSubImage2D(t.id, 0, 0, 0, int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, p)
		return t
	default:
		imgRGBA := image.NewRGBA(img.Bounds())
		draw.Draw(imgRGBA, imgRGBA.Bounds(), img, img.Bounds().Min, draw.Over)
		return NewTexture2DFromImage(filter, wrap, format, imgRGBA)
	}
}

func ReadTexture2D(filter FilterMode, wrap WrapMode, format uint32, filename string) (*Texture2D, error) {
	img, err := readImage(filename)
	if err != nil {
		return nil, err
	}
	return NewTexture2DFromImage(filter, wrap, format, img), nil
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
	gl.TextureParameterfv(t.id, gl.TEXTURE_BORDER_COLOR, &rgba[0])
}

func NewCubeMap(filter FilterMode, format uint32, width, height int) *CubeMap {
	var t CubeMap
	t.Width = width
	t.Height = height
	gl.CreateTextures(gl.TEXTURE_CUBE_MAP, 1, &t.id)

	gl.TextureParameteri(t.id, gl.TEXTURE_MIN_FILTER, int32(filter))
	gl.TextureParameteri(t.id, gl.TEXTURE_MAG_FILTER, int32(filter))
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	gl.TextureStorage2D(t.id, 1, format, int32(width), int32(height))

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
		gl.TextureSubImage3D(t.id, 0, 0, 0, int32(i), int32(w), int32(h), 1, gl.RGBA, gl.UNSIGNED_BYTE, p)
	}

	return t
}

func ReadCubeMap(filter FilterMode, filename1, filename2, filename3, filename4, filename5, filename6 string) *CubeMap {
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
	return NewCubeMapFromImages(filter, imgs[0], imgs[1], imgs[2], imgs[3], imgs[4], imgs[5])
}

func ReadCubeMapFromDir(filter FilterMode, dir string) *CubeMap {
	names := []string{"posx.jpg", "negx.jpg", "posy.jpg", "negy.jpg", "posz.jpg", "negz.jpg"}
	var filenames [6]string
	for i, name := range names {
		filenames[i] = path.Join(dir, name)
		println(filenames[i])
	}
	return ReadCubeMap(filter, filenames[0], filenames[1], filenames[2], filenames[3], filenames[4], filenames[5])
}

func (c *CubeMap) Face(layer CubeMapLayer) *CubeMapFace {
	var f CubeMapFace
	f.CubeMap = c
	f.layer = layer
	return &f
}
