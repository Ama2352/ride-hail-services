package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Dispatcher manages the waterfall dispatch algorithm
type Dispatcher struct {
	rdb        *redis.Client
	pool       *DriverPool
	dispatches map[int64]*DispatchState // Active dispatch states
	mutex      sync.Mutex
}

// NewDispatcher creates a new dispatcher
func NewDispatcher(rdb *redis.Client) *Dispatcher {
	return &Dispatcher{
		rdb:        rdb,
		pool:       NewDriverPool(rdb),
		dispatches: make(map[int64]*DispatchState),
	}
}

// DispatchRide implements waterfall algorithm:
// 1. Find 5 nearest drivers
// 2. Offer to closest driver with 10-second timeout
// 3. If rejected/timeout, move to next driver
// 4. Continue until driver accepts or all drivers exhausted
func (d *Dispatcher) DispatchRide(ctx context.Context, ride *RideRequestedEvent) {
	log.Printf("Starting dispatch for ride %d from passenger %s", ride.RideID, ride.PassengerID)

	// Find 5 nearest drivers
	driverIDs, err := d.pool.FindNearbyDrivers(ctx, ride.PickupLat, ride.PickupLng, 5)
	if err != nil {
		log.Printf("Error finding nearby drivers: %v", err)
		return
	}

	if len(driverIDs) == 0 {
		log.Printf("No drivers available for ride %d", ride.RideID)
		return
	}

	// Store dispatch state
	state := &DispatchState{
		RideID:      ride.RideID,
		PassengerID: ride.PassengerID,
		DriverIDs:   driverIDs,
		StartTime:   time.Now(),
	}

	d.mutex.Lock()
	d.dispatches[ride.RideID] = state
	d.mutex.Unlock()

	// Start waterfall dispatch
	go d.waterfallDispatch(ctx, ride, state)
}

// waterfallDispatch implements the cascade logic with timeouts
func (d *Dispatcher) waterfallDispatch(ctx context.Context, ride *RideRequestedEvent, state *DispatchState) {
	timeoutDuration := 10 * time.Second

	for state.CurrentIndex < len(state.DriverIDs) {
		driverID := state.DriverIDs[state.CurrentIndex]

		// Check if driver is not busy
		busy, err := d.pool.IsDriverBusy(ctx, driverID)
		if err != nil || busy {
			log.Printf("Driver %s is busy, moving to next", driverID)
			state.CurrentIndex++
			continue
		}

		// Publish Ride.Offered event
		event := RideOfferEvent{
			EventType:   "Ride.Offered",
			RideID:      ride.RideID,
			PassengerID: ride.PassengerID,
			DriverID:    driverID,
			PickupLat:   ride.PickupLat,
			PickupLng:   ride.PickupLng,
			DropoffLat:  ride.DropoffLat,
			DropoffLng:  ride.DropoffLng,
			Timestamp:   time.Now(),
		}

		eventJSON, _ := json.Marshal(event)
		log.Printf("Publishing Ride.Offered for ride %d to driver %s: %s", ride.RideID, driverID, string(eventJSON))
		d.publishEvent("ride-events", "Ride.Offered", eventJSON)

		// Wait for driver response or timeout
		done := make(chan bool)
		go d.waitForDriverResponse(ctx, ride.RideID, driverID, done)

		// Timeout + response handling
		select {
		case accepted := <-done:
			if accepted {
				// Driver accepted! Publish Ride.Assigned event
				assignEvent := RideAssignedEvent{
					EventType:   "Ride.Assigned",
					RideID:      ride.RideID,
					PassengerID: ride.PassengerID,
					DriverID:    driverID,
					PickupLat:   ride.PickupLat,
					PickupLng:   ride.PickupLng,
					DropoffLat:  ride.DropoffLat,
					DropoffLng:  ride.DropoffLng,
					Timestamp:   time.Now(),
				}
				assignJSON, _ := json.Marshal(assignEvent)
				log.Printf("Publishing Ride.Assigned for ride %d to driver %s", ride.RideID, driverID)
				d.publishEvent("ride-events", "Ride.Assigned", assignJSON)

				// Mark driver as busy
				d.pool.BusyDriver(ctx, driverID)

				// Clean up dispatch state
				d.mutex.Lock()
				delete(d.dispatches, ride.RideID)
				d.mutex.Unlock()
				return
			}
		case <-time.After(timeoutDuration):
			log.Printf("Timeout for driver %s on ride %d, moving to next", driverID, ride.RideID)
		}

		// Move to next driver
		state.CurrentIndex++
	}

	log.Printf("All drivers rejected or unavailable for ride %d", ride.RideID)
	d.mutex.Lock()
	delete(d.dispatches, ride.RideID)
	d.mutex.Unlock()
}

// waitForDriverResponse polls Redis for driver acceptance (placeholder)
// In real system, this would be via WebSocket or message queue
func (d *Dispatcher) waitForDriverResponse(ctx context.Context, rideID int64, driverID string, done chan bool) {
	key := "ride:" + string(rune(rideID)) + ":response:" + driverID
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if driver responded
			response, _ := d.rdb.Get(ctx, key).Result()
			if response == "accepted" {
				d.rdb.Del(ctx, key)
				done <- true
				return
			} else if response == "rejected" {
				d.rdb.Del(ctx, key)
				done <- false
				return
			}
		case <-ctx.Done():
			done <- false
			return
		case <-time.After(15 * time.Second):
			done <- false
			return
		}
	}
}

// publishEvent publishes event to Redis Stream
func (d *Dispatcher) publishEvent(streamKey, eventType string, eventJSON []byte) error {
	ctx := context.Background()
	return d.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		MaxLen: 10000,
		Values: []interface{}{
			"event_type", eventType,
			"payload", string(eventJSON),
			"timestamp", time.Now().Unix(),
		},
	}).Err()
}

// HandleDriverAcceptance is called when driver accepts ride (via API/WebSocket)
func (d *Dispatcher) HandleDriverAcceptance(ctx context.Context, rideID int64, driverID string) {
	key := "ride:" + string(rune(rideID)) + ":response:" + driverID
	d.rdb.Set(ctx, key, "accepted", 15*time.Second)
}

// HandleDriverRejection is called when driver rejects ride
func (d *Dispatcher) HandleDriverRejection(ctx context.Context, rideID int64, driverID string) {
	key := "ride:" + string(rune(rideID)) + ":response:" + driverID
	d.rdb.Set(ctx, key, "rejected", 15*time.Second)
}
