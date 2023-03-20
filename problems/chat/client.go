package chat

import (
	"bufio"
	"github.com/rs/zerolog/log"
	"net"
	"regexp"
	"strings"
)

type Client struct {
	username string
	conn     net.Conn
	outgoing chan string
	state    ConnectionState

	hub *Hub

	close chan struct{}
}

var (
	alphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]*$`)
)

func ConnectClient(conn net.Conn, hub *Hub) {
	c := Client{
		conn:     conn,
		state:    StateAwaitingName,
		outgoing: make(chan string),
		hub:      hub,
		close:    make(chan struct{}),
	}

	// send welcome message
	log.Info().Str("remote_addr", conn.RemoteAddr().String()).Msg("sending welcome")

	conn.Write([]byte("What should I call you?\n"))

	go c.ReadRoutine()
	go c.WriteRoutine()

	<-c.close

	log.Info().Str("remote_addr", conn.RemoteAddr().String()).Msg("closing")
	conn.Close()
}

func (c *Client) WriteRoutine() {
	for curr := range c.outgoing {
		_, err := c.conn.Write([]byte(curr))
		if err != nil {
			log.Error().Err(err).Msg("could not write")
			c.Close()
			return
		}
	}
}

func (c *Client) ReadRoutine() {
	read := bufio.NewReader(c.conn)
	for {
		line, err := read.ReadString('\n')
		if err != nil {
			log.Error().Err(err).Msg("could not read")
			c.Close()
			return
		}

		line = strings.TrimSpace(line)

		if c.state == StateAwaitingName {
			if !isNameLegal(line) {
				c.conn.Write([]byte("Invalid username. Bye!\n"))
				c.Close()
				return
			}
			c.username = line
			c.state = StateConnected
			log.Info().Str("username", c.username).Msg("set name")

			c.hub.registrationRequests <- c
		} else {
			log.Info().Str("line", line).Msg("got line")
			c.hub.messages <- &Message{
				message: line,
				sender:  c,
			}
		}
	}

}

func (c *Client) Close() {
	c.hub.deregistrationRequests <- c
	close(c.close)
}

func isNameLegal(name string) bool {
	return alphanumericRegex.MatchString(name)
}
