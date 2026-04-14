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

// Global variables for dispatcher and stream consumer
var (
	dispatcher      *Dispatcher
	streamConsumer  *StreamConsumer
	ctx             context.Context
	cancel          context.CancelFunc
)

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
		// Skip wrapping endpoints that rely on special ResponseWriter interfaces.
		if r.URL.Path == "/metrics" || r.URL.Path == "/ws" {
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

	initTracking()

	// Initialize dispatcher for waterfall dispatch algorithm
	dispatcher = NewDispatcher(rdb)
	streamConsumer = NewStreamConsumer(rdb, dispatcher)

	// Start stream consumer in background goroutine
	ctx, cancel = context.WithCancel(context.Background())
	go streamConsumer.StartConsuming(ctx)

	// Prometheus metrics endpoint (required for monitoring)
	mux.Handle("/metrics", promhttp.Handler())

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:    "healthy",
			Service:   config.ServiceName,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	// GPS tracking WebSocket endpoint
	mux.HandleFunc("/ws", trackingHandler)

	// Dispatch status endpoints (for demo/monitoring)
	mux.HandleFunc("/dispatch/active", func(w http.ResponseWriter, r *http.Request) {
		dispatcher.mutex.Lock()
		count := len(dispatcher.dispatches)
		dispatcher.mutex.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{
			"active_dispatches": count,
		})
	})

	// Driver acceptance endpoint
	mux.HandleFunc("/dispatch/rides/accept", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
			return
		}

		rideIDStr := r.URL.Query().Get("rideId")
		driverID := r.URL.Query().Get("driverId")

		if rideIDStr == "" || driverID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "rideId and driverId required"})
			return
		}

		var rideID int64
		_, err := fmt.Sscanf(rideIDStr, "%d", &rideID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid rideId format"})
			return
		}

		dispatcher.HandleDriverAcceptance(ctx, rideID, driverID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "accepted",
			"rideId": rideIDStr,
			"driverId": driverID,
		})
	})

	// Driver rejection endpoint
	mux.HandleFunc("/dispatch/rides/reject", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
			return
		}

		rideIDStr := r.URL.Query().Get("rideId")
		driverID := r.URL.Query().Get("driverId")

		if rideIDStr == "" || driverID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "rideId and driverId required"})
			return
		}

		var rideID int64
		_, err := fmt.Sscanf(rideIDStr, "%d", &rideID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid rideId format"})
			return
		}

		dispatcher.HandleDriverRejection(ctx, rideID, driverID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "rejected",
			"rideId": rideIDStr,
			"driverId": driverID,
		})
	})

	return mux
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
