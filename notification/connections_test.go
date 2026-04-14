package main

import (
	"testing"
	"time"
)

// TestConnectionManagerStart tests the connection manager initialization
func TestConnectionManagerStart(t *testing.T) {
	cm := NewConnectionManager()
	cm.Start()

	// Should not panic
	time.Sleep(100 * time.Millisecond)
}

// TestRegisterClient tests client registration
func TestRegisterClient(t *testing.T) {
	cm := NewConnectionManager()
	cm.Start()

	conn := &ClientConnection{
		ID:       "test-conn-1",
		UserID:   "driver-123",
		UserType: "driver",
		RideIDs:  make(map[int64]bool),
		Send:     make(chan *WebSocketMessage, 10),
	}

	cm.register <- conn
	time.Sleep(50 * time.Millisecond)

	if len(cm.clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(cm.clients))
	}
}

// TestUnregisterClient tests client unregistration
func TestUnregisterClient(t *testing.T) {
	cm := NewConnectionManager()
	cm.Start()

	conn := &ClientConnection{
		ID:       "test-conn-1",
		UserID:   "driver-123",
		UserType: "driver",
		RideIDs:  make(map[int64]bool),
		Send:     make(chan *WebSocketMessage, 10),
	}

	cm.register <- conn
	time.Sleep(50 * time.Millisecond)

	cm.unregister <- conn
	time.Sleep(50 * time.Millisecond)

	if len(cm.clients) != 0 {
		t.Errorf("Expected 0 clients after unregister, got %d", len(cm.clients))
	}
}

// TestBroadcastMessage tests broadcasting to relevant clients
func TestBroadcastMessage(t *testing.T) {
	cm := NewConnectionManager()
	cm.Start()

	// Register a driver
	driverConn := &ClientConnection{
		ID:       "driver-conn",
		UserID:   "driver-1",
		UserType: "driver",
		RideIDs:  make(map[int64]bool),
		Send:     make(chan *WebSocketMessage, 10),
	}

	cm.register <- driverConn
	time.Sleep(100 * time.Millisecond)

	// Drain any initial connection acknowledgment
	select {
	case <-driverConn.Send:
		// Consume the first message (typically connection_ack)
	case <-time.After(100 * time.Millisecond):
		// No message to drain, continue
	}

	// Broadcast ride offered event
	msg := &WebSocketMessage{
		Type:   "ride_offered",
		RideID: 1001,
		Data: RideEvent{
			EventType:   "Ride.Offered",
			RideID:      1001,
			PassengerID: "passenger-1",
			DriverID:    "driver-1",
		},
		SentAt: time.Now(),
	}

	cm.broadcast <- msg
	time.Sleep(100 * time.Millisecond)

	// Check if message was received
	select {
	case received := <-driverConn.Send:
		if received.Type != "ride_offered" {
			t.Errorf("Expected 'ride_offered', got '%s'", received.Type)
		}
	case <-time.After(1 * time.Second):
		t.Error("Expected to receive broadcast message")
	}
}

// TestAddRideSubscription tests adding ride-specific subscriptions
func TestAddRideSubscription(t *testing.T) {
	cm := NewConnectionManager()
	cm.Start()

	conn := &ClientConnection{
		ID:       "test-conn-1",
		UserID:   "driver-123",
		UserType: "driver",
		RideIDs:  make(map[int64]bool),
		Send:     make(chan *WebSocketMessage, 10),
	}

	cm.register <- conn
	time.Sleep(50 * time.Millisecond)

	cm.AddRideSubscription("test-conn-1", 1001)
	time.Sleep(50 * time.Millisecond)

	// Verify subscription was added
	cm.mutex.RLock()
	if client, ok := cm.clients["test-conn-1"]; ok {
		if !client.RideIDs[1001] {
			t.Error("Ride 1001 should be in client's subscriptions")
		}
	} else {
		t.Error("Client not found")
	}
	cm.mutex.RUnlock()
}

// TestRemoveRideSubscription tests removing ride subscriptions
func TestRemoveRideSubscription(t *testing.T) {
	cm := NewConnectionManager()
	cm.Start()

	conn := &ClientConnection{
		ID:       "test-conn-1",
		UserID:   "driver-123",
		UserType: "driver",
		RideIDs:  make(map[int64]bool),
		Send:     make(chan *WebSocketMessage, 10),
	}
	conn.RideIDs[1001] = true

	cm.register <- conn
	time.Sleep(50 * time.Millisecond)

	cm.RemoveRideSubscription("test-conn-1", 1001)
	time.Sleep(50 * time.Millisecond)

	// Verify subscription was removed
	cm.mutex.RLock()
	if client, ok := cm.clients["test-conn-1"]; ok {
		if client.RideIDs[1001] {
			t.Error("Ride 1001 should not be in client's subscriptions")
		}
	}
	cm.mutex.RUnlock()
}

// TestGetConnectionCount tests retrieving connection count
func TestGetConnectionCount(t *testing.T) {
	cm := NewConnectionManager()
	cm.Start()

	if cm.GetConnectionCount() != 0 {
		t.Errorf("Expected 0 connections initially, got %d", cm.GetConnectionCount())
	}

	conn := &ClientConnection{
		ID:       "test-conn-1",
		UserID:   "driver-123",
		UserType: "driver",
		RideIDs:  make(map[int64]bool),
		Send:     make(chan *WebSocketMessage, 10),
	}

	cm.register <- conn
	time.Sleep(50 * time.Millisecond)

	if cm.GetConnectionCount() != 1 {
		t.Errorf("Expected 1 connection, got %d", cm.GetConnectionCount())
	}
}

// TestGetClientsByUser tests retrieving clients by user ID
func TestGetClientsByUser(t *testing.T) {
	cm := NewConnectionManager()
	cm.Start()

	conn := &ClientConnection{
		ID:       "test-conn-1",
		UserID:   "driver-123",
		UserType: "driver",
		RideIDs:  make(map[int64]bool),
		Send:     make(chan *WebSocketMessage, 10),
	}

	cm.register <- conn
	time.Sleep(50 * time.Millisecond)

	clients := cm.GetClientsByUser("driver-123")
	if len(clients) != 1 {
		t.Errorf("Expected 1 client for user, got %d", len(clients))
	}

	if clients[0].ID != "test-conn-1" {
		t.Errorf("Expected conn ID 'test-conn-1', got '%s'", clients[0].ID)
	}
}

// TestBroadcastRideAssigned tests assignment broadcast to specific parties
func TestBroadcastRideAssigned(t *testing.T) {
	cm := NewConnectionManager()
	cm.Start()

	// Register passenger subscribed to ride 1001
	passengerConn := &ClientConnection{
		ID:       "passenger-conn",
		UserID:   "passenger-1",
		UserType: "passenger",
		RideIDs:  make(map[int64]bool),
		Send:     make(chan *WebSocketMessage, 10),
	}
	passengerConn.RideIDs[1001] = true

	// Register driver subscribed to ride 1001
	driverConn := &ClientConnection{
		ID:       "driver-conn",
		UserID:   "driver-1",
		UserType: "driver",
		RideIDs:  make(map[int64]bool),
		Send:     make(chan *WebSocketMessage, 10),
	}
	driverConn.RideIDs[1001] = true

	cm.register <- passengerConn
	cm.register <- driverConn
	time.Sleep(50 * time.Millisecond)

	// Broadcast ride assigned
	msg := &WebSocketMessage{
		Type:   "ride_assigned",
		RideID: 1001,
		Data: RideEvent{
			EventType:   "Ride.Assigned",
			RideID:      1001,
			PassengerID: "passenger-1",
			DriverID:    "driver-1",
		},
		SentAt: time.Now(),
	}

	cm.broadcast <- msg
	time.Sleep(100 * time.Millisecond)

	// Both should receive
	select {
	case <-passengerConn.Send:
		// Received
	case <-time.After(500 * time.Millisecond):
		t.Error("Passenger should receive ride_assigned")
	}

	select {
	case <-driverConn.Send:
		// Received
	case <-time.After(500 * time.Millisecond):
		t.Error("Driver should receive ride_assigned")
	}
}

// TestIsSubscribedToRide tests ride subscription checking
func TestIsSubscribedToRide(t *testing.T) {
	// Test with no filters (all rides)
	conn1 := &ClientConnection{
		RideIDs: make(map[int64]bool),
	}
	if !isSubscribedToRide(conn1, 1001) {
		t.Error("Should receive all rides when filter is empty")
	}

	// Test with specific ride
	conn2 := &ClientConnection{
		RideIDs: make(map[int64]bool),
	}
	conn2.RideIDs[1001] = true
	if !isSubscribedToRide(conn2, 1001) {
		t.Error("Should receive ride 1001 when subscribed")
	}
	if isSubscribedToRide(conn2, 1002) {
		t.Error("Should not receive ride 1002 when not subscribed")
	}
}
