package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/wxvn/golang-messenger/internal/logger"
	"github.com/wxvn/golang-messenger/internal/server"
	server_response "github.com/wxvn/golang-messenger/internal/server/response"
	"go.uber.org/zap"
)

const RequestIDHeader = "X-Request-ID"

type contextKey string

const requestIDKey contextKey = "request_id"

func RequestID() server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get(RequestIDHeader)

			if reqID == "" {
				reqID = uuid.NewString()
			}

			ctx := context.WithValue(r.Context(), requestIDKey, reqID)

			w.Header().Set(RequestIDHeader, reqID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func requestIDFromContext(ctx context.Context) string {
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}

	return ""
}

func Logger(log *logger.Logger) server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := requestIDFromContext(r.Context())

			l := log.With(
				zap.String("request_id", reqID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)

			ctx := logger.ToContext(r.Context(), l)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Trace() server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.FromContext(r.Context())
			rw := server_response.NewResponseWriter(w)

			before := time.Now()

			log.Debug(
				">>> incoming HTTP request",
				zap.String("http_method", r.Method),
				zap.Time("time", before.UTC()),
			)

			next.ServeHTTP(rw, r)

			log.Debug(
				"<<< done HTTP request",
				zap.Int("status code", rw.GetStatusCode()),
				zap.Duration("latency", time.Since(before)),
			)
		})
	}
}

func Panic() server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := logger.FromContext(ctx)
			responseHandler := server_response.NewHTTPResponseHandler(log, w)

			defer func() {
				if p := recover(); p != nil {
					responseHandler.PanicResponse(p, "during handle Http request got nexpected panic")
				}
			}()

			next.ServeHTTP(w, r)
		})

	}
}
