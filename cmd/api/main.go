package main

import (
	"log"

	"github.com/r3dp4nd/go-clean-api/internal/app"
	"github.com/r3dp4nd/go-clean-api/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	application := app.New(cfg)

	if err := application.Run(); err != nil {
		log.Fatalf("error running application: %v", err)
	}
}
