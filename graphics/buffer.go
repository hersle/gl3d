package graphics

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"reflect"
	"unsafe"
)

type Buffer struct {
	id   int
	size int
}

type VertexBuffer struct {
	Buffer
	vertex reflect.Type
}

type IndexBuffer struct {
	Buffer
	index reflect.Type
}

func NewBuffer() *Buffer {
	var b Buffer
	var id uint32
	gl.CreateBuffers(1, &id)
	b.id = int(id)
	b.size = 0
	return &b
}

func (b *Buffer) Size() int {
	return b.size
}

func (b *Buffer) Bytes(i, j int) []byte {
	if b.size == 0 {
		return nil
	}
	if j < i {
		panic("invalid buffer data selection")
	}

	size := j-i
	bytes := make([]byte, size)
	ptr := unsafe.Pointer(&bytes[0])
	gl.GetNamedBufferSubData(uint32(b.id), i, int32(size), ptr)
	return bytes
}

func (b *Buffer) Allocate(size int) {
	b.size = size
	gl.NamedBufferData(uint32(b.id), int32(b.size), nil, gl.STREAM_DRAW)
}

func (b *Buffer) Reallocate(size int) {
	println("reallocating buffer")
	if size < b.size {
		return
	}

	// TODO: very slow. make faster and better!
	bytes := b.Bytes(0, b.size)
	b.Allocate(size)
	b.SetBytes(bytes, 0)
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

func (b *Buffer) SetData(data interface{}, byteOffset int) {
	bytes := byteSlice(data)
	b.SetBytes(bytes, byteOffset)
}

func (b *Buffer) SetBytes(bytes []byte, byteOffset int) {
	if len(bytes) == 0 {
		return
	}

	size := byteOffset + len(bytes)
	p := unsafe.Pointer(&bytes[0])
	if size > b.size {
		b.Reallocate(size)
	}
	gl.NamedBufferSubData(uint32(b.id), byteOffset, int32(len(bytes)), p)
}

func NewVertexBuffer() *VertexBuffer {
	var b VertexBuffer
	b.Buffer = *NewBuffer()
	return &b
}

func (b *VertexBuffer) SetData(data interface{}, byteOffset int) {
	b.vertex = reflect.TypeOf(data).Elem()
	b.Buffer.SetData(data, byteOffset)
}

func (b *VertexBuffer) Offset(i int) int {
	if b.vertex == nil {
		panic("queried vertex buffer with unknown vertex type")
	}

	switch b.vertex.Kind() {
	case reflect.Struct:
		return int(b.vertex.Field(i).Offset)
	case reflect.Slice, reflect.Array:
		return int(b.vertex.Elem().Size()) * i
	default:
		panic("invalid vertex type")
	}
}

func (b *VertexBuffer) ElementSize() int {
	return int(b.vertex.Size())
}

func NewIndexBuffer() *IndexBuffer {
	var b IndexBuffer
	b.Buffer = *NewBuffer()
	return &b
}

func (b *IndexBuffer) SetData(data interface{}, byteOffset int) {
	b.index = reflect.TypeOf(data).Elem()
	b.Buffer.SetData(data, byteOffset)
}

func (b *IndexBuffer) elementGlType() uint32 {
	bits := b.index.Size() * 8
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

func (b *IndexBuffer) ElementSize() int {
	return int(b.index.Size())
}
