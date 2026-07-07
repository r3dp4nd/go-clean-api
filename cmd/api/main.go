package main

import (
	"log"

	"github.com/r3dp4nd/go-clean-api/internal/app"
	"github.com/r3dp4nd/go-clean-api/internal/config"
	applogger "github.com/r3dp4nd/go-clean-api/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	logger := applogger.New(cfg.Log)

	application := app.New(cfg, logger)

	if err := application.Run(); err != nil {
		logger.Error("error running application", "error", err)
	}
}
