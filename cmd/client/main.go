package main

import (
	"log"

	"github.com/itallix/gophkeeper/internal/logger"
)

func main() {
	if err := logger.Initialize("debug"); err != nil {
		log.Fatalf("Cannot instantiate zap logger: %s", err)
	}
	logger.Log().Info("client")
}
