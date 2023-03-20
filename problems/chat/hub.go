package chat

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"strings"
)

type ConnectionState int

const (
	StateAwaitingName ConnectionState = iota
	StateConnected
)

type Hub struct {
	clients map[string]*Client

	registrationRequests   chan *Client
	deregistrationRequests chan *Client
	messages               chan *Message
}

func NewHub() *Hub {
	h := &Hub{
		clients:                map[string]*Client{},
		registrationRequests:   make(chan *Client),
		deregistrationRequests: make(chan *Client),
		messages:               make(chan *Message),
	}

	go h.registrationRoutine()
	go h.deregistrationRoutine()
	go h.messageRoutine()

	return h
}

func (h *Hub) registrationRoutine() {
	for curr := range h.registrationRequests {
		log.Info().Str("username", curr.username).Msg("got registration request")

		curr.outgoing <- fmt.Sprintf("* The room contains: %s\n", h.GetUserList())
		h.clients[curr.username] = curr
		h.broadcast(fmt.Sprintf("* %s has joined!\n", curr.username), curr)
	}
}

func (h *Hub) GetUserList() string {
	// there's probably some data race here
	var names []string
	for k := range h.clients {
		names = append(names, k)
	}

	return strings.Join(names, ", ")
}

func (h *Hub) deregistrationRoutine() {
	for curr := range h.deregistrationRequests {
		log.Info().Str("username", curr.username).Msg("got deregistration request")
		if curr.username == "" {
			// dumb hack to ignore users that haven't fully joined
			log.Warn().Msg("got deregitration from unnamed client")
			continue
		}
		delete(h.clients, curr.username)
		h.broadcast(fmt.Sprintf("* %s has left!\n", curr.username), curr)
	}
}

func (h *Hub) messageRoutine() {
	for curr := range h.messages {
		log.Info().Str("username", curr.sender.username).Msg("got message")
		h.broadcast(fmt.Sprintf("[%s] %s\n", curr.sender.username, curr.message), curr.sender)
	}
}

func (h *Hub) broadcast(m string, from *Client) {
	for _, v := range h.clients {
		if v == from {
			log.Debug().Str("username", v.username).Msg("skipping")
			continue
		}
		v.outgoing <- m
	}
}

type Message struct {
	sender  *Client
	message string
}
