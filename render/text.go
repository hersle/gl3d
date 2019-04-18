package render

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/hersle/gl3d/graphics"
	"github.com/hersle/gl3d/math"
	"github.com/hersle/gl3d/object"
	"golang.org/x/image/font/basicfont"
)

type TextShaderProgram struct {
	*graphics.ShaderProgram
	Atlas    *graphics.UniformSampler
	Position *graphics.Attrib
	TexCoord *graphics.Attrib
}

type TextRenderer struct {
	sp          *TextShaderProgram
	tex         *graphics.Texture2D
	vbo         *graphics.Buffer
	ibo         *graphics.Buffer
	renderState *graphics.RenderState
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

	sp.Atlas = sp.UniformSampler("fontAtlas")
	sp.Position = sp.Attrib("position")
	sp.TexCoord = sp.Attrib("texCoordV")

	sp.Position.SetFormat(gl.FLOAT, false)
	sp.TexCoord.SetFormat(gl.FLOAT, false)

	return &sp
}

func NewTextRenderer() *TextRenderer {
	var r TextRenderer

	r.sp = NewTextShaderProgram()

	r.vbo = graphics.NewBuffer()
	r.ibo = graphics.NewBuffer()

	r.SetAttribs(r.vbo, r.ibo)

	img := basicfont.Face7x13.Mask
	r.tex = graphics.NewTexture2DFromImage(graphics.NearestFilter, graphics.EdgeClampWrap, gl.RGBA8, img)

	r.renderState = graphics.NewRenderState()
	r.renderState.Program = r.sp.ShaderProgram
	r.renderState.BlendSourceFactor = graphics.OneMinusDestinationColorBlendFactor
	r.renderState.BlendDestinationFactor = graphics.OneMinusSourceColorBlendFactor

	return &r
}

func (r *TextRenderer) SetAtlas(tex *graphics.Texture2D) {
	r.sp.Atlas.Set2D(tex)
}

func (r *TextRenderer) SetAttribs(vbo, ibo *graphics.Buffer) {
	var v object.Vertex
	r.sp.Position.SetSource(vbo, v.PositionOffset(), v.Size())
	r.sp.TexCoord.SetSource(vbo, v.TexCoordOffset(), v.Size())
	r.sp.SetAttribIndexBuffer(ibo)
}

func (r *TextRenderer) Render(tl math.Vec2, text string, height float32) {
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
				br := math.NewVec2(tl.X()+w, tl.Y()-h)
				tr := math.NewVec2(br.X(), tl.Y())
				bl := math.NewVec2(tl.X(), br.Y())

				normal := math.NewVec3(0, 0, 0)
				vert1 := object.NewVertex(bl.Vec3(0), math.NewVec2(texX1, texY2), normal, math.Vec3{})
				vert2 := object.NewVertex(br.Vec3(0), math.NewVec2(texX2, texY2), normal, math.Vec3{})
				vert3 := object.NewVertex(tr.Vec3(0), math.NewVec2(texX2, texY1), normal, math.Vec3{})
				vert4 := object.NewVertex(tl.Vec3(0), math.NewVec2(texX1, texY1), normal, math.Vec3{})
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
			tl = math.NewVec2(x0, tl.Y()-h)
		} else if char == '\t' {
			tl = tl.Add(math.NewVec2(4*float32(face.Advance)*h/float32(subImgH), 0))
		} else {
			tl = tl.Add(math.NewVec2(float32(face.Advance)*h/float32(subImgH), 0))
		}
	}

	r.SetAtlas(r.tex)
	r.vbo.SetData(verts, 0)
	r.ibo.SetData(inds, 0)
	graphics.NewRenderCommand(graphics.Triangle, len(inds), 0, r.renderState).Execute()
}
