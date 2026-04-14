package com.ridehailing.ride.controller;

import com.ridehailing.ride.dto.CreateRideRequest;
import com.ridehailing.ride.dto.RideResponse;
import com.ridehailing.ride.dto.UpdateRideStatusRequest;
import com.ridehailing.ride.entity.RideStatus;
import com.ridehailing.ride.service.RideService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

/**
 * Integration tests for RideController
 */
@SpringBootTest
@AutoConfigureMockMvc
class RideControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @MockBean
    private RideService rideService;

    private RideResponse mockRideResponse;

    @BeforeEach
    void setUp() {
        mockRideResponse = new RideResponse();
        mockRideResponse.setId(1L);
        mockRideResponse.setPassengerId("pax-123");
        mockRideResponse.setStatus(RideStatus.REQUESTED);
        mockRideResponse.setPickupLat(10.123);
        mockRideResponse.setPickupLng(106.456);
        mockRideResponse.setDropoffLat(10.789);
        mockRideResponse.setDropoffLng(106.789);
    }

    @Test
    void testCreateRide() throws Exception {
        when(rideService.createRide(any(CreateRideRequest.class))).thenReturn(mockRideResponse);

        String requestBody = """
                {
                  "passengerId": "pax-123",
                  "pickupLat": 10.123,
                  "pickupLng": 106.456,
                  "dropoffLat": 10.789,
                  "dropoffLng": 106.789
                }
                """;

        mockMvc.perform(post("/rides")
                .contentType(MediaType.APPLICATION_JSON)
                .content(requestBody))
                .andExpect(status().isCreated())
                .andExpect(jsonPath("$.id").value(1))
                .andExpect(jsonPath("$.passengerId").value("pax-123"));
    }

    @Test
    void testGetRideById() throws Exception {
        when(rideService.getRideById(1L)).thenReturn(mockRideResponse);

        mockMvc.perform(get("/rides/1"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.id").value(1))
                .andExpect(jsonPath("$.passengerId").value("pax-123"));
    }

    @Test
    void testGetHealth() throws Exception {
        mockMvc.perform(get("/rides/health"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.status").value("healthy"))
                .andExpect(jsonPath("$.service").value("ride-service"));
    }
}
