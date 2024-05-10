package utils

import (
	"encoding/binary"
	"io"
	"math"
)

// Writes the 4 byte array corresponding the the given int value to the given reader
func WriteIntAs4ByteNumber(i int, w io.Writer) error {
	b := make([]uint8, 4)
	b[0] = uint8(i)
	b[1] = uint8(i >> 8)
	b[2] = uint8(i >> 16)
	b[3] = uint8(i >> 24)
	_, err := w.Write(b)
	return err
}

// Writes the 4 byte array corresponding the the given int value to the given reader
func WriteUint16As2ByteShort(i uint16, w io.Writer) error {
	b := make([]uint8, 2)
	b[0] = uint8(i)
	b[1] = uint8(i >> 8)
	_, err := w.Write(b)
	return err
}

// Writes a float64 number as a float32 to the given writer
func WriteTruncateFloat64ToFloat32(n float64, w io.Writer) error {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, math.Float32bits(float32(n)))
	_, err := w.Write(bytes)
	return err
}
