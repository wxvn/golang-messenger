package users

import (
	"fmt"
	"net/http"

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

// UpdateMe godoc
// @Summary Update current user
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body PatchUser true "update user"
// @Success 200 {object} User
// @Failure 400 {object} server_response.ErrorResponse
// @Failure 401 {object} server_response.ErrorResponse
// @Failure 404 {object} server_response.ErrorResponse
// @Failure 409 {object} server_response.ErrorResponse
// @Router /users/me [patch]
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
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

	var req PatchUser
	if err := server_request.DecodeAndValidateRequest(r, &req); err != nil {
		rw.ErrorResponse(err, "failed to decode")
		return
	}

	user, err := h.service.UpdateUser(ctx, userID, req)
	if err != nil {
		rw.ErrorResponse(err, "update user")
		return
	}

	rw.JSONResponse(user, http.StatusOK)
}

// DeleteMe godoc
// @Summary Delete current user
// @Tags users
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 401 {object} server_response.ErrorResponse
// @Failure 404 {object} server_response.ErrorResponse
// @Router /users/me [delete]
func (h *Handler) DeleteMe(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.DeleteUser(ctx, userID); err != nil {
		rw.ErrorResponse(err, "delete user")
		return
	}

	rw.NoContentResponse()
}

// GetMe godoc
// @Summary Get current user
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} User
// @Failure 401 {object} server_response.ErrorResponse
// @Failure 404 {object} server_response.ErrorResponse
// @Router /users/me [get]
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.service.GetUser(ctx, userID)
	if err != nil {
		rw.ErrorResponse(err, "get me")
		return
	}

	rw.JSONResponse(user, http.StatusOK)
}

// GetUser godoc
// @Summary Get user by id
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path string true "user id"
// @Success 200 {object} User
// @Failure 400 {object} server_response.ErrorResponse
// @Failure 401 {object} server_response.ErrorResponse
// @Failure 404 {object} server_response.ErrorResponse
// @Router /users/{id} [get]
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rw := server_response.NewHTTPResponseHandler(log, w)

	userID, err := server_request.GetUUIDPathParam(r, "id")
	if err != nil {
		rw.ErrorResponse(err, "get path")
		return
	}

	user, err := h.service.GetUser(ctx, userID)
	if err != nil {
		rw.ErrorResponse(err, "get user")
		return
	}

	rw.JSONResponse(user, http.StatusOK)
}

// GetUsers godoc
// @Summary List users
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param username query string false "filter by username"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {array} User
// @Failure 400 {object} server_response.ErrorResponse
// @Failure 401 {object} server_response.ErrorResponse
// @Router /users [get]
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rw := server_response.NewHTTPResponseHandler(log, w)

	username, limit, offset, err := getUsersQueryParams(r)
	if err != nil {
		rw.ErrorResponse(err, "failed to get username/limit/offset query params")
		return
	}

	users, err := h.service.GetUsers(ctx, username, limit, offset)
	if err != nil {
		rw.ErrorResponse(err, "failed to get users")
		return
	}

	rw.JSONResponse(users, http.StatusOK)
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
