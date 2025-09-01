package proxy

import (
	"bufio"
	"net"
)

func getServerInfo(addr *net.TCPAddr) (ClientBoundStatusRes, error) {
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return ClientBoundStatusRes{}, err
	}
	defer conn.Close()

	rd := bufio.NewReader(conn)

	{
		data, err := EncodeMessage(&ServerBoundHandshaking{
			ProtocolVersion: 0,
			ServerAddress:   addr.IP.String(),
			ServerPort:      uint16(addr.Port),
			Intent:          HandshakingIntentStatus,
		})
		if err != nil {
			return ClientBoundStatusRes{}, err
		}

		err = WritePacket(conn, Packet{
			ID:   ServerBoundHandshakingID,
			Data: data,
		})
		if err != nil {
			return ClientBoundStatusRes{}, err
		}
	}

	packet, err := ReadPacket(rd)
	if err != nil {
		return ClientBoundStatusRes{}, err
	}

	var data ClientBoundStatusRes
	err = data.DecodeMCP(NewDecoder(packet.Data))
	return data, err
}
