import React, { useState, useCallback } from 'react';
import { apiClient } from '../api/client';
import type { RideRequest } from '../types';
import './RideRequestForm.css';

interface RideRequestFormProps {
  onRequestSubmitted: (ride: RideRequest) => void;
}

export const RideRequestForm: React.FC<RideRequestFormProps> = ({ onRequestSubmitted }) => {
  const [pickupLat, setPickupLat] = useState('40.7128');
  const [pickupLng, setPickupLng] = useState('-74.0060');
  const [dropoffLat, setDropoffLat] = useState('40.7580');
  const [dropoffLng, setDropoffLng] = useState('-73.9855');
  const [rideType, setRideType] = useState<'economy' | 'premium'>('economy');
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      setError(null);
      setIsLoading(true);

      try {
        const ride = await apiClient.requestRide(
          parseFloat(pickupLat),
          parseFloat(pickupLng),
          parseFloat(dropoffLat),
          parseFloat(dropoffLng),
          rideType
        );
        onRequestSubmitted(ride);
        // Reset form
        setPickupLat('40.7128');
        setPickupLng('-74.0060');
        setDropoffLat('40.7580');
        setDropoffLng('-73.9855');
      } catch (err: any) {
        setError(err.response?.data?.message || 'Failed to request ride. Please try again.');
      } finally {
        setIsLoading(false);
      }
    },
    [pickupLat, pickupLng, dropoffLat, dropoffLng, rideType, onRequestSubmitted]
  );

  return (
    <form className="ride-request-form" onSubmit={handleSubmit}>
      <h2>Book Your Ride</h2>

      <div className="form-section">
        <h3>Pickup Location</h3>
        <div className="form-group">
          <label htmlFor="pickup-lat">Latitude</label>
          <input
            id="pickup-lat"
            type="number"
            step="0.0001"
            value={pickupLat}
            onChange={(e) => setPickupLat(e.target.value)}
            placeholder="40.7128"
            required
            disabled={isLoading}
          />
        </div>
        <div className="form-group">
          <label htmlFor="pickup-lng">Longitude</label>
          <input
            id="pickup-lng"
            type="number"
            step="0.0001"
            value={pickupLng}
            onChange={(e) => setPickupLng(e.target.value)}
            placeholder="-74.0060"
            required
            disabled={isLoading}
          />
        </div>
      </div>

      <div className="form-section">
        <h3>Dropoff Location</h3>
        <div className="form-group">
          <label htmlFor="dropoff-lat">Latitude</label>
          <input
            id="dropoff-lat"
            type="number"
            step="0.0001"
            value={dropoffLat}
            onChange={(e) => setDropoffLat(e.target.value)}
            placeholder="40.7580"
            required
            disabled={isLoading}
          />
        </div>
        <div className="form-group">
          <label htmlFor="dropoff-lng">Longitude</label>
          <input
            id="dropoff-lng"
            type="number"
            step="0.0001"
            value={dropoffLng}
            onChange={(e) => setDropoffLng(e.target.value)}
            placeholder="-73.9855"
            required
            disabled={isLoading}
          />
        </div>
      </div>

      <div className="form-section">
        <h3>Ride Type</h3>
        <div className="ride-type-select">
          <label>
            <input
              type="radio"
              value="economy"
              checked={rideType === 'economy'}
              onChange={(e) => setRideType(e.target.value as 'economy' | 'premium')}
              disabled={isLoading}
            />
            <span>Economy - Affordable & Reliable</span>
          </label>
          <label>
            <input
              type="radio"
              value="premium"
              checked={rideType === 'premium'}
              onChange={(e) => setRideType(e.target.value as 'economy' | 'premium')}
              disabled={isLoading}
            />
            <span>Premium - Comfort & Quality</span>
          </label>
        </div>
      </div>

      {error && <div className="error-message">{error}</div>}

      <button type="submit" disabled={isLoading} className="submit-btn">
        {isLoading ? 'Requesting Ride...' : 'Request Ride'}
      </button>
    </form>
  );
};
