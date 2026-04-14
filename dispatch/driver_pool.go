package main

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

// DriverPool manages driver geographic locations and availability
type DriverPool struct {
	rdb *redis.Client
}

// NewDriverPool creates a new driver pool manager
func NewDriverPool(rdb *redis.Client) *DriverPool {
	return &DriverPool{rdb: rdb}
}

// FindNearbyDrivers returns list of drivers within specified radius from pickup point
// Returns up to maxDrivers closest drivers sorted by distance
func (dp *DriverPool) FindNearbyDrivers(ctx context.Context, pickupLat, pickupLng float64, maxDrivers int64) ([]string, error) {
	// Query Redis Geo: find all drivers within 10km radius, sorted by distance
	radiusCmd := dp.rdb.GeoRadius(ctx, "driver_locations", pickupLng, pickupLat, &redis.GeoRadiusQuery{
		Radius:      10, // 10 km radius
		Unit:        "km",
		WithCoord:   false,
		WithDist:    true,
		Count:       int(maxDrivers),
		Sort:        "ASC", // Closest first
	})

	drivers, err := radiusCmd.Val(), radiusCmd.Err()
	if err != nil {
		log.Printf("Error querying nearby drivers: %v", err)
		return nil, err
	}

	var driverIDs []string
	for _, d := range drivers {
		driverIDs = append(driverIDs, d.Name)
	}

	log.Printf("Found %d nearby drivers for pickup (%.4f, %.4f)", len(driverIDs), pickupLat, pickupLng)
	return driverIDs, nil
}

// RemoveDriver removes driver from available pool (when assigned to ride)
// We remove by storing driver in a "busy" set instead of deleting from geo
func (dp *DriverPool) BusyDriver(ctx context.Context, driverID string) error {
	// Add to busy set with current time
	err := dp.rdb.Set(ctx, "driver:busy:"+driverID, "1", 0).Err()
	if err != nil {
		log.Printf("Error marking driver %s as busy: %v", driverID, err)
		return err
	}
	return nil
}

// ReleaseDriver returns driver to available pool
func (dp *DriverPool) ReleaseDriver(ctx context.Context, driverID string) error {
	// Remove from busy set
	err := dp.rdb.Del(ctx, "driver:busy:"+driverID).Err()
	if err != nil {
		log.Printf("Error releasing driver %s: %v", driverID, err)
		return err
	}
	return nil
}

// IsDriverBusy checks if driver is marked as busy
func (dp *DriverPool) IsDriverBusy(ctx context.Context, driverID string) (bool, error) {
	exists, err := dp.rdb.Exists(ctx, "driver:busy:"+driverID).Result()
	return exists > 0, err
}

// AddDriver adds driver to geo pool (when connected via WebSocket)
func (dp *DriverPool) AddDriver(ctx context.Context, driverID string, lat, lng float64) error {
	err := dp.rdb.GeoAdd(ctx, "driver_locations", &redis.GeoLocation{
		Name:      driverID,
		Longitude: lng,
		Latitude:  lat,
	}).Err()
	if err != nil {
		log.Printf("Error adding driver to pool: %v", err)
		return err
	}
	return nil
}
