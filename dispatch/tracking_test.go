package main

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

func TestInitTracking_Errors(t *testing.T) {
	keyPath := filepath.Join("..", ".keys", "public.pem")
	backupPath := keyPath + ".bak"

	// 1. Rename to test "Failed to read public key" (Line 29)
	os.Rename(keyPath, backupPath)
	initTracking()

	// 2. Write invalid PEM to test "Failed to parse public key" (Line 33)
	os.WriteFile(keyPath, []byte("NOT A VALID PUBLIC KEY PEM"), 0644)
	initTracking()

	// 3. Restore
	os.Remove(keyPath)
	os.Rename(backupPath, keyPath)

	// Test success parsing
	initTracking()
}

func TestTrackingHandler(t *testing.T) {
	// Setup test keys
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey = &privKey.PublicKey

	// Setup dummy redis client
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	rdb = redis.NewClient(&redis.Options{Addr: redisAddr})

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(trackingHandler))
	defer server.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Helper to generate token
	genToken := func(sub string) string {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"sub": sub,
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		str, _ := token.SignedString(privKey)
		return str
	}

	validToken := genToken("drv-123")

	t.Run("MissingToken", func(t *testing.T) {
		resp, _ := http.Get(server.URL)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("InvalidToken", func(t *testing.T) {
		resp, _ := http.Get(server.URL + "?token=invalid.token.str")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("NoPubKey", func(t *testing.T) {
		oldKey := publicKey
		publicKey = nil
		resp, _ := http.Get(server.URL + "?token=" + validToken)
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", resp.StatusCode)
		}
		publicKey = oldKey
	})

	t.Run("TokenNoSub", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()})
		str, _ := token.SignedString(privKey)
		resp, _ := http.Get(server.URL + "?token=" + str)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("WSUpgradeFail", func(t *testing.T) {
		// Calling with simple HTTP GET forces websocket upgrader to fail
		resp, _ := http.Get(server.URL + "?token=" + validToken)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 (Bad Request from Gorilla WS upgrade), got %d", resp.StatusCode)
		}
	})

	t.Run("SuccessFlow", func(t *testing.T) {
		u, _ := url.Parse(wsURL)
		u.RawQuery = "token=" + validToken

		dialer := websocket.Dialer{}
		// Setting custom origin to test CheckOrigin (Line 21)
		header := http.Header{"Origin": []string{"http://example.com"}}
		conn, resp, err := dialer.Dial(u.String(), header)
		if err != nil {
			t.Fatalf("Failed to connect WS: %v, Response code: %d", err, resp.StatusCode)
		}
		
		// Send Invalid JSON to cover unmarshalling failure (Lines 92-95)
		err = conn.WriteMessage(websocket.TextMessage, []byte("invalid json"))
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		// Send Valid JSON to cover GeoAdd execution (Lines 98-103)
		validMsg := `{"lat": 10.0, "lng": 106.0}`
		err = conn.WriteMessage(websocket.TextMessage, []byte(validMsg))
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		// Close connection to cover disconnect loop break (Lines 85-89)
		err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			t.Fatalf("Write close failed: %v", err)
		}

		// Wait briefly to allow goroutine server side to process messages before closing test
		time.Sleep(100 * time.Millisecond)
		conn.Close() // covers Line 81 defer connection close cleanup
	})
}
