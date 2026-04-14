// =============================================================================
// Application Entry Point — Notification Service
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

	"github.com/redis/go-redis/v9"
)

func main() {
	config := Config{
		Port:        getEnv("PORT", "8080"),
		ServiceName: getEnv("SERVICE_NAME", "notification-service"),
		Version:     getEnv("VERSION", "1.0.0"),
	}

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_URL", "localhost:6379"),
	})

	mux := setupRouter(config, rdb)

	// Wrap mux with metrics middleware
	handler := metricsMiddleware(mux)

	// Handle graceful shutdown
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
		<-sigch

		log.Printf("Shutting down %s...", config.ServiceName)
		rdb.Close()
		os.Exit(0)
	}()

	log.Printf("Starting %s on port %s", config.ServiceName, config.Port)
	if err := http.ListenAndServe(":"+config.Port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
