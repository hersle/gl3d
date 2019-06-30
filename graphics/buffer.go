package graphics

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	_ "github.com/hersle/gl3d/window" // initialize graphics
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

func NewBuffer() *Buffer {
	var b Buffer
	var id uint32
	gl.CreateBuffers(1, &id)
	b.id = int(id)
	b.size = 0
	return &b
}

func (b *Buffer) Allocate(size int) {
	b.size = size
	gl.NamedBufferData(uint32(b.id), int32(b.size), nil, gl.STREAM_DRAW)
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
	size := len(bytes)
	p := unsafe.Pointer(&bytes[0])
	if size > b.size {
		b.Allocate(size)
	}
	gl.NamedBufferSubData(uint32(b.id), byteOffset, int32(size), p)
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

func (b *VertexBuffer) Stride() int {
	return int(b.vertex.Size())
}
