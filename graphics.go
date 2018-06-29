package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"unsafe"
	"image"
	"image/color"
)

type RGBAColor [4]uint8

type Vertex struct {
	pos Vec3
	color RGBAColor
	texCoord Vec2
}

type Renderer struct {
	prog Program
	vbo *Buffer
	ibo *Buffer
	vaoId uint32
	posLoc uint32
	colorLoc uint32
	texCoordLoc uint32
	projViewModelMatLoc uint32
	verts []Vertex
	inds []int32
}

func NewColor(r, g, b, a uint8) RGBAColor {
	return RGBAColor{r, g, b, a}
}

func NewRenderer(win *Window) (*Renderer, error) {
	var r Renderer

	win.MakeContextCurrent()

	err := gl.Init()
	if err != nil {
		return nil, err
	}

	gl.Enable(gl.DEPTH_TEST)

	r.prog, err = NewProgramFromFiles("vshader.glsl", "fshader.glsl")
	if err != nil {
		return nil, err
	}
	r.prog.use()

	r.posLoc, err = r.prog.attribLocation("position")
	if err != nil {
		println(err.Error())
	}
	r.colorLoc, err = r.prog.attribLocation("colorV")
	if err != nil {
		println(err.Error())
	}
	r.texCoordLoc, err = r.prog.attribLocation("texCoordV")
	if err != nil {
		println(err.Error())
	}
	r.projViewModelMatLoc, err = r.prog.uniformLocation("projectionViewModelMatrix")
	if err != nil {
		println(err.Error())
	}

	gl.GenVertexArrays(1, &r.vaoId)
	gl.BindVertexArray(r.vaoId)

	r.vbo = NewBuffer(gl.ARRAY_BUFFER)
	r.vbo.bind()

	r.ibo = NewBuffer(gl.ELEMENT_ARRAY_BUFFER)
	r.ibo.bind()

	stride := int32(unsafe.Sizeof(Vertex{}))
	offset := gl.PtrOffset(int(unsafe.Offsetof(Vertex{}.pos)))
	gl.VertexAttribPointer(r.posLoc, 3, gl.FLOAT, false, stride, offset)
	gl.EnableVertexAttribArray(r.posLoc)

	offset = gl.PtrOffset(int(unsafe.Offsetof(Vertex{}.color)))
	gl.VertexAttribPointer(r.colorLoc, 4, gl.UNSIGNED_BYTE, true, stride, offset)
	gl.EnableVertexAttribArray(r.colorLoc)

	offset = gl.PtrOffset(int(unsafe.Offsetof(Vertex{}.texCoord)))
	gl.VertexAttribPointer(r.texCoordLoc, 2, gl.FLOAT, false, stride, offset)
	gl.EnableVertexAttribArray(r.texCoordLoc)

	tex := NewTexture2D()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{0xff, 0x00, 0xff, 0xff})
	img.Set(0, 1, color.RGBA{0x00, 0xff, 0x00, 0xff})
	img.Set(1, 0, color.RGBA{0x00, 0x00, 0xff, 0xff})
	img.Set(1, 1, color.RGBA{0xff, 0xff, 0x00, 0xff})
	tex.SetImage(img)

	return &r, nil
}

func (r *Renderer) Clear() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	r.verts = r.verts[:0]
	r.inds = r.inds[:0]
}

func (r *Renderer) renderMesh(m *Mesh, c *Camera) {
	r.SetProjectionViewModelMatrix(c.ProjectionViewMatrix())
	for _, i := range m.faces {
		r.inds = append(r.inds, int32(len(r.verts) + i))
	}
	for _, vert := range m.verts {
		r.verts = append(r.verts, vert)
	}
}

func (r *Renderer) Render(s *Scene, c *Camera) {
	for _, m := range s.meshes {
		r.renderMesh(m, c)
	}
}

func (r *Renderer) SetProjectionViewModelMatrix(m *Mat4) {
	gl.UniformMatrix4fv(int32(r.projViewModelMatLoc), 1, true, &m[0])
}

func (r *Renderer) Flush() {
	r.vbo.SetData(r.verts, 0)
	r.ibo.SetData(r.inds, 0)
	gl.DrawElements(gl.TRIANGLES, int32(len(r.inds)), gl.UNSIGNED_INT, nil)
}

func (r *Renderer) SetViewport(l, b, w, h int) {
	gl.Viewport(int32(l), int32(b), int32(w), int32(h))
}

func (r *Renderer) SetFullViewport(win *Window) {
	w, h := win.Size()
	r.SetViewport(0, 0, w, h)
}
