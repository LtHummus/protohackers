package lrcp

import (
	"errors"
	"github.com/rs/zerolog/log"
	"net"
	"sync"
	"time"
)

const (
	RetransmissionTimeout    = 3 * time.Second
	SessionExpirationTimeout = 1 * time.Minute
)

type recievedPacket struct {
	p    Packet
	from net.Addr
}

type Listener struct {
	conn net.PacketConn

	incomingConnections chan *Conn
	incomingPackets     chan recievedPacket

	address     string
	connections map[int]*Conn
	ticker      *time.Ticker
}

func Listen(address string) (*Listener, error) {

	conn, err := net.ListenPacket("udp", address)
	if err != nil {
		return nil, err
	}

	l := &Listener{
		conn:                conn,
		incomingConnections: make(chan *Conn),
		incomingPackets:     make(chan recievedPacket),
		address:             address,
		connections:         map[int]*Conn{},
	}

	go l.handleIncomingPackets()

	return l, nil
}

func (l *Listener) handleIncomingPackets() {
	go l.readIncomingPackets()
	l.ticker = time.NewTicker(1 * time.Second)
	defer l.ticker.Stop()

	for {
		select {
		case in := <-l.incomingPackets:
			l.handlePacket(in.p, in.from)
		case <-l.ticker.C:
			l.handlePendingRetransmissions()
		}
	}
}

func (l *Listener) Close() {
	l.conn.Close()
}

func (l *Listener) readIncomingPackets() {
	buf := make([]byte, 1024)
	for {
		n, raddr, err := l.conn.ReadFrom(buf)
		if err != nil {
			log.Error().Err(err).Msg("could not read from socket")
			return
		}
		log.Trace().Int("bytes_read", n).Stringer("remote_address", raddr).Msg("packet incomming")
		packet, err := decodePacket(buf[:n])
		if err != nil {
			log.Warn().Stringer("remote_address", raddr).Msg("invalid packet")
			continue
		}

		log.Debug().Stringer("packet", packet).Stringer("remote_address", raddr).Msg("packet decoded")
		l.incomingPackets <- recievedPacket{p: packet, from: raddr}
	}
}

func (l *Listener) Accept() (*Conn, error) {
	conn, ok := <-l.incomingConnections
	if ok {
		return conn, nil
	}

	return nil, errors.New("could not read from incoming connections")
}

func (l *Listener) handlePacket(packet Packet, addr net.Addr) {
	switch p := packet.(type) {
	case *ConnectPacket:
		l.handeConnectPacket(p, addr)
	case *DataPacket:
		l.handleDataPacket(p, addr)
	case *AckPacket:
		l.handleAckPacket(p, addr)
	case *ClosePacket:
		l.handleClosePacket(p, addr)
	}
}

func (l *Listener) handeConnectPacket(p *ConnectPacket, addr net.Addr) {
	log.Debug().Int("session", p.Session).Stringer("remote_address", addr).Msg("handling connection packet")
	conn := l.connections[p.Session]
	var err error
	if conn != nil {
		err = l.sendPacket(&AckPacket{
			Session: p.Session,
			Length:  0,
		}, addr)
	} else {
		conn = &Conn{
			sessionID: p.Session,
			listener:  l,
			remote:    addr,
			recvLock:  *sync.NewCond(&sync.Mutex{}),
			lastAck:   time.Now(),
		}
		err = l.sendPacket(&AckPacket{
			Session: p.Session,
			Length:  0,
		}, addr)
		log.Info().Int("session", p.Session).Msg("creating session")
		l.connections[p.Session] = conn
		l.incomingConnections <- conn
	}

	if err != nil {
		log.Error().Err(err).Msg("could not send ack")
	}
}

func (l *Listener) handleDataPacket(p *DataPacket, addr net.Addr) {
	log.Debug().Int("session", p.Session).Stringer("remote_address", addr).Msg("handling data packet")
	conn := l.connections[p.Session]
	if conn == nil {
		log.Warn().Int("session", p.Session).Msg("unknown session")
		l.sendPacket(&ClosePacket{Session: p.Session}, addr)
		return
	}

	if conn.recvCount >= (p.Position + len(p.Data)) {
		log.Warn().Int("session", p.Session).Msg("data retransmit")
	} else if conn.recvCount >= p.Position {
		log.Info().Int("session", p.Session).Msg("got new data")
		offset := conn.recvCount - p.Position
		conn.recvLock.L.Lock()
		conn.recvBuff.Write(p.Data[offset:])
		conn.recvLock.L.Unlock()
		conn.recvLock.Signal()

		conn.recvCount = p.Position + len(p.Data)
		l.sendPacket(&AckPacket{Session: p.Session, Length: conn.recvCount}, addr)
	} else {
		// behind in data
		l.sendPacket(&AckPacket{Session: p.Session, Length: conn.recvCount}, addr)
	}
}

func (l *Listener) handleAckPacket(p *AckPacket, addr net.Addr) {
	log.Debug().Int("session", p.Session).Stringer("remote_address", addr).Msg("handling ack packet")
	conn := l.connections[p.Session]
	if conn == nil {
		log.Warn().Int("session", p.Session).Msg("unknown session")
		l.sendPacket(&ClosePacket{Session: p.Session}, addr)
		return
	}

	if conn.ackCount > conn.bytesSent {
		// they got more data than we sent?!?
		log.Warn().Int("ack_count", conn.ackCount).Int("packet_count", p.Length).Msg("more data claimed than accounted for")
		l.sendPacket(&ClosePacket{Session: p.Session}, addr)
		delete(l.connections, p.Session)
		conn.setClose()
	} else {
		log.Info().Int("session", p.Session).Int("ack", p.Length).Msg("noting ack")
		conn.sendLock.Lock()
		defer conn.sendLock.Unlock()
		newAck := p.Length - conn.ackCount
		conn.ackCount = p.Length
		conn.lastAck = time.Now()
		if newAck < 0 {
			log.Error().Msg("negative ack count")
		}
		conn.sendBuff.Next(newAck)
		conn.retransmit()
	}
}

func (l *Listener) handleClosePacket(p *ClosePacket, addr net.Addr) {
	log.Debug().Int("session", p.Session).Stringer("remote_address", addr).Msg("handling close packet")
	conn := l.connections[p.Session]
	if conn == nil {
		log.Warn().Int("session", p.Session).Msg("unknown session on close packet")
		l.sendPacket(&ClosePacket{Session: p.Session}, addr)
		return
	}

	log.Info().Int("session", p.Session).Int("ackd", conn.ackCount).Int("bytes_sent", conn.bytesSent).Msg("sending close response")
	l.sendPacket(&ClosePacket{Session: p.Session}, addr)
	delete(l.connections, p.Session)
	conn.setClose()
}
func (l *Listener) handlePendingRetransmissions() {
	now := time.Now()
	for _, curr := range l.connections {
		if curr.ackCount < curr.bytesSent && now.After(curr.lastAck.Add(RetransmissionTimeout)) {
			curr.retransmit()
		}
	}
}

func (l *Listener) sendPacket(p Packet, destination net.Addr) error {
	payload := p.Serialize()
	log.Debug().Stringer("packet", p).Stringer("remote_address", destination).Msg("sending packet")
	_, err := l.conn.WriteTo(payload, destination)
	return err
}
