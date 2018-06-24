package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"unsafe"
)

type RGBAColor [4]uint8

type Vertex struct {
	pos Vec3
	color RGBAColor
}

type Renderer struct {
	prog Program
	vbo *Buffer
	ibo *Buffer
	vaoId uint32
	posLoc uint32
	colorLoc uint32
	verts []Vertex
	inds []int32
}

func NewColor(r, g, b, a uint8) RGBAColor {
	return RGBAColor{r, g, b, a}
	//return RGBAColor((a << 0) | (b << 8) | (g << 16) | (r << 24))
}

func NewRenderer(win *Window) (*Renderer, error) {
	var r Renderer

	win.MakeContextCurrent()

	err := gl.Init()
	if err != nil {
		return nil, err
	}

	r.prog, err = NewProgramFromFiles("vshader.glsl", "fshader.glsl")
	if err != nil {
		return nil, err
	}
	r.prog.use()

	r.posLoc, err = r.prog.attribLocation("position")
	if err != nil {
		return nil, err
	}
	r.colorLoc, err = r.prog.attribLocation("colorV")
	if err != nil {
		return nil, err
	}

	gl.GenVertexArrays(1, &r.vaoId)
	gl.BindVertexArray(r.vaoId)

	r.vbo = NewBuffer(gl.ARRAY_BUFFER)
	r.vbo.bind()

	r.ibo = NewBuffer(gl.ELEMENT_ARRAY_BUFFER)
	r.ibo.bind()

	stride := int32(unsafe.Sizeof(Vertex{}))
	offset := gl.PtrOffset(int(unsafe.Offsetof(Vertex{}.pos)))
	gl.VertexAttribPointer(r.posLoc, 3, gl.DOUBLE, false, stride, offset)
	gl.EnableVertexAttribArray(r.posLoc)

	offset = gl.PtrOffset(int(unsafe.Offsetof(Vertex{}.color)))
	gl.VertexAttribPointer(r.colorLoc, 4, gl.UNSIGNED_BYTE, true, stride, offset)
	gl.EnableVertexAttribArray(r.colorLoc)

	return &r, nil
}

func (r *Renderer) Clear() {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	r.verts = r.verts[:0]
	r.inds = r.inds[:0]
}

func (r *Renderer) RenderTriangle(p1, p2, p3 Vec3, color RGBAColor) {
	r.inds = append(r.inds, int32(len(r.verts) + 0))
	r.inds = append(r.inds, int32(len(r.verts) + 1))
	r.inds = append(r.inds, int32(len(r.verts) + 2))

	r.verts = append(r.verts, Vertex{p1, color})
	r.verts = append(r.verts, Vertex{p2, color})
	r.verts = append(r.verts, Vertex{p3, color})
}

func (r *Renderer) RenderMesh(m *Mesh) {
	for _, i := range m.faces {
		r.inds = append(r.inds, int32(len(r.verts) + i))
	}
	for _, vert := range m.verts {
		r.verts = append(r.verts, vert)
	}
}

func (r *Renderer) Flush() {
	r.vbo.SetData(r.verts, 0)
	r.ibo.SetData(r.inds, 0)
	gl.DrawElements(gl.TRIANGLES, int32(len(r.inds)), gl.UNSIGNED_INT, nil)
}

/*
func drawTriangle(point1, point2, point3 glmath.Vec2, clr color) {
	points := []glmath.Vec2{point1, point2, point3}
	drawConvexPolygon(points, clr)
}

func aspectRatio() float64 {
	w, h := window.GetSize()
	return float64(w) / float64(h)
}

func setMatrix(m *glmath.Mat3) {
	// simply upload to GPU
	gl.UniformMatrix3dv(int32(matLoc), 1, true, &m[0])
}

func setViewport(l, b, r, t int) {
	gl.Viewport(int32(l), int32(b), int32(r), int32(t))
}
*/
