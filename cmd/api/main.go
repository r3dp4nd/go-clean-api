package main

import (
	"log"

	"github.com/r3dp4nd/go-clean-api/internal/app"
)

func main() {
	application := app.New("go-clean-api", "v0.1.0", ":8080")

	if err := application.Run(); err != nil {
		log.Fatalf("error running application: %v", err)
	}
}
