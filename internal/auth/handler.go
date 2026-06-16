package auth

import (
	"fmt"
	"net/http"

	"github.com/wxvn/golang-messenger/internal/logger"
	"github.com/wxvn/golang-messenger/internal/middleware"
	"github.com/wxvn/golang-messenger/internal/server"
	server_request "github.com/wxvn/golang-messenger/internal/server/request"
	server_response "github.com/wxvn/golang-messenger/internal/server/response"
)

func (h *Handler) Routes(authMW server.Middleware) []server.Route {
	return []server.Route{
		{Method: http.MethodPost, Path: "/auth/signup", Handler: h.SignUp},
		{Method: http.MethodPost, Path: "/auth/signin", Handler: h.SignIn},
		{Method: http.MethodPost, Path: "/auth/logout", Handler: h.Logout, Middlewares: []server.Middleware{authMW}},
		{Method: http.MethodPost, Path: "/auth/refresh", Handler: h.Refresh},
	}
}

type Handler struct {
	service *Service
}

func NewAuthHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

// SignUp godoc
// @Summary Register new user
// @Description Creates new account and returns user + tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SignRequest true "signup request"
// @Success 200 {object} SignResponse
// @Failure 400 {object} server_response.ErrorResponse
// @Failure 500 {object} server_response.ErrorResponse
// @Router /auth/signup [post]
func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rw := server_response.NewHTTPResponseHandler(log, w)

	var req SignRequest
	if err := server_request.DecodeAndValidateRequest(r, &req); err != nil {
		rw.ErrorResponse(err, "failed to decode")
		return
	}

	res, err := h.service.SignUp(r.Context(), req.Username, req.Password)
	if err != nil {
		rw.ErrorResponse(err, "failed to sign up")
		return
	}

	rw.JSONResponse(res, http.StatusOK)
}

// SignIn godoc
// @Summary Login user
// @Description Authenticates user and returns tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SignRequest true "signin request"
// @Success 200 {object} Tokens
// @Failure 400 {object} server_response.ErrorResponse
// @Failure 401 {object} server_response.ErrorResponse
// @Router /auth/signin [post]
func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rw := server_response.NewHTTPResponseHandler(log, w)

	var req SignRequest
	if err := server_request.DecodeAndValidateRequest(r, &req); err != nil {
		rw.ErrorResponse(err, "failed to decode")
		return
	}

	response, err := h.service.SignIn(ctx, req.Username, req.Password)
	if err != nil {
		rw.ErrorResponse(err, "failed to sign in")
		return
	}

	rw.JSONResponse(response, http.StatusOK)
}

// Logout godoc
// @Summary Logout user
// @Description Revokes refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requestRefreshToken true "refresh token"
// @Success 204 "No Content"
// @Failure 400 {object} server_response.ErrorResponse
// @Failure 401 {object} server_response.ErrorResponse
// @Router /auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
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

	var req requestRefreshToken
	if err := server_request.DecodeAndValidateRequest(r, &req); err != nil {
		rw.ErrorResponse(err, "failed to decode")
		return
	}

	if err := h.service.Logout(ctx, req.RefreshToken, userID); err != nil {
		rw.ErrorResponse(err, "failed to logout")
		return
	}

	rw.NoContentResponse()
}

// Refresh godoc
// @Summary Refresh access token
// @Description Generates new access + refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body Tokens true "refresh request"
// @Success 200 {object} Tokens
// @Failure 400 {object} server_response.ErrorResponse
// @Failure 401 {object} server_response.ErrorResponse
// @Router /auth/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rw := server_response.NewHTTPResponseHandler(log, w)

	var req Tokens
	if err := server_request.DecodeAndValidateRequest(r, &req); err != nil {
		rw.ErrorResponse(err, "failet to decode")
		return
	}

	tokens, err := h.service.Refresh(ctx, req)
	if err != nil {
		rw.ErrorResponse(err, "failet to refresh")
		return
	}

	rw.JSONResponse(tokens, http.StatusOK)
}
