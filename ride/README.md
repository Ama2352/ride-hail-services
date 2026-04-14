# Ride Service

Microservice responsible for **Ride Lifecycle Management** and **Event Publishing** using Redis Streams.

The Ride Service is the **Source of Truth** for all ride state and ensures consistency across the distributed system.

## Technology Stack

- **Framework:** Spring Boot 3.3.0
- **Language:** Java 21
- **Database:** PostgreSQL (Ride history & state)
- **Message Queue:** Redis Streams (Event publishing)
- **Monitoring:** Prometheus metrics (via Spring Actuator)

## Architecture

### Key Responsibilities

1. **REST API for Ride Management**
   - `POST /rides` - Create new ride request
   - `GET /rides/{id}` - Fetch ride details
   - `PUT /rides/{id}/status` - Update ride status and publish events
   - `GET /rides/passenger/{id}` - List rides for passenger
   - `GET /rides/driver/{id}` - List rides for driver

2. **Ride Lifecycle State Machine**
   ```
   REQUESTED → ASSIGNED → IN_PROGRESS → COMPLETED
           ↘     CANCELLED
   ```

3. **Event Publishing**
   - Publishes events to Redis Stream (`ride-events`)
   - Event Types: `Ride.Requested`, `Ride.Assigned`, `Ride.Started`, `Ride.Completed`, `Ride.Cancelled`
   - Event payload includes full ride context (coordinates, IDs, timestamps)

4. **Database Persistence**
   - Stores ride records in PostgreSQL
   - Maintains audit trail with timestamps
   - Each ride linked to passenger/driver IDs

### Environment Variables

Environment variables should be set at runtime via deployment (Kubernetes, CI/CD, or system environment):

```bash
# Database
DB_HOST=postgres                    # PostgreSQL host
DB_PORT=5432                        # PostgreSQL port
DB_NAME=ridehailing                # Database name
DB_USER=postgres                    # Database user
DB_PASSWORD=<secure-password>       # Database password (from secure vault)
REDIS_HOST=redis-master             # Redis host
REDIS_PORT=6379                     # Redis port
PORT=8082                           # HTTP server port
```

**Note:** Do NOT hardcode sensitive values in Dockerfile or application code. Use secure secrets management (Kubernetes Secrets, DockerCompose env files, CI/CD vaults, etc.)

## API Examples

### Create a Ride Request

```bash
POST /rides
Content-Type: application/json

{
  "passengerId": "pax-123",
  "pickupLat": 10.7769,
  "pickupLng": 106.7009,
  "dropoffLat": 10.8141,
  "dropoffLng": 106.6269
}

Response:
{
  "id": 1,
  "passengerId": "pax-123",
  "status": "REQUESTED",
  "estimatedFare": 8.50,
  "createdAt": "2026-04-14T12:00:00"
}
```

### Update Ride Status (Dispatch assigns driver)

```bash
PUT /rides/1/status
Content-Type: application/json

{
  "status": "ASSIGNED",
  "driverId": "drv-456"
}

# Publishes: Ride.Assigned event to Redis Stream
```

### Get Ride Details

```bash
GET /rides/1

Response:
{
  "id": 1,
  "passengerId": "pax-123",
  "driverId": "drv-456",
  "status": "ASSIGNED",
  "pickupLat": 10.7769,
  "pickupLng": 106.7009,
  "dropoffLat": 10.8141,
  "dropoffLng": 106.6269,
  "createdAt": "2026-04-14T12:00:00"
}
```

### Health Check

```bash
GET /rides/health

Response:
{
  "status": "healthy",
  "service": "ride-service"
}
```

## Testing

```bash
# Run all tests
mvn test

# Run specific test class
mvn test -Dtest=RideControllerTest

# Run with coverage
mvn test jacoco:report
```

## Docker

### Build Docker Image

```bash
docker build -t ridehailing/ride-service:1.0.0 .

# Tag latest
docker tag ridehailing/ride-service:1.0.0 ridehailing/ride-service:latest
```

### Run with Docker

```bash
# Provide environment variables at runtime
docker run -d \
  --name ride-service \
  -p 8082:8082 \
  -e DB_HOST=postgres-host \
  -e DB_USER=postgres \
  -e DB_PASSWORD=<secure-password> \
  -e REDIS_HOST=redis-host \
  ridehailing/ride-service:latest
```

## Metrics & Monitoring

Prometheus metrics exposed at: `GET /actuator/prometheus`

Key metrics:
- `http_requests_total` - Total HTTP requests by endpoint
- `http_request_duration_seconds` - Request latency histogram
- `jpa_sessions_open` - Database connection pool status

## Integration with Other Services

### Dispatch Service
- Consumes `Ride.Requested` events to find nearby drivers
- Publishes `Ride.Assigned` events back

### Notification Service
- Consumes all ride events to push notifications to drivers/passengers

### User Service
- Validates passenger/driver credentials via JWT tokens

## Database Schema

### rides table
```sql
CREATE TABLE rides (
  id BIGSERIAL PRIMARY KEY,
  passenger_id VARCHAR(255) NOT NULL,
  driver_id VARCHAR(255),
  status VARCHAR(50) NOT NULL,
  pickup_lat DOUBLE PRECISION NOT NULL,
  pickup_lng DOUBLE PRECISION NOT NULL,
  dropoff_lat DOUBLE PRECISION NOT NULL,
  dropoff_lng DOUBLE PRECISION NOT NULL,
  estimated_fare DOUBLE PRECISION,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP,
  started_at TIMESTAMP,
  completed_at TIMESTAMP
);
```

## Troubleshooting

### Connection refused to PostgreSQL

```bash
# Ensure PostgreSQL is running
docker-compose ps postgres

# Check database connection
psql -h localhost -U postgres -c "SELECT 1;"
```

### Redis connection issues

```bash
# Test Redis connectivity
redis-cli PING

# Check Redis Streams
redis-cli XRANGE ride-events - +
```

### Build failures

```bash
# Clean Maven cache
mvn clean

# Force dependency update
mvn clean dependency:resolve -U
```

## Contributing

- Follow Spring Boot best practices
- Add unit tests for new features
- Use consistent naming: `RideService`, `RideRepository`, `RideController`
- Keep business logic in services, not controllers
