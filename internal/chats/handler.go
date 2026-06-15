package chats

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/wxvn/golang-messenger/internal/logger"
	"github.com/wxvn/golang-messenger/internal/middleware"
	"github.com/wxvn/golang-messenger/internal/server"
	server_request "github.com/wxvn/golang-messenger/internal/server/request"
	server_response "github.com/wxvn/golang-messenger/internal/server/response"
)

func (h *Handler) Routes() []server.Route {
	return []server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/chats",
			Handler: h.GetChats,
		},
		{
			Method:  http.MethodGet,
			Path:    "/chats/{id}",
			Handler: h.GetChat,
		},
		{
			Method:  http.MethodPost,
			Path:    "/chats",
			Handler: h.CreateChatGroup,
		},
		{
			Method:  http.MethodPatch,
			Path:    "/chats/{id}",
			Handler: h.UpdateChatGroup,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/chats/{id}",
			Handler: h.DeleteChat,
		},
	}
}

type Handler struct {
	service *ChatsService
}

func NewChatsHandler(service *ChatsService) *Handler {
	return &Handler{
		service: service,
	}
}

// GetChat godoc
// @Summary Get chat by id
// @Tags chats
// @Security BearerAuth
// @Produce json
// @Param id path string true "Chat ID"
// @Success 200 {object} Chat
// @Failure 400 {object} server_response.ErrorResponse
// @Failure 404 {object} server_response.ErrorResponse
// @Router /chats/{id} [get]
func (h *Handler) GetChat(w http.ResponseWriter, r *http.Request) {
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

	chatID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		rw.ErrorResponse(
			err,
			"parse chat id",
		)
		return
	}

	chat, err := h.service.GetChat(ctx, chatID, userID)
	if err != nil {
		rw.ErrorResponse(err, "get chat")
		return
	}

	rw.JSONResponse(chat, http.StatusOK)
}

// GetChats godoc
// @Summary Get user chats
// @Tags chats
// @Security BearerAuth
// @Produce json
// @Success 200 {array} Chat
// @Failure 401 {object} server_response.ErrorResponse
// @Router /chats [get]
func (h *Handler) GetChats(w http.ResponseWriter, r *http.Request) {
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

	chats, err := h.service.GetChats(ctx, userID)
	if err != nil {
		rw.ErrorResponse(err, "get chats")
		return
	}

	rw.JSONResponse(chats, http.StatusOK)
}

// CreateChatGroup godoc
// @Summary Create group chat
// @Tags chats
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateChatRequest true "create chat"
// @Success 201 {object} Chat
// @Failure 400 {object} server_response.ErrorResponse
// @Failure 401 {object} server_response.ErrorResponse
// @Router /chats [post]
func (h *Handler) CreateChatGroup(w http.ResponseWriter, r *http.Request) {
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

	var req CreateChatRequest

	if err := server_request.DecodeAndValidateRequest(r, &req); err != nil {
		rw.ErrorResponse(err, "decode request")
		return
	}

	chat, err := h.service.CreateChatGroup(ctx, userID, req.ChatName, req.MemberIDs)
	if err != nil {
		rw.ErrorResponse(err, "create chat")
		return
	}

	rw.JSONResponse(chat, http.StatusCreated)
}

// UpdateChatGroup godoc
// @Summary Update chat group
// @Tags chats
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Chat ID"
// @Param request body PatchChat true "update chat"
// @Success 200 {object} Chat
// @Failure 400 {object} server_response.ErrorResponse
// @Failure 403 {object} server_response.ErrorResponse
// @Router /chats/{id} [patch]
func (h *Handler) UpdateChatGroup(w http.ResponseWriter, r *http.Request) {
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

	chatID, err := server_request.GetUUIDPathParam(r, "id")
	if err != nil {
		rw.ErrorResponse(err, "get path")
		return
	}

	var req PatchChat
	if err := server_request.DecodeAndValidateRequest(r, &req); err != nil {
		rw.ErrorResponse(err, "failed to decode")
		return
	}

	chat, err := h.service.UpdateGroupChat(ctx, req, userID, chatID)
	if err != nil {
		rw.ErrorResponse(err, "update group chat")
		return
	}

	rw.JSONResponse(chat, http.StatusOK)
}

// DeleteChat godoc
// @Summary Delete chat
// @Tags chats
// @Security BearerAuth
// @Param id path string true "Chat ID"
// @Success 204 "No Content"
// @Failure 400 {object} server_response.ErrorResponse
// @Failure 403 {object} server_response.ErrorResponse
// @Router /chats/{id} [delete]
func (h *Handler) DeleteChat(w http.ResponseWriter, r *http.Request) {
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

	chatID, err := server_request.GetUUIDPathParam(r, "id")
	if err != nil {
		rw.ErrorResponse(err, "get path")
		return
	}

	if err := h.service.DeleteChat(ctx, userID, chatID); err != nil {
		rw.ErrorResponse(err, "delete chat")
		return
	}

	rw.NoContentResponse()
}
