package main

import (
	"log"

	"gophkeeper.com/internal/logger"
)

func main() {
	if err := logger.Initialize("debug"); err != nil {
		log.Fatalf("Cannot instantiate zap logger: %s", err)
	}
	logger.Log().Info("client")
}
