package ws

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/wxvn/golang-messenger/internal/models"
)

type IncomingMessage struct {
	Type     string     `json:"type"`
	ChatID   *uuid.UUID `json:"chat_id"`
	ToUserID *uuid.UUID `json:"to_user_id"`
	Text     string     `json:"text"`
}

type Client struct {
	UserID  uuid.UUID
	Conn    *websocket.Conn
	Send    chan models.OutgoingMessage
	handler *Handler
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		var msg IncomingMessage

		if err := c.Conn.ReadJSON(&msg); err != nil {
			return
		}

		if err := c.handler.HandleIncoming(c.UserID, msg); err != nil {
			return
		}
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()

	for msg := range c.Send {
		if err := c.Conn.WriteJSON(msg); err != nil {
			return
		}
	}
}
