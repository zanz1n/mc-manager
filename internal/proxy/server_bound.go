package proxy

import (
	"fmt"
	"time"
)

var _ Encodable = (HandshakingIntent)(0)
var _ Decodable = (*HandshakingIntent)(nil)

type HandshakingIntent uint8

const (
	HandshakingIntentStatus   = 1
	HandshakingIntentLogin    = 2
	HandshakingIntentTransfer = 3
)

// EncodeMCP implements Encodable.
func (h HandshakingIntent) EncodeMCP(e *Encoder) error {
	return e.WriteVarInt(int64(h))
}

// DecodeMCP implements Decodable.
func (h *HandshakingIntent) DecodeMCP(d *Decoder) error {
	i, err := d.ReadVarInt()
	if err != nil {
		return err
	}

	switch i {
	case 1, 2, 3:
		*h = HandshakingIntent(i)
		return nil
	default:
		return fmt.Errorf(
			"invalid handshaking intent %d expected 1, 2 or 3",
			i,
		)
	}
}

var _ Encodable = (*ServerBoundHandshaking)(nil)
var _ Decodable = (*ServerBoundHandshaking)(nil)

const ServerBoundHandshakingID = 0x00

type ServerBoundHandshaking struct {
	ProtocolVersion int32
	ServerAddress   string
	ServerPort      uint16
	Intent          HandshakingIntent
}

// EncodeMCP implements Encodable.
func (p *ServerBoundHandshaking) EncodeMCP(e *Encoder) error {
	err := e.WriteVarInt(int64(p.ProtocolVersion))
	if err != nil {
		return err
	}

	if err = e.WriteString(p.ServerAddress); err != nil {
		return err
	}

	if err = e.WriteUint16(p.ServerPort); err != nil {
		return err
	}

	return p.Intent.EncodeMCP(e)
}

// DecodeMCP implements Decodable.
func (p *ServerBoundHandshaking) DecodeMCP(d *Decoder) error {
	protocolVersion, err := d.ReadVarInt()
	if err != nil {
		return err
	}
	p.ProtocolVersion = int32(protocolVersion)

	p.ServerAddress, err = d.ReadString()
	if err != nil {
		return err
	}

	p.ServerPort, err = d.ReadUint16()
	if err != nil {
		return err
	}

	return p.Intent.DecodeMCP(d)
}

var _ Encodable = (*ServerBoundStatusReq)(nil)
var _ Decodable = (*ServerBoundStatusReq)(nil)

const ServerBoundStatusReqID = 0x00

type ServerBoundStatusReq struct{}

// EncodeMCP implements Encodable.
func (p *ServerBoundStatusReq) EncodeMCP(*Encoder) error {
	return nil
}

// DecodeMCP implements Decodable.
func (p *ServerBoundStatusReq) DecodeMCP(d *Decoder) error {
	return nil
}

var _ Encodable = (*ServerBoundStatusPingReq)(nil)
var _ Decodable = (*ServerBoundStatusPingReq)(nil)

const ServerBoundStatusPingReqID = 0x01

type ServerBoundStatusPingReq struct {
	Timestamp time.Time
}

// EncodeMCP implements Encodable.
func (p *ServerBoundStatusPingReq) EncodeMCP(e *Encoder) error {
	return e.WriteUint64(uint64(p.Timestamp.UnixMilli()))
}

// DecodeMCP implements Decodable.
func (p *ServerBoundStatusPingReq) DecodeMCP(d *Decoder) error {
	unix, err := d.ReadUint64()
	if err != nil {
		return err
	}

	p.Timestamp = time.UnixMilli(int64(unix))
	return nil
}
