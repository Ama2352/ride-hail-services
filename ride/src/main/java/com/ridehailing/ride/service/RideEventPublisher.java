package com.ridehailing.ride.service;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.ridehailing.ride.event.RideEvent;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.HashMap;
import java.util.Map;

/**
 * Service for publishing ride events to Redis Stream.
 * Event types: Ride.Requested, Ride.Offered, Ride.Assigned, Ride.Started,
 * Ride.Completed, Ride.Cancelled
 */
@Service
public class RideEventPublisher {
    private static final Logger logger = LoggerFactory.getLogger(RideEventPublisher.class);
    private static final String RIDE_EVENTS_STREAM = "ride-events";

    @Autowired
    private RedisTemplate<String, Object> redisTemplate;

    @Autowired
    private ObjectMapper objectMapper;

    /**
     * Publish a ride event to Redis Stream.
     * 
     * @param eventType Type of event (e.g., "Ride.Requested", "Ride.Assigned")
     * @param rideEvent The event payload
     */
    public void publishEvent(String eventType, RideEvent rideEvent) {
        try {
            rideEvent.setEventType(eventType);
            rideEvent.setTimestamp(LocalDateTime.now());

            Map<String, String> eventData = new HashMap<>();
            eventData.put("event_type", eventType);
            eventData.put("ride_id", String.valueOf(rideEvent.getRideId()));
            eventData.put("passenger_id", rideEvent.getPassengerId());
            eventData.put("driver_id", rideEvent.getDriverId() != null ? rideEvent.getDriverId() : "null");
            eventData.put("pickup_lat", String.valueOf(rideEvent.getPickupLat()));
            eventData.put("pickup_lng", String.valueOf(rideEvent.getPickupLng()));
            eventData.put("dropoff_lat", String.valueOf(rideEvent.getDropoffLat()));
            eventData.put("dropoff_lng", String.valueOf(rideEvent.getDropoffLng()));
            eventData.put("timestamp", rideEvent.getTimestamp().toString());

            // Add to Redis Stream
            redisTemplate.opsForStream().add(RIDE_EVENTS_STREAM, eventData);

            logger.info("Published event to Redis Stream. Type: {}, RideID: {}", eventType, rideEvent.getRideId());
        } catch (Exception e) {
            logger.error("Failed to publish event to Redis Stream: {}", e.getMessage(), e);
            throw new RuntimeException("Failed to publish event", e);
        }
    }

    /**
     * Publish Ride.Requested event (when passenger requests a ride)
     */
    public void publishRideRequested(Long rideId, String passengerId, Double pickupLat, Double pickupLng,
            Double dropoffLat, Double dropoffLng) {
        RideEvent event = new RideEvent("Ride.Requested", rideId, passengerId, null,
                pickupLat, pickupLng, dropoffLat, dropoffLng, LocalDateTime.now());
        publishEvent("Ride.Requested", event);
    }

    /**
     * Publish Ride.Assigned event (when driver accepts the ride)
     */
    public void publishRideAssigned(Long rideId, String passengerId, String driverId,
            Double pickupLat, Double pickupLng,
            Double dropoffLat, Double dropoffLng) {
        RideEvent event = new RideEvent("Ride.Assigned", rideId, passengerId, driverId,
                pickupLat, pickupLng, dropoffLat, dropoffLng, LocalDateTime.now());
        publishEvent("Ride.Assigned", event);
    }

    /**
     * Publish Ride.Started event (when ride begins)
     */
    public void publishRideStarted(Long rideId, String passengerId, String driverId,
            Double pickupLat, Double pickupLng,
            Double dropoffLat, Double dropoffLng) {
        RideEvent event = new RideEvent("Ride.Started", rideId, passengerId, driverId,
                pickupLat, pickupLng, dropoffLat, dropoffLng, LocalDateTime.now());
        publishEvent("Ride.Started", event);
    }

    /**
     * Publish Ride.Completed event (when ride finishes)
     */
    public void publishRideCompleted(Long rideId, String passengerId, String driverId,
            Double pickupLat, Double pickupLng,
            Double dropoffLat, Double dropoffLng) {
        RideEvent event = new RideEvent("Ride.Completed", rideId, passengerId, driverId,
                pickupLat, pickupLng, dropoffLat, dropoffLng, LocalDateTime.now());
        publishEvent("Ride.Completed", event);
    }

    /**
     * Publish Ride.Cancelled event (when ride is cancelled)
     */
    public void publishRideCancelled(Long rideId, String passengerId, String driverId,
            Double pickupLat, Double pickupLng,
            Double dropoffLat, Double dropoffLng) {
        RideEvent event = new RideEvent("Ride.Cancelled", rideId, passengerId, driverId,
                pickupLat, pickupLng, dropoffLat, dropoffLng, LocalDateTime.now());
        publishEvent("Ride.Cancelled", event);
    }
}
