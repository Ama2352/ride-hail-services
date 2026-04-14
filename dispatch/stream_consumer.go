package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// StreamConsumer listens to Redis Stream for ride events
type StreamConsumer struct {
	rdb        *redis.Client
	dispatcher *Dispatcher
}

// NewStreamConsumer creates a new stream consumer
func NewStreamConsumer(rdb *redis.Client, dispatcher *Dispatcher) *StreamConsumer {
	return &StreamConsumer{
		rdb:        rdb,
		dispatcher: dispatcher,
	}
}

// StartConsuming begins listening to ride-events stream
func (sc *StreamConsumer) StartConsuming(ctx context.Context) {
	log.Println("Starting Redis Stream consumer...")

	lastID := "0" // Start from beginning of stream

	for {
		select {
		case <-ctx.Done():
			log.Println("Stream consumer stopped")
			return
		default:
		}

		// Read from stream with blocking
		streams, err := sc.rdb.XRead(ctx, &redis.XReadArgs{
			Streams: []string{"ride-events", lastID},
			Count:   1,
			Block:   5 * time.Second,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				// Timeout, no messages - continue waiting
				continue
			}
			log.Printf("Error reading from stream: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Process each message stream
		for _, stream := range streams {
			// Process all messages in this stream
			for _, msg := range stream.Messages {
				lastID = msg.ID

				// Extract event data from message values
				// msg.Values is a map[string]interface{}
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

				log.Printf("Received event from stream: %s (ID: %s)", eventType, msg.ID)

				// Parse payload and process based on event type
				if eventType == "Ride.Requested" {
					var event RideRequestedEvent
					if err := json.Unmarshal([]byte(payload), &event); err != nil {
						log.Printf("Error parsing Ride.Requested: %v", err)
						continue
					}
					sc.handleRideRequested(ctx, event)
				} else if eventType == "Ride.Cancelled" {
					var event RideCancelledEvent
					if err := json.Unmarshal([]byte(payload), &event); err != nil {
						log.Printf("Error parsing Ride.Cancelled: %v", err)
						continue
					}
					sc.handleRideCancelled(ctx, event)
				} else if eventType == "Ride.Completed" {
					var event RideCompletedEvent
					if err := json.Unmarshal([]byte(payload), &event); err != nil {
						log.Printf("Error parsing Ride.Completed: %v", err)
						continue
					}
					sc.handleRideCompleted(ctx, event)
				}
			}
		}
	}
}

// handleRideRequested processes Ride.Requested events
func (sc *StreamConsumer) handleRideRequested(ctx context.Context, event RideRequestedEvent) {
	log.Printf("Processing Ride.Requested: rideID=%d, passenger=%s, pickup=(%.4f,%.4f)", 
		event.RideID, event.PassengerID, event.PickupLat, event.PickupLng)

	// Dispatch ride to drivers
	sc.dispatcher.DispatchRide(ctx, &event)
}

// handleRideCancelled processes Ride.Cancelled events
// RideCancelledEvent needs to be added to events.go if not present
func (sc *StreamConsumer) handleRideCancelled(ctx context.Context, event RideCancelledEvent) {
	log.Printf("Processing Ride.Cancelled: rideID=%d", event.RideID)

	// Release the driver from the ride if one was assigned
	sc.dispatcher.mutex.Lock()
	if state, ok := sc.dispatcher.dispatches[event.RideID]; ok {
		if state.CurrentIndex < len(state.DriverIDs) {
			driverID := state.DriverIDs[state.CurrentIndex]
			sc.dispatcher.pool.ReleaseDriver(ctx, driverID)
		}
		delete(sc.dispatcher.dispatches, event.RideID)
	}
	sc.dispatcher.mutex.Unlock()
}

// handleRideCompleted processes Ride.Completed events
// RideCompletedEvent needs to be added to events.go if not present
func (sc *StreamConsumer) handleRideCompleted(ctx context.Context, event RideCompletedEvent) {
	log.Printf("Processing Ride.Completed: rideID=%d, driver=%s", event.RideID, event.DriverID)

	// Release the driver back to available pool
	sc.dispatcher.pool.ReleaseDriver(ctx, event.DriverID)

	// Clean up dispatch state
	sc.dispatcher.mutex.Lock()
	delete(sc.dispatcher.dispatches, event.RideID)
	sc.dispatcher.mutex.Unlock()
}
