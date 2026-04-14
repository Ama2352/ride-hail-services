package com.ridehailing.ride.entity;

/**
 * RideStatus Enum - Lifecycle states for a ride
 * 
 * Flow: REQUESTED → OFFERED → ASSIGNED → IN_PROGRESS → COMPLETED
 * REQUESTED → OFFERED → CANCELLED
 */
public enum RideStatus {
    REQUESTED, // Passenger has requested a ride
    OFFERED, // Ride is being offered to drivers
    ASSIGNED, // Driver has accepted the ride
    IN_PROGRESS, // Driver is en route / ride started
    COMPLETED, // Ride finished successfully
    CANCELLED // Ride cancelled by passenger or timeout
}
