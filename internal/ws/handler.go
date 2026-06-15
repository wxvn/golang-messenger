package ws

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	errs "github.com/wxvn/golang-messenger/internal/errors"
	"github.com/wxvn/golang-messenger/internal/logger"
	"github.com/wxvn/golang-messenger/internal/messages"
	"github.com/wxvn/golang-messenger/internal/middleware"
	"github.com/wxvn/golang-messenger/internal/models"
	"github.com/wxvn/golang-messenger/internal/server"
	server_response "github.com/wxvn/golang-messenger/internal/server/response"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Handler struct {
	hub     *Hub
	service *messages.MessagesService
}

func NewHandler(hub *Hub, service *messages.MessagesService) *Handler {
	return &Handler{
		hub:     hub,
		service: service,
	}
}

func (h *Handler) Routes() []server.Route {
	return []server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/ws",
			Handler: h.ServeWS,
		},
	}
}

// ServeWS godoc
// @Summary WebSocket connection
// @Description Real-time messaging WebSocket connection (upgrade to ws protocol)
// @Description
// @Description  ### How to connect
// @Description  `ws://localhost:8080/api/v1/ws`
// @Description
// @Description  ### Auth
// @Description  Requires `Authorization: Bearer <JWT>` header in the initial handshake.
// @Description
// @Description  ### Data Format
// @Description  **Client → Server (Incoming):**
// @Description  ```json
// @Description  {"type": "message", "chat_id": "uuid", "text": "hello"}
// @Description  ```
// @Description  **Server → Client (Outgoing):**
// @Description  ```json
// @Description  {"type": "message", "chat_id": "uuid", "message_id": "uuid", "text": "hello"}
// @Description  ```
// @Tags ws
// @Security BearerAuth
// @Produce json
// @Success 101 "Switching Protocols"
// @Failure 401 {object} server_response.ErrorResponse
// @Failure 400 {object} server_response.ErrorResponse
// @Router /ws [get]
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rw := server_response.NewHTTPResponseHandler(log, w)

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		rw.ErrorResponse(
			fmt.Errorf("user id not found in context"),
			"get user id from context",
		)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &Client{
		UserID:  userID,
		Conn:    conn,
		Send:    make(chan models.OutgoingMessage, 128),
		handler: h,
	}

	h.hub.register <- client

	go client.readPump(h.hub)
	go client.writePump()
}

func (h *Handler) HandleIncoming(userID uuid.UUID, msg IncomingMessage) error {
	switch msg.Type {
	case "message":
		created, members, err := h.service.SendMessage(
			context.Background(),
			models.SendMessageRequest{
				ChatID:   msg.ChatID,
				ToUserID: msg.ToUserID,
				Text:     msg.Text,
			},
			userID,
		)
		if err != nil {
			return err
		}

		out := models.OutgoingMessage{
			Type:      "message",
			ChatID:    created.ChatID,
			MessageID: created.ID,
			SenderID:  created.SenderID,
			Text:      created.Text,
		}

		for _, uid := range members {
			h.hub.Send(uid, out)
		}

		return nil
	default:
		return fmt.Errorf("%w: unknown message type %s", errs.ErrInvalidArgument, msg.Type)
	}
}
