import axios, { AxiosInstance } from 'axios';
import type { RideRequest, Trip, UserProfile } from '../types';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });
  }

  // Ride endpoints
  async requestRide(pickupLat: number, pickupLng: number, dropoffLat: number, dropoffLng: number, rideType: 'economy' | 'premium'): Promise<RideRequest> {
    const response = await this.client.post('/api/rides', {
      pickupLocation: { latitude: pickupLat, longitude: pickupLng },
      dropoffLocation: { latitude: dropoffLat, longitude: dropoffLng },
      rideType,
    });
    return response.data;
  }

  async getRideStatus(rideId: string): Promise<RideRequest> {
    const response = await this.client.get(`/api/rides/${rideId}`);
    return response.data;
  }

  async cancelRide(rideId: string): Promise<void> {
    await this.client.post(`/api/rides/${rideId}/cancel`);
  }

  // Trip endpoints
  async getTripDetails(rideId: string): Promise<Trip> {
    const response = await this.client.get(`/api/trips/ride/${rideId}`);
    return response.data;
  }

  async rateTripAndFeedback(tripId: string, rating: number, feedback: string): Promise<void> {
    await this.client.post(`/api/trips/${tripId}/rating`, {
      rating,
      feedback,
    });
  }

  // User endpoints
  async getUserProfile(): Promise<UserProfile> {
    const response = await this.client.get('/api/users/profile');
    return response.data;
  }

  async getRideHistory(limit: number = 10): Promise<RideRequest[]> {
    const response = await this.client.get('/api/rides/history', {
      params: { limit },
    });
    return response.data;
  }

  // Health check
  async healthCheck(): Promise<boolean> {
    try {
      const response = await this.client.get('/health');
      return response.status === 200;
    } catch {
      return false;
    }
  }
}

export const apiClient = new ApiClient();
