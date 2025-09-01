package proxy

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"

	"github.com/google/uuid"
	"github.com/zanz1n/mc-manager/internal/utils"
)

type Decodable interface {
	DecodeMCP(*Decoder) error
}

type Decoder struct {
	buf         *bytes.Buffer
	maxByteSize uint64
}

func NewDecoder(b []byte) *Decoder {
	return &Decoder{
		buf:         bytes.NewBuffer(b),
		maxByteSize: math.MaxUint32,
	}
}

func (d *Decoder) ReadBool() (bool, error) {
	b, err := d.buf.ReadByte()
	return b != 0, err
}

func (d *Decoder) ReadByte() (byte, error) {
	return d.buf.ReadByte()
}

func (d *Decoder) ReadUint16() (uint16, error) {
	b, err := d.readn(2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b), nil
}

func (d *Decoder) ReadUint32() (uint32, error) {
	b, err := d.readn(4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b), nil
}

func (d *Decoder) ReadUint64() (uint64, error) {
	b, err := d.readn(8)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b), nil
}

func (d *Decoder) ReadFloat32() (float32, error) {
	i, err := d.ReadUint32()
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(i), nil
}

func (d *Decoder) ReadFloat64() (float64, error) {
	i, err := d.ReadUint64()
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(i), nil
}

func (d *Decoder) ReadVarInt() (int64, error) {
	return binary.ReadVarint(d.buf)
}

func (d *Decoder) ReadVarUint() (uint64, error) {
	return binary.ReadUvarint(d.buf)
}

func (d *Decoder) ReadBytes() ([]byte, error) {
	size, err := d.ReadVarUint()
	if err != nil {
		return nil, err
	}

	if size > d.maxByteSize {
		return nil, io.ErrShortBuffer
	}

	return d.readn(int(size))
}

func (d *Decoder) ReadString() (string, error) {
	b, err := d.ReadBytes()
	if err != nil {
		return "", err
	}
	return utils.UnsafeString(b), nil
}

func (d *Decoder) ReadLastBytes() []byte {
	return d.buf.Bytes()
}

func (d *Decoder) ReadUUID() (uuid.UUID, error) {
	b, err := d.readn(16)
	if err != nil {
		return uuid.Nil, err
	}

	id := uuid.UUID{}
	copy(id[:], b)

	return id, nil
}

func (d *Decoder) readn(n int) ([]byte, error) {
	b := make([]byte, n)
	n2, err := d.buf.Read(b)
	if err != nil {
		return b, err
	}

	if n2 != n {
		return b, io.ErrShortBuffer
	}
	return b, nil
}
