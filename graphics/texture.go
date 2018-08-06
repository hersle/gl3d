package graphics

import (
	_ "github.com/hersle/gl3d/window" // initialize graphics
	"github.com/go-gl/gl/v4.5-core/gl"
	"os"
	"github.com/hersle/gl3d/math"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	_ "github.com/ftrvxmtrx/tga"
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
				img2.SetRGBA(x, img.Bounds().Max.Y-y, img.RGBAAt(x, y))
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
