# Ride Hailing Frontend

React-based rider-facing UI for the ride-hailing platform.

## Features

- **Ride Request Form** - Book rides with pickup/dropoff locations and ride type selection
- **Trip Tracking** - Real-time tracking of driver location and trip status
- **Feedback System** - Rate trips and provide feedback after completion
- **Responsive Design** - Works on desktop and mobile devices

## Project Structure

```
frontend/
├── public/              # Static assets
├── src/
│   ├── components/      # React components
│   │   ├── RideRequestForm.tsx   # Ride booking form
│   │   └── TripTracking.tsx      # Trip tracking display
│   ├── api/
│   │   └── client.ts    # API client for backend services
│   ├── types/
│   │   └── index.ts     # TypeScript type definitions
│   ├── App.tsx          # Main app component
│   ├── main.tsx         # Application entry point
│   └── index.css        # Global styles
├── package.json
├── tsconfig.json
└── Dockerfile
```

## Setup

### Prerequisites
- Node.js 18+
- npm or yarn

### Development

```bash
# Install dependencies
npm install

# Start development server
npm start
# Open http://localhost:3000
```

### Build for Production

```bash
# Build the application
npm run build
```

## Environment Variables

```bash
REACT_APP_API_URL=http://localhost:8080
```

## Docker

```bash
# Build
docker build -t ride-hailing-frontend:latest .

# Run
docker run -d \
  -p 3000:3000 \
  -e REACT_APP_API_URL=http://ride-service:8080 \
  ride-hailing-frontend:latest
```

## API Integration

The frontend communicates with the Ride Service API at the configured `REACT_APP_API_URL`.

### Key Endpoints

- `POST /api/rides` - Request new ride
- `GET /api/rides/:rideId` - Get ride status
- `GET /api/trips/ride/:rideId` - Get trip details
- `POST /api/trips/:tripId/rating` - Submit trip feedback

## Architecture Notes

- **State Management**: React hooks (useState, useEffect)
- **API Client**: Axios with base URL configuration
- **Styling**: CSS modules for component isolation
- **Type Safety**: Full TypeScript typing for all components and API calls
