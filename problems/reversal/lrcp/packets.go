package lrcp

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"strconv"
)

const (
	PacketKindConnect = "connect"
	PacketKindAck     = "ack"
	PacketKindData    = "data"
	PacketKindClose   = "close"
)

type Packet interface {
	fmt.Stringer
	Serialize() []byte
	Kind() string
	SessionID() int
}

type ConnectPacket struct {
	Session int
}

func (cp *ConnectPacket) SessionID() int {
	return cp.Session
}

func (cp *ConnectPacket) Serialize() []byte {
	return []byte(fmt.Sprintf("/connect/%d/", cp.Session))
}

func (cp *ConnectPacket) Kind() string {
	return PacketKindConnect
}

func (cp *ConnectPacket) String() string {
	return fmt.Sprintf("ConnectPacket{Session: %d}", cp.Session)
}

type DataPacket struct {
	Session  int
	Position int
	Data     []byte
}

func (dp *DataPacket) Serialize() []byte {
	packet := []byte(fmt.Sprintf("/data/%d/%d/", dp.Session, dp.Position))
	packet = append(packet, escape(dp.Data)...)
	packet = append(packet, '/')
	return packet
}

func (dp *DataPacket) SessionID() int {
	return dp.Session
}

func (dp *DataPacket) Kind() string {
	return PacketKindData
}

func (dp *DataPacket) String() string {
	return fmt.Sprintf("DataPacket{Session: %d, Position: %d, Data: (%d bytes)}", dp.Session, dp.Position, len(dp.Data))
}

type AckPacket struct {
	Session int
	Length  int
}

func (ap *AckPacket) SessionID() int {
	return ap.Session
}

func (ap *AckPacket) Serialize() []byte {
	return []byte(fmt.Sprintf("/ack/%d/%d/", ap.Session, ap.Length))
}

func (ap *AckPacket) Kind() string {
	return PacketKindAck
}

func (ap *AckPacket) String() string {
	return fmt.Sprintf("AckPacket{Session: %d, Length: %d}", ap.Session, ap.Length)
}

type ClosePacket struct {
	Session int
}

func (cp *ClosePacket) SessionID() int {
	return cp.Session
}

func (cp *ClosePacket) Serialize() []byte {
	return []byte(fmt.Sprintf("/close/%d/", cp.Session))
}

func (cp *ClosePacket) Kind() string {
	return PacketKindClose
}

func (cp *ClosePacket) String() string {
	return fmt.Sprintf("ClosePacket{Session: %d}", cp.Session)
}

func decodePacket(data []byte) (Packet, error) {
	log.Debug().Str("raw_packet", string(data)).Msg("beginning raw packet decode")
	if len(data) < 2 || len(data) > 1000 {
		return nil, errors.New("invalid size")
	}

	if data[0] != '/' || data[len(data)-1] != '/' {
		return nil, errors.New("invalid preamble or postamble")
	}

	parts := bytes.SplitN(data[1:len(data)-1], []byte{'/'}, 4)
	if len(parts) < 2 {
		return nil, errors.New("could not split parts")
	}

	packetKind := string(parts[0])
	session, err := parseSession(parts[1])
	if err != nil {
		return nil, err
	}

	switch packetKind {
	case PacketKindConnect:
		if len(parts) != 2 {
			return nil, errors.New("invalid part count for connect packet")
		}
		return &ConnectPacket{
			Session: session,
		}, nil
	case PacketKindAck:
		if len(parts) != 3 {
			return nil, errors.New("invalid part count for ack packet")
		}

		parsedLen, err := strconv.Atoi(string(parts[2]))
		if err != nil {
			log.Warn().Str("length", string(parts[2])).Msg("invalid length")
			return nil, errors.New("invalid length")
		}

		if parsedLen < 0 {
			return nil, errors.New("negative length")
		}
		return &AckPacket{
			Session: session,
			Length:  parsedLen,
		}, nil
	case PacketKindClose:
		if len(parts) != 2 {
			return nil, errors.New("invalid part count for close packet")
		}
		return &ClosePacket{
			Session: session,
		}, nil
	case PacketKindData:
		if len(parts) != 4 {
			return nil, errors.New("invalid part count for data packet")
		}
		pos, err := strconv.Atoi(string(parts[2]))
		if err != nil {
			return nil, errors.New("invalid post")
		}
		if pos < 0 {
			return nil, errors.New("negative pos")
		}

		return &DataPacket{
			Session:  session,
			Position: pos,
			Data:     unescape(parts[3]),
		}, nil
	default:
		log.Warn().Str("kind", packetKind).Msg("invalid packet time")
		return nil, errors.New("unknown packet type")
	}

}

func parseSession(data []byte) (int, error) {
	id, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, errors.New("session is not a number")
	}

	return id, nil
}

func escape(x []byte) []byte {
	res := make([]byte, 0)
	for _, c := range x {
		if c == '/' {
			res = append(res, `\/`...)
		} else if c == '\\' {
			res = append(res, `\\`...)
		} else {
			res = append(res, c)
		}
	}

	return res
}

func unescape(x []byte) []byte {
	res := make([]byte, 0)
	for i := 0; i < len(x); i++ {
		// skip the escape character
		if x[i] == '\\' {
			i++
		}
		res = append(res, x[i])
	}
	return res
}
