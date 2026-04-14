// Type definitions for the ride-hailing system

export interface Location {
  latitude: number;
  longitude: number;
}

export interface RideRequest {
  rideId: string;
  userId: string;
  pickupLocation: Location;
  dropoffLocation: Location;
  status: 'requested' | 'accepted' | 'driver_arrived' | 'in_progress' | 'completed' | 'cancelled';
  rideType: 'economy' | 'premium';
  requestedAt: string;
  createdAt: string;
  updatedAt: string;
}

export interface Driver {
  driverId: string;
  name: string;
  rating: number;
  vehiclePlate: string;
  currentLocation: Location;
  status: 'available' | 'on_trip' | 'offline';
}

export interface Trip {
  tripId: string;
  rideId: string;
  driver: Driver;
  status: 'assigned' | 'pick_up_in_progress' | 'in_progress' | 'drop_off_in_progress' | 'completed';
  pickupLocation: Location;
  dropoffLocation: Location;
  currentLocation: Location;
  estimatedArrivalTime: number; // seconds
  fare: {
    amount: number;
    currency: string;
  };
  rating?: number;
  feedback?: string;
}

export interface UserProfile {
  userId: string;
  name: string;
  email: string;
  phone: string;
  homeLocation?: Location;
  workLocation?: Location;
  rideCount: number;
  averageRating: number;
}
