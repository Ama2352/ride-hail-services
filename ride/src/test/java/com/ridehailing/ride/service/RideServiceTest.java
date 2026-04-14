package test.java.com.ridehailing.ride.service;

import com.ridehailing.ride.dto.CreateRideRequest;
import com.ridehailing.ride.dto.RideResponse;
import com.ridehailing.ride.dto.UpdateRideStatusRequest;
import com.ridehailing.ride.entity.Ride;
import com.ridehailing.ride.entity.RideStatus;
import com.ridehailing.ride.repository.RideRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.time.LocalDateTime;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

/**
 * Unit tests for RideService
 * Tests cover: create, get, update, start, complete, and cancel operations
 */
@ExtendWith(MockitoExtension.class)
public class RideServiceTest {

    @Mock
    private RideRepository rideRepository;

    @Mock
    private RideEventPublisher eventPublisher;

    @InjectMocks
    private RideService rideService;

    private CreateRideRequest createRideRequest;
    private Ride mockRide;

    @BeforeEach
    void setUp() {
        // Initialize test data
        createRideRequest = new CreateRideRequest();
        createRideRequest.setPassengerId("passenger-123");
        createRideRequest.setPickupLat(10.7769);
        createRideRequest.setPickupLng(106.7009);
        createRideRequest.setDropoffLat(10.7890);
        createRideRequest.setDropoffLng(106.7100);

        mockRide = new Ride(
                "passenger-123",
                10.7769,
                106.7009,
                10.7890,
                106.7100);
        mockRide.setId(1001L);
        mockRide.setStatus(RideStatus.REQUESTED);
        mockRide.setCreatedAt(LocalDateTime.now());
        mockRide.setUpdatedAt(LocalDateTime.now());
    }

    @Test
    void testCreateRide() {
        when(rideRepository.save(any(Ride.class))).thenReturn(mockRide);

        RideResponse response = rideService.createRide(createRideRequest);

        assertNotNull(response);
        assertEquals(1001L, response.getId());
        assertEquals("passenger-123", response.getPassengerId());
        assertEquals(RideStatus.REQUESTED, response.getStatus());

        verify(rideRepository, times(1)).save(any(Ride.class));
        verify(eventPublisher, times(1)).publishRideRequested(
                eq(1001L),
                eq("passenger-123"),
                eq(10.7769),
                eq(106.7009),
                eq(10.7890),
                eq(106.7100));
    }

    @Test
    void testGetRideById_Success() {
        when(rideRepository.findById(1001L)).thenReturn(Optional.of(mockRide));

        RideResponse response = rideService.getRideById(1001L);

        assertNotNull(response);
        assertEquals(1001L, response.getId());
        verify(rideRepository, times(1)).findById(1001L);
    }

    @Test
    void testGetRideById_NotFound() {
        when(rideRepository.findById(9999L)).thenReturn(Optional.empty());

        assertThrows(IllegalArgumentException.class, () -> rideService.getRideById(9999L));
        verify(rideRepository, times(1)).findById(9999L);
    }

    @Test
    void testGetRidesByPassenger() {
        List<Ride> rides = Arrays.asList(mockRide);
        when(rideRepository.findByPassengerId("passenger-123")).thenReturn(rides);

        List<RideResponse> responses = rideService.getRidesByPassenger("passenger-123");

        assertNotNull(responses);
        assertEquals(1, responses.size());
        verify(rideRepository, times(1)).findByPassengerId("passenger-123");
    }

    @Test
    void testGetRidesByDriver() {
        mockRide.setDriverId("driver-456");
        List<Ride> rides = Arrays.asList(mockRide);
        when(rideRepository.findByDriverId("driver-456")).thenReturn(rides);

        List<RideResponse> responses = rideService.getRidesByDriver("driver-456");

        assertNotNull(responses);
        assertEquals(1, responses.size());
        verify(rideRepository, times(1)).findByDriverId("driver-456");
    }

    @Test
    void testUpdateRideStatus_ToAssigned() {
        mockRide.setStatus(RideStatus.REQUESTED);
        UpdateRideStatusRequest request = new UpdateRideStatusRequest();
        request.setStatus(RideStatus.ASSIGNED);
        request.setDriverId("driver-456");

        when(rideRepository.findById(1001L)).thenReturn(Optional.of(mockRide));
        when(rideRepository.save(any(Ride.class))).thenReturn(mockRide);

        RideResponse response = rideService.updateRideStatus(1001L, request);

        assertNotNull(response);
        assertEquals("driver-456", mockRide.getDriverId());
        verify(rideRepository, times(1)).findById(1001L);
        verify(rideRepository, times(1)).save(any(Ride.class));
    }

    @Test
    void testUpdateRideStatus_ToInProgress() {
        mockRide.setStatus(RideStatus.ASSIGNED);
        mockRide.setDriverId("driver-456");
        UpdateRideStatusRequest request = new UpdateRideStatusRequest();
        request.setStatus(RideStatus.IN_PROGRESS);

        when(rideRepository.findById(1001L)).thenReturn(Optional.of(mockRide));
        when(rideRepository.save(any(Ride.class))).thenReturn(mockRide);

        RideResponse response = rideService.updateRideStatus(1001L, request);

        assertNotNull(response);
        assertNotNull(mockRide.getStartedAt());
        verify(eventPublisher, times(1)).publishRideStarted(
                eq(1001L),
                anyString(),
                anyString(),
                anyDouble(),
                anyDouble(),
                anyDouble(),
                anyDouble());
    }

    @Test
    void testUpdateRideStatus_Completed() {
        mockRide.setStatus(RideStatus.IN_PROGRESS);
        mockRide.setDriverId("driver-456");
        UpdateRideStatusRequest request = new UpdateRideStatusRequest();
        request.setStatus(RideStatus.COMPLETED);

        when(rideRepository.findById(1001L)).thenReturn(Optional.of(mockRide));
        when(rideRepository.save(any(Ride.class))).thenReturn(mockRide);

        RideResponse response = rideService.updateRideStatus(1001L, request);

        assertNotNull(response);
        assertNotNull(mockRide.getCompletedAt());
        verify(eventPublisher, times(1)).publishRideCompleted(
                eq(1001L),
                anyString(),
                anyString(),
                anyDouble(),
                anyDouble(),
                anyDouble(),
                anyDouble());
    }

    @Test
    void testStartRide_Success() {
        mockRide.setStatus(RideStatus.ASSIGNED);
        mockRide.setDriverId("driver-456");

        when(rideRepository.findById(1001L)).thenReturn(Optional.of(mockRide));
        when(rideRepository.save(any(Ride.class))).thenReturn(mockRide);

        RideResponse response = rideService.startRide(1001L);

        assertNotNull(response);
        assertEquals(RideStatus.IN_PROGRESS, mockRide.getStatus());
        assertNotNull(mockRide.getStartedAt());
        verify(eventPublisher, times(1)).publishRideStarted(anyLong(), anyString(), anyString(),
                anyDouble(), anyDouble(), anyDouble(), anyDouble());
    }

    @Test
    void testStartRide_NotAssigned() {
        mockRide.setStatus(RideStatus.REQUESTED);

        when(rideRepository.findById(1001L)).thenReturn(Optional.of(mockRide));

        assertThrows(IllegalArgumentException.class, () -> rideService.startRide(1001L),
                "Ride must be in ASSIGNED status to start");
    }

    @Test
    void testStartRide_NotFound() {
        when(rideRepository.findById(9999L)).thenReturn(Optional.empty());

        assertThrows(IllegalArgumentException.class, () -> rideService.startRide(9999L),
                "Ride not found");
    }

    @Test
    void testCompleteRide_Success() {
        mockRide.setStatus(RideStatus.IN_PROGRESS);
        mockRide.setDriverId("driver-456");
        mockRide.setStartedAt(LocalDateTime.now().minusMinutes(10));

        when(rideRepository.findById(1001L)).thenReturn(Optional.of(mockRide));
        when(rideRepository.save(any(Ride.class))).thenReturn(mockRide);

        RideResponse response = rideService.completeRide(1001L);

        assertNotNull(response);
        assertEquals(RideStatus.COMPLETED, mockRide.getStatus());
        assertNotNull(mockRide.getCompletedAt());
        verify(eventPublisher, times(1)).publishRideCompleted(anyLong(), anyString(), anyString(),
                anyDouble(), anyDouble(), anyDouble(), anyDouble());
    }

    @Test
    void testCompleteRide_NotInProgress() {
        mockRide.setStatus(RideStatus.ASSIGNED);

        when(rideRepository.findById(1001L)).thenReturn(Optional.of(mockRide));

        assertThrows(IllegalArgumentException.class, () -> rideService.completeRide(1001L),
                "Ride must be in IN_PROGRESS status to complete");
    }

    @Test
    void testCancelRide_FromRequested() {
        mockRide.setStatus(RideStatus.REQUESTED);

        when(rideRepository.findById(1001L)).thenReturn(Optional.of(mockRide));
        when(rideRepository.save(any(Ride.class))).thenReturn(mockRide);

        RideResponse response = rideService.cancelRide(1001L);

        assertNotNull(response);
        assertEquals(RideStatus.CANCELLED, mockRide.getStatus());
        assertNotNull(mockRide.getCompletedAt());
        verify(eventPublisher, times(1)).publishRideCancelled(anyLong(), anyString(), anyString(),
                anyDouble(), anyDouble(), anyDouble(), anyDouble());
    }

    @Test
    void testCancelRide_FromAssigned() {
        mockRide.setStatus(RideStatus.ASSIGNED);
        mockRide.setDriverId("driver-456");

        when(rideRepository.findById(1001L)).thenReturn(Optional.of(mockRide));
        when(rideRepository.save(any(Ride.class))).thenReturn(mockRide);

        RideResponse response = rideService.cancelRide(1001L);

        assertNotNull(response);
        assertEquals(RideStatus.CANCELLED, mockRide.getStatus());
    }

    @Test
    void testCancelRide_AlreadyCompleted() {
        mockRide.setStatus(RideStatus.COMPLETED);

        when(rideRepository.findById(1001L)).thenReturn(Optional.of(mockRide));

        assertThrows(IllegalArgumentException.class, () -> rideService.cancelRide(1001L),
                "Cannot cancel ride in COMPLETED status");
    }

    @Test
    void testCancelRide_AlreadyCancelled() {
        mockRide.setStatus(RideStatus.CANCELLED);

        when(rideRepository.findById(1001L)).thenReturn(Optional.of(mockRide));

        assertThrows(IllegalArgumentException.class, () -> rideService.cancelRide(1001L),
                "Cannot cancel ride in CANCELLED status");
    }
}
