package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
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

// =============================================================================
// Prometheus Metrics
// =============================================================================
var (
	// Counter: total HTTP requests by method, path, and status
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed",
		},
		[]string{"method", "path", "status"},
	)

	// Histogram: request duration in seconds
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Ensure metrics are only registered once
	metricsOnce sync.Once
)

// Global stream consumer and connection manager
var (
	streamConsumer   *StreamConsumer
	connManager      *ConnectionManager
	ctx              context.Context
	cancel           context.CancelFunc
)

// WebSocket upgrader configuration
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, validate the Origin header
		return true
	},
}

func init() {
	registerMetrics()
}

// registerMetrics safely registers all Prometheus metrics once
func registerMetrics() {
	metricsOnce.Do(func() {
		// Register custom metrics
		prometheus.MustRegister(httpRequestsTotal)
		prometheus.MustRegister(httpRequestDuration)
	})
}

// =============================================================================
// Middleware: Request Instrumentation
// =============================================================================

// statusWriter wraps ResponseWriter to capture the status code
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Hijack forwards websocket/connection upgrade support when available.
func (w *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response does not implement http.Hijacker")
	}
	return hijacker.Hijack()
}

// metricsMiddleware instruments HTTP requests with Prometheus metrics
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip instrumenting the metrics endpoint itself
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		wrapped := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		status := http.StatusText(wrapped.status)

		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

// setupRouter creates and configures the HTTP router
func setupRouter(config Config, rdb *redis.Client) *http.ServeMux {
	mux := http.NewServeMux()

	// Initialize connection manager and stream consumer
	connManager = NewConnectionManager()
	streamConsumer = NewStreamConsumer(rdb, connManager)

	// Create context for consumer lifecycle
	ctx, cancel = context.WithCancel(context.Background())

	// Start stream consumer in background
	go streamConsumer.StartConsuming(ctx)

	// Start connection manager
	connManager.Start()

	// Prometheus metrics endpoint (legacy + namespaced)
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/notification/metrics", promhttp.Handler())

	// Health check endpoint handler (shared by legacy + namespaced routes)
	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:    "healthy",
			Service:   config.ServiceName,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/notification/health", healthHandler)

	// WebSocket endpoint for notifications (legacy + namespaced)
	mux.HandleFunc("/notifications", websocketHandler)
	mux.HandleFunc("/notification/notifications", websocketHandler)
	mux.HandleFunc("/notification/ws", websocketHandler)

	// Stats endpoint handler (shared by legacy + namespaced routes)
	statsHandler := func(w http.ResponseWriter, r *http.Request) {
		stats := map[string]interface{}{
			"connected_clients": connManager.GetConnectionCount(),
			"timestamp":         time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
	mux.HandleFunc("/stats", statsHandler)
	mux.HandleFunc("/notification/stats", statsHandler)

	return mux
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
