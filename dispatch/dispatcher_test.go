package main

import (
	"testing"
	"time"
)

// TestNewDispatcher verifies dispatcher initialization
func TestNewDispatcher(t *testing.T) {
	dispatcher := &Dispatcher{
		rdb:        nil,
		pool:       nil,
		dispatches: make(map[int64]*DispatchState),
	}

	if len(dispatcher.dispatches) != 0 {
		t.Error("Dispatches map should be empty on creation")
	}
}

// TestDispatchState validates dispatch state structure
func TestDispatchState(t *testing.T) {
	state := &DispatchState{
		RideID:       1001,
		PassengerID:  "passenger-1",
		DriverIDs:    []string{"driver-1", "driver-2", "driver-3"},
		CurrentIndex: 0,
		StartTime:    time.Now(),
	}

	if state.RideID != 1001 {
		t.Errorf("Expected RideID 1001, got %d", state.RideID)
	}

	if len(state.DriverIDs) != 3 {
		t.Errorf("Expected 3 drivers, got %d", len(state.DriverIDs))
	}

	if state.CurrentIndex != 0 {
		t.Errorf("Expected current index 0, got %d", state.CurrentIndex)
	}
}

// TestDispatchStateAddition tests adding dispatch state
func TestDispatchStateAddition(t *testing.T) {
	dispatcher := &Dispatcher{
		rdb:        nil,
		dispatches: make(map[int64]*DispatchState),
	}

	state := &DispatchState{
		RideID:       1001,
		PassengerID:  "passenger-1",
		DriverIDs:    []string{"driver-1", "driver-2"},
		CurrentIndex: 0,
		StartTime:    time.Now(),
	}

	dispatcher.mutex.Lock()
	dispatcher.dispatches[1001] = state
	dispatcher.mutex.Unlock()

	// Verify it was added
	dispatcher.mutex.Lock()
	stored, exists := dispatcher.dispatches[1001]
	dispatcher.mutex.Unlock()

	if !exists {
		t.Error("Dispatch state should exist for ride 1001")
	}

	if stored.RideID != 1001 {
		t.Errorf("Expected RideID 1001, got %d", stored.RideID)
	}
}

// TestDispatchStateRemoval tests removing dispatch state
func TestDispatchStateRemoval(t *testing.T) {
	dispatcher := &Dispatcher{
		rdb:        nil,
		dispatches: make(map[int64]*DispatchState),
	}

	state := &DispatchState{
		RideID:      1001,
		PassengerID: "passenger-1",
		DriverIDs:   []string{"driver-1", "driver-2"},
		StartTime:   time.Now(),
	}

	// Add it
	dispatcher.mutex.Lock()
	dispatcher.dispatches[1001] = state
	dispatcher.mutex.Unlock()

	// Verify it exists
	dispatcher.mutex.Lock()
	if _, exists := dispatcher.dispatches[1001]; !exists {
		t.Error("Dispatch state should exist for ride 1001")
	}
	dispatcher.mutex.Unlock()

	// Remove it
	dispatcher.mutex.Lock()
	delete(dispatcher.dispatches, 1001)
	dispatcher.mutex.Unlock()

	// Verify it's gone
	dispatcher.mutex.Lock()
	if _, exists := dispatcher.dispatches[1001]; exists {
		t.Error("Dispatch state should not exist for ride 1001 after deletion")
	}
	dispatcher.mutex.Unlock()
}

// TestCurrentIndexIncrement tests driver cascade logic
func TestCurrentIndexIncrement(t *testing.T) {
	state := &DispatchState{
		RideID:       2001,
		PassengerID:  "passenger-2",
		DriverIDs:    []string{"driver-1", "driver-2", "driver-3", "driver-4", "driver-5"},
		CurrentIndex: 0,
		StartTime:    time.Now(),
	}

	// Simulate moving through drivers on rejection
	for i := 0; i < 3; i++ {
		if state.CurrentIndex < len(state.DriverIDs) {
			state.CurrentIndex++
		}
	}

	if state.CurrentIndex != 3 {
		t.Errorf("Expected current index 3, got %d", state.CurrentIndex)
	}
}

// TestRideRequestedEventStruct validates event fields
func TestRideRequestedEventStruct(t *testing.T) {
	event := &RideRequestedEvent{
		EventType:   "Ride.Requested",
		RideID:      3001,
		PassengerID: "passenger-3",
		PickupLat:   10.7769,
		PickupLng:   106.7009,
		DropoffLat:  10.7890,
		DropoffLng:  106.7100,
		Timestamp:   time.Now(),
	}

	if event.EventType != "Ride.Requested" {
		t.Errorf("Expected event type 'Ride.Requested', got %s", event.EventType)
	}

	if event.RideID != 3001 {
		t.Errorf("Expected RideID 3001, got %d", event.RideID)
	}

	if event.PassengerID != "passenger-3" {
		t.Errorf("Expected PassengerID 'passenger-3', got %s", event.PassengerID)
	}

	if event.PickupLat != 10.7769 {
		t.Errorf("Expected PickupLat 10.7769, got %f", event.PickupLat)
	}
}

// TestRideAssignedEventStruct validates assigned event
func TestRideAssignedEventStruct(t *testing.T) {
	event := &RideAssignedEvent{
		EventType:   "Ride.Assigned",
		RideID:      4001,
		PassengerID: "passenger-4",
		DriverID:    "driver-100",
		PickupLat:   10.7769,
		PickupLng:   106.7009,
		Timestamp:   time.Now(),
	}

	if event.EventType != "Ride.Assigned" {
		t.Errorf("Expected event type 'Ride.Assigned', got %s", event.EventType)
	}

	if event.DriverID != "driver-100" {
		t.Errorf("Expected DriverID 'driver-100', got %s", event.DriverID)
	}
}

