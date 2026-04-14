package main.java.com.ridehailing.ride.controller;

import com.ridehailing.ride.dto.CreateRideRequest;
import com.ridehailing.ride.dto.RideResponse;
import com.ridehailing.ride.dto.UpdateRideStatusRequest;
import com.ridehailing.ride.service.RideService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * REST Controller for Ride Service.
 * Exposes endpoints for managing rides.
 */
@RestController
@RequestMapping("/rides")
public class RideController {
    private static final Logger logger = LoggerFactory.getLogger(RideController.class);

    @Autowired
    private RideService rideService;

    /**
     * POST /rides - Create a new ride request
     */
    @PostMapping
    public ResponseEntity<RideResponse> createRide(@RequestBody CreateRideRequest request) {
        logger.info("POST /rides - Creating new ride for passenger: {}", request.getPassengerId());
        try {
            RideResponse response = rideService.createRide(request);
            return ResponseEntity.status(HttpStatus.CREATED).body(response);
        } catch (Exception e) {
            logger.error("Error creating ride: {}", e.getMessage(), e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    /**
     * GET /rides/{id} - Get ride by ID
     */
    @GetMapping("/{id}")
    public ResponseEntity<RideResponse> getRideById(@PathVariable Long id) {
        logger.info("GET /rides/{} - Fetching ride details", id);
        try {
            RideResponse response = rideService.getRideById(id);
            return ResponseEntity.ok(response);
        } catch (IllegalArgumentException e) {
            logger.warn("Ride not found: {}", id);
            return ResponseEntity.status(HttpStatus.NOT_FOUND).build();
        } catch (Exception e) {
            logger.error("Error fetching ride: {}", e.getMessage(), e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    /**
     * GET /rides/passenger/{passengerId} - Get all rides for a passenger
     */
    @GetMapping("/passenger/{passengerId}")
    public ResponseEntity<List<RideResponse>> getRidesByPassenger(@PathVariable String passengerId) {
        logger.info("GET /rides/passenger/{} - Fetching rides for passenger", passengerId);
        try {
            List<RideResponse> rides = rideService.getRidesByPassenger(passengerId);
            return ResponseEntity.ok(rides);
        } catch (Exception e) {
            logger.error("Error fetching passenger rides: {}", e.getMessage(), e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    /**
     * GET /rides/driver/{driverId} - Get all rides for a driver
     */
    @GetMapping("/driver/{driverId}")
    public ResponseEntity<List<RideResponse>> getRidesByDriver(@PathVariable String driverId) {
        logger.info("GET /rides/driver/{} - Fetching rides for driver", driverId);
        try {
            List<RideResponse> rides = rideService.getRidesByDriver(driverId);
            return ResponseEntity.ok(rides);
        } catch (Exception e) {
            logger.error("Error fetching driver rides: {}", e.getMessage(), e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    /**
     * PUT /rides/{id}/status - Update ride status
     */
    @PutMapping("/{id}/status")
    public ResponseEntity<RideResponse> updateRideStatus(@PathVariable Long id,
            @RequestBody UpdateRideStatusRequest request) {
        logger.info("PUT /rides/{}/status - Updating to: {}", id, request.getStatus());
        try {
            RideResponse response = rideService.updateRideStatus(id, request);
            return ResponseEntity.ok(response);
        } catch (IllegalArgumentException e) {
            logger.warn("Ride not found: {}", id);
            return ResponseEntity.status(HttpStatus.NOT_FOUND).build();
        } catch (Exception e) {
            logger.error("Error updating ride status: {}", e.getMessage(), e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    // ============================================================================
    // Phase 4: Trip Lifecycle Endpoints
    // ============================================================================

    /**
     * POST /trips/{id}/start - Start a ride (ASSIGNED → IN_PROGRESS)
     * Driver calls this when beginning the trip
     */
    @PostMapping("/trips/{id}/start")
    public ResponseEntity<RideResponse> startRide(@PathVariable Long id) {
        logger.info("POST /trips/{}/start - Starting trip", id);
        try {
            RideResponse response = rideService.startRide(id);
            return ResponseEntity.ok(response);
        } catch (IllegalArgumentException e) {
            logger.warn("Cannot start ride: {}", e.getMessage());
            return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(null);
        } catch (Exception e) {
            logger.error("Error starting ride: {}", e.getMessage(), e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    /**
     * POST /trips/{id}/complete - Complete a ride (IN_PROGRESS → COMPLETED)
     * Driver calls this when finishing the trip
     */
    @PostMapping("/trips/{id}/complete")
    public ResponseEntity<RideResponse> completeRide(@PathVariable Long id) {
        logger.info("POST /trips/{}/complete - Completing trip", id);
        try {
            RideResponse response = rideService.completeRide(id);
            return ResponseEntity.ok(response);
        } catch (IllegalArgumentException e) {
            logger.warn("Cannot complete ride: {}", e.getMessage());
            return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(null);
        } catch (Exception e) {
            logger.error("Error completing ride: {}", e.getMessage(), e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    /**
     * POST /trips/{id}/cancel - Cancel a ride (REQUESTED/ASSIGNED → CANCELLED)
     * Passenger or driver can cancel before trip starts
     */
    @PostMapping("/trips/{id}/cancel")
    public ResponseEntity<RideResponse> cancelRide(@PathVariable Long id) {
        logger.info("POST /trips/{}/cancel - Cancelling trip", id);
        try {
            RideResponse response = rideService.cancelRide(id);
            return ResponseEntity.ok(response);
        } catch (IllegalArgumentException e) {
            logger.warn("Cannot cancel ride: {}", e.getMessage());
            return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(null);
        } catch (Exception e) {
            logger.error("Error cancelling ride: {}", e.getMessage(), e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    /**
     * GET /health - Health check endpoint
     */
    @GetMapping("/health")
    public ResponseEntity<Map<String, String>> health() {
        Map<String, String> response = new HashMap<>();
        response.put("status", "healthy");
        response.put("service", "ride-service");
        return ResponseEntity.ok(response);
    }
}
