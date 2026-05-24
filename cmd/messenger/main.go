package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/wxvn/golang-messenger/internal/auth"
	"github.com/wxvn/golang-messenger/internal/config"
	"github.com/wxvn/golang-messenger/internal/jwt"
	"github.com/wxvn/golang-messenger/internal/logger"
	"github.com/wxvn/golang-messenger/internal/middleware"
	"github.com/wxvn/golang-messenger/internal/postgres"
	"github.com/wxvn/golang-messenger/internal/server"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	cfg := config.Load()

	log, err := logger.NewLogger(cfg.Logger)
	if err != nil {
		slog.Error("logger init failed", "error", err)
		os.Exit(1)
	}
	defer log.Close()

	pool, err := postgres.NewPool(ctx, cfg.Postgres)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	tokenManager := jwt.NewTokenManager(cfg.JWTSecret)

	authRepo := auth.NewRepository(pool)
	authService := auth.NewService(authRepo, tokenManager)
	authDelivery := auth.NewDelivery(authService)

	httpServer := server.New(
		cfg.Addr,
		cfg.ShutdownTimeout,

		middleware.RequestID(),
		middleware.Logger(log),
		middleware.Trace(),
		middleware.Panic(),
	)

	httpServer.RegisterVersion("v1", authDelivery.Routes())

	if err := httpServer.Run(ctx); err != nil {
		slog.Error("http server error", "error", err)
		os.Exit(1)
	}
}
