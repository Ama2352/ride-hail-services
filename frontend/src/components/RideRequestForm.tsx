import React, { useState, useCallback } from 'react';
import { apiClient } from '../api/client';
import type { RideRequest } from '../types';
import { Map, MapPin, Navigation, Car, Sparkles } from 'lucide-react';
import './RideRequestForm.css';

interface RideRequestFormProps {
  onRequestSubmitted: (ride: RideRequest) => void;
}

export const RideRequestForm: React.FC<RideRequestFormProps> = ({ onRequestSubmitted }) => {
  const [pickupLat, setPickupLat] = useState('10.762622');
  const [pickupLng, setPickupLng] = useState('106.660172');
  const [dropoffLat, setDropoffLat] = useState('10.771511');
  const [dropoffLng, setDropoffLng] = useState('106.698387');
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
      } catch (err: any) {
        setError(err.response?.data?.message || 'Failed to request ride. Please try again.');
      } finally {
        setIsLoading(false);
      }
    },
    [pickupLat, pickupLng, dropoffLat, dropoffLng, rideType, onRequestSubmitted]
  );

  return (
    <div className="bottom-sheet">
      <div className="bottom-sheet-handle"></div>
      <h2 className="sheet-title">Where to?</h2>

      <form className="ride-form" onSubmit={handleSubmit}>
        <div className="location-inputs">
          <div className="location-track">
            <div className="dot-green"></div>
            <div className="line-vertical"></div>
            <div className="dot-red"></div>
          </div>
          
          <div className="inputs-wrapper">
            <div className="input-group">
              <input
                type="text"
                value={`${pickupLat}, ${pickupLng}`}
                onChange={() => {}}
                placeholder="Current Location"
                readOnly
                className="location-input"
              />
            </div>
            <div className="divider"></div>
            <div className="input-group">
              <input
                type="text"
                value={`${dropoffLat}, ${dropoffLng}`}
                onChange={() => {}}
                placeholder="Destination"
                readOnly
                className="location-input destination-input"
              />
            </div>
          </div>
        </div>

        <div className="ride-options">
          <div 
            className={`ride-option ${rideType === 'economy' ? 'active' : ''}`}
            onClick={() => setRideType('economy')}
          >
            <div className="option-icon"><Car size={24} /></div>
            <div className="option-details">
              <h4>Economy</h4>
              <p>Affordable, everyday rides</p>
            </div>
            <div className="option-price">~ $5.00</div>
          </div>

          <div 
            className={`ride-option ${rideType === 'premium' ? 'active' : ''}`}
            onClick={() => setRideType('premium')}
          >
            <div className="option-icon premium"><Sparkles size={24} /></div>
            <div className="option-details">
              <h4>Premium</h4>
              <p>Comfort and extra space</p>
            </div>
            <div className="option-price">~ $8.50</div>
          </div>
        </div>

        {error && <div className="error-message">{error}</div>}

        <button type="submit" disabled={isLoading} className="submit-btn grab-primary-btn">
          {isLoading ? 'Finding driver...' : 'Book Ride'}
        </button>
      </form>
    </div>
  );
};
