// =============================================================================
// Application Entry Point — Dispatch Service
// This file contains only boilerplate bootstrapping code.
// All business logic is in server.go for better testability.
// =============================================================================

package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config := Config{
		Port:        getEnv("PORT", "8080"),
		ServiceName: getEnv("SERVICE_NAME", "dispatch-service"),
		Version:     getEnv("VERSION", "1.0.0"),
	}

	mux := setupRouter(config)

	// Handle graceful shutdown
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
		<-sigch

		log.Printf("Shutting down %s...", config.ServiceName)
		os.Exit(0)
	}()

	log.Printf("Starting %s on port %s", config.ServiceName, config.Port)
	if err := http.ListenAndServe(":"+config.Port, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
