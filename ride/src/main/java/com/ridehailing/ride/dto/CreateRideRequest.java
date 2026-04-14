package com.ridehailing.ride.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * Request DTO for creating a new ride
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class CreateRideRequest {
    private String passengerId;
    private Double pickupLat;
    private Double pickupLng;
    private Double dropoffLat;
    private Double dropoffLng;
}
