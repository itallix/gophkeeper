package main

import (
	"fmt"
	"os"

	"gophkeeper.com/internal/client"
)

func main() {
	if err := client.Execute(); err != nil {
		fmt.Printf("Error starting client: %v", err)
		os.Exit(1)
	}
}
