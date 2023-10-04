package db

import (
	"context"
	"sensord/internal/models"
	"time"
)

// SensorsDb is a generic DB interface
type SensorsDb interface {
	Connect(ctx context.Context) error
	Close()
	StoreMeasurement(ctx context.Context, day time.Time, sensorId int, value float64)
	GetMeasurementStatsForDay(ctx context.Context, day time.Time, sensorId int) (*models.MeasurementRec, error)
	GetMeasurementPeriodStatsTotal(ctx context.Context, periodStart, periodEnd time.Time) (*models.MeasurementRec, error)
	GetMeasurementPeriodStatsForEachSensor(ctx context.Context, periodStart, periodEnd time.Time) ([]*models.MeasurementRec, error)
	GetMeasurementPeriodStatsForEachSensorAndDay(ctx context.Context, periodStart, periodEnd time.Time) ([]*models.MeasurementRec, error)
	Cleanup(ctx context.Context)
}
