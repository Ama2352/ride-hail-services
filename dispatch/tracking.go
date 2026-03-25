package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
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
	rdb       *redis.Client
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func initTracking() {
	keyPath := filepath.Join("..", ".keys", "public.pem")
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		log.Printf("Warning: Failed to read public key at %s, skipping JWT init for tests: %v", keyPath, err)
	} else {
		publicKey, err = jwt.ParseRSAPublicKeyFromPEM(keyData)
		if err != nil {
			log.Printf("Warning: Failed to parse public key: %v", err)
		}
	}

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	log.Println("Tracking module initialized")
}

type LocationUpdate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func trackingHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	if publicKey == nil {
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
		rdb.GeoAdd(ctx, "driver_locations", &redis.GeoLocation{
			Name:      userID,
			Longitude: loc.Lng,
			Latitude:  loc.Lat,
		})
	}
}
