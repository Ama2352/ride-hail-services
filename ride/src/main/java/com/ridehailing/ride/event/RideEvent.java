package com.ridehailing.ride.event;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import java.time.LocalDateTime;

/**
 * Event model for Redis Stream publishing.
 * All ride lifecycle events follow this structure.
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class RideEvent {
    @JsonProperty("event_type")
    private String eventType;

    @JsonProperty("ride_id")
    private Long rideId;

    @JsonProperty("passenger_id")
    private String passengerId;

    @JsonProperty("driver_id")
    private String driverId;

    @JsonProperty("pickup_lat")
    private Double pickupLat;

    @JsonProperty("pickup_lng")
    private Double pickupLng;

    @JsonProperty("dropoff_lat")
    private Double dropoffLat;

    @JsonProperty("dropoff_lng")
    private Double dropoffLng;

    @JsonProperty("timestamp")
    private LocalDateTime timestamp;
}
