package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config holds service configuration
type Config struct {
	Port        string
	ServiceName string
	Version     string
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

// setupRouter creates and configures the HTTP router
func setupRouter(config Config) *http.ServeMux {
	mux := http.NewServeMux()

	// Prometheus metrics endpoint (legacy + namespaced)
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/dispatch/metrics", promhttp.Handler())

	// Health check endpoint handler (shared by legacy + namespaced routes)
	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:    "healthy service is running",
			Service:   config.ServiceName,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/dispatch/health", healthHandler)

	return mux
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
