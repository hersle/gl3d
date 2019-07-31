package graphics

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"reflect"
	"unsafe"
	"log"
)

type buffer struct {
	id   uint32
	size int
}

type VertexBuffer struct {
	buffer
	vertex reflect.Type
}

type IndexBuffer struct {
	buffer
	index reflect.Type
}

func newBuffer() *buffer {
	var buf buffer
	gl.CreateBuffers(1, &buf.id)
	buf.size = 0
	return &buf
}

func (buf *buffer) Size() int {
	return buf.size
}

func (buf *buffer) Bytes(i, j int) []byte {
	if buf.size == 0 {
		return nil
	}
	if j < i {
		panic("invalid buffer data selection")
	}

	size := j-i
	bytes := make([]byte, size)
	ptr := unsafe.Pointer(&bytes[0])
	gl.GetNamedBufferSubData(buf.id, i, size, ptr)
	return bytes
}

func (buf *buffer) Allocate(size int) {
	buf.size = size
	gl.NamedBufferData(buf.id, buf.size, nil, gl.STREAM_DRAW)
}

func (buf *buffer) Reallocate(size int) {
	log.Print("reallocating buffer from ", buf.size, " to ", size, " bytes")
	if size < buf.size {
		return
	}

	// TODO: very slow. make faster and better!
	bytes := buf.Bytes(0, buf.size)
	buf.Allocate(size)
	buf.SetBytes(bytes, 0)
}

func (buf *buffer) SetData(data interface{}, byteOffset int) {
	bytes := byteSlice(data)
	buf.SetBytes(bytes, byteOffset)
}

func (buf *buffer) SetBytes(bytes []byte, byteOffset int) {
	if len(bytes) == 0 {
		return
	}

	size := byteOffset + len(bytes)
	p := unsafe.Pointer(&bytes[0])
	if size > buf.size {
		buf.Reallocate(size)
	}
	gl.NamedBufferSubData(buf.id, byteOffset, len(bytes), p)
}

func NewVertexBuffer() *VertexBuffer {
	var buf VertexBuffer
	buf.buffer = *newBuffer()
	return &buf
}

func (buf *VertexBuffer) SetData(data interface{}, byteOffset int) {
	buf.vertex = reflect.TypeOf(data).Elem()
	buf.buffer.SetData(data, byteOffset)
}

func (buf *VertexBuffer) Offset(i int) int {
	if buf.vertex == nil {
		panic("queried vertex buffer with unknown vertex type")
	}

	switch buf.vertex.Kind() {
	case reflect.Struct:
		return int(buf.vertex.Field(i).Offset)
	case reflect.Slice, reflect.Array:
		return int(buf.vertex.Elem().Size()) * i
	default:
		panic("invalid vertex type")
	}
}

func (buf *VertexBuffer) ElementSize() int {
	return int(buf.vertex.Size())
}

func NewIndexBuffer() *IndexBuffer {
	var buf IndexBuffer
	buf.buffer = *newBuffer()
	return &buf
}

func (buf *IndexBuffer) SetData(data interface{}, byteOffset int) {
	buf.index = reflect.TypeOf(data).Elem()
	buf.buffer.SetData(data, byteOffset)
}

func (buf *IndexBuffer) ElementSize() int {
	return int(buf.index.Size())
}

func (buf *IndexBuffer) elementGlType() uint32 {
	bits := buf.index.Size() * 8
	switch bits {
	case 8:
		return gl.UNSIGNED_BYTE
	case 16:
		return gl.UNSIGNED_SHORT
	case 32:
		return gl.UNSIGNED_INT
	default:
		panic("invalid index buffer type")
	}
}

func byteSlice(data interface{}) []byte {
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		return []byte{}
	}
	size := val.Len() * int(val.Type().Elem().Size())
	p := unsafe.Pointer(val.Index(0).UnsafeAddr())
	bytes := (*(*[1 << 31]byte)(p))[:size]
	return bytes
}
