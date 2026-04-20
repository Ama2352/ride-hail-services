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
    return <div className="bottom-sheet loading">Loading tracking details...</div>;
  }

  if (error || !trip) {
    return <div className="bottom-sheet error">{error || 'Trip details not found'}</div>;
  }

  const isCompleted = trip.status === 'completed';

  return (
    <div className="bottom-sheet">
      <div className="bottom-sheet-handle"></div>
      
      <div className="trip-status-header">
        <h2 className="sheet-title">
          {isCompleted ? 'You have arrived' : trip.status === 'in_progress' ? 'Heading to destination' : 'Driver is on the way'}
        </h2>
        {trip.status !== 'completed' && (
          <div className="eta-badge">{Math.ceil(trip.estimatedArrivalTime / 60)} min</div>
        )}
      </div>

      <div className="driver-card">
        <div className="driver-avatar">
          <img src={`https://ui-avatars.com/api/?name=${trip.driver.name}&background=0D8ABC&color=fff`} alt="Driver" />
        </div>
        <div className="driver-details">
          <h4>{trip.driver.name}</h4>
          <span className="driver-rating">⭐ {trip.driver.rating.toFixed(1)}</span>
        </div>
        <div className="vehicle-plate">{trip.driver.vehiclePlate}</div>
      </div>

      <div className="trip-info-card">
        <div className="location-inputs">
          <div className="location-track">
            <div className="dot-green"></div>
            <div className="line-vertical"></div>
            <div className="dot-red"></div>
          </div>
          <div className="inputs-wrapper">
             <div className="info-text">Pickup Location</div>
             <div className="divider"></div>
             <div className="info-text font-bold">Destination</div>
          </div>
        </div>
      </div>

      {isCompleted ? (
        <div className="feedback-section">
          <h3>Rate your trip</h3>
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
          <textarea
            className="feedback-textarea"
            placeholder="Add a comment (Optional)"
            value={feedback}
            onChange={(e) => setFeedback(e.target.value)}
            rows={2}
          />
          <button className="grab-primary-btn" onClick={handleRateTrip} disabled={rating === 0}>
            Submit & Complete
          </button>
        </div>
      ) : (
         <button className="grab-secondary-btn" onClick={onTripComplete}>
            (Demo) Finish Booking Flow
         </button>
      )}
    </div>
  );
};
