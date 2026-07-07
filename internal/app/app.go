package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/r3dp4nd/go-clean-api/internal/config"
	"github.com/r3dp4nd/go-clean-api/internal/server"
)

type App struct {
	config *config.Config
	logger *slog.Logger
}

func New(cfg *config.Config, logger *slog.Logger) *App {
	return &App{
		config: cfg,
		logger: logger,
	}
}

func (a *App) Run() error {
	a.logger.Info(
		"application starting",
		"app_name", a.config.App.Name,
		"app_version", a.config.App.Version,
		"environment", a.config.App.Environment,
		"shutdown_timeout_seconds", a.config.HTTP.ShutdownTimeoutSeconds,
	)

	httpServer := server.New(server.Options{
		Addr:              a.config.HTTP.Addr,
		ReadHeaderTimeout: a.config.HTTP.ReadHeaderTimeout,
		ReadTimeout:       a.config.HTTP.ReadTimeout,
		WriteTimeout:      a.config.HTTP.WriteTimeout,
		IdleTimeout:       a.config.HTTP.IdleTimeout,
		Logger:            a.logger,
	})

	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- httpServer.Start()
	}()

	shutdownContext, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	select {
	case err := <-serverErrors:
		return err

	case <-shutdownContext.Done():
		a.logger.Info("shutdown signal received")
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.config.HTTP.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		a.logger.Error("graceful shutdown failed", "error", err)
		return err
	}

	a.logger.Info("application stopped gracefully")

	return nil
}
