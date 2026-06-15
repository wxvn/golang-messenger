package swagger

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/wxvn/golang-messenger/docs"
	"github.com/wxvn/golang-messenger/internal/server"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Routes() []server.Route {
	return []server.Route{
		{
			Method: http.MethodGet,
			Path:   "/swagger/",
			Handler: httpSwagger.Handler(
				httpSwagger.URL("/api/v1/swagger/doc.json"),
			),
		},
	}
}
