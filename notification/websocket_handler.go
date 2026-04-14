package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// websocketHandler handles WebSocket connections for real-time notifications
func websocketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create client connection
	clientID := generateConnectionID()
	client := &ClientConnection{
		ID:        clientID,
		Conn:      conn,
		Send:      make(chan *WebSocketMessage, 10),
		Close:     make(chan bool),
		Connected: time.Now(),
		RideIDs:   make(map[int64]bool),
	}

	// Register client
	connManager.register <- client

	// Start goroutines to handle reading/writing
	go handleClientRead(client, connManager)
	go handleClientWrite(client, connManager)
}

// handleClientRead reads messages from WebSocket client
func handleClientRead(client *ClientConnection, connMgr *ConnectionManager) {
	defer func() {
		connMgr.unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg ConnectionMessage
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		switch msg.Action {
		case "subscribe":
			handleSubscription(client, msg, connMgr)
		case "ping":
			// Respond with pong
			pong := &WebSocketMessage{
				Type:   "pong",
				SentAt: time.Now(),
			}
			client.Send <- pong
		default:
			log.Printf("Unknown action from client %s: %s", client.ID, msg.Action)
		}
	}
}

// handleClientWrite writes messages to WebSocket client
func handleClientWrite(client *ClientConnection, connMgr *ConnectionManager) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case msg := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := client.Conn.WriteJSON(msg)
			if err != nil {
				log.Printf("Error writing to client %s: %v", client.ID, err)
				return
			}

		case <-client.Close:
			log.Printf("Client %s closing connection", client.ID)
			return

		case <-ticker.C:
			// Send heartbeat ping
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := client.Conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				log.Printf("Error sending ping to client %s: %v", client.ID, err)
				return
			}
		}
	}
}

// handleSubscription processes subscription requests from clients
func handleSubscription(client *ClientConnection, msg ConnectionMessage, connMgr *ConnectionManager) {
	// Update client info
	client.UserID = msg.UserID
	client.UserType = msg.UserType

	// Set ride filter if provided
	if len(msg.RideIDs) > 0 {
		for _, rideID := range msg.RideIDs {
			client.RideIDs[rideID] = true
		}
		log.Printf("Client %s (user=%s, type=%s) subscribed to %d rides",
			client.ID, client.UserID, client.UserType, len(msg.RideIDs))
	} else {
		log.Printf("Client %s (user=%s, type=%s) subscribed to all events",
			client.ID, client.UserID, client.UserType)
	}

	// Re-register to pick up new filters
	connMgr.register <- client
}

// generateConnectionID creates a unique connection ID
func generateConnectionID() string {
	return strings.ToLower(
		time.Now().Format("20060102150405") +
		"-" + randomString(8),
	)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
