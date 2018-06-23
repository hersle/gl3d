package main

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"unsafe"
)

type Vertex struct {
	pos Vec3
}

type Renderer struct {
	prog Program
	vboId, vaoId, iboId uint32
	vboSize, iboSize int
	posLoc uint32
	verts []Vertex
	inds []int32
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

	gl.GenVertexArrays(1, &r.vaoId)
	gl.BindVertexArray(r.vaoId)

	gl.GenBuffers(1, &r.vboId)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vboId)
	r.vboSize = 0

	gl.GenBuffers(1, &r.iboId)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.iboId)
	r.iboSize = 0

	stride := int32(unsafe.Sizeof(Vertex{}))
	offset := gl.PtrOffset(int(unsafe.Offsetof(Vertex{}.pos)))
	gl.VertexAttribPointer(r.posLoc, 3, gl.DOUBLE, false, stride, offset)
	gl.EnableVertexAttribArray(r.posLoc)

	return &r, nil
}

func (r *Renderer) Clear() {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	r.verts = r.verts[:0]
	r.inds = r.inds[:0]
}

func (r *Renderer) DrawTriangle(p1, p2, p3 Vec3) {
	r.inds = append(r.inds, int32(len(r.verts) + 0))
	r.inds = append(r.inds, int32(len(r.verts) + 1))
	r.inds = append(r.inds, int32(len(r.verts) + 2))

	r.verts = append(r.verts, Vertex{p1})
	r.verts = append(r.verts, Vertex{p2})
	r.verts = append(r.verts, Vertex{p3})

	println(len(r.inds), len(r.verts))
}

/*
func (r *Renderer) Render(m *Mesh) {
}
*/

func (r *Renderer) Flush() {
	// if > 0 tests necessary since gl.Ptr() fails on slices with zero length

	//gl.BindBuffer(gl.ARRAY_BUFFER, r.vboId)
	//gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.iboId)

	size := int(unsafe.Sizeof(Vertex{})) * len(r.verts)
	if size > r.vboSize {
		// reallocate TODO: improve reallocation (*2)?
		gl.BufferData(gl.ARRAY_BUFFER, size, nil, gl.STREAM_DRAW)
		r.vboSize = size
	}
	println("VBO size", size, "bytes")
	if size > 0 {
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, size, gl.Ptr(r.verts))
	}

	size = int(unsafe.Sizeof(int(0))) * len(r.inds)
	if size > r.iboSize {
		// reallocate
		// TODO: improve reallocation (*2)?
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, size, nil, gl.STREAM_DRAW)
		r.iboSize = size
	}
	println("IBO size", size, "bytes")
	if size > 0 {
		gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, 0, size, gl.Ptr(r.inds))
		gl.DrawElements(gl.TRIANGLES, int32(len(r.inds)), gl.UNSIGNED_INT, nil)
		println("drawing", len(r.inds))
	}
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
