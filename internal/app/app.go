package app

import (
	"log/slog"

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
	)

	httpServer := server.New(server.Options{
		Addr:              a.config.HTTP.Addr,
		ReadHeaderTimeout: a.config.HTTP.ReadHeaderTimeout,
		ReadTimeout:       a.config.HTTP.ReadTimeout,
		WriteTimeout:      a.config.HTTP.WriteTimeout,
		IdleTimeout:       a.config.HTTP.IdleTimeout,
		Logger:            a.logger,
	})

	return httpServer.Start()
}
