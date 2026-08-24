// @title Messenger API
// @version 1.0
// @description Messenger backend
// @BasePath /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
//
// @security BearerAuth
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wxvn/golang-messenger/internal/auth"
	"github.com/wxvn/golang-messenger/internal/chats"
	"github.com/wxvn/golang-messenger/internal/config"
	"github.com/wxvn/golang-messenger/internal/jwt"
	"github.com/wxvn/golang-messenger/internal/logger"
	"github.com/wxvn/golang-messenger/internal/messages"
	"github.com/wxvn/golang-messenger/internal/middleware"
	"github.com/wxvn/golang-messenger/internal/postgres"
	redisclient "github.com/wxvn/golang-messenger/internal/redis"
	"github.com/wxvn/golang-messenger/internal/server"
	"github.com/wxvn/golang-messenger/internal/swagger"
	"github.com/wxvn/golang-messenger/internal/users"
	"github.com/wxvn/golang-messenger/internal/ws"
	"go.uber.org/zap"
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

	redisClient, err := redisclient.NewClient(ctx, cfg.Redis)
	if err != nil {
		slog.Error("redis connection failed", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	userCache := users.NewRedisUserCache(redisClient)

	hub := ws.NewHub()
	go hub.Run()

	tokenManager := jwt.NewTokenManager(cfg.JWTSecret, time.Minute*15)

	authRepo := auth.NewRepository(pool)
	authService := auth.NewService(authRepo, tokenManager)
	authDelivery := auth.NewAuthHandler(authService)

	usersRepo := users.NewUserRepository(pool)
	usersService := users.NewUserService(usersRepo, userCache)
	usersHandler := users.NewUsersHandler(usersService)

	chatsRepo := chats.NewChatsRepository(pool)
	chatsService := chats.NewChatsService(chatsRepo)
	chatsHandler := chats.NewChatsHandler(chatsService)

	messagesRepo := messages.NewMessagesRepository(pool)
	messagesService := messages.NewMessagesService(messagesRepo, chatsRepo)
	messagesHandler := messages.NewMessagesHandler(messagesService, hub)

	wsHandler := ws.NewHandler(hub, messagesService)

	swaggerHandler := swagger.NewHandler()

	httpServer := server.New(
		cfg.Addr,
		cfg.ShutdownTimeout,

		middleware.RequestID(),
		middleware.Logger(log),
		middleware.Trace(),
		middleware.Panic(),
	)
	authMW := middleware.Auth(tokenManager)

	httpServer.RegisterVersion("v1",
		server.RouteGroup{
			Routes: authDelivery.Routes(authMW),
		},
		server.RouteGroup{
			Middlewares: []server.Middleware{authMW},
			Routes:      usersHandler.Routes(),
		},
		server.RouteGroup{
			Middlewares: []server.Middleware{authMW},
			Routes:      chatsHandler.Routes(),
		},
		server.RouteGroup{
			Middlewares: []server.Middleware{authMW},
			Routes:      messagesHandler.Routes(),
		},
		server.RouteGroup{
			Middlewares: []server.Middleware{authMW},
			Routes:      wsHandler.Routes(),
		},
		server.RouteGroup{
			Routes: swaggerHandler.Routes(),
		},
	)

	if err := httpServer.Run(ctx); err != nil {
		log.Error("http server error", zap.Error(err))
		os.Exit(1)
	}
}
