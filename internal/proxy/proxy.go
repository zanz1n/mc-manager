package proxy

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"sync/atomic"

	"github.com/zanz1n/mc-manager/internal/dto"
)

type Proxy struct {
	Players    atomic.Int32
	MaxPlayers int32
	Active     atomic.Bool

	id                dto.Snowflake
	version           string
	protocolVersion   int32
	description       json.RawMessage
	favIcon           string
	enforceSecureChat bool

	endpoint   net.TCPAddr
	launched   atomic.Bool
	loadedData atomic.Bool

	ln *net.TCPListener
}

func New(
	maxPlayers int32,
	id dto.Snowflake,
	endpoint net.TCPAddr,
) (*Proxy, error) {
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: endpoint.Port,
	})
	if err != nil {
		return nil, err
	}

	return &Proxy{
		MaxPlayers: maxPlayers,
		id:         id,
		endpoint:   endpoint,
		ln:         ln,
	}, nil
}

func (p *Proxy) LoadServerData() error {
	data, err := getServerInfo(&p.endpoint)
	if err != nil {
		return err
	}

	p.loadedData.Store(true)

	p.version = data.Version.Name
	p.protocolVersion = data.Version.Protocol
	p.description = data.Description
	p.favIcon = data.FavIcon
	p.enforceSecureChat = data.EnforceSecureChat

	return nil
}

func (p *Proxy) Close() error {
	return p.ln.Close()
}

func (p *Proxy) Launch() {
	if p.launched.Load() {
		return
	}
	p.launched.Store(true)

	slog.Info("Proxy: Listening", "id", p.id)

	for {
		conn, err := p.ln.AcceptTCP()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				slog.Error(
					"Proxy: Closed listener unexpectedly",
					"id", p.id,
					"error", err,
				)
				p.ln.Close()
			} else {
				slog.Info("Proxy: Closed listener", "id", p.id)
			}
			break
		}

		go func() {
			var err error
			if p.Active.Load() {
				err = p.handleOnline(conn)
			} else {
				err = p.handleOffline(conn)
			}
			if err != nil {
				slog.Warn(
					"Proxy: Failed to handle conn",
					"addr", conn.RemoteAddr(),
					"error", err,
				)
			}
		}()
	}
}

func (p *Proxy) handleOnline(conn *net.TCPConn) error {
	defer conn.Close()

	serverConn, err := net.DialTCP("tcp", nil, &p.endpoint)
	if err != nil {
		return err
	}
	defer serverConn.Close()

	go io.Copy(conn, serverConn)
	_, err = io.Copy(serverConn, conn)
	return err
}

func (p *Proxy) handleOffline(conn *net.TCPConn) error {
	defer conn.Close()
	if !p.loadedData.Load() {
		return errors.New("server data not loaded yet")
	}

	bufread := bufio.NewReader(conn)

	packet, err := ReadPacket(bufread)
	if err != nil {
		return err
	}

	var handshake ServerBoundHandshaking
	err = handshake.DecodeMCP(NewDecoder(packet.Data))
	if err != nil {
		return err
	}

	if handshake.Intent == HandshakingIntentStatus {
		var status ClientBoundStatusRes
		status.Version.Name = p.version
		status.Version.Protocol = p.protocolVersion
		status.Players.Max = p.MaxPlayers
		status.Players.Online = 0
		status.Description = p.description
		status.FavIcon = p.favIcon
		status.EnforceSecureChat = p.enforceSecureChat

		data, err := EncodeMessage(&status)
		if err != nil {
			return err
		}

		err = WritePacket(conn, Packet{
			ID:   ClientBoundStatusResID,
			Data: data,
		})
		if err != nil {
			return err
		}
	} else {
		data, err := EncodeMessage(&ClientBoundLoginDisconect{
			Message: []byte(`{"text":"Starting server ..."}`),
		})
		if err != nil {
			return err
		}

		err = WritePacket(conn, Packet{
			ID:   ClientBoundLoginDisconectID,
			Data: data,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
