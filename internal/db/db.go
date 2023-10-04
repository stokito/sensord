package db

import (
	"context"
	"time"
)

// SensorsDb is a generic DB interface
type SensorsDb interface {
	Connect(ctx context.Context) error
	Close()
	StoreMeasurement(ctx context.Context, day time.Time, sensorId int, value float64)
	GetMeasurementStatsForDay(ctx context.Context, day time.Time, sensorId int) (*MeasurementRec, error)
	GetMeasurementPeriodStatsTotal(ctx context.Context, periodStart, periodEnd time.Time) (*MeasurementRec, error)
	GetMeasurementPeriodStatsForEachSensor(ctx context.Context, periodStart, periodEnd time.Time) ([]*MeasurementRec, error)
	GetMeasurementPeriodStatsForEachSensorAndDay(ctx context.Context, periodStart, periodEnd time.Time) ([]*MeasurementRec, error)
	Cleanup(ctx context.Context)
}
