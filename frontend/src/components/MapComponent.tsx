import React, { useEffect, useState } from 'react';
import { MapContainer, TileLayer, Marker, Popup, Polyline } from 'react-leaflet';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';

interface Location {
  lat: number;
  lng: number;
}

interface MapComponentProps {
  pickupStatus?: Location | null;
  dropoffStatus?: Location | null;
  currentLocation?: Location | null;
  className?: string;
  onClick?: (e: L.LeafletMouseEvent) => void;
}

// Marker Icons setup (using standard leaflet icons but tailored)
const customIcon = (color: string) => new L.Icon({
  iconUrl: `https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-${color}.png`,
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/0.7.7/images/marker-shadow.png',
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41]
});

const pickupIcon = customIcon('green');
const dropoffIcon = customIcon('red');
const carIcon = customIcon('blue'); // Simulating a car

export const MapComponent: React.FC<MapComponentProps> = ({ pickupStatus, dropoffStatus, currentLocation, className, onClick }) => {
  const [center, setCenter] = useState<[number, number]>([10.762622, 106.660172]); // Ho Chi Minh City Default

  useEffect(() => {
    if (pickupStatus) setCenter([pickupStatus.lat, pickupStatus.lng]);
    else if (currentLocation) setCenter([currentLocation.lat, currentLocation.lng]);
  }, [pickupStatus, currentLocation]);

  return (
    <div className={`map-wrapper ${className || ''}`}>
      <MapContainer center={center} zoom={14} style={{ height: '100%', width: '100%' }} zoomControl={false} onClick={onClick}>
        <TileLayer
          url="https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png"
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> &copy; <a href="https://carto.com/">CARTO</a>'
        />
        
        {pickupStatus && (
          <Marker position={[pickupStatus.lat, pickupStatus.lng]} icon={pickupIcon}>
            <Popup>Pickup Location</Popup>
          </Marker>
        )}

        {dropoffStatus && (
          <Marker position={[dropoffStatus.lat, dropoffStatus.lng]} icon={dropoffIcon}>
            <Popup>Dropoff Location</Popup>
          </Marker>
        )}

        {currentLocation && (
          <Marker position={[currentLocation.lat, currentLocation.lng]} icon={carIcon}>
            <Popup>Driver</Popup>
          </Marker>
        )}

        {pickupStatus && dropoffStatus && (
          <Polyline positions={[[pickupStatus.lat, pickupStatus.lng], [dropoffStatus.lat, dropoffStatus.lng]]} color="#00B14F" weight={4} dashArray="8, 8"/>
        )}
      </MapContainer>
    </div>
  );
};
