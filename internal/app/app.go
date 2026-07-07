package app

import (
	"log"

	"github.com/r3dp4nd/go-clean-api/internal/config"
	"github.com/r3dp4nd/go-clean-api/internal/server"
)

type App struct {
	config *config.Config
}

func New(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

func (a *App) Run() error {
	log.Printf(
		"%s %s starting in %s mode",
		a.config.App.Name,
		a.config.App.Version,
		a.config.App.Environment,
	)

	httpServer := server.New(server.Options{
		Addr:              a.config.HTTP.Addr,
		ReadHeaderTimeout: a.config.HTTP.ReadHeaderTimeout,
		ReadTimeout:       a.config.HTTP.ReadTimeout,
		WriteTimeout:      a.config.HTTP.WriteTimeout,
		IdleTimeout:       a.config.HTTP.IdleTimeout,
	})

	return httpServer.Start()
}
