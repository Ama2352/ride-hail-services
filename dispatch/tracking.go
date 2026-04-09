package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var (
	publicKey *rsa.PublicKey
	publicKeyPath string
	rdb       *redis.Client
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

const (
	defaultRedisAddrLocal = "localhost:6379"
	defaultRedisAddrK8s   = "redis-master.redis.svc.cluster.local:6379"
)

func resolvePublicKeyPath() string {
	if keyPath := os.Getenv("PUBLIC_KEY_PATH"); keyPath != "" {
		return keyPath
	}
	return filepath.Join("..", ".keys", "public.pem")
}

func resolveRedisAddr() string {
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		return redisAddr
	}

	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return defaultRedisAddrK8s
	}

	return defaultRedisAddrLocal
}

func tryLoadPublicKey() error {
	publicKeyPath = resolvePublicKeyPath()
	keyData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key at %s: %w", publicKeyPath, err)
	}
	parsedKey, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return fmt.Errorf("failed to parse public key at %s: %w", publicKeyPath, err)
	}
	publicKey = parsedKey
	return nil
}

func ensurePublicKey() error {
	if publicKey != nil {
		return nil
	}
	return tryLoadPublicKey()
}

func initTracking() {
	if err := tryLoadPublicKey(); err != nil {
		// Keep service alive for tests; handler will retry and log concrete failure.
		log.Printf("Warning: %v", err)
	}

	redisAddr := resolveRedisAddr()
	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	log.Printf("Tracking module initialized with Redis at %s", redisAddr)
}

type LocationUpdate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func storeDriverLocation(ctx context.Context, userID string, loc LocationUpdate) error {
	if rdb == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	return rdb.GeoAdd(ctx, "driver_locations", &redis.GeoLocation{
		Name:      userID,
		Longitude: loc.Lng,
		Latitude:  loc.Lat,
	}).Err()
}

func trackingHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	if err := ensurePublicKey(); err != nil {
		log.Printf("Auth key unavailable for /ws: %v", err)
		http.Error(w, "Server not configured for Auth", http.StatusInternalServerError)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userID, ok := token.Claims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		http.Error(w, "Invalid token subject", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WS Upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Printf("Driver %s connected to tracking socket", userID)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Driver %s disconnected", userID)
			break
		}

		var loc LocationUpdate
		if err := json.Unmarshal(msg, &loc); err != nil {
			log.Println("Invalid location JSON:", string(msg))
			continue
		}

		ctx := context.Background()
		if err := storeDriverLocation(ctx, userID, loc); err != nil {
			log.Printf("Failed to store location for driver %s: %v", userID, err)
		}
	}
}
