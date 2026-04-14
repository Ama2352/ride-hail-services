import React, { useState } from 'react';
import { RideRequestForm } from './components/RideRequestForm';
import { TripTracking } from './components/TripTracking';
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

  return (
    <div className="app">
      <header className="app-header">
        <div className="header-content">
          <h1 className="app-title">🚗 Ride Hailing</h1>
          <p className="app-subtitle">Quick, Safe, and Affordable Rides</p>
        </div>
      </header>

      <main className="app-main">
        <div className="container">
          {!currentRide ? (
            <RideRequestForm onRequestSubmitted={handleRideRequested} />
          ) : (
            <div>
              <TripTracking rideId={currentRide.rideId} onTripComplete={handleNewRide} />
              <button className="new-ride-btn" onClick={handleNewRide}>
                Book Another Ride
              </button>
            </div>
          )}
        </div>
      </main>

      <footer className="app-footer">
        <p>&copy; 2026 Ride Hailing Platform | Safe &amp; Reliable Transportation</p>
      </footer>
    </div>
  );
};
