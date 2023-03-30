package lrcp

import (
	"bytes"
	"github.com/rs/zerolog/log"
	"net"
	"sync"
	"time"
)

type Conn struct {
	sessionID int
	isClosed  bool
	listener  *Listener

	remote net.Addr

	recvCount int
	recvLock  sync.Cond
	recvBuff  bytes.Buffer

	ackCount  int
	lastAck   time.Time
	sendLock  sync.Mutex
	sendBuff  bytes.Buffer
	bytesSent int
}

func (c *Conn) SessionID() int {
	return c.sessionID
}

func (c *Conn) Read(b []byte) (int, error) {
	c.recvLock.L.Lock()
	if c.recvBuff.Len() == 0 {
		c.recvLock.Wait()
		if c.isClosed {
			return 0, net.ErrClosed
		}
	}
	defer c.recvLock.L.Unlock()
	return c.recvBuff.Read(b)
}

func (c *Conn) Write(b []byte) (int, error) {
	c.sendLock.Lock()
	defer c.sendLock.Unlock()

	c.sendBuff.Write(b)
	c.sendData(b, c.bytesSent)
	c.bytesSent += len(b)
	return len(b), nil
}

func (c *Conn) Close() error {
	c.isClosed = true
	return nil
}

func (c *Conn) setClose() {
	c.isClosed = true
	c.recvLock.Signal()
}

func (c *Conn) sendData(b []byte, pos int) {
	if c.isClosed {
		log.Warn().Int("session", c.sessionID).Msg("not sending on closed session")
		return
	}
	size := 800
	for i := 0; i < len(b); i += size {
		end := i + size
		if end > len(b) {
			end = len(b)
		}

		c.listener.sendPacket(&DataPacket{
			Session:  c.sessionID,
			Position: pos + i,
			Data:     b[i:end],
		}, c.remote)
	}
}

func (c *Conn) retransmit() {
	if time.Now().After(c.lastAck.Add(SessionExpirationTimeout)) {
		// guess they're not coming back :(
		log.Warn().Int("session", c.sessionID).Msg("timing out session")
		delete(c.listener.connections, c.sessionID)
		c.setClose()
		return
	}

	if c.ackCount < c.bytesSent {
		c.sendData(c.sendBuff.Bytes(), c.ackCount)
	}
}
