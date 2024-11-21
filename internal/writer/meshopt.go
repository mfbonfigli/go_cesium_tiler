package writer

/*
#cgo CFLAGS: -I./meshoptimizer
#cgo LDFLAGS: -L./meshoptimizer -lmeshoptimizer
#include "meshoptimizer.h"
#include <stdlib.h>
*/

// #cgo pkg-config: meshoptimizer
// #include "meshoptimizer.h"
import "C"
import (
	"errors"
	"unsafe"
)

// EncodeVertexBufferBound computes the maximum size of the encoded buffer.
func EncodeVertexBufferBound(vertexCount, vertexSize int) int {
	return int(C.meshopt_encodeVertexBufferBound(C.size_t(vertexCount), C.size_t(vertexSize)))
}

// EncodeVertexBuffer compresses a vertex buffer into the destination buffer.
// It returns the number of bytes written or an error if the buffer is too small.
func EncodeVertexBuffer(dest []byte, vertices []byte, vertexCount, vertexSize int) (int, error) {
	if len(dest) < EncodeVertexBufferBound(vertexCount, vertexSize) {
		return 0, errors.New("destination buffer too small")
	}

	destPtr := (*C.uchar)(unsafe.Pointer(&dest[0]))
	srcPtr := unsafe.Pointer(&vertices[0])
	written := C.meshopt_encodeVertexBuffer(
		destPtr,
		C.size_t(len(dest)),
		srcPtr,
		C.size_t(vertexCount),
		C.size_t(vertexSize),
	)

	if written == 0 {
		return 0, errors.New("encoding failed")
	}
	return int(written), nil
}
