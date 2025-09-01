package proxy

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

var _ Encodable = (*ClientBoundStatusRes)(nil)
var _ Decodable = (*ClientBoundStatusRes)(nil)

const ClientBoundStatusResID = 0x00

type ClientBoundStatusRes struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int32  `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int32                `json:"max"`
		Online int32                `json:"online"`
		Sample []StatusPlayerSample `json:"sample"`
	} `json:"players"`
	Description       json.RawMessage `json:"description"`
	FavIcon           string          `json:"favicon"`
	EnforceSecureChat bool            `json:"enforcesSecureChat"`
}

type StatusPlayerSample struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// EncodeMCP implements Encodable.
func (p *ClientBoundStatusRes) EncodeMCP(e *Encoder) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return e.WriteBytes(b)
}

// DecodeMCP implements Decodable.
func (p *ClientBoundStatusRes) DecodeMCP(d *Decoder) error {
	b, err := d.ReadBytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(b, p)
}

var _ Encodable = (*ClientBoundStatusPongRes)(nil)
var _ Decodable = (*ClientBoundStatusPongRes)(nil)

const ClientBoundStatusPongResID = 0x01

type ClientBoundStatusPongRes struct {
	Timestamp time.Time
}

// EncodeMCP implements Encodable.
func (p *ClientBoundStatusPongRes) EncodeMCP(e *Encoder) error {
	return e.WriteUint64(uint64(p.Timestamp.UnixMilli()))
}

// DecodeMCP implements Decodable.
func (p *ClientBoundStatusPongRes) DecodeMCP(d *Decoder) error {
	unix, err := d.ReadUint64()
	if err != nil {
		return err
	}

	p.Timestamp = time.UnixMilli(int64(unix))
	return nil
}

var _ Encodable = (*ClientBoundLoginDisconect)(nil)
var _ Decodable = (*ClientBoundLoginDisconect)(nil)

const ClientBoundLoginDisconectID = 0x00

type ClientBoundLoginDisconect struct {
	Message json.RawMessage
}

// EncodeMCP implements Encodable.
func (p *ClientBoundLoginDisconect) EncodeMCP(e *Encoder) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return e.WriteBytes(b)
}

// DecodeMCP implements Decodable.
func (p *ClientBoundLoginDisconect) DecodeMCP(d *Decoder) error {
	b, err := d.ReadBytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(b, p)
}
