package main

import "time"

// RideEvent is the base notification event
type RideEvent struct {
	EventType   string    `json:"event_type"`
	RideID      int64     `json:"ride_id"`
	PassengerID string    `json:"passenger_id"`
	DriverID    string    `json:"driver_id,omitempty"`
	PickupLat   float64   `json:"pickup_lat,omitempty"`
	PickupLng   float64   `json:"pickup_lng,omitempty"`
	DropoffLat  float64   `json:"dropoff_lat,omitempty"`
	DropoffLng  float64   `json:"dropoff_lng,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// WebSocketMessage represents a message pushed to clients via WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`      // Event type: ride_offered, ride_assigned, ride_started, etc.
	RideID  int64       `json:"ride_id"`
	Data    interface{} `json:"data"`      // Full event data
	SentAt  time.Time   `json:"sent_at"`
}

// NotificationSubscriber represents a client that wants to receive notifications
type NotificationSubscriber struct {
	UserID    string   // Unique user ID (driver or passenger)
	UserType  string   // "driver" or "passenger"
	RideIDs   []int64  // Specific ride IDs to filter (if empty, receive all)
}

// ConnectionMessage is sent over WebSocket to establish subscription
type ConnectionMessage struct {
	Action    string   `json:"action"`    // "subscribe" or "ping"
	UserID    string   `json:"user_id"`
	UserType  string   `json:"user_type"` // "driver" or "passenger"
	RideIDs   []int64  `json:"ride_ids"`  // Optional: specific rides to filter
}

// ConnectionAck is sent back to confirm subscription
type ConnectionAck struct {
	Status    string    `json:"status"`
	UserID    string    `json:"user_id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
