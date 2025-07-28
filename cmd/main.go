package main

import (
	"270725/internal/config"
	v1 "270725/internal/rest/v1"
	"270725/internal/services"
	"270725/internal/storage/inmemory"
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()
	logger := setupLogger(cfg)

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := newServer(cfg, logger)
	go run(logger, server)

	logger.Info("starting server", slog.String("addr", server.Addr))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	ctx, cancel := context.WithTimeout(rootCtx, 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("failed to gracefully shutdown the server", slog.String("error", err.Error()))
	}
}

func run(logger *slog.Logger, server *http.Server) {
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("failed to listen and serve", slog.String("err", err.Error()))
	}
}

func newServer(cfg config.Config, logger *slog.Logger) *http.Server {
	repo := inmemory.NewMemory()
	logger.Info("starting repository")

	requester := services.NewRequesterService(int(cfg.TasksBufferSize * cfg.LinksInTask))
	archiver, err := services.NewZipper(cfg.ArchivesDir)
	if err != nil {
		panic(fmt.Errorf("failed to create archiver: %w", err))
	}

	taskService := services.NewTaskService(cfg, logger, repo, requester, archiver)
	logger.Info("starting task service")

	handler := v1.NewHandler(logger, taskService)

	e := echo.New()
	v1.RegisterHandler(e, handler)

	server := &http.Server{
		Handler:           e,
		Addr:              net.JoinHostPort(cfg.Host, cfg.Port),
		ReadHeaderTimeout: cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	return server
}

func setupLogger(cfg config.Config) *slog.Logger {
	var log *slog.Logger
	switch cfg.LogLevel {
	case config.LogLevelDebug:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case config.LogLevelInfo:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case config.LogLevelWarn:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	default:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	}
	return log
}
