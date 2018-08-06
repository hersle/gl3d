package graphics

import (
	_ "github.com/hersle/gl3d/window" // initialize graphics
	"github.com/go-gl/gl/v4.5-core/gl"
	"reflect"
	"unsafe"
)

type Buffer struct {
	id   uint32
	size int
}

func NewBuffer() *Buffer {
	var b Buffer
	gl.CreateBuffers(1, &b.id)
	b.size = 0
	return &b
}

func (b *Buffer) Allocate(size int) {
	b.size = size
	gl.NamedBufferData(b.id, int32(b.size), nil, gl.STREAM_DRAW)
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
	gl.NamedBufferSubData(b.id, byteOffset, int32(size), p)
}
