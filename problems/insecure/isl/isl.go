package isl

import (
	"github.com/rs/zerolog/log"
	"net"
	"time"
)

type Listener struct {
	l net.Listener
}

type Conn struct {
	c net.Conn

	inPos  uint64
	outPos uint64

	chain *cipherChain
}

func Listen(address string) (*Listener, error) {
	underlyingListener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return &Listener{
		l: underlyingListener,
	}, nil
}

func (l *Listener) Accept() (*Conn, error) {
	c, err := l.l.Accept()
	if err != nil {
		return nil, err
	}

	return &Conn{
		c: c,
	}, nil
}

func (l *Listener) Close() error {
	return l.l.Close()
}

func (c *Conn) parseCipherChain() error {
	if c.chain == nil {
		chain, err := buildCipherChain(c.c)
		if err != nil {
			log.Warn().Err(err).Msg("could not build cipher chain")
			c.Close()
			return err
		}

		c.chain = chain
		log.Info().Int("chain_length", len(chain.ciphers)).Msg("built cipher chain")
	}

	return nil
}

func (c *Conn) Read(b []byte) (n int, err error) {
	err = c.parseCipherChain()
	if err != nil {
		return 0, err
	}
	n, err = c.c.Read(b)
	c.chain.Decrypt(b[:n], c.inPos)
	c.inPos += uint64(n)
	return
}

func (c *Conn) Write(b []byte) (n int, err error) {
	err = c.parseCipherChain()
	if err != nil {
		return 0, err
	}
	buff := make([]byte, len(b))
	copy(buff, b)
	c.chain.Encrypt(buff, c.outPos)
	c.outPos += uint64(len(b))
	return c.c.Write(buff)
}

func (c *Conn) Close() error {
	return c.c.Close()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.c.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.c.RemoteAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.c.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.c.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.c.SetWriteDeadline(t)
}
