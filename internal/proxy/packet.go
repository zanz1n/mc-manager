package proxy

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
)

type Packet struct {
	ID   int32
	Data []byte
}

func WritePacket(w io.Writer, p Packet) error {
	idb := binary.AppendVarint(nil, int64(p.ID))

	totalLen := binary.AppendVarint(nil, int64(len(idb)+len(p.Data)))

	buf := bytes.NewBuffer(make([]byte, 0, len(totalLen)+len(idb)+len(p.Data)))
	buf.Write(totalLen)
	buf.Write(idb)
	buf.Write(p.Data)

	_, err := w.Write(buf.Bytes())
	return err
}

func ReadPacket(r *bufio.Reader) (Packet, error) {
	size, err := binary.ReadVarint(r)
	if err != nil {
		return Packet{}, err
	}

	packedId, err := binary.ReadVarint(r)
	if err != nil {
		return Packet{}, err
	}

	dataLen := int(size) - len(binary.AppendVarint(nil, packedId))
	if dataLen == 0 {
		return Packet{ID: int32(packedId)}, nil
	}

	data := make([]byte, dataLen)
	n, err := r.Read(data)
	if err != nil {
		return Packet{}, err
	}

	if n != dataLen {
		return Packet{}, io.ErrUnexpectedEOF
	}

	return Packet{
		ID:   int32(packedId),
		Data: data,
	}, nil
}
