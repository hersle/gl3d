package render

import (
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"golang.org/x/image/font/basicfont"
)

type TextProgram struct {
	*graphics.Program
	Atlas    *graphics.Uniform
	Position *graphics.Input
	TexCoord *graphics.Input
	Color    *graphics.Input

	OutColor *graphics.Output
}

type TextRenderer struct {
	sp          *TextProgram
	tex         *graphics.Texture2D
	vbo         *graphics.VertexBuffer
	ibo         *graphics.IndexBuffer
	renderOpts  *graphics.RenderOptions
}

type Justification int

const (
	BottomLeft Justification = iota
	BottomRight
	TopRight
	TopLeft
)

func NewTextProgram() *TextProgram {
	var sp TextProgram

	vShaderFilename := "render/shaders/textvshader.glsl" // TODO: make independent from executable directory
	fShaderFilename := "render/shaders/textfshader.glsl" // TODO: make independent from executable directory
	sp.Program = graphics.ReadProgram(vShaderFilename, fShaderFilename, "")

	sp.Atlas = sp.UniformByName("fontAtlas")
	sp.Position = sp.InputByName("position")
	sp.TexCoord = sp.InputByName("texCoordV")
	sp.Color = sp.InputByName("colorV")

	sp.OutColor = sp.OutputColorByName("fragColor")

	return &sp
}

func NewTextRenderer() *TextRenderer {
	var r TextRenderer

	r.sp = NewTextProgram()

	r.vbo = graphics.NewVertexBuffer()
	r.ibo = graphics.NewIndexBuffer()

	img := basicfont.Face7x13.Mask
	r.tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.NearestFilter, graphics.EdgeClampWrap, img, false)

	r.renderOpts = graphics.NewRenderOptions()
	r.renderOpts.Primitive = graphics.Triangles

	return &r
}

func (r *TextRenderer) SetAtlas(tex *graphics.Texture2D) {
	r.sp.Atlas.Set(tex)
}

func (r *TextRenderer) SetInputs(vbo *graphics.VertexBuffer, ibo *graphics.IndexBuffer) {
	r.sp.Position.SetSourceVertex(vbo, 0)
	r.sp.TexCoord.SetSourceVertex(vbo, 1)
	r.sp.SetIndices(ibo)
	r.sp.Color.SetSourceVertex(vbo, 2) // normal TODO: don't abuse normal
}

func (r *TextRenderer) Render(org math.Vec2, text string, height float32, color math.Vec3, just Justification, target *graphics.Texture2D) {
	var verts []object.Vertex
	var inds []int32

	face := basicfont.Face7x13

	imgW, imgH := face.Mask.Bounds().Dx(), face.Mask.Bounds().Dy()
	subImgW, subImgH := face.Width, face.Ascent+face.Descent
	h := height
	w := h * float32(subImgW) / float32(subImgH)

	lineWidths := []float32{0}
	totalHeight := float32(h)
	i := 0

	for _, char := range text {
		switch char {
		case '\n':
			totalHeight += h
			lineWidths = append(lineWidths, 0)
			i++
		case '\t':
			lineWidths[i] += 4 * float32(face.Advance) * h / float32(subImgH)
		default:
			lineWidths[i] += float32(face.Advance) * h / float32(subImgH)
		}
	}

	totalWidth := lineWidths[0]
	for _, lineWidth := range lineWidths {
		if lineWidth > totalWidth {
			totalWidth = lineWidth
		}
	}

	tls := make([]math.Vec2, len(lineWidths))
	for i := 0; i < len(lineWidths); i++ {
		var x float32
		switch just {
		case BottomLeft, TopLeft: // left
			x = org.X()
		case BottomRight, TopRight: // right
			x = org.X() - lineWidths[i]
		default:
			panic("invalid justification")
		}

		var y float32
		switch just {
		case TopLeft, TopRight: // top
			y = org.Y() - float32(i) * h
		case BottomLeft, BottomRight: // bottom
			y = org.Y() + totalHeight - float32(i) * h
		default:
			panic("invalid justification")
		}

		tl := math.Vec2{x, y}
		tls[i] = tl
	}

	i = 0
	tl := tls[i]
	for _, char := range text {
		for _, runeRange := range face.Ranges {
			lo, hi, offset := runeRange.Low, runeRange.High, runeRange.Offset
			if char >= lo && char < hi {
				imgX1, imgY1 := 0, imgH-(int(char-lo)+offset)*subImgH
				imgX2, imgY2 := imgX1+subImgW, imgY1-subImgH
				texX1 := float32(imgX1) / float32(imgW) // left
				texY1 := float32(imgY1) / float32(imgH) // top
				texX2 := float32(imgX2) / float32(imgW) // right
				texY2 := float32(imgY2) / float32(imgH) // bottom
				br := math.Vec2{tl.X() + w, tl.Y() - h}
				tr := math.Vec2{br.X(), tl.Y()}
				bl := math.Vec2{tl.X(), br.Y()}

				normal := math.Vec3{0, 0, 0}
				vert1 := object.NewVertex(bl.Vec3(0), math.Vec2{texX1, texY2}, normal, math.Vec3{})
				vert2 := object.NewVertex(br.Vec3(0), math.Vec2{texX2, texY2}, normal, math.Vec3{})
				vert3 := object.NewVertex(tr.Vec3(0), math.Vec2{texX2, texY1}, normal, math.Vec3{})
				vert4 := object.NewVertex(tl.Vec3(0), math.Vec2{texX1, texY1}, normal, math.Vec3{})
				vert1.Normal = color
				vert2.Normal = color
				vert3.Normal = color
				vert4.Normal = color
				inds = append(inds, int32(len(verts)+0))
				inds = append(inds, int32(len(verts)+1))
				inds = append(inds, int32(len(verts)+2))
				inds = append(inds, int32(len(verts)+0))
				inds = append(inds, int32(len(verts)+2))
				inds = append(inds, int32(len(verts)+3))
				verts = append(verts, vert1, vert2, vert3, vert4)
				break
			}
		}

		if char == '\n' {
			i++
			tl = tls[i]
		} else if char == '\t' {
			tl = tl.Add(math.Vec2{4 * float32(face.Advance) * h / float32(subImgH), 0})
		} else {
			tl = tl.Add(math.Vec2{float32(face.Advance) * h / float32(subImgH), 0})
		}
	}

	r.SetAtlas(r.tex)
	r.vbo.SetData(verts, 0)
	r.ibo.SetData(inds, 0)
	r.SetInputs(r.vbo, r.ibo)
	r.sp.OutColor.Set(target)
	r.sp.Render(len(inds), r.renderOpts)
}
