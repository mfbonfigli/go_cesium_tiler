package golas

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

// bufferedReadSeeker wraps a io.ReadSeeker adding a buffering layer on top of it.
// Seek operations reset the buffer
type bufferedReadSeeker struct {
	r io.ReadSeeker
	b bufio.Reader
}

func newBufferedReadSeeker(r io.ReadSeeker) *bufferedReadSeeker {
	return &bufferedReadSeeker{
		r: r,
		b: *bufio.NewReaderSize(r, 64*1024),
	}
}

func (b *bufferedReadSeeker) Read(p []byte) (n int, err error) {
	return b.b.Read(p)
}

func (b *bufferedReadSeeker) Seek(offset int64, whence int) (int64, error) {
	defer b.b.Reset(b.r)
	return b.r.Seek(offset, whence)
}

func readString(r io.Reader, n int) (string, error) {
	data, err := readBytes(r, n)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(data), "\u0000"), err
}

func readInt8(r io.Reader) (int8, error) {
	var data int8
	err := binary.Read(r, binary.LittleEndian, &data)
	return data, err
}

func readUint8(r io.Reader) (uint8, error) {
	var data uint8
	err := binary.Read(r, binary.LittleEndian, &data)
	return data, err
}

func readShort(r io.Reader) (int16, error) {
	var data int16
	err := binary.Read(r, binary.LittleEndian, &data)
	return data, err
}

func readUnsignedShort(r io.Reader) (uint16, error) {
	var data uint16
	err := binary.Read(r, binary.LittleEndian, &data)
	return data, err
}

func readLong(r io.Reader) (int32, error) {
	var data int32
	err := binary.Read(r, binary.LittleEndian, &data)
	return data, err
}

func readUnsignedLong(r io.Reader) (uint32, error) {
	var data uint32
	err := binary.Read(r, binary.LittleEndian, &data)
	return data, err
}

func readUnsignedLong64(r io.Reader) (uint64, error) {
	var data uint64
	err := binary.Read(r, binary.LittleEndian, &data)
	return data, err
}

func readFloat32(r io.Reader) (float32, error) {
	var data float32
	err := binary.Read(r, binary.LittleEndian, &data)
	return data, err
}

func readFloat64(r io.Reader) (float64, error) {
	var data float64
	err := binary.Read(r, binary.LittleEndian, &data)
	return data, err
}

func readUnsignedLongArray(r io.Reader, n int) ([]uint32, error) {
	out := make([]uint32, n)
	for i := 0; i < n; i++ {
		data, err := readUnsignedLong(r)
		if err != nil {
			return out, err
		}
		out[i] = data
	}
	return out, nil
}

func readUnsignedLong64Array(r io.Reader, n int) ([]uint64, error) {
	out := make([]uint64, n)
	for i := 0; i < n; i++ {
		data, err := readUnsignedLong64(r)
		if err != nil {
			return out, err
		}
		out[i] = data
	}
	return out, nil
}

func readBytes(r io.Reader, n int) ([]byte, error) {
	data := make([]byte, n)
	nRead, err := io.ReadFull(r, data)
	if err != nil {
		return nil, err
	}
	if nRead != n {
		return nil, errors.New("unexpected number of bytes read")
	}
	return data, nil
}
