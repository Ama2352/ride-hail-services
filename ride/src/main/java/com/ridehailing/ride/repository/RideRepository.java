package com.ridehailing.ride.repository;

import com.ridehailing.ride.entity.Ride;
import com.ridehailing.ride.entity.RideStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

/**
 * JPA Repository for Ride entity.
 * Provides database access layer.
 */
@Repository
public interface RideRepository extends JpaRepository<Ride, Long> {

    /**
     * Find rides by passenger ID
     */
    List<Ride> findByPassengerId(String passengerId);

    /**
     * Find rides by driver ID
     */
    List<Ride> findByDriverId(String driverId);

    /**
     * Find rides by status
     */
    List<Ride> findByStatus(RideStatus status);

    /**
     * Find a specific ride by ID
     */
    Optional<Ride> findById(Long id);
}
