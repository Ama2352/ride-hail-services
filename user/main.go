package main

import (
	"crypto/rsa"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var privateKey *rsa.PrivateKey

func init() {
	keyPath := filepath.Join("..", ".keys", "private.pem")
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		log.Fatalf("Failed to read private key at %s: %v", keyPath, err)
	}
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}
	log.Println("Successfully loaded private key from", keyPath)
}

type LoginRequest struct {
	UserID string `json:"user_id"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req LoginRequest
	if err := json.Unmarshal(body, &req); err != nil || req.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": req.UserID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		log.Printf("Failed to sign token: %v\n", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func main() {
	http.HandleFunc("/login", loginHandler)
	log.Println("Starting user service on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}
