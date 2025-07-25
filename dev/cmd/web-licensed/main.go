package main

import (
	"log"

	"isxcli/internal/app"
)

func main() {
	// Create application instance
	application, err := app.NewApplication()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Start application
	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}