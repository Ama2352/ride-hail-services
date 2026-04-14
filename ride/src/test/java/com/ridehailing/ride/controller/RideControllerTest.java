package com.ridehailing.ride.controller;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.ridehailing.ride.dto.CreateRideRequest;
import com.ridehailing.ride.dto.RideResponse;
import com.ridehailing.ride.dto.UpdateRideStatusRequest;
import com.ridehailing.ride.entity.RideStatus;
import com.ridehailing.ride.service.RideService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import java.time.LocalDateTime;

import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

/**
 * Integration tests for RideController
 * Tests cover: create, read, update, start, complete, cancel, and health
 * endpoints
 */
@WebMvcTest(RideController.class)
public class RideControllerTest {

        @Autowired
        private MockMvc mockMvc;

        @Autowired
        private ObjectMapper objectMapper;

        @MockBean
        private RideService rideService;

        private RideResponse mockRideResponse;

        @BeforeEach
        void setUp() {
                mockRideResponse = new RideResponse();
                mockRideResponse.setId(1001L);
                mockRideResponse.setPassengerId("passenger-123");
                mockRideResponse.setDriverId("driver-456");
                mockRideResponse.setStatus(RideStatus.REQUESTED);
                mockRideResponse.setPickupLat(10.7769);
                mockRideResponse.setPickupLng(106.7009);
                mockRideResponse.setDropoffLat(10.7890);
                mockRideResponse.setDropoffLng(106.7100);
                mockRideResponse.setEstimatedFare(12.50);
                mockRideResponse.setCreatedAt(LocalDateTime.now());
                mockRideResponse.setUpdatedAt(LocalDateTime.now());
        }

        @Test
        void testCreateRide() throws Exception {
                CreateRideRequest request = new CreateRideRequest();
                request.setPassengerId("passenger-123");
                request.setPickupLat(10.7769);
                request.setPickupLng(106.7009);
                request.setDropoffLat(10.7890);
                request.setDropoffLng(106.7100);

                when(rideService.createRide(any(CreateRideRequest.class)))
                                .thenReturn(mockRideResponse);

                mockMvc.perform(post("/rides")
                                .contentType(MediaType.APPLICATION_JSON)
                                .content(objectMapper.writeValueAsString(request)))
                                .andExpect(status().isCreated())
                                .andExpect(jsonPath("$.id").value(1001L))
                                .andExpect(jsonPath("$.passengerId").value("passenger-123"))
                                .andExpect(jsonPath("$.status").value("REQUESTED"));
        }

        @Test
        void testGetRideById() throws Exception {
                when(rideService.getRideById(1001L))
                                .thenReturn(mockRideResponse);

                mockMvc.perform(get("/rides/1001"))
                                .andExpect(status().isOk())
                                .andExpect(jsonPath("$.id").value(1001L))
                                .andExpect(jsonPath("$.passengerId").value("passenger-123"));
        }

        @Test
        void testGetRideById_NotFound() throws Exception {
                when(rideService.getRideById(9999L))
                                .thenThrow(new IllegalArgumentException("Ride not found"));

                mockMvc.perform(get("/rides/9999"))
                                .andExpect(status().isNotFound());
        }

        @Test
        void testUpdateRideStatus() throws Exception {
                mockRideResponse.setStatus(RideStatus.ASSIGNED);

                UpdateRideStatusRequest request = new UpdateRideStatusRequest();
                request.setStatus(RideStatus.ASSIGNED);
                request.setDriverId("driver-456");

                when(rideService.updateRideStatus(eq(1001L), any(UpdateRideStatusRequest.class)))
                                .thenReturn(mockRideResponse);

                mockMvc.perform(put("/rides/1001/status")
                                .contentType(MediaType.APPLICATION_JSON)
                                .content(objectMapper.writeValueAsString(request)))
                                .andExpect(status().isOk())
                                .andExpect(jsonPath("$.status").value("ASSIGNED"));
        }

        // ============================================================================
        // Phase 4: Trip Lifecycle Endpoint Tests
        // ============================================================================

        @Test
        void testStartRide() throws Exception {
                mockRideResponse.setStatus(RideStatus.IN_PROGRESS);
                mockRideResponse.setStartedAt(LocalDateTime.now());

                when(rideService.startRide(1001L))
                                .thenReturn(mockRideResponse);

                mockMvc.perform(post("/rides/trips/1001/start"))
                                .andExpect(status().isOk())
                                .andExpect(jsonPath("$.status").value("IN_PROGRESS"))
                                .andExpect(jsonPath("$.id").value(1001L));
        }

        @Test
        void testStartRide_InvalidStatus() throws Exception {
                when(rideService.startRide(1001L))
                                .thenThrow(new IllegalArgumentException("Ride must be in ASSIGNED status"));

                mockMvc.perform(post("/rides/trips/1001/start"))
                                .andExpect(status().isBadRequest());
        }

        @Test
        void testCompleteRide() throws Exception {
                mockRideResponse.setStatus(RideStatus.COMPLETED);
                mockRideResponse.setCompletedAt(LocalDateTime.now());

                when(rideService.completeRide(1001L))
                                .thenReturn(mockRideResponse);

                mockMvc.perform(post("/rides/trips/1001/complete"))
                                .andExpect(status().isOk())
                                .andExpect(jsonPath("$.status").value("COMPLETED"))
                                .andExpect(jsonPath("$.id").value(1001L));
        }

        @Test
        void testCompleteRide_InvalidStatus() throws Exception {
                when(rideService.completeRide(1001L))
                                .thenThrow(new IllegalArgumentException("Ride must be in IN_PROGRESS status"));

                mockMvc.perform(post("/rides/trips/1001/complete"))
                                .andExpect(status().isBadRequest());
        }

        @Test
        void testCancelRide() throws Exception {
                mockRideResponse.setStatus(RideStatus.CANCELLED);
                mockRideResponse.setCompletedAt(LocalDateTime.now());

                when(rideService.cancelRide(1001L))
                                .thenReturn(mockRideResponse);

                mockMvc.perform(post("/rides/trips/1001/cancel"))
                                .andExpect(status().isOk())
                                .andExpect(jsonPath("$.status").value("CANCELLED"))
                                .andExpect(jsonPath("$.id").value(1001L));
        }

        @Test
        void testCancelRide_AlreadyCompleted() throws Exception {
                when(rideService.cancelRide(1001L))
                                .thenThrow(new IllegalArgumentException("Cannot cancel ride in COMPLETED status"));

                mockMvc.perform(post("/rides/trips/1001/cancel"))
                                .andExpect(status().isBadRequest());
        }

        @Test
        void testStartRide_NotFound() throws Exception {
                when(rideService.startRide(9999L))
                                .thenThrow(new IllegalArgumentException("Ride not found"));

                mockMvc.perform(post("/rides/trips/9999/start"))
                                .andExpect(status().isBadRequest());
        }

        @Test
        void testCompleteRide_NotFound() throws Exception {
                when(rideService.completeRide(9999L))
                                .thenThrow(new IllegalArgumentException("Ride not found"));

                mockMvc.perform(post("/rides/trips/9999/complete"))
                                .andExpect(status().isBadRequest());
        }

        @Test
        void testCancelRide_NotFound() throws Exception {
                when(rideService.cancelRide(9999L))
                                .thenThrow(new IllegalArgumentException("Ride not found"));

                mockMvc.perform(post("/rides/trips/9999/cancel"))
                                .andExpect(status().isBadRequest());
        }

        @Test
        void testHealth() throws Exception {
                mockMvc.perform(get("/rides/health"))
                                .andExpect(status().isOk())
                                .andExpect(jsonPath("$.status").value("healthy"))
                                .andExpect(jsonPath("$.service").value("ride-service"));
        }
}
