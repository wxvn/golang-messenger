package auth

import (
	"net/http"

	"github.com/wxvn/golang-messenger/internal/logger"
	"github.com/wxvn/golang-messenger/internal/server"
	server_request "github.com/wxvn/golang-messenger/internal/server/request"
	server_response "github.com/wxvn/golang-messenger/internal/server/response"
)

func (h *Handler) Routes() []server.Route {
	return []server.Route{
		{Method: http.MethodPost, Path: "/auth/signup", Handler: h.SignUp},
		{Method: http.MethodPost, Path: "/auth/signin", Handler: h.SignIn},
		{Method: http.MethodPost, Path: "/auth/loguot", Handler: h.Logout},
		{Method: http.MethodPost, Path: "/auth/refresh", Handler: h.Refresh},
	}
}

type Handler struct {
	service *Service
}

func NewDelivery(svc *Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rw := server_response.NewHTTPResponseHandler(log, w)

	var req SignUpRequest
	if err := server_request.DecodeAndValidateRequest(r, &req); err != nil {
		rw.ErrorRespone(err, "failed to decode")
		return
	}

	res, err := h.service.SignUp(r.Context(), req.Username, req.Password)
	if err != nil {
		rw.ErrorRespone(err, "failed to sign up")
		return
	}

	rw.JSONResponse(res, http.StatusOK)
}

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rw := server_response.NewHTTPResponseHandler(log, w)

	var req SignUpRequest
	if err := server_request.DecodeAndValidateRequest(r, &req); err != nil {
		rw.ErrorRespone(err, "failed to decode")
		return
	}

	response, err := h.service.SignIn(ctx, req.Username, req.Password)
	if err != nil {
		rw.ErrorRespone(err, "failed to sign in")
	}

	rw.JSONResponse(response, http.StatusOK)

}

type requestRefreshToken struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rw := server_response.NewHTTPResponseHandler(log, w)

	var req requestRefreshToken
	if err := server_request.DecodeAndValidateRequest(r, &req); err != nil {
		rw.ErrorRespone(err, "failed to decode")
		return
	}

	if err := h.service.Logout(ctx, req.RefreshToken); err != nil {
		rw.ErrorRespone(err, "failed to logout")
		return
	}

	rw.NoContentResponse()

}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rw := server_response.NewHTTPResponseHandler(log, w)

	var req Tokens
	if err := server_request.DecodeAndValidateRequest(r, &req); err != nil {
		rw.ErrorRespone(err, "failet to decode")
		return
	}

	tokens, err := h.service.Refresh(ctx, req)
	if err != nil {
		rw.ErrorRespone(err, "failet to refresh")
		return
	}

	rw.JSONResponse(tokens, http.StatusOK)

}
