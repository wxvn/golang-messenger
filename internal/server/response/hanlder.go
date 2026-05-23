package server_response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	errs "github.com/wxvn/golang-messenger/internal/errors"
	"github.com/wxvn/golang-messenger/internal/logger"
	"go.uber.org/zap"
)

type HTTPResponseHandler struct {
	log *logger.Logger
	rw  http.ResponseWriter
}

func NewHTTPResponseHandler(log *logger.Logger, w http.ResponseWriter) *HTTPResponseHandler {
	return &HTTPResponseHandler{
		log: log,
		rw:  w,
	}
}

func (h *HTTPResponseHandler) JSONResponse(responseBode any, statusCode int) {
	h.rw.WriteHeader(statusCode)

	if err := json.NewEncoder(h.rw).Encode(responseBode); err != nil {
		h.log.Error("Write HTTP response", zap.Error(err))
	}
}

func (h *HTTPResponseHandler) NoContentResponse() {
	h.rw.WriteHeader(http.StatusNoContent)
}

func (h *HTTPResponseHandler) ErrorRespone(err error, msg string) {
	var (
		statusCode int
		logFunc    func(string, ...zap.Field)
	)

	switch {
	case errors.Is(err, errs.ErrInvalidArgument):
		statusCode = http.StatusBadRequest
		logFunc = h.log.Warn

	case errors.Is(err, errs.ErrNotFound):
		statusCode = http.StatusNotFound
		logFunc = h.log.Debug

	case errors.Is(err, errs.ErrConflict):
		statusCode = http.StatusConflict
		logFunc = h.log.Warn

	default:
		statusCode = http.StatusInternalServerError
		logFunc = h.log.Error
	}

	logFunc(msg, zap.Error(err))

	h.errorRespone(statusCode, err, msg)
}

func (h *HTTPResponseHandler) PanicResponse(p any, msg string) {
	statusCode := http.StatusInternalServerError
	err := fmt.Errorf("unexpected panic : %v", p)

	h.log.Error(msg, zap.Error(err))

	h.errorRespone(statusCode, err, msg)

}

func (h *HTTPResponseHandler) errorRespone(statusCode int, err error, msg string) {

	response := map[string]string{
		"message": msg,
		"error":   err.Error(),
	}

	h.JSONResponse(
		response,
		statusCode,
	)
}
