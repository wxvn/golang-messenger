package server_response

import "net/http"

type ResponseWriter struct {
	http.ResponseWriter
	stausCode int
}

var (
	StatusCodeUninitialized = -1
)

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		stausCode:      StatusCodeUninitialized,
	}
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.stausCode = statusCode
}

func (rw *ResponseWriter) GetStatusCode() int {
	if rw.stausCode == StatusCodeUninitialized {
		return http.StatusOK
	}
	return rw.stausCode
}
