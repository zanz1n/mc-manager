package proxy

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/google/uuid"
	"github.com/zanz1n/mc-manager/internal/utils"
)

type Encodable interface {
	EncodeMCP(*Encoder) error
}

type Encoder struct {
	buf *bytes.Buffer
}

func EncodeMessage(v Encodable) ([]byte, error) {
	enc := NewEncoder()
	err := v.EncodeMCP(enc)
	if err != nil {
		return nil, err
	}
	return enc.Bytes(), nil
}

func NewEncoder() *Encoder {
	return &Encoder{buf: bytes.NewBuffer(nil)}
}

func (e *Encoder) Bytes() []byte {
	return e.buf.Bytes()
}

func (e *Encoder) WriteBool(b bool) error {
	var by byte = 0
	if b {
		by = 1
	}
	return e.buf.WriteByte(by)
}

func (e *Encoder) WriteByte(b byte) error {
	return e.buf.WriteByte(b)
}

func (e *Encoder) WriteUint16(i uint16) error {
	b := binary.BigEndian.AppendUint16(make([]byte, 0, 2), i)
	return e.write(b)
}

func (e *Encoder) WriteUint32(i uint32) error {
	b := binary.BigEndian.AppendUint32(make([]byte, 0, 4), i)
	return e.write(b)
}

func (e *Encoder) WriteUint64(i uint64) error {
	b := binary.BigEndian.AppendUint64(make([]byte, 0, 8), i)
	return e.write(b)
}

func (e *Encoder) WriteFloat32(f float32) error {
	b := binary.BigEndian.AppendUint32(
		make([]byte, 0, 4),
		math.Float32bits(f),
	)
	return e.write(b)
}

func (e *Encoder) WriteFloat64(f float64) error {
	b := binary.BigEndian.AppendUint64(
		make([]byte, 0, 8),
		math.Float64bits(f),
	)
	return e.write(b)
}

func (e *Encoder) WriteVarInt(i int64) error {
	b := binary.AppendVarint(nil, i)
	return e.write(b)
}

func (e *Encoder) WriteVarUint(i uint64) error {
	b := binary.AppendUvarint(nil, i)
	return e.write(b)
}

func (e *Encoder) WriteBytes(b []byte) error {
	err := e.WriteVarInt(int64(len(b)))
	if err != nil {
		return err
	}

	return e.write(b)
}

func (e *Encoder) WriteString(s string) error {
	return e.WriteBytes(utils.UnsafeBytes(s))
}

func (e *Encoder) WriteLastBytes(b []byte) error {
	return e.write(b)
}

func (e *Encoder) WriteUUID(id uuid.UUID) error {
	return e.write(id[:])
}

func (e *Encoder) write(b []byte) error {
	_, err := e.buf.Write(b)
	return err
}
