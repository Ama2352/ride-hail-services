package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// StreamConsumer listens to Redis Stream for ride events and broadcasts via WebSocket
type StreamConsumer struct {
	rdb                *redis.Client
	connManager        *ConnectionManager
	consumerGroupName  string
}

// NewStreamConsumer creates a new stream consumer
func NewStreamConsumer(rdb *redis.Client, connManager *ConnectionManager) *StreamConsumer {
	return &StreamConsumer{
		rdb:               rdb,
		connManager:       connManager,
		consumerGroupName: "notification-service",
	}
}

// StartConsuming begins listening to ride-events stream and broadcasting events
func (sc *StreamConsumer) StartConsuming(ctx context.Context) {
	log.Println("Starting Redis Stream consumer for notifications...")

	// Initialize consumer group (idempotent, won't fail if exists)
	err := sc.rdb.XGroupCreateMkStream(ctx, "ride-events", sc.consumerGroupName, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer already exists" {
		// BUSYGROUP is expected if group already exists, other errors are problems
		if err.Error() != "BUSYGROUP Consumer already exists" {
			log.Printf("Info: creating consumer group: %v", err)
		}
	}

	lastProcessed := "0"
	
	for {
		select {
		case <-ctx.Done():
			log.Println("Stream consumer stopped")
			return
		default:
		}

		// Read messages from the stream
		streams, err := sc.rdb.XRead(ctx, &redis.XReadArgs{
			Streams: []string{"ride-events", lastProcessed},
			Count:   10,
			Block:   5 * time.Second,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				// Timeout during BLOCK - continue waiting
				continue
			}
			log.Printf("Error reading from stream: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Process each message
		for _, stream := range streams {
			for _, msg := range stream.Messages {
				lastProcessed = msg.ID

				// Extract event data
				eventType := ""
				payload := ""

				if val, ok := msg.Values["event_type"]; ok {
					eventType = val.(string)
				}
				if val, ok := msg.Values["payload"]; ok {
					payload = val.(string)
				}

				if eventType == "" {
					log.Printf("Invalid event_type in message %s", msg.ID)
					continue
				}

				log.Printf("Consuming event from stream: %s (ID: %s)", eventType, msg.ID)

				// Parse payload and broadcast to WebSocket clients
				if payload != "" {
					var event RideEvent
					if err := json.Unmarshal([]byte(payload), &event); err != nil {
						log.Printf("Error parsing event payload: %v", err)
						continue
					}

					// Determine WebSocket message type based on event type
					wsType := sc.mapEventTypeToWSType(eventType)

					// Broadcast to subscribers
					wsMsg := &WebSocketMessage{
						Type:   wsType,
						RideID: event.RideID,
						Data:   event,
						SentAt: time.Now(),
					}

					// Non-blocking send to broadcast channel
					select {
					case sc.connManager.broadcast <- wsMsg:
						log.Printf("Broadcasted %s (rideID=%d) to WebSocket clients", wsType, event.RideID)
					case <-time.After(100 * time.Millisecond):
						log.Printf("Warning: broadcast channel full for event %s", eventType)
					}
				}
			}
		}
	}
}

// mapEventTypeToWSType converts Redis event type to WebSocket message type
func (sc *StreamConsumer) mapEventTypeToWSType(eventType string) string {
	switch eventType {
	case "Ride.Requested":
		return "ride_requested"
	case "Ride.Offered":
		return "ride_offered"
	case "Ride.Assigned":
		return "ride_assigned"
	case "Ride.Started":
		return "ride_started"
	case "Ride.InProgress":
		return "ride_in_progress"
	case "Ride.Completed":
		return "ride_completed"
	case "Ride.Cancelled":
		return "ride_cancelled"
	default:
		log.Printf("Unknown event type: %s", eventType)
		return "unknown_event"
	}
}
