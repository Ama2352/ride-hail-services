package main

import "time"

// RideRequestedEvent represents event from Ride Service
type RideRequestedEvent struct {
	EventType   string    `json:"event_type"`
	RideID      int64     `json:"ride_id"`
	PassengerID string    `json:"passenger_id"`
	PickupLat   float64   `json:"pickup_lat"`
	PickupLng   float64   `json:"pickup_lng"`
	DropoffLat  float64   `json:"dropoff_lat"`
	DropoffLng  float64   `json:"dropoff_lng"`
	Timestamp   time.Time `json:"timestamp"`
}

// RideOfferEvent represents Ride.Offered event to be published
type RideOfferEvent struct {
	EventType   string    `json:"event_type"`
	RideID      int64     `json:"ride_id"`
	PassengerID string    `json:"passenger_id"`
	DriverID    string    `json:"driver_id"`
	PickupLat   float64   `json:"pickup_lat"`
	PickupLng   float64   `json:"pickup_lng"`
	DropoffLat  float64   `json:"dropoff_lat"`
	DropoffLng  float64   `json:"dropoff_lng"`
	Timestamp   time.Time `json:"timestamp"`
}

// RideAssignedEvent represents Ride.Assigned event to be published
type RideAssignedEvent struct {
	EventType   string    `json:"event_type"`
	RideID      int64     `json:"ride_id"`
	PassengerID string    `json:"passenger_id"`
	DriverID    string    `json:"driver_id"`
	PickupLat   float64   `json:"pickup_lat"`
	PickupLng   float64   `json:"pickup_lng"`
	DropoffLat  float64   `json:"dropoff_lat"`
	DropoffLng  float64   `json:"dropoff_lng"`
	Timestamp   time.Time `json:"timestamp"`
}

// DispatchState tracks in-progress dispatch
type DispatchState struct {
	RideID       int64
	PassengerID  string
	CurrentIndex int              // Current driver index in cascade
	DriverIDs    []string         // List of candidate drivers
	Timeout      <-chan time.Time // Timeout for current offer
	StartTime    time.Time
}

// RideCancelledEvent represents Ride.Cancelled event
type RideCancelledEvent struct {
	EventType   string    `json:"event_type"`
	RideID      int64     `json:"ride_id"`
	PassengerID string    `json:"passenger_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// RideCompletedEvent represents Ride.Completed event
type RideCompletedEvent struct {
	EventType   string    `json:"event_type"`
	RideID      int64     `json:"ride_id"`
	PassengerID string    `json:"passenger_id"`
	DriverID    string    `json:"driver_id"`
	Timestamp   time.Time `json:"timestamp"`
}
