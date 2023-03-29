package reversal

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
	Serialize() []byte
	Kind() string
	String() string
}

type ConnectPacket struct {
	Session int
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
	packet := []byte(fmt.Sprintf("/data/%s/%d/", dp.Session, dp.Position))
	packet = append(packet, dp.Data...)
	packet = append(packet, '/')
	return packet
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

func (ap *AckPacket) Serialize() []byte {
	return []byte(fmt.Sprintf("/connect/%d/", ap.Session))
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

func (cp *ClosePacket) Serialize() []byte {
	return []byte(fmt.Sprintf("/connect/%d/", cp.Session))
}

func (cp *ClosePacket) Kind() string {
	return PacketKindClose
}

func (cp *ClosePacket) String() string {
	return fmt.Sprintf("ClosePacket{Session: %d}", cp.Session)
}

func decodePacket(data []byte) (Packet, error) {
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
			Data:     parts[3],
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
