package main

import (
	"log"

	"gophkeeper.com/internal/client"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	if err := client.Execute(buildVersion, buildDate, buildCommit); err != nil {
		log.Fatalf("Error starting client: %s", err)
	}
}
