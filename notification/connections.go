package main

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ClientConnection represents a connected WebSocket client
type ClientConnection struct {
	ID        string           // Unique connection ID
	UserID    string           // User ID (driver or passenger)
	UserType  string           // "driver" or "passenger"
	RideIDs   map[int64]bool   // Set of ride IDs to filter (if empty, all rides)
	Conn      *websocket.Conn
	Send      chan *WebSocketMessage // Messages to send to client
	Close     chan bool               // Signal to close connection
	Connected time.Time              // When connection was established
}

// ConnectionManager manages all active WebSocket connections
type ConnectionManager struct {
	clients      map[string]*ClientConnection // Map of connection ID -> client
	byUser       map[string][]*ClientConnection // Map of userID -> connections (one user can have multiple)
	byRide       map[int64][]*ClientConnection // Map of rideID -> connections
	register     chan *ClientConnection
	unregister   chan *ClientConnection
	broadcast    chan *WebSocketMessage
	mutex        sync.RWMutex
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		clients:    make(map[string]*ClientConnection),
		byUser:     make(map[string][]*ClientConnection),
		byRide:     make(map[int64][]*ClientConnection),
		register:   make(chan *ClientConnection, 10),
		unregister: make(chan *ClientConnection, 10),
		broadcast:  make(chan *WebSocketMessage, 100),
	}
}

// Start begins accepting connections and broadcasts
func (cm *ConnectionManager) Start() {
	go func() {
		for {
			select {
			case client := <-cm.register:
				cm.registerClient(client)
			case client := <-cm.unregister:
				cm.unregisterClient(client)
			case msg := <-cm.broadcast:
				cm.broadcastMessage(msg)
			}
		}
	}()
}

// registerClient adds a new connected client
func (cm *ConnectionManager) registerClient(client *ClientConnection) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.clients[client.ID] = client
	cm.byUser[client.UserID] = append(cm.byUser[client.UserID], client)

	// Index by ride IDs if filtering
	for rideID := range client.RideIDs {
		cm.byRide[rideID] = append(cm.byRide[rideID], client)
	}

	log.Printf("Client registered: ID=%s, UserID=%s, UserType=%s, RideFilter=%d",
		client.ID, client.UserID, client.UserType, len(client.RideIDs))

	// Send connection acknowledgement
	ack := &WebSocketMessage{
		Type:   "connection_ack",
		RideID: 0,
		Data: ConnectionAck{
			Status:    "connected",
			UserID:    client.UserID,
			Message:   "Connected to notification service",
			Timestamp: time.Now(),
		},
		SentAt: time.Now(),
	}
	client.Send <- ack
}

// unregisterClient removes a disconnected client
func (cm *ConnectionManager) unregisterClient(client *ClientConnection) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if _, exists := cm.clients[client.ID]; !exists {
		return
	}

	delete(cm.clients, client.ID)

	// Remove from user index
	if conns, ok := cm.byUser[client.UserID]; ok {
		for i, c := range conns {
			if c.ID == client.ID {
				cm.byUser[client.UserID] = append(conns[:i], conns[i+1:]...)
				break
			}
		}
		if len(cm.byUser[client.UserID]) == 0 {
			delete(cm.byUser, client.UserID)
		}
	}

	// Remove from ride indexes
	for rideID := range client.RideIDs {
		if conns, ok := cm.byRide[rideID]; ok {
			for i, c := range conns {
				if c.ID == client.ID {
					cm.byRide[rideID] = append(conns[:i], conns[i+1:]...)
					break
				}
			}
			if len(cm.byRide[rideID]) == 0 {
				delete(cm.byRide, rideID)
			}
		}
	}

	close(client.Send)
	log.Printf("Client unregistered: ID=%s, UserID=%s", client.ID, client.UserID)
}

// broadcastMessage sends a message to relevant clients
func (cm *ConnectionManager) broadcastMessage(msg *WebSocketMessage) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	recipients := make(map[string]*ClientConnection)

	// Find all relevant clients based on event type and ride
	switch msg.Type {
	case "ride_offered":
		// Drivers receive offer events (anyone subscribed to drivers or this ride)
		for _, client := range cm.clients {
			if client.UserType == "driver" {
				// Send to all drivers (they can accept/reject)
				recipients[client.ID] = client
			}
			// Also send if client explicitly subscribed to this ride
			if isSubscribedToRide(client, msg.RideID) {
				recipients[client.ID] = client
			}
		}

	case "ride_assigned":
		// Passenger and assigned driver get notified
		// Find the driver and passenger from clients
		for _, client := range cm.clients {
			if isSubscribedToRide(client, msg.RideID) {
				recipients[client.ID] = client
			}
		}

	case "ride_started", "ride_in_progress":
		// Passenger and driver both notified
		for _, client := range cm.clients {
			if isSubscribedToRide(client, msg.RideID) {
				recipients[client.ID] = client
			}
		}

	case "ride_completed":
		// Both parties notified, then subscription can be cleaned up
		for _, client := range cm.clients {
			if isSubscribedToRide(client, msg.RideID) {
				recipients[client.ID] = client
			}
		}

	case "ride_cancelled":
		// Both parties notified of cancellation
		for _, client := range cm.clients {
			if isSubscribedToRide(client, msg.RideID) {
				recipients[client.ID] = client
			}
		}

	default:
		log.Printf("Unknown message type: %s", msg.Type)
		return
	}

	// Send to all recipients
	for _, client := range recipients {
		select {
		case client.Send <- msg:
			// Message queued
		case <-time.After(1 * time.Second):
			// Client send buffer full, log warning but don't block
			log.Printf("Warning: client %s send buffer full for message type %s", client.ID, msg.Type)
		}
	}

	log.Printf("Broadcast %s to %d clients for ride %d", msg.Type, len(recipients), msg.RideID)
}

// AddRideSubscription adds a ride to client's filter
func (cm *ConnectionManager) AddRideSubscription(clientID string, rideID int64) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if client, ok := cm.clients[clientID]; ok {
		if client.RideIDs == nil {
			client.RideIDs = make(map[int64]bool)
		}
		client.RideIDs[rideID] = true
		cm.byRide[rideID] = append(cm.byRide[rideID], client)
		log.Printf("Client %s subscribed to ride %d", clientID, rideID)
	}
}

// RemoveRideSubscription removes a ride from client's filter
func (cm *ConnectionManager) RemoveRideSubscription(clientID string, rideID int64) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if client, ok := cm.clients[clientID]; ok {
		delete(client.RideIDs, rideID)

		// Remove from byRide index
		if conns, ok := cm.byRide[rideID]; ok {
			for i, c := range conns {
				if c.ID == clientID {
					cm.byRide[rideID] = append(conns[:i], conns[i+1:]...)
					break
				}
			}
			if len(cm.byRide[rideID]) == 0 {
				delete(cm.byRide, rideID)
			}
		}
	}
}

// GetConnectionCount returns number of active connections
func (cm *ConnectionManager) GetConnectionCount() int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return len(cm.clients)
}

// GetClientsByUser returns all connections for a user
func (cm *ConnectionManager) GetClientsByUser(userID string) []*ClientConnection {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	if conns, ok := cm.byUser[userID]; ok {
		result := make([]*ClientConnection, len(conns))
		copy(result, conns)
		return result
	}
	return nil
}

// isSubscribedToRide checks if client is subscribed to a ride
func isSubscribedToRide(client *ClientConnection, rideID int64) bool {
	if len(client.RideIDs) == 0 {
		// No ride filter = receive all events
		return true
	}
	_, subscribed := client.RideIDs[rideID]
	return subscribed
}
