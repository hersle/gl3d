package render

import (
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"golang.org/x/image/font/basicfont"
)

type TextShaderProgram struct {
	*graphics.ShaderProgram
	Atlas    *graphics.Uniform
	Position *graphics.Attrib
	TexCoord *graphics.Attrib
}

type TextRenderer struct {
	sp          *TextShaderProgram
	tex         *graphics.Texture2D
	vbo         *graphics.VertexBuffer
	ibo         *graphics.IndexBuffer
	renderState *graphics.State
}

func NewTextShaderProgram() *TextShaderProgram {
	var sp TextShaderProgram
	var err error

	vShaderFilename := "render/shaders/textvshader.glsl" // TODO: make independent from executable directory
	fShaderFilename := "render/shaders/textfshader.glsl" // TODO: make independent from executable directory
	sp.ShaderProgram, err = graphics.ReadShaderProgram(vShaderFilename, fShaderFilename, "")
	if err != nil {
		panic(err)
	}

	sp.Atlas = sp.Uniform("fontAtlas")
	sp.Position = sp.Attrib("position")
	sp.TexCoord = sp.Attrib("texCoordV")

	return &sp
}

func NewTextRenderer() *TextRenderer {
	var r TextRenderer

	r.sp = NewTextShaderProgram()

	r.vbo = graphics.NewVertexBuffer()
	r.ibo = graphics.NewIndexBuffer()

	img := basicfont.Face7x13.Mask
	r.tex = graphics.LoadTexture2D(graphics.ColorTexture, graphics.NearestFilter, graphics.EdgeClampWrap, img)

	r.renderState = graphics.NewState()
	r.renderState.Program = r.sp.ShaderProgram
	r.renderState.PrimitiveType = graphics.Triangle

	return &r
}

func (r *TextRenderer) SetAtlas(tex *graphics.Texture2D) {
	r.sp.Atlas.Set(tex)
}

func (r *TextRenderer) SetAttribs(vbo *graphics.VertexBuffer, ibo *graphics.IndexBuffer) {
	r.sp.Position.SetSourceVertex(vbo, 0)
	r.sp.TexCoord.SetSourceVertex(vbo, 1)
	r.sp.SetAttribIndexBuffer(ibo)
}

func (r *TextRenderer) Render(tl math.Vec2, text string, height float32, fb *graphics.Framebuffer) {
	var verts []object.Vertex
	var inds []int32

	face := basicfont.Face7x13

	x0 := tl.X()
	imgW, imgH := face.Mask.Bounds().Dx(), face.Mask.Bounds().Dy()
	subImgW, subImgH := face.Width, face.Ascent+face.Descent
	h := height
	w := h * float32(subImgW) / float32(subImgH)

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
			tl = math.Vec2{x0, tl.Y() - h}
		} else if char == '\t' {
			tl = tl.Add(math.Vec2{4 * float32(face.Advance) * h / float32(subImgH), 0})
		} else {
			tl = tl.Add(math.Vec2{float32(face.Advance) * h / float32(subImgH), 0})
		}
	}

	r.SetAtlas(r.tex)
	r.vbo.SetData(verts, 0)
	r.ibo.SetData(inds, 0)
	r.SetAttribs(r.vbo, r.ibo)
	r.renderState.Framebuffer = fb
	r.renderState.Render(len(inds))
}
