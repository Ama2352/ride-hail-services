import React, { useState } from 'react';
import { RideRequestForm } from './components/RideRequestForm';
import { TripTracking } from './components/TripTracking';
import { MapComponent } from './components/MapComponent';
import type { RideRequest } from './types';
import './App.css';

export const App: React.FC = () => {
  const [currentRide, setCurrentRide] = useState<RideRequest | null>(null);

  const handleRideRequested = (ride: RideRequest) => {
    setCurrentRide(ride);
  };

  const handleNewRide = () => {
    setCurrentRide(null);
  };

  // Mocking stats for Map
  const mapData = currentRide ? {
    pickupStatus: { lat: 10.762622, lng: 106.660172 },
    dropoffStatus: { lat: 10.771511, lng: 106.698387 },
  } : {
    pickupStatus: { lat: 10.762622, lng: 106.660172 }
  };

  return (
    <div className="app-container">
      <header className="app-header">
        <div className="logo-container">
          <div className="logo-dot"></div>
          <span className="logo-text">RideHailing</span>
        </div>
        <div className="profile-btn">
          <img src="https://ui-avatars.com/api/?name=User&background=0D8ABC&color=fff" alt="User" />
        </div>
      </header>

      <main className="app-main">
        <MapComponent 
          pickupStatus={mapData.pickupStatus} 
          dropoffStatus={mapData.dropoffStatus}
          className="map-wrapper"
        />

        {!currentRide ? (
          <RideRequestForm onRequestSubmitted={handleRideRequested} />
        ) : (
          <TripTracking rideId={currentRide.rideId} onTripComplete={handleNewRide} />
        )}
      </main>
    </div>
  );
};
