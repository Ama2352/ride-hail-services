package com.ridehailing.ride.dto;

import com.ridehailing.ride.entity.RideStatus;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import java.time.LocalDateTime;

/**
 * Response DTO for ride queries and creation
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class RideResponse {
    private Long id;
    private String passengerId;
    private String driverId;
    private RideStatus status;
    private Double pickupLat;
    private Double pickupLng;
    private Double dropoffLat;
    private Double dropoffLng;
    private Double estimatedFare;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
    private LocalDateTime startedAt;
    private LocalDateTime completedAt;
}
