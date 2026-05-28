package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

type RouteGroup struct {
	Middlewares []Middleware
	Routes      []Route
}

type Route struct {
	Method      string
	Path        string
	Handler     http.HandlerFunc
	Middlewares []Middleware
}

type HTTPServer struct {
	mux        *http.ServeMux
	addr       string
	timeout    time.Duration
	middleware []Middleware
}

func New(addr string, timeout time.Duration, mws ...Middleware) *HTTPServer {
	return &HTTPServer{
		mux:        http.NewServeMux(),
		addr:       addr,
		timeout:    timeout,
		middleware: mws,
	}
}

func (s *HTTPServer) RegisterVersion(version string, groups ...RouteGroup) {
	prefix := "/api/" + version
	versionMux := http.NewServeMux()

	for _, group := range groups {
		for _, r := range group.Routes {

			pattern := fmt.Sprintf("%s %s", r.Method, r.Path)
			handler := http.Handler(r.Handler)

			// route middleware
			for i := len(r.Middlewares) - 1; i >= 0; i-- {
				handler = r.Middlewares[i](handler)
			}

			// group middleware
			for i := len(group.Middlewares) - 1; i >= 0; i-- {
				handler = group.Middlewares[i](handler)
			}

			versionMux.Handle(pattern, handler)
		}
	}

	s.mux.Handle(prefix+"/", http.StripPrefix(prefix, versionMux))
}

func (s *HTTPServer) Run(ctx context.Context) error {
	var handler http.Handler = s.mux
	for i := len(s.middleware) - 1; i >= 0; i-- {
		handler = s.middleware[i](handler)
	}

	srv := &http.Server{
		Addr:    s.addr,
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("Starting HTTP server", "addr", s.addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		slog.Info("Shutting down HTTP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.timeout)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
