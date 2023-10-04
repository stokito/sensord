package models

import "time"

// MeasurementRec used in stats reporting
type MeasurementRec struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
	SensorId    int
	TotalCount  int64
	TotalSum    float64
	AvgValue    float64
	MinValue    float64
	MaxValue    float64
}
