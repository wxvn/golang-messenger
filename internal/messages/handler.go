package messages

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/wxvn/golang-messenger/internal/logger"
	"github.com/wxvn/golang-messenger/internal/middleware"
	"github.com/wxvn/golang-messenger/internal/models"
	"github.com/wxvn/golang-messenger/internal/server"
	server_request "github.com/wxvn/golang-messenger/internal/server/request"
	server_response "github.com/wxvn/golang-messenger/internal/server/response"
)

type HubPublisher interface {
	Send(userID uuid.UUID, msg models.OutgoingMessage)
}

type Handler struct {
	service *MessagesService
	hub     HubPublisher
}

func NewMessagesHandler(service *MessagesService, hub HubPublisher) *Handler {
	return &Handler{
		service: service,
		hub:     hub,
	}
}

func (h *Handler) Routes() []server.Route {
	return []server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/messages",
			Handler: h.SendMessage,
		},
		{
			Method:  http.MethodGet,
			Path:    "/messages",
			Handler: h.GetMessages,
		},
		{
			Method:  http.MethodPatch,
			Path:    "/messages/{id}",
			Handler: h.UpdateMessage,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/messages/{id}",
			Handler: h.DeleteMessage,
		},
	}
}

// SendMessage godoc
// @Summary Send message
// @Tags messages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.SendMessageRequest true "send message"
// @Success 201 {object} Message
// @Failure 400 {object} server_response.ErrorResponse
// @Failure 401 {object} server_response.ErrorResponse
// @Router /messages [post]
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rw := server_response.NewHTTPResponseHandler(log, w)

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		rw.ErrorResponse(fmt.Errorf("user id not found"), "auth")
		return
	}

	var req models.SendMessageRequest
	if err := server_request.DecodeAndValidateRequest(r, &req); err != nil {
		rw.ErrorResponse(err, "decode request")
		return
	}

	msg, members, err := h.service.SendMessage(ctx, req, userID)
	if err != nil {
		rw.ErrorResponse(err, "send message")
		return
	}

	out := models.OutgoingMessage{
		Type:      "message",
		ChatID:    msg.ChatID,
		MessageID: msg.ID,
		SenderID:  msg.SenderID,
		Text:      msg.Text,
	}

	for _, uid := range members {
		h.hub.Send(uid, out)
	}

	rw.JSONResponse(msg, http.StatusCreated)
}

// GetMessages godoc
// @Summary Get messages by chat
// @Tags messages
// @Security BearerAuth
// @Produce json
// @Param chat_id query string true "chat id"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {array} Message
// @Failure 400 {object} server_response.ErrorResponse
// @Failure 401 {object} server_response.ErrorResponse
// @Router /messages [get]
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rw := server_response.NewHTTPResponseHandler(log, w)

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		rw.ErrorResponse(fmt.Errorf("user id not found"), "auth")
		return
	}

	chatIDStr := r.URL.Query().Get("chat_id")
	if chatIDStr == "" {
		rw.ErrorResponse(fmt.Errorf("missing chat_id"), "query")
		return
	}

	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		rw.ErrorResponse(err, "invalid chat_id")
		return
	}

	limit, offset, err := getMessagesQueryParams(r)
	if err != nil {
		rw.ErrorResponse(err, "query params")
		return
	}

	msgs, err := h.service.GetMessages(ctx, chatID, userID, limit, offset)
	if err != nil {
		rw.ErrorResponse(err, "get messages")
		return
	}

	rw.JSONResponse(msgs, http.StatusOK)
}

// UpdateMessage godoc
// @Summary Update message
// @Tags messages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "message id"
// @Param request body PatchMessage true "update message"
// @Success 200 {object} Message
// @Failure 400 {object} server_response.ErrorResponse
// @Failure 401 {object} server_response.ErrorResponse
// @Failure 403 {object} server_response.ErrorResponse
// @Router /messages/{id} [patch]
func (h *Handler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
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

	messageID, err := server_request.GetUUIDPathParam(r, "id")
	if err != nil {
		rw.ErrorResponse(err, "get path")
		return
	}

	var req PatchMessage
	if err := server_request.DecodeAndValidateRequest(r, &req); err != nil {
		rw.ErrorResponse(err, "failed to decode")
		return
	}

	message, err := h.service.UpdateMessage(ctx, req, userID, messageID)
	if err != nil {
		rw.ErrorResponse(err, "update messages chat")
		return
	}

	rw.JSONResponse(message, http.StatusOK)
}

// DeleteMessage godoc
// @Summary Delete message
// @Tags messages
// @Security BearerAuth
// @Param id path string true "message id"
// @Success 204 "No Content"
// @Failure 400 {object} server_response.ErrorResponse
// @Failure 401 {object} server_response.ErrorResponse
// @Failure 403 {object} server_response.ErrorResponse
// @Router /messages/{id} [delete]
func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
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

	messageID, err := server_request.GetUUIDPathParam(r, "id")
	if err != nil {
		rw.ErrorResponse(err, "get path")
		return
	}

	if err := h.service.DeleteMessage(ctx, userID, messageID); err != nil {
		rw.ErrorResponse(err, "delete chat")
		return
	}

	rw.NoContentResponse()
}

func getMessagesQueryParams(r *http.Request) (*int, *int, error) {
	limit, err := server_request.GetIntQueryParam(r, "limit")
	if err != nil {
		return nil, nil, err
	}

	offset, err := server_request.GetIntQueryParam(r, "offset")
	if err != nil {
		return nil, nil, err
	}

	return limit, offset, nil
}
