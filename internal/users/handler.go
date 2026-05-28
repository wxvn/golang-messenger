package users

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
			Path:    "/users",
			Handler: h.GetUsers,
		},
		{
			Method:  http.MethodGet,
			Path:    "/users/{id}",
			Handler: h.GetUser,
		},
		{
			Method:  http.MethodGet,
			Path:    "/users/me",
			Handler: h.GetMe,
		},
		{
			Method:  http.MethodPatch,
			Path:    "/users/me",
			Handler: h.UpdateMe,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/users/me",
			Handler: h.DeleteMe,
		},
	}
}

type Handler struct {
	service *UserService
}

func NewUsersHandler(service *UserService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rh := server_response.NewHTTPResponseHandler(log, w)

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		rh.ErrorRespone(
			fmt.Errorf("user id not found in context"),
			"get user id from context",
		)
		return
	}

	var req PatchUser
	if err := server_request.DecodeAndValidateRequest(r, &req); err != nil {
		rh.ErrorRespone(err, "failed to decode")
		return
	}

	user, err := h.service.UpdateUser(ctx, userID, req)
	if err != nil {
		rh.ErrorRespone(err, "update user")
		return
	}

	rh.JSONResponse(user, http.StatusOK)
}

func (h *Handler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rh := server_response.NewHTTPResponseHandler(log, w)

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		rh.ErrorRespone(
			fmt.Errorf("user id not found in context"),
			"get user id from context",
		)
		return
	}

	if err := h.service.DeleteUser(ctx, userID); err != nil {
		rh.ErrorRespone(err, "delete user")
		return
	}

	rh.NoContentResponse()
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rh := server_response.NewHTTPResponseHandler(log, w)

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		rh.ErrorRespone(
			fmt.Errorf("user id not found in context"),
			"get user id from context",
		)
		return
	}

	user, err := h.service.GetUser(ctx, userID)
	if err != nil {
		rh.ErrorRespone(err, "get me")
		return
	}

	rh.JSONResponse(user, http.StatusOK)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rh := server_response.NewHTTPResponseHandler(log, w)

	userIDStr, err := server_request.GetPathValue(r, "id")
	if err != nil {
		rh.ErrorRespone(err, "get path")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		rh.ErrorRespone(err, "invalid user id")
		return
	}
	user, err := h.service.GetUser(ctx, userID)
	if err != nil {
		rh.ErrorRespone(err, "get user")
		return
	}

	rh.JSONResponse(user, http.StatusOK)
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rh := server_response.NewHTTPResponseHandler(log, w)

	username, limit, offset, err := getUsersQueryParams(r)
	if err != nil {
		rh.ErrorRespone(err, "failed to get username/limit/offset query params")
		return
	}

	users, err := h.service.GetUsers(ctx, username, limit, offset)
	if err != nil {
		rh.ErrorRespone(err, "failed to get users")
		return
	}

	rh.JSONResponse(users, http.StatusOK)
}

func getUsersQueryParams(r *http.Request) (*string, *int, *int, error) {
	const (
		usernameKey = "username"
		limitKey    = "limit"
		offsetKey   = "offset"
	)

	username, err := server_request.GetStringQueryParam(r, usernameKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get username: %w", err)
	}

	limit, err := server_request.GetIntQueryParam(r, limitKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get limit: %w", err)
	}

	offset, err := server_request.GetIntQueryParam(r, offsetKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get offset: %w", err)
	}

	return username, limit, offset, nil
}
