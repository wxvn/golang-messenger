package ws

import (
	"github.com/google/uuid"
	"github.com/wxvn/golang-messenger/internal/models"
)

type delivery struct {
	UserID uuid.UUID
	Msg    models.OutgoingMessage
}

type Hub struct {
	clients map[uuid.UUID]map[*Client]struct{}

	register   chan *Client
	unregister chan *Client
	deliver    chan delivery
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]struct{}),
		register:   make(chan *Client, 128),
		unregister: make(chan *Client, 128),
		deliver:    make(chan delivery, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {

		case c := <-h.register:
			if h.clients[c.UserID] == nil {
				h.clients[c.UserID] = make(map[*Client]struct{})
			}
			h.clients[c.UserID][c] = struct{}{}

		case c := <-h.unregister:
			h.removeClient(c)

		case d := <-h.deliver:
			for c := range h.clients[d.UserID] {
				select {
				case c.Send <- d.Msg:
				default:
					h.removeClient(c)
				}
			}
		}
	}
}

func (h *Hub) Send(userID uuid.UUID, msg models.OutgoingMessage) {
	h.deliver <- delivery{
		UserID: userID,
		Msg:    msg,
	}
}

func (h *Hub) removeClient(c *Client) {
	conns := h.clients[c.UserID]
	if conns == nil {
		return
	}

	if _, exists := conns[c]; !exists {
		return
	}

	delete(conns, c)
	close(c.Send)

	if len(conns) == 0 {
		delete(h.clients, c.UserID)
	}
}
