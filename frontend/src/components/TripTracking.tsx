import React, { useState, useEffect } from 'react';
import { apiClient } from '../api/client';
import type { Trip } from '../types';
import './TripTracking.css';

interface TripTrackingProps {
  rideId: string;
  onTripComplete?: () => void;
}

export const TripTracking: React.FC<TripTrackingProps> = ({ rideId, onTripComplete }) => {
  const [trip, setTrip] = useState<Trip | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [feedback, setFeedback] = useState('');
  const [rating, setRating] = useState<number>(0);

  useEffect(() => {
    const fetchTrip = async () => {
      try {
        const tripData = await apiClient.getTripDetails(rideId);
        setTrip(tripData);
      } catch (err: any) {
        setError(err.message || 'Failed to load trip details');
      } finally {
        setLoading(false);
      }
    };

    fetchTrip();
    const interval = setInterval(fetchTrip, 3000); // Poll every 3 seconds

    return () => clearInterval(interval);
  }, [rideId]);

  const handleRateTrip = async () => {
    if (!trip) return;

    try {
      await apiClient.rateTripAndFeedback(trip.tripId, rating, feedback);
      alert('Thank you for your feedback!');
      onTripComplete?.();
    } catch (err: any) {
      setError(err.message || 'Failed to submit feedback');
    }
  };

  if (loading) {
    return <div className="trip-tracking loading">Loading trip details...</div>;
  }

  if (error || !trip) {
    return <div className="trip-tracking error">{error || 'Trip details not found'}</div>;
  }

  const isCompleted = trip.status === 'completed';

  return (
    <div className="trip-tracking">
      <div className="trip-header">
        <h2>Trip Details</h2>
        <span className={`status-badge ${trip.status}`}>{trip.status.replace(/_/g, ' ')}</span>
      </div>

      <div className="driver-info">
        <h3>Driver Information</h3>
        <div className="driver-details">
          <div className="driver-name">{trip.driver.name}</div>
          <div className="driver-rating">⭐ {trip.driver.rating.toFixed(1)}</div>
          <div className="vehicle-plate">{trip.driver.vehiclePlate}</div>
        </div>
      </div>

      <div className="trip-info">
        <h3>Trip Progress</h3>
        <div className="location-info">
          <div className="location-item">
            <span className="label">Pickup:</span>
            <span className="coords">
              {trip.pickupLocation.latitude.toFixed(4)}, {trip.pickupLocation.longitude.toFixed(4)}
            </span>
          </div>
          <div className="location-item">
            <span className="label">Current:</span>
            <span className="coords">
              {trip.currentLocation.latitude.toFixed(4)}, {trip.currentLocation.longitude.toFixed(4)}
            </span>
          </div>
          <div className="location-item">
            <span className="label">Dropoff:</span>
            <span className="coords">
              {trip.dropoffLocation.latitude.toFixed(4)}, {trip.dropoffLocation.longitude.toFixed(4)}
            </span>
          </div>
        </div>

        {trip.status !== 'completed' && (
          <div className="eta-info">
            <span className="label">Estimated Arrival:</span>
            <span className="time">{Math.ceil(trip.estimatedArrivalTime / 60)} min</span>
          </div>
        )}

        <div className="fare-info">
          <span className="label">Fare:</span>
          <span className="amount">
            {trip.fare.currency} {trip.fare.amount.toFixed(2)}
          </span>
        </div>
      </div>

      {isCompleted && (
        <div className="feedback-section">
          <h3>Rate Your Trip</h3>
          <div className="rating-input">
            <div className="star-rating">
              {[1, 2, 3, 4, 5].map((star) => (
                <button
                  key={star}
                  className={`star ${rating >= star ? 'filled' : ''}`}
                  onClick={() => setRating(star)}
                >
                  ★
                </button>
              ))}
            </div>
          </div>
          <textarea
            className="feedback-textarea"
            placeholder="How was your ride? (Optional)"
            value={feedback}
            onChange={(e) => setFeedback(e.target.value)}
            rows={3}
          />
          <button className="submit-feedback-btn" onClick={handleRateTrip} disabled={rating === 0}>
            Submit Feedback
          </button>
        </div>
      )}
    </div>
  );
};
