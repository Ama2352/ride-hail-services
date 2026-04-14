package com.ridehailing.ride.dto;

import com.ridehailing.ride.entity.RideStatus;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * Request DTO for updating ride status
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class UpdateRideStatusRequest {
    private RideStatus status;
    private String driverId; // Optional, for ASSIGNED status
}
