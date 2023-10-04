package db

import (
	"context"
	"time"
)

// SensorsDb is a generic DB interface
type SensorsDb interface {
	Connect(ctx context.Context) error
	Close()
	StoreMeasurement(ctx context.Context, measureTime time.Time, sensorId int, value float64)
	Cleanup(ctx context.Context)
}
