package main

import (
	"fmt"
	"os"

	"gophkeeper.com/internal/client"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	if err := client.Execute(buildVersion, buildDate, buildCommit); err != nil {
		fmt.Printf("Error starting client: %v", err)
		os.Exit(1)
	}
}
