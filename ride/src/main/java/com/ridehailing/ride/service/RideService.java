package com.ridehailing.ride.service;

import com.ridehailing.ride.dto.CreateRideRequest;
import com.ridehailing.ride.dto.RideResponse;
import com.ridehailing.ride.dto.UpdateRideStatusRequest;
import com.ridehailing.ride.entity.Ride;
import com.ridehailing.ride.entity.RideStatus;
import com.ridehailing.ride.repository.RideRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

/**
 * Service for managing ride lifecycle.
 * Handles ride creation, status updates, and publishes events to Redis Stream.
 */
@Service
public class RideService {
    private static final Logger logger = LoggerFactory.getLogger(RideService.class);

    @Autowired
    private RideRepository rideRepository;

    @Autowired
    private RideEventPublisher eventPublisher;

    /**
     * Create a new ride request
     */
    @Transactional
    public RideResponse createRide(CreateRideRequest request) {
        logger.info("Creating new ride for passenger: {}", request.getPassengerId());

        Ride ride = new Ride(
                request.getPassengerId(),
                request.getPickupLat(),
                request.getPickupLng(),
                request.getDropoffLat(),
                request.getDropoffLng());

        // Set estimated fare (simple calculation: 5 + 2 per km approximation)
        ride.setEstimatedFare(5.0 + Math.random() * 10.0);

        // Save to database
        Ride savedRide = rideRepository.save(ride);
        logger.info("Ride created with ID: {}", savedRide.getId());

        // Publish event to Redis Stream
        eventPublisher.publishRideRequested(
                savedRide.getId(),
                savedRide.getPassengerId(),
                savedRide.getPickupLat(),
                savedRide.getPickupLng(),
                savedRide.getDropoffLat(),
                savedRide.getDropoffLng());

        return convertToResponse(savedRide);
    }

    /**
     * Get ride by ID
     */
    public RideResponse getRideById(Long id) {
        Optional<Ride> ride = rideRepository.findById(id);
        if (ride.isEmpty()) {
            throw new IllegalArgumentException("Ride not found: " + id);
        }
        return convertToResponse(ride.get());
    }

    /**
     * Get all rides for a passenger
     */
    public List<RideResponse> getRidesByPassenger(String passengerId) {
        return rideRepository.findByPassengerId(passengerId)
                .stream()
                .map(this::convertToResponse)
                .collect(Collectors.toList());
    }

    /**
     * Get all rides for a driver
     */
    public List<RideResponse> getRidesByDriver(String driverId) {
        return rideRepository.findByDriverId(driverId)
                .stream()
                .map(this::convertToResponse)
                .collect(Collectors.toList());
    }

    /**
     * Update ride status and publish corresponding event
     */
    @Transactional
    public RideResponse updateRideStatus(Long rideId, UpdateRideStatusRequest request) {
        logger.info("Updating ride {} to status: {}", rideId, request.getStatus());

        Ride ride = rideRepository.findById(rideId)
                .orElseThrow(() -> new IllegalArgumentException("Ride not found: " + rideId));

        RideStatus oldStatus = ride.getStatus();
        ride.setStatus(request.getStatus());

        // Set driver if status is ASSIGNED
        if (request.getStatus() == RideStatus.ASSIGNED && request.getDriverId() != null) {
            ride.setDriverId(request.getDriverId());
        }

        // Track timestamps
        if (request.getStatus() == RideStatus.IN_PROGRESS) {
            ride.setStartedAt(LocalDateTime.now());
        } else if (request.getStatus() == RideStatus.COMPLETED) {
            ride.setCompletedAt(LocalDateTime.now());
        }

        Ride updatedRide = rideRepository.save(ride);
        logger.info("Ride {} updated from {} to {}", rideId, oldStatus, request.getStatus());

        // Publish event based on status
        publishStatusChangeEvent(updatedRide, request.getStatus());

        return convertToResponse(updatedRide);
    }

    /**
     * Publish event based on ride status change
     */
    private void publishStatusChangeEvent(Ride ride, RideStatus status) {
        switch (status) {
            case ASSIGNED ->
                eventPublisher.publishRideAssigned(
                        ride.getId(),
                        ride.getPassengerId(),
                        ride.getDriverId(),
                        ride.getPickupLat(),
                        ride.getPickupLng(),
                        ride.getDropoffLat(),
                        ride.getDropoffLng());
            case IN_PROGRESS ->
                eventPublisher.publishRideStarted(
                        ride.getId(),
                        ride.getPassengerId(),
                        ride.getDriverId(),
                        ride.getPickupLat(),
                        ride.getPickupLng(),
                        ride.getDropoffLat(),
                        ride.getDropoffLng());
            case COMPLETED ->
                eventPublisher.publishRideCompleted(
                        ride.getId(),
                        ride.getPassengerId(),
                        ride.getDriverId(),
                        ride.getPickupLat(),
                        ride.getPickupLng(),
                        ride.getDropoffLat(),
                        ride.getDropoffLng());
            case CANCELLED ->
                eventPublisher.publishRideCancelled(
                        ride.getId(),
                        ride.getPassengerId(),
                        ride.getDriverId(),
                        ride.getPickupLat(),
                        ride.getPickupLng(),
                        ride.getDropoffLat(),
                        ride.getDropoffLng());
            default -> logger.debug("No event published for status: {}", status);
        }
    }

    /**
     * Convert Ride entity to RideResponse DTO
     */
    private RideResponse convertToResponse(Ride ride) {
        RideResponse response = new RideResponse();
        response.setId(ride.getId());
        response.setPassengerId(ride.getPassengerId());
        response.setDriverId(ride.getDriverId());
        response.setStatus(ride.getStatus());
        response.setPickupLat(ride.getPickupLat());
        response.setPickupLng(ride.getPickupLng());
        response.setDropoffLat(ride.getDropoffLat());
        response.setDropoffLng(ride.getDropoffLng());
        response.setEstimatedFare(ride.getEstimatedFare());
        response.setCreatedAt(ride.getCreatedAt());
        response.setUpdatedAt(ride.getUpdatedAt());
        response.setStartedAt(ride.getStartedAt());
        response.setCompletedAt(ride.getCompletedAt());
        return response;
    }

    /**
     * Start a ride (transition from ASSIGNED to IN_PROGRESS)
     * Phase 4 Endpoint: POST /trips/{id}/start
     */
    @Transactional
    public RideResponse startRide(Long rideId) {
        logger.info("Starting ride: {}", rideId);

        Ride ride = rideRepository.findById(rideId)
                .orElseThrow(() -> new IllegalArgumentException("Ride not found: " + rideId));

        if (ride.getStatus() != RideStatus.ASSIGNED) {
            throw new IllegalArgumentException(
                    "Ride must be in ASSIGNED status to start. Current status: " + ride.getStatus());
        }

        ride.setStatus(RideStatus.IN_PROGRESS);
        ride.setStartedAt(LocalDateTime.now());

        Ride updatedRide = rideRepository.save(ride);
        logger.info("Ride {} started successfully", rideId);

        // Publish event
        eventPublisher.publishRideStarted(
                updatedRide.getId(),
                updatedRide.getPassengerId(),
                updatedRide.getDriverId(),
                updatedRide.getPickupLat(),
                updatedRide.getPickupLng(),
                updatedRide.getDropoffLat(),
                updatedRide.getDropoffLng());

        return convertToResponse(updatedRide);
    }

    /**
     * Complete a ride (transition from IN_PROGRESS to COMPLETED)
     * Phase 4 Endpoint: POST /trips/{id}/complete
     */
    @Transactional
    public RideResponse completeRide(Long rideId) {
        logger.info("Completing ride: {}", rideId);

        Ride ride = rideRepository.findById(rideId)
                .orElseThrow(() -> new IllegalArgumentException("Ride not found: " + rideId));

        if (ride.getStatus() != RideStatus.IN_PROGRESS) {
            throw new IllegalArgumentException(
                    "Ride must be in IN_PROGRESS status to complete. Current status: " + ride.getStatus());
        }

        ride.setStatus(RideStatus.COMPLETED);
        ride.setCompletedAt(LocalDateTime.now());

        Ride updatedRide = rideRepository.save(ride);
        logger.info("Ride {} completed successfully", rideId);

        // Publish event
        eventPublisher.publishRideCompleted(
                updatedRide.getId(),
                updatedRide.getPassengerId(),
                updatedRide.getDriverId(),
                updatedRide.getPickupLat(),
                updatedRide.getPickupLng(),
                updatedRide.getDropoffLat(),
                updatedRide.getDropoffLng());

        return convertToResponse(updatedRide);
    }

    /**
     * Cancel a ride (transition from REQUESTED or ASSIGNED to CANCELLED)
     * Phase 4 Endpoint: POST /trips/{id}/cancel
     */
    @Transactional
    public RideResponse cancelRide(Long rideId) {
        logger.info("Cancelling ride: {}", rideId);

        Ride ride = rideRepository.findById(rideId)
                .orElseThrow(() -> new IllegalArgumentException("Ride not found: " + rideId));

        // Can cancel from REQUESTED or ASSIGNED states
        if (ride.getStatus() == RideStatus.COMPLETED || ride.getStatus() == RideStatus.CANCELLED) {
            throw new IllegalArgumentException(
                    "Cannot cancel ride in " + ride.getStatus() + " status");
        }

        ride.setStatus(RideStatus.CANCELLED);
        ride.setCompletedAt(LocalDateTime.now());

        Ride updatedRide = rideRepository.save(ride);
        logger.info("Ride {} cancelled successfully", rideId);

        // Publish event
        eventPublisher.publishRideCancelled(
                updatedRide.getId(),
                updatedRide.getPassengerId(),
                updatedRide.getDriverId() != null ? updatedRide.getDriverId() : "null",
                updatedRide.getPickupLat(),
                updatedRide.getPickupLng(),
                updatedRide.getDropoffLat(),
                updatedRide.getDropoffLng());

        return convertToResponse(updatedRide);
    }
}
